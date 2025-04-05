// Package parser implements a parser for Stable source files.
package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/stable-lang/stlang/ast"
	"github.com/stable-lang/stlang/lexer"
	"github.com/stable-lang/stlang/token"
)

// ParseFile of a single Stable source file and returns the corresponding [ast.File] node.
// The source code may be provided via the filename of the source file, or via the src parameter.
func ParseFile(fset *token.FileSet, filename string, src any) (f *ast.File, err error) {
	if fset == nil {
		panic("parser.ParseFile: no token.FileSet provided")
	}

	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	file := fset.AddFile(filename, -1, len(text))

	var p parser
	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}

		if f == nil {
			f = &ast.File{
				PkgName: new(ast.Ident),
			}
		}

		// Ensure the start/end are consistent, whether parsing succeeded or not.
		f.FileStart = token.Pos(file.Base())
		f.FileEnd = token.Pos(file.Base() + file.Size())

		p.errors.sort()
		err = p.errors.Err()
	}()

	p.init(file, text)
	f = p.parseFile()

	return f, err
}

func readSource(filename string, src any) ([]byte, error) {
	if src != nil {
		switch src := src.(type) {
		case string:
			return []byte(src), nil
		case []byte:
			return src, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if src != nil {
				return src.Bytes(), nil
			}
			return nil, errors.New("invalid source")
		case io.Reader:
			return io.ReadAll(src)
		case fs.FS:
			return fs.ReadFile(src, filename)
		}
	}
	return os.ReadFile(filename)
}

type parser struct {
	file    *token.File
	errors  ErrorList
	scanner *lexer.Lexer

	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	syncPos   token.Pos // last synchronization position
	syncCount int       // number of parser.advance calls without progress

	exprLevel int  // < 0: in control clause, >= 0: in expression
	inRHS     bool // if set, the parser is parsing a RHS expression
	nestLevel int  // nestLevel is used to track and limit the recursion depth during parsing.
}

func (p *parser) init(file *token.File, src []byte) {
	p.file = file
	errFn := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner = lexer.NewLexer(p.file, src, errFn)

	p.next()
}

func (p *parser) parseFile() *ast.File {
	// likely not a Stable source file at all.
	if p.errors.Len() != 0 {
		return nil
	}

	doc, pos, ident := p.parsePackageDecl()
	if p.errors.Len() != 0 {
		return nil
	}

	var decls []ast.Decl
	var imports []*ast.ImportDecl

	for p.tok == token.Import {
		decl := p.parseImportDecl()
		decls = append(decls, decl)
		imports = append(imports, decl)
	}

	prev := token.Import
	for p.tok != token.EOF {
		// accept imports but complain.
		if p.tok == token.Import && prev != token.Import {
			p.error(p.pos, "imports must appear before other declarations")
		}
		prev = p.tok

		decl := p.parseDecl(declStart)
		decls = append(decls, decl)

		if imp, ok := decl.(*ast.ImportDecl); ok {
			imports = append(imports, imp)
		}
	}

	return &ast.File{
		Doc:     doc,
		Package: pos,
		PkgName: ident,
		Decls:   decls,
		// File{Start,End} are set by the defer in the caller.
		Imports:  imports,
		Comments: p.comments,
	}
}

// advance to the next non-comment token.
// In the process, collect any comment groups encountered,
// and remember the last lead and line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is stored in the AST.
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos

	p.next0()
	if p.tok != token.Comment {
		return
	}

	var comment *ast.CommentGroup
	var endline int
	if p.file.Line(p.pos) == p.file.Line(prev) {
		// The comment is on same line as the previous token;
		// it cannot be a lead comment but may be a line comment.
		comment, endline = p.consumeCommentGroup(0)
		if p.file.Line(p.pos) != endline || p.tok == token.Semicolon || p.tok == token.EOF {
			// The next token is on a different line, thus
			// the last comment group is a line comment.
			p.lineComment = comment
		}
	}

	// consume successor comments, if any
	endline = -1
	for p.tok == token.Comment {
		comment, endline = p.consumeCommentGroup(1)
	}

	if endline+1 == p.file.Line(p.pos) {
		// The next token is following on the line immediately after the
		// comment group, thus the last comment group is a lead comment.
		p.leadComment = comment
	}
}

// Advance to the next token.
func (p *parser) next0() {
	for {
		p.pos, p.tok, p.lit = p.scanner.Scan()
		if p.tok != token.Comment {
			break
		}
	}
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.Comment && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return comments, endline
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{
		Slash: p.pos,
		Text:  p.lit,
	}
	p.next0()

	return comment, endline
}

// expectSemi consumes a semicolon and returns the applicable line comment.
func (p *parser) expectSemi() *ast.CommentGroup {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RightParen && p.tok != token.RightBrace {
		switch p.tok {
		case token.Comma:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.Semicolon:
			var comment *ast.CommentGroup
			if p.lit == ";" {
				// explicit semicolon
				p.next()
				comment = p.lineComment // use following comments
			} else {
				// artificial semicolon
				comment = p.lineComment // use preceding comments
				p.next()
			}
			return comment
		default:
			p.errorExpected(p.pos, "';'")
			p.advance(stmtStart)
		}
	}
	return nil
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or token.EOF. For error recovery.
func (p *parser) advance(to map[token.Token]bool) {
	for ; p.tok != token.EOF; p.next() {
		if !to[p.tok] {
			continue
		}

		// Return only if parser made some progress since last
		// sync or if it has not reached 10 advance calls without
		// progress. Otherwise consume at least one token to
		// avoid an endless parser loop (it is possible that
		// both parseOperand and parseStmt call advance and
		// correctly do not advance, thus the need for the
		// invocation limit p.syncCount).
		if p.pos == p.syncPos && p.syncCount < 10 {
			p.syncCount++
			return
		}
		if p.pos > p.syncPos {
			p.syncPos = p.pos
			p.syncCount = 0
			return
		}
		// Reaching here indicates a parser bug, likely an
		// incorrect token list in this function, but it only
		// leads to skipping of possibly correct code if a
		// previous error is present, and thus is preferred
		// over a non-terminating parse.
	}
}

var declStart = map[token.Token]bool{
	token.Const:   true,
	token.Func:    true,
	token.Import:  true,
	token.Struct:  true,
	token.Typedef: true,
	token.Var:     true,
}

var stmtStart = map[token.Token]bool{
	token.Break:       true,
	token.Const:       true,
	token.Continue:    true,
	token.Defer:       true,
	token.Fallthrough: true,
	token.For:         true,
	token.Goto:        true,
	token.If:          true,
	token.Return:      true,
	token.Switch:      true,
	token.Typedef:     true,
	token.Var:         true,
}

var exprEnd = map[token.Token]bool{
	token.Comma:      true,
	token.Colon:      true,
	token.Semicolon:  true,
	token.RightParen: true,
	token.RightBrack: true,
	token.RightBrace: true,
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok == token.Semicolon && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) error(pos token.Pos, msg string, args ...any) {
	epos := p.file.Position(pos)

	p.errors.Add(epos, fmt.Sprintf(msg, args...))
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}
