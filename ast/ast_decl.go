package ast

import "github.com/stable-lang/stlang/token"

// BadDecl node is a placeholder for a declaration containing syntax errors
// for which a correct declaration node cannot be created.
type BadDecl struct {
	From, To token.Pos // position range of bad declaration.
}

// ConstDecl node represents a constant declaration.
type ConstDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // constant name
	Type    Expr          // constant type; or nil
	Value   Expr          // initial value
	Comment *CommentGroup // line comments; or nil
}

// FuncDecl node represents a function declaration.
type FuncDecl struct {
	Doc  *CommentGroup // associated documentation; or nil
	Recv *Ident        // receiver (for methods); or nil (for functions)
	Name *Ident        // function/method name
	Type *FuncType     // function signature: type and value parameters, results, and position of "func" keyword
	Body *BlockStmt    // function body; or nil for external (non-Stable) function
}

// ImportDecl node represents a single package import.
type ImportDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // local package name (including "."); or nil
	Path    *BasicLit     // import path
	Comment *CommentGroup // line comments; or nil
	EndPos  token.Pos     // end of decl (overrides Path.Pos if nonzero)
}

// StructDecl node represents a structure declaration.
type StructDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // struct name
	Fields  *FieldList    // list of field declarations
	Comment *CommentGroup // line comments; or nil
}

// TypedefDecl node represents a type definition.
type TypedefDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // type name
	Assign  token.Pos     // position of '=', if any
	Type    Expr          // *SelectorExpr, *StarExpr, or any of the *XxxTypes
	Comment *CommentGroup // line comments; or nil
}

// VarDecl node represents a variable declaration.
type VarDecl struct {
	Doc     *CommentGroup // associated documentation; or nil
	Name    *Ident        // variable name
	Type    Expr          // variable type; or nil
	Value   Expr          // initial value
	Comment *CommentGroup // line comments; or nil
}

func (d *BadDecl) Pos() token.Pos   { return d.From }
func (d *ConstDecl) Pos() token.Pos { return d.Name.Pos() }
func (d *FuncDecl) Pos() token.Pos  { return d.Type.Pos() }
func (d *ImportDecl) Pos() token.Pos {
	if d.Name != nil {
		return d.Name.Pos()
	}
	return d.Path.Pos()
}
func (d *StructDecl) Pos() token.Pos  { return d.Name.Pos() }
func (d *TypedefDecl) Pos() token.Pos { return d.Name.Pos() }
func (d *VarDecl) Pos() token.Pos     { return d.Name.Pos() }

func (d *BadDecl) End() token.Pos   { return d.To }
func (d *ConstDecl) End() token.Pos { return d.Value.End() }
func (d *FuncDecl) End() token.Pos {
	if d.Body != nil {
		return d.Body.End()
	}
	return d.Type.End()
}

func (d *ImportDecl) End() token.Pos {
	if d.EndPos != 0 {
		return d.EndPos
	}
	return d.Path.End()
}
func (d *StructDecl) End() token.Pos  { return d.Fields.End() }
func (d *TypedefDecl) End() token.Pos { return d.Type.End() }
func (d *VarDecl) End() token.Pos     { return d.Value.End() }

func (*BadDecl) declNode()     {}
func (*ConstDecl) declNode()   {}
func (*FuncDecl) declNode()    {}
func (*ImportDecl) declNode()  {}
func (*StructDecl) declNode()  {}
func (*TypedefDecl) declNode() {}
func (*VarDecl) declNode()     {}

var _ = []Node{
	&BadDecl{},
	&ConstDecl{},
	&FuncDecl{},
	&ImportDecl{},
	&StructDecl{},
	&TypedefDecl{},
	&VarDecl{},
}
