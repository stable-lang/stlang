package lexer

import (
	"path/filepath"
	"testing"

	"github.com/stable-lang/stlang/token"
)

var fset = token.NewFileSet()

type elt struct {
	tok   token.Token
	lit   string
	class int
}

var tokens = []elt{
	// Special tokens
	{token.Comment, "/* a comment */", special},
	{token.Comment, "// a comment \n", special},
	{token.Comment, "/*\r*/", special},
	{token.Comment, "/**\r/*/", special},
	{token.Comment, "/**\r\r/*/", special},
	{token.Comment, "//\r\n", special},

	// Identifiers and basic type literals
	{token.Ident, "a", literal},
	{token.Ident, "main", literal},
	{token.Ident, "foobar", literal},
	{token.Ident, "foo_", literal},
	{token.Ident, "_bar", literal},
	{token.Ident, "err2", literal},

	{token.Int, "0", literal},
	{token.Int, "1", literal},
	{token.Int, "123456789012345678890", literal},
	{token.Int, "01234567", literal},
	{token.Int, "000123", literal},
	{token.Int, "0b111000", literal},
	{token.Int, "0o1234567", literal},
	{token.Int, "0xcafebabe", literal},

	{token.Float, "0.0", literal},
	{token.Float, "0.1", literal},
	{token.Float, "1.0", literal},
	{token.Float, "3.14159265", literal},
	{token.Float, "12345.54321", literal},

	{token.Char, "'a'", literal},
	{token.Char, "'\\000'", literal},
	{token.Char, "'\\xFF'", literal},
	{token.Char, "'\\uff16'", literal},
	{token.Char, "'\\U0000ff16'", literal},

	{token.String, `"a۰۱۸"`, literal},
	{token.String, `"foo६४"`, literal},
	{token.String, `"bar９８７６"`, literal},
	{token.String, `"ŝ"`, literal},
	{token.String, `"ŝfoo"`, literal},
	{token.String, "`foobar`", literal},
	{
		token.String, "`" + `foo
	                        bar` +
			"`",
		literal,
	},
	{token.String, "`\r`", literal},
	{token.String, "`foo\r\nbar`", literal},

	{token.True, "true", literal},
	{token.Nil, "nil", literal},
	{token.False, "false", literal},

	// Operators and delimiters
	{token.Add, "+", operator},
	{token.Sub, "-", operator},
	{token.Mul, "*", operator},
	{token.Quo, "/", operator},
	{token.Rem, "%", operator},
	{token.And, "&", operator},
	{token.Or, "|", operator},
	{token.Xor, "^", operator},
	{token.AndNot, "&^", operator},
	{token.Shl, "<<", operator},
	{token.Shr, ">>", operator},
	{token.Concat, "++", operator},

	{token.Assign, "=", operator},
	{token.AddAssign, "+=", operator},
	{token.SubAssign, "-=", operator},
	{token.MulAssign, "*=", operator},
	{token.QuoAssign, "/=", operator},
	{token.RemAssign, "%=", operator},
	{token.AndAssign, "&=", operator},
	{token.OrAssign, "|=", operator},
	{token.XorAssign, "^=", operator},
	{token.AndNotAssign, "&^=", operator},
	{token.ShlAssign, "<<=", operator},
	{token.ShrAssign, ">>=", operator},
	{token.ConcatAssign, "++=", operator},

	{token.LogicAnd, "&&", operator},
	{token.LogicOr, "||", operator},
	{token.LogicNot, "!", operator},
	{token.Equal, "==", operator},
	{token.NotEqual, "!=", operator},
	{token.Less, "<", operator},
	{token.Greater, ">", operator},
	{token.LessEqual, "<=", operator},
	{token.GreaterEqual, ">=", operator},

	{token.Define, ":=", operator},
	{token.Ellipsis, "...", operator},
	{token.Comma, ",", operator},
	{token.Period, ".", operator},
	{token.Colon, ":", operator},
	{token.Semicolon, ";", operator},

	{token.LeftParen, "(", operator},
	{token.LeftBrack, "[", operator},
	{token.LeftBrace, "{", operator},
	{token.RightParen, ")", operator},
	{token.RightBrack, "]", operator},
	{token.RightBrace, "}", operator},

	// Keywords
	{token.Any, "any", keyword},
	{token.Bool, "bool", keyword},
	{token.Break, "break", keyword},
	{token.Case, "case", keyword},
	{token.Const, "const", keyword},
	{token.Continue, "continue", keyword},
	{token.Defer, "defer", keyword},
	{token.Else, "else", keyword},
	{token.Fallthrough, "fallthrough", keyword},
	{token.For, "for", keyword},
	{token.Func, "func", keyword},
	{token.Goto, "goto", keyword},
	{token.If, "if", keyword},
	{token.Import, "import", keyword},
	{token.Package, "package", keyword},
	{token.Return, "return", keyword},
	{token.Struct, "struct", keyword},
	{token.Switch, "switch", keyword},
	{token.Typedef, "typedef", keyword},
	{token.Var, "var", keyword},
	{token.Void, "void", keyword},
}

func TestScan(t *testing.T) {
	wslines := newlineCount(whitespace)

	file := fset.AddFile("", fset.Base(), len(testSource))
	s := NewLexer(file, testSource, func(_ token.Position, msg string) {
		t.Errorf("error handler called (msg = %s)", msg)
	})
	s.noNewSemi = true

	// set up expected position
	epos := token.Position{
		Filename: "",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	index := 0
	for {
		pos, tok, lit := s.Scan()

		// check position
		if tok == token.EOF {
			// correction for EOF
			epos.Line = newlineCount(string(testSource))
			epos.Column = 2
		}
		checkPos(t, lit, pos, epos)

		// check token
		e := elt{
			tok:   token.EOF,
			lit:   "",
			class: special,
		}
		if index < len(tokens) {
			e = tokens[index]
			index++
		}
		if tok != e.tok {
			t.Errorf("bad token for %q: have %s, want %s", lit, tok, e.tok)
		}
		if tokenclass(tok) != e.class {
			t.Errorf("bad class for %q: have %d, want %d", lit, tokenclass(tok), e.class)
		}

		// check literal
		elit := ""
		switch e.tok {
		case token.Comment:
			// no CRs in comments
			elit = string(stripCR([]byte(e.lit), e.lit[1] == '*'))
			//-style comment literal doesn't contain newline
			if elit[1] == '/' {
				elit = elit[0 : len(elit)-1]
			}
		case token.Ident:
			elit = e.lit
		case token.Semicolon:
			elit = ";"
		default:
			if e.tok.IsLiteral() {
				// no CRs in raw string literals
				elit = e.lit
				if elit[0] == '`' {
					elit = string(stripCR([]byte(elit), false))
				}
			}
			if e.tok.IsKeyword() {
				elit = e.lit
			}
		}
		if lit != elit {
			t.Errorf("bad literal for %q: have %q, want %q", lit, lit, elit)
		}

		if tok == token.EOF {
			break
		}

		// update position
		epos.Offset += len(e.lit) + len(whitespace)
		epos.Line += newlineCount(e.lit) + wslines
	}

	if s.errCount != 0 {
		t.Errorf("found %d errors", s.errCount)
	}
}

func checkPos(t *testing.T, lit string, p token.Pos, want token.Position) {
	pos := fset.Position(p)

	// Check cleaned filenames so that we don't have to worry about
	// different os.PathSeparator values.
	if pos.Filename != want.Filename &&
		filepath.Clean(pos.Filename) != filepath.Clean(want.Filename) {
		t.Errorf("bad filename for %q: have %s, want %s", lit, pos.Filename, want.Filename)
	}
	if pos.Offset != want.Offset {
		t.Errorf("bad position for %q: have %d, want %d", lit, pos.Offset, want.Offset)
	}
	if pos.Line != want.Line {
		t.Errorf("bad line for %q: have %d, want %d", lit, pos.Line, want.Line)
	}
	if pos.Column != want.Column {
		t.Errorf("bad column for %q: have %d, want %d", lit, pos.Column, want.Column)
	}
}

const (
	special  = 0
	literal  = 1
	operator = 2
	keyword  = 3
)

func tokenclass(tok token.Token) int {
	switch {
	case tok.IsLiteral():
		return literal
	case tok.IsOperator():
		return operator
	case tok.IsKeyword():
		return keyword
	default:
		return special
	}
}

func newlineCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

const whitespace = "  \t  \n\n\n" // to separate tokens

var testSource = func() []byte {
	var src []byte
	for _, t := range tokens {
		src = append(src, t.lit...)
		src = append(src, whitespace...)
	}
	return src
}()
