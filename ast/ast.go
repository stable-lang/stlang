// Package declares types used to represent syntax trees for Stable source code.
package ast

import (
	"strings"

	"github.com/stable-lang/stlang/token"
)

// Node interface is implemented by all node types.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node.
	End() token.Pos // position of first character immediately after the node.
}

// Decl interface is implemented by all declaration nodes.
type Decl interface {
	Node
	declNode()
}

// Stmt interface is implemented by all statement nodes.
type Stmt interface {
	Node
	stmtNode()
}

// Expr interface is implemented by all expression nodes.
type Expr interface {
	Node
	exprNode()
}

// IsExported reports whether name starts with an upper-case letter.
func IsExported(name string) bool {
	return token.IsExported(name)
}

// Unparen returns the expression with any enclosing parentheses removed.
func Unparen(e Expr) Expr {
	for {
		paren, ok := e.(*ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}

// File node represents a Stable source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.
type File struct {
	FileStart token.Pos // start of the entire file
	FileEnd   token.Pos // end of the entire file

	Doc     *CommentGroup // associated documentation; or nil
	Package token.Pos     // position of "package" keyword
	PkgName *Ident        // package name

	Imports  []*ImportDecl   // imports in this file
	Decls    []Decl          // top-level declarations; or nil
	Comments []*CommentGroup // list of all comments in the source file
}

// Pos returns the position of the package declaration.
// It may be invalid, for example in an empty file.
// Use FileStart for the start of the entire file. It is always valid.
func (f *File) Pos() token.Pos { return f.Package }

// End returns the end of the last declaration in the file.
// It may be invalid, for example in an empty file.
// Use FileEnd for the end of the entire file. It is always valid.
func (f *File) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.PkgName.End()
}

// Comment node represents a single //-style or /*-style comment.
type Comment struct {
	Slash token.Pos // position of "/" starting the comment.
	Text  string    // comment text (excluding '\n' for //-style comments).
}

func (c *Comment) Pos() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

// CommentGroup node represents a sequence of comments
// with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (cg *CommentGroup) Pos() token.Pos { return cg.List[0].Pos() }
func (cg *CommentGroup) End() token.Pos { return cg.List[len(cg.List)-1].End() }

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed.
// Multiple empty lines are reduced to one, and trailing space on lines is trimmed.
// Unless the result is empty, it is newline-terminated.
func (cg *CommentGroup) Text() string {
	if cg == nil {
		return ""
	}

	comments := make([]string, len(cg.List))
	for i, c := range cg.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines

	for _, c := range comments {
		// Remove comment markers. The parser has given us exactly the comment text.
		switch c[1] {
		case '/': //-style comment (no newline at the end)
			c = c[2:]
			if c == "" {
				break
			}
			if c[0] == ' ' {
				// strip first space - required for Example tests
				c = c[1:]
				break
			}
		case '*': /*-style comment */
			c = c[2 : len(c)-2]
		}

		cl := strings.Split(c, "\n")
		for _, l := range cl {
			lines = append(lines, stripTrailingWS(l))
		}
	}

	// Remove leading blank lines.
	// Convert runs of interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func stripTrailingWS(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

func isWhitespace(ch byte) bool {
	return ch == ' ' ||
		ch == '\t' ||
		ch == '\n' ||
		ch == '\r'
}

// Field represents a Field declaration list in a struct type,
// or a parameter/result declaration in a function signature.
type Field struct {
	Doc     *CommentGroup // associated documentation; or nil
	Names   []*Ident      // field/parameter names; or nil
	Type    Expr          // field/parameter type; or nil
	Comment *CommentGroup // line comments; or nil
}

func (f *Field) Pos() token.Pos {
	switch {
	case len(f.Names) > 0:
		return f.Names[0].Pos()
	case f.Type != nil:
		return f.Type.Pos()
	default:
		return token.NoPos
	}
}

func (f *Field) End() token.Pos {
	switch {
	case f.Type != nil:
		return f.Type.End()
	case len(f.Names) > 0:
		return f.Names[len(f.Names)-1].End()
	default:
		return token.NoPos
	}
}

// FieldList represents a list of Fields, enclosed by parentheses, curly braces, or square brackets.
type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace/bracket, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace/bracket, if any
}

func (f *FieldList) Pos() token.Pos {
	switch {
	case f.Opening.IsValid():
		return f.Opening
	// the list should not be empty in this case; be conservative and guard against bad ASTs
	case len(f.List) > 0:
		return f.List[0].Pos()
	default:
		return token.NoPos
	}
}

func (f *FieldList) End() token.Pos {
	switch {
	case f.Closing.IsValid():
		return f.Closing + 1
		// the list should not be empty in this case; be conservative and guard against bad ASTs
	case len(f.List) > 0:
		return f.List[len(f.List)-1].End()
	default:
		return token.NoPos
	}
}

var _ = []Node{
	&File{},
	&Comment{},
	&CommentGroup{},
	&Field{},
	&FieldList{},
}
