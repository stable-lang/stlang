package ast

import "github.com/stable-lang/stlang/token"

// BadExpr node is a placeholder for an expression containing syntax errors
// for which a correct expression node cannot be created.
type BadExpr struct {
	From, To token.Pos // position range of bad expression
}

// Ident node represents an identifier.
type Ident struct {
	NamePos token.Pos // identifier position
	Name    string    // identifier name
}

// IsExported reports whether id starts with an upper-case letter.
func (id *Ident) IsExported() bool { return IsExported(id.Name) }

func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// BasicLit node represents a literal of basic type.
type BasicLit struct {
	ValuePos token.Pos   // literal position
	Kind     token.Token // token.Int, token.Float, token.Char, or token.String
	Value    string      // literal string; e.g. 42, 0x7f, 3.14, 'a', '\x7f', "foo" or `\m\n\o`
}

// CompositeLit node represents a composite literal.
type CompositeLit struct {
	Type       Expr      // literal type; or nil
	LeftBrace  token.Pos // position of "{"
	ElemTypes  []Expr    // list of composite elements; or nil
	RightBrace token.Pos // position of "}"
	Incomplete bool      // true if (source) expressions are missing in the ElemTypes list
}

// FuncLit node represents a function literal.
type FuncLit struct {
	Type *FuncType  // function type
	Body *BlockStmt // function body
}

// expressions

// BinaryExpr node represents a binary expression.
type BinaryExpr struct {
	X     Expr        // left operand
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	Y     Expr        // right operand
}

// CallExpr node represents an expression followed by an argument list.
type CallExpr struct {
	Fun        Expr      // function expression
	LeftParen  token.Pos // position of "("
	Args       []Expr    // function arguments; or nil
	Ellipsis   token.Pos // position of "..." (token.NoPos if there is no "...")
	RightParen token.Pos // position of ")"
}

// Ellipsis node stands for the "..." type in a parameter list.
type Ellipsis struct {
	Ellipsis token.Pos // position of "..."
	ElemType Expr      // ellipsis element type (parameter lists only); or nil
}

// IndexExpr node represents an expression followed by an index.
type IndexExpr struct {
	X          Expr      // expression
	LeftBrack  token.Pos // position of "["
	Index      Expr      // index expression
	RightBrack token.Pos // position of "]"
}

// IndexListExpr node represents an expression followed by multiple indices.
type IndexListExpr struct {
	X          Expr      // expression
	LeftBrack  token.Pos // position of "["
	Indices    []Expr    // index expressions
	RightBrack token.Pos // position of "]"
}

// KeyValueExpr node represents (key : value) pairs in composite literals.
type KeyValueExpr struct {
	Key   Expr
	Colon token.Pos // position of ":"
	Value Expr
}

// ParenExpr node represents a parenthesized expression.
type ParenExpr struct {
	LeftParen  token.Pos // position of "("
	X          Expr      // parenthesized expression
	RightParen token.Pos // position of ")"
}

// SelectorExpr node represents an expression followed by a selector.
type SelectorExpr struct {
	X   Expr   // expression
	Sel *Ident // field selector
}

// SliceExpr node represents an expression followed by slice indices.
type SliceExpr struct {
	X          Expr      // expression
	LeftBrack  token.Pos // position of "["
	Low        Expr      // begin of slice range; or nil
	High       Expr      // end of slice range; or nil
	Max        Expr      // maximum capacity of slice; or nil
	Slice3     bool      // true if 3-index slice (2 colons present)
	RightBrack token.Pos // position of "]"
}

// StarExpr node represents an expression of the form "*" Expression.
// Semantically it could be a unary "*" expression, or a pointer type.
type StarExpr struct {
	Star token.Pos // position of "*"
	X    Expr      // operand
}

// UnaryExpr node represents a unary expression.
// Unary "*" expressions are represented via StarExpr nodes.
type UnaryExpr struct {
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	X     Expr        // operand
}

// types

// ArrayType node represents an array type.
type ArrayType struct {
	LeftBrack token.Pos // position of "["
	Len       Expr      // Ident or Expr for size
	ElemType  Expr      // element type
}

// FuncType node represents a function type.
type FuncType struct {
	Func    token.Pos  // position of "func" keyword (token.NoPos if there is no "func")
	Params  *FieldList // (incoming) parameters; non-nil
	Results *FieldList // (outgoing) results; or nil
}

// MapType node represents a map type.
type MapType struct {
	LeftBrack token.Pos // position of "["
	KeyType   Expr
	ValueType Expr
}

// SliceType node represents a slice type.
type SliceType struct {
	LeftBrack token.Pos // position of "["
	ElemType  Expr      // element type
}

// StructType node represents a struct type.
type StructType struct {
	Struct token.Pos  // position of "struct" keyword
	Fields *FieldList // list of field declarations
}

func (x *BadExpr) Pos() token.Pos  { return x.From }
func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *CompositeLit) Pos() token.Pos {
	if x.Type != nil {
		return x.Type.Pos()
	}
	return x.LeftBrace
}
func (x *FuncLit) Pos() token.Pos       { return x.Type.Pos() }
func (x *BinaryExpr) Pos() token.Pos    { return x.X.Pos() }
func (x *CallExpr) Pos() token.Pos      { return x.Fun.Pos() }
func (x *Ellipsis) Pos() token.Pos      { return x.Ellipsis }
func (x *IndexExpr) Pos() token.Pos     { return x.X.Pos() }
func (x *IndexListExpr) Pos() token.Pos { return x.X.Pos() }
func (x *KeyValueExpr) Pos() token.Pos  { return x.Key.Pos() }
func (x *ParenExpr) Pos() token.Pos     { return x.LeftParen }
func (x *SelectorExpr) Pos() token.Pos  { return x.X.Pos() }
func (x *SliceExpr) Pos() token.Pos     { return x.X.Pos() }
func (x *StarExpr) Pos() token.Pos      { return x.Star }
func (x *UnaryExpr) Pos() token.Pos     { return x.OpPos }
func (x *ArrayType) Pos() token.Pos     { return x.LeftBrack }
func (x *FuncType) Pos() token.Pos {
	if x.Func.IsValid() || x.Params == nil { // see issue 3870
		return x.Func
	}
	return x.Params.Pos() // interface method declarations have no "func" keyword
}
func (x *MapType) Pos() token.Pos    { return x.LeftBrack }
func (x *SliceType) Pos() token.Pos  { return x.LeftBrack }
func (x *StructType) Pos() token.Pos { return x.Struct }

func (x *BadExpr) End() token.Pos      { return x.To }
func (x *Ident) End() token.Pos        { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() token.Pos     { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *CompositeLit) End() token.Pos { return x.RightBrace + 1 }
func (x *FuncLit) End() token.Pos      { return x.Body.End() }
func (x *BinaryExpr) End() token.Pos   { return x.Y.End() }
func (x *CallExpr) End() token.Pos     { return x.RightParen + 1 }
func (x *Ellipsis) End() token.Pos {
	if x.ElemType != nil {
		return x.ElemType.End()
	}
	return x.Ellipsis + token.Pos(len("..."))
}
func (x *IndexExpr) End() token.Pos     { return x.RightBrack + 1 }
func (x *IndexListExpr) End() token.Pos { return x.RightBrack + 1 }
func (x *KeyValueExpr) End() token.Pos  { return x.Value.End() }
func (x *ParenExpr) End() token.Pos     { return x.RightParen + 1 }
func (x *SelectorExpr) End() token.Pos  { return x.Sel.End() }
func (x *SliceExpr) End() token.Pos     { return x.RightBrack + 1 }
func (x *StarExpr) End() token.Pos      { return x.X.End() }
func (x *UnaryExpr) End() token.Pos     { return x.X.End() }
func (x *ArrayType) End() token.Pos     { return x.ElemType.End() }
func (x *FuncType) End() token.Pos {
	if x.Results != nil {
		return x.Results.End()
	}
	return x.Params.End()
}
func (x *MapType) End() token.Pos    { return x.ValueType.End() }
func (x *SliceType) End() token.Pos  { return x.ElemType.End() }
func (x *StructType) End() token.Pos { return x.Fields.End() }

func (*BadExpr) exprNode()       {}
func (*Ident) exprNode()         {}
func (*BasicLit) exprNode()      {}
func (*CompositeLit) exprNode()  {}
func (*FuncLit) exprNode()       {}
func (*BinaryExpr) exprNode()    {}
func (*CallExpr) exprNode()      {}
func (*Ellipsis) exprNode()      {}
func (*IndexExpr) exprNode()     {}
func (*IndexListExpr) exprNode() {}
func (*KeyValueExpr) exprNode()  {}
func (*ParenExpr) exprNode()     {}
func (*SelectorExpr) exprNode()  {}
func (*SliceExpr) exprNode()     {}
func (*StarExpr) exprNode()      {}
func (*UnaryExpr) exprNode()     {}
func (*ArrayType) exprNode()     {}
func (*FuncType) exprNode()      {}
func (*MapType) exprNode()       {}
func (*SliceType) exprNode()     {}
func (*StructType) exprNode()    {}

var _ = []Node{
	&BadExpr{},
	&Ident{},
	&BasicLit{},
	&CompositeLit{},
	&FuncLit{},

	&BinaryExpr{},
	&CallExpr{},
	&Ellipsis{},
	&IndexExpr{},
	&IndexListExpr{},
	&KeyValueExpr{},
	&ParenExpr{},
	&SelectorExpr{},
	&SliceExpr{},
	&StarExpr{},
	&UnaryExpr{},

	&ArrayType{},
	&FuncType{},
	&MapType{},
	&SliceType{},
	&StructType{},
}
