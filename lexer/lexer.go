// Package implements a lexer for Stable source text.
package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/stable-lang/stlang/token"
)

const (
	bom = 0xFEFF // byte order mark, only permitted as a very first character
	eof = -1     // end of file
)

// ErrorHandler is an error handler for the [Lexer].
// The position points to the beginning of the offending token.
type ErrorHandler func(pos token.Position, msg string)

// Lexer reads the Stable source text.
type Lexer struct {
	file     *token.File
	src      []byte
	errFn    ErrorHandler
	errCount int

	ch         rune      // current character
	offset     int       // character offset
	readOffset int       // reading offset (position after current character)
	lineOffset int       // current line offset
	insertSemi bool      // insert a semicolon before next newline
	nlPos      token.Pos // position of newline in preceding comment

	noNewSemi bool // used only for testing
}

// NewLexer creates a new [Lexer].
func NewLexer(file *token.File, src []byte, err ErrorHandler) *Lexer {
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}

	l := &Lexer{
		file:  file,
		src:   src,
		errFn: err,
		ch:    ' ',
	}

	l.next()
	if l.ch == bom {
		l.next() // ignore BOM at file beginning
	}
	return l
}

// Scan the next token and returns the token position, the token and its literal string if applicable.
// The source end is indicated by [token.EOF].
func (l *Lexer) Scan() (pos token.Pos, tok token.Token, lit string) {
	if l.nlPos.IsValid() {
		// Return artificial ';' token after /*...*/ comment
		// containing newline, at position of first newline.
		pos, tok, lit = l.nlPos, token.Semicolon, "\n"
		l.nlPos = token.NoPos
		return pos, tok, lit
	}

	l.skipWhitespace()

	pos = l.file.Pos(l.offset)
	insertSemi := false

	switch ch := l.ch; {
	case isLetter(ch):
		lit = l.scanIdent()
		tok = token.Lookup(lit)
		switch tok {
		case token.Ident, token.Break,
			token.Continue, token.Fallthrough,
			token.Return:
			insertSemi = true
		}

	case isDecimal(ch) || (ch == '.' && isDecimal(rune(l.peek()))):
		insertSemi = true
		tok, lit = l.scanNumber()

	default:
		l.next() // always make progress

		switch ch {
		case eof:
			if l.insertSemi {
				l.insertSemi = false // EOF consumed
				return pos, token.Semicolon, "\n"
			}
			tok = token.EOF
		case '\n':
			// we only reach here if l.insertSemi was
			// set in the first place and exited early
			// from s.skipWhitespace()
			l.insertSemi = false // newline consumed
			return pos, token.Semicolon, "\n"
		case '"':
			insertSemi = true
			tok = token.String
			lit = l.scanString()
		case '\'':
			insertSemi = true
			tok = token.Char
			lit = l.scanRune()
		case '`':
			insertSemi = true
			tok = token.String
			lit = l.scanRawString()

		case ',':
			tok = token.Comma
		case '.':
			tok = token.Period
			if l.ch == '.' && l.peek() == '.' {
				l.next()
				l.next() // consume last '.'
				tok = token.Ellipsis
			}
		case ':':
			tok = l.switch2(token.Colon, token.Define)
		case ';':
			tok = token.Semicolon
			lit = ";"

		case '(':
			tok = token.LeftParen
		case '[':
			tok = token.LeftBrack
		case '{':
			tok = token.LeftBrace

		case ')':
			insertSemi = true
			tok = token.RightParen
		case ']':
			insertSemi = true
			tok = token.RightBrack
		case '}':
			insertSemi = true
			tok = token.RightBrace

		case '+':
			tok = l.switch4(token.Add, token.AddAssign, '+', token.Concat, token.ConcatAssign)
		case '-':
			tok = l.switch2(token.Sub, token.SubAssign)
		case '*':
			tok = l.switch2(token.Mul, token.MulAssign)
		case '/':
			if l.ch == '/' || l.ch == '*' {
				// comment
				comment, nlOffset := l.scanComment()
				if l.insertSemi && nlOffset != 0 {
					// For /*...*/ containing \n, return
					// COMMENT then artificial SEMICOLON.
					l.nlPos = l.file.Pos(nlOffset)
					l.insertSemi = false
				} else {
					insertSemi = l.insertSemi // preserve insertSemi info
				}
				tok = token.Comment
				lit = comment
			} else {
				// division
				tok = l.switch2(token.Quo, token.QuoAssign)
			}
		case '%':
			tok = l.switch2(token.Rem, token.RemAssign)
		case '^':
			tok = l.switch2(token.Xor, token.XorAssign)
		case '<':
			tok = l.switch4(token.Less, token.LessEqual, '<', token.Shl, token.ShlAssign)
		case '>':
			tok = l.switch4(token.Greater, token.GreaterEqual, '>', token.Shr, token.ShrAssign)
		case '=':
			tok = l.switch2(token.Assign, token.Equal)
		case '!':
			tok = l.switch2(token.LogicNot, token.NotEqual)
		case '&':
			if l.ch == '^' {
				l.next()
				tok = l.switch2(token.AndNot, token.AndNotAssign)
			} else {
				tok = l.switch3(token.And, token.AndAssign, '&', token.LogicAnd)
			}
		case '|':
			tok = l.switch3(token.Or, token.OrAssign, '|', token.LogicOr)

		default:
			// next reports unexpected BOMs - don't repeat
			if ch != bom {
				// Report an informative error for U+201[CD] quotation
				// marks, which are easily introduced via copy and paste.
				if ch == '“' || ch == '”' {
					l.errorf(l.file.Offset(pos), "curly quotation mark %q (use neutral %q)", ch, '"')
				} else {
					l.errorf(l.file.Offset(pos), "illegal character %#U", ch)
				}
			}
			insertSemi = l.insertSemi // preserve insertSemi info
			tok = token.Illegal
			lit = string(ch)
		}
	}

	if !l.noNewSemi {
		l.insertSemi = insertSemi
	}
	return pos, tok, lit
}

// next Unicode char into l.ch, l.ch < 0 means end-of-file.
func (l *Lexer) next() {
	if l.readOffset >= len(l.src) {
		l.offset = len(l.src)
		if l.ch == '\n' {
			l.lineOffset = l.offset
			l.file.AddLine(l.offset)
		}
		l.ch = eof
		return
	}

	l.offset = l.readOffset
	if l.ch == '\n' {
		l.lineOffset = l.offset
		l.file.AddLine(l.offset)
	}

	r, w := rune(l.src[l.readOffset]), 1
	switch {
	case r == 0:
		l.error(l.offset, "illegal character NUL")

	case r >= utf8.RuneSelf: // not ASCII
		r, w = utf8.DecodeRune(l.src[l.readOffset:])
		switch {
		case r == utf8.RuneError && w == 1:
			in := l.src[l.readOffset:]

			// U+FEFF BOM at start of file, encoded as big- or little-endian UCS-2 (i.e. 2-byte UTF-16).
			if l.offset == 0 && len(in) >= 2 &&
				(in[0] == 0xFF && in[1] == 0xFE || in[0] == 0xFE && in[1] == 0xFF) {
				l.error(l.offset, "illegal UTF-8 encoding (got UTF-16)")
				l.readOffset += len(in) // consume all input to avoid error cascade
			} else {
				l.error(l.offset, "illegal UTF-8 encoding")
			}
		case r == bom && l.offset > 0:
			l.error(l.offset, "illegal byte order mark")
		}
	}
	l.readOffset += w
	l.ch = r
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (l *Lexer) peek() byte {
	if l.readOffset < len(l.src) {
		return l.src[l.readOffset]
	}
	return 0
}

// scanIdent reads the string of valid identifier characters at l.offset.
// It must only be called when l.ch is known to be a valid letter.
func (l *Lexer) scanIdent() string {
	offs := l.offset
	for isLetter(l.ch) || isDecimal(l.ch) {
		l.next()
	}
	return string(l.src[offs:l.offset])
}

func (l *Lexer) scanNumber() (token.Token, string) {
	offs := l.offset
	tok := token.Int
	base := 10

	// determine base by prefix (if present)
	if l.ch == '0' {
		switch l.peek() {
		case 'b', 'B':
			base = 2
			l.next()
			l.next()
		case 'o', 'O':
			base = 8
			l.next()
			l.next()
		case 'x', 'X':
			base = 16
			l.next()
			l.next()
		}
	}

	// scan number
	invalid := -1 // index of invalid digit in literal, or < 0.
	digsep := l.scanDigits(base, &invalid)
	if digsep.IsEmpty() {
		l.error(l.offset, litname(base)+" has no digits")
	}

	var digsepFrac digSep
	// scan fractional part
	if l.ch == '.' {
		l.next()
		tok = token.Float
		if base != 10 {
			l.error(offs, "only decimal floats are possible")
		}

		digsepFrac = l.scanDigits(10, &invalid)
	}

	lit := string(l.src[offs:l.offset])
	if tok == token.Int && invalid >= 0 {
		l.errorf(invalid, "invalid digit %q in %s", lit[invalid-offs], litname(base))
	}
	if digsep.HasSep() || digsepFrac.HasSep() {
		if i := invalidSep(lit); i >= 0 {
			l.error(offs+i, "'_' must separate successive digits")
		}
	}
	if tok == token.Float && digsepFrac.IsEmpty() {
		l.error(offs, "no fraction part for the float")
	}
	return tok, lit
}

type digSep byte

func (d digSep) IsEmpty() bool { return d&1 == 0 }
func (d digSep) HasDig() bool  { return d&1 == 1 }
func (d digSep) HasSep() bool  { return d&2 != 0 }

// digits accepts the sequence { digit | '_' }.
// If base <= 10, digits accepts any decimal digit but records
// the offset (relative to the source start) of a digit >= base
// in *invalid, if *invalid < 0.
//
// digits returns a bitset describing whether the sequence contained
// digits (bit 0 is set), or separators '_' (bit 1 is set).
func (l *Lexer) scanDigits(base int, invalid *int) digSep {
	var digsep digSep

	if base <= 10 {
		maxDigit := rune('0' + base)
		for isHex(l.ch) || l.ch == '_' {
			ds := 1
			switch {
			case l.ch == '_':
				ds = 2
			case l.ch >= maxDigit && *invalid < 0:
				*invalid = l.offset // record invalid rune offset
			}
			digsep |= digSep(ds)
			l.next()
		}
	} else {
		for isHex(l.ch) || l.ch == '_' {
			ds := 1
			if l.ch == '_' {
				ds = 2
			}
			digsep |= digSep(ds)
			l.next()
		}
	}
	return digsep
}

// invalidSep returns the index of the first invalid separator in x, or -1.
func invalidSep(x string) int {
	x1 := ' ' // prefix char, we only care if it's 'x'
	d := '.'  // digit, one of '_', '0' (a digit), or '.' (anything else)
	i := 0

	// a prefix counts as a digit
	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		switch x1 {
		case 'b', 'o', 'x':
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}
	return -1
}

// scanComment returns the text of the comment and (if nonzero)
// the offset of the first newline within it, which implies a
// /*...*/ comment.
func (l *Lexer) scanComment() (string, int) {
	// initial '/' already consumed; l.ch == '/' || l.ch == '*'
	offs := l.offset - 1 // position of initial '/'
	next := -1           // position immediately following the comment; < 0 means invalid comment
	numCR := 0
	nlOffset := 0 // offset of first newline within /*...*/ comment

	if l.ch == '/' {
		//-style comment
		// (the final '\n' is not considered part of the comment)
		l.next()
		for l.ch != '\n' && l.ch >= 0 {
			if l.ch == '\r' {
				numCR++
			}
			l.next()
		}
		// if we are at '\n', the position following the comment is afterwards
		next = l.offset
		if l.ch == '\n' {
			next++
		}
		goto exit
	}

	/*-style comment */
	l.next()
	for l.ch >= 0 {
		ch := l.ch
		if ch == '\r' {
			numCR++
		} else if ch == '\n' && nlOffset == 0 {
			nlOffset = l.offset
		}
		l.next()
		if ch == '*' && l.ch == '/' {
			l.next()
			next = l.offset
			goto exit
		}
	}

	l.error(offs, "comment not terminated")

exit:
	lit := l.src[offs:l.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	if numCR > 0 {
		lit = stripCR(lit, lit[1] == '*')
	}
	return string(lit), nlOffset
}

func (l *Lexer) scanString() string {
	// '"' opening already consumed
	offs := l.offset - 1

	for {
		ch := l.ch
		if ch == '\n' || ch < 0 {
			l.error(offs, "string literal not terminated")
			break
		}
		l.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			l.scanEscape('"')
		}
	}
	return string(l.src[offs:l.offset])
}

func (l *Lexer) scanRune() string {
	// '\'' opening already consumed
	offs := l.offset - 1

	valid := true
	n := 0
	for {
		ch := l.ch
		if ch == '\n' || ch < 0 {
			// only report error if we don't have one already
			if valid {
				l.error(offs, "rune literal not terminated")
				valid = false
			}
			break
		}
		l.next()
		if ch == '\'' {
			break
		}
		n++
		if ch == '\\' {
			if !l.scanEscape('\'') {
				valid = false
			}
			// continue to read to closing quote
		}
	}

	if valid && n != 1 {
		l.error(offs, "illegal rune literal")
	}
	return string(l.src[offs:l.offset])
}

func (l *Lexer) scanRawString() string {
	// '`' opening already consumed
	offs := l.offset - 1

	hasCR := false
	for {
		ch := l.ch
		if ch < 0 {
			l.error(offs, "raw string literal not terminated")
			break
		}
		l.next()
		if ch == '`' {
			break
		}
		if ch == '\r' {
			hasCR = true
		}
	}

	lit := l.src[offs:l.offset]
	if hasCR {
		lit = stripCR(lit, false)
	}
	return string(lit)
}

// scanEscape parses an escape sequence where rune is the accepted escaped quote.
// In case of a syntax error, it stops at the offending character (without consuming it) and returns false.
// Otherwise it returns true.
func (l *Lexer) scanEscape(quote rune) bool {
	offs := l.offset

	var n int
	var base, maxValue uint32
	switch l.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		l.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, maxValue = 3, 8, 255
	case 'x':
		l.next()
		n, base, maxValue = 2, 16, 255
	case 'u':
		l.next()
		n, base, maxValue = 4, 16, utf8.MaxRune
	case 'U':
		l.next()
		n, base, maxValue = 8, 16, utf8.MaxRune
	default:
		msg := "unknown escape sequence"
		if l.ch < 0 {
			msg = "escape sequence not terminated"
		}
		l.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(l.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", l.ch)
			if l.ch < 0 {
				msg = "escape sequence not terminated"
			}
			l.error(l.offset, msg)
			return false
		}
		x = x*base + d
		l.next()
		n--
	}

	if x > maxValue || 0xD800 <= x && x < 0xE000 {
		l.error(offs, "escape sequence is invalid Unicode code point")
		return false
	}
	return true
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' ||
		l.ch == '\t' ||
		l.ch == '\n' && !l.insertSemi ||
		l.ch == '\r' {
		l.next()
	}
}

// Helper functions for scanning multi-byte tokens such as >> += >>= .

func (l *Lexer) switch2(tok0, tok1 token.Token) token.Token {
	if l.ch == '=' {
		l.next()
		return tok1
	}
	return tok0
}

func (l *Lexer) switch3(tok0, tok1 token.Token, ch2 rune, tok2 token.Token) token.Token {
	if l.ch == '=' {
		l.next()
		return tok1
	}
	if l.ch == ch2 {
		l.next()
		return tok2
	}
	return tok0
}

func (l *Lexer) switch4(tok0, tok1 token.Token, ch2 rune, tok2, tok3 token.Token) token.Token {
	if l.ch == '=' {
		l.next()
		return tok1
	}
	if l.ch == ch2 {
		l.next()
		if l.ch == '=' {
			l.next()
			return tok3
		}
		return tok2
	}
	return tok0
}

func (l *Lexer) errorf(offs int, format string, args ...any) {
	l.error(offs, fmt.Sprintf(format, args...))
}

func (l *Lexer) error(offs int, msg string) {
	l.errCount++
	if l.errFn != nil {
		l.errFn(l.file.Position(l.file.Pos(offs)), msg)
	}
}

func stripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl.
		// sequences of \r from *\r\r...\r/) since the resulting
		// */ would terminate the comment too early unless the \r
		// is immediately following the opening /* in which case
		// it's ok because /*/ is not closed yet (issue #11151).
		if ch != '\r' ||
			comment &&
				i > len("/*") && c[i-1] == '*' &&
				j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	default:
		return 16 // larger than any legal digit val
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' ||
		'A' <= ch && ch <= 'Z' ||
		ch == '_'
}

// returns lower-case ch iff ch is ASCII letter
func lower(ch rune) rune     { return ('a' - 'A') | ch }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool {
	return '0' <= ch && ch <= '9' ||
		'a' <= ch && ch <= 'f' ||
		'A' <= ch && ch <= 'F'
}

func litname(base int) string {
	switch base {
	case 2:
		return "binary literal"
	case 8:
		return "octal literal"
	case 16:
		return "hexadecimal literal"
	default:
		return "decimal literal"
	}
}
