package ast

import "github.com/stable-lang/stlang/token"

// BadStmt node is a placeholder for statements containing syntax errors
// for which no correct statement nodes can be created.
type BadStmt struct {
	From, To token.Pos // position range of bad statement
}

// AssignStmt node represents an assignment or a short variable declaration.
type AssignStmt struct {
	LHS    []Expr
	TokPos token.Pos   // position of Tok
	Tok    token.Token // assignment token, Define
	RHS    []Expr
}

// BlockStmt node represents a braced statement list.
type BlockStmt struct {
	LeftBrace  token.Pos // position of "{"
	List       []Stmt
	RightBrace token.Pos // position of "}", if any (may be absent due to syntax error)
}

// BranchStmt node represents a break, continue, goto or fallthrough statement.
type BranchStmt struct {
	TokPos token.Pos   // position of Tok
	Tok    token.Token // keyword token (break, continue, goto, fallthrough)
	Label  *Ident      // label name; or nil
}

// CaseStmt represents a case of an expression or type switch statement.
type CaseStmt struct {
	Case  token.Pos // position of "case" or "else" keyword
	List  []Expr    // list of expressions or types; nil means "else" case
	Colon token.Pos // position of ":"
	Body  []Stmt    // statement list; or nil
}

// DeclStmt node represents a declaration in a statement list.
type DeclStmt struct {
	Decl Decl // All Decl except ImportDecl token
}

// DeferStmt node represents a defer statement.
type DeferStmt struct {
	Defer token.Pos // position of "defer" keyword
	Body  *BlockStmt
}

// EmptyStmt node represents an empty statement.
// The "position" of the empty statement is the position
// of the immediately following (explicit or implicit) semicolon.
type EmptyStmt struct {
	Semicolon token.Pos // position of following ";"
	Implicit  bool      // if set, ";" was omitted in the source
}

// ExprStmt node represents a (stand-alone) expression in a statement list.
type ExprStmt struct {
	X Expr // expression
}

// ForStmt represents a for statement.
type ForStmt struct {
	For  token.Pos // position of "for" keyword
	Init Stmt      // initialization statement; or nil
	Cond Expr      // condition; or nil
	Post Stmt      // post iteration statement; or nil
	Body *BlockStmt
}

// IfStmt node represents an if statement.
type IfStmt struct {
	If   token.Pos // position of "if" keyword
	Init Stmt      // initialization statement; or nil
	Cond Expr      // condition
	Body *BlockStmt
	Else Stmt // else branch; or nil
}

// LabeledStmt node represents a labeled statement.
type LabeledStmt struct {
	Label *Ident
	Colon token.Pos // position of ":"
	Stmt  Stmt
}

// ReturnStmt node represents a return statement.
type ReturnStmt struct {
	Return  token.Pos // position of "return" keyword
	Results []Expr    // result expressions; or nil
}

// SwitchStmt node represents an expression switch statement.
type SwitchStmt struct {
	Switch token.Pos  // position of "switch" keyword
	Init   Stmt       // initialization statement; or nil
	Tag    Expr       // tag expression; or nil
	Body   *BlockStmt // CaseStmts only
}

func (s *BadStmt) Pos() token.Pos     { return s.From }
func (s *AssignStmt) Pos() token.Pos  { return s.LHS[0].Pos() }
func (s *BlockStmt) Pos() token.Pos   { return s.LeftBrace }
func (s *BranchStmt) Pos() token.Pos  { return s.TokPos }
func (s *CaseStmt) Pos() token.Pos    { return s.Case }
func (s *DeclStmt) Pos() token.Pos    { return s.Decl.Pos() }
func (s *DeferStmt) Pos() token.Pos   { return s.Defer }
func (s *EmptyStmt) Pos() token.Pos   { return s.Semicolon }
func (s *ExprStmt) Pos() token.Pos    { return s.X.Pos() }
func (s *ForStmt) Pos() token.Pos     { return s.For }
func (s *IfStmt) Pos() token.Pos      { return s.If }
func (s *LabeledStmt) Pos() token.Pos { return s.Label.Pos() }
func (s *ReturnStmt) Pos() token.Pos  { return s.Return }
func (s *SwitchStmt) Pos() token.Pos  { return s.Switch }

func (s *BadStmt) End() token.Pos    { return s.To }
func (s *AssignStmt) End() token.Pos { return s.RHS[len(s.RHS)-1].End() }
func (s *BlockStmt) End() token.Pos {
	if s.RightBrace.IsValid() {
		return s.RightBrace + 1
	}
	if n := len(s.List); n > 0 {
		return s.List[n-1].End()
	}
	return s.LeftBrace + 1
}

func (s *BranchStmt) End() token.Pos {
	if s.Label != nil {
		return s.Label.End()
	}
	return token.Pos(int(s.TokPos) + len(s.Tok.String()))
}

func (s *CaseStmt) End() token.Pos {
	if n := len(s.Body); n > 0 {
		return s.Body[n-1].End()
	}
	return s.Colon + 1
}
func (s *DeclStmt) End() token.Pos  { return s.Decl.End() }
func (s *DeferStmt) End() token.Pos { return s.Body.End() }
func (s *EmptyStmt) End() token.Pos {
	if s.Implicit {
		return s.Semicolon
	}
	return s.Semicolon + token.Pos(len(";"))
}
func (s *ExprStmt) End() token.Pos { return s.X.End() }
func (s *ForStmt) End() token.Pos  { return s.Body.End() }
func (s *IfStmt) End() token.Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}
func (s *LabeledStmt) End() token.Pos { return s.Stmt.End() }
func (s *ReturnStmt) End() token.Pos {
	if n := len(s.Results); n > 0 {
		return s.Results[n-1].End()
	}
	return s.Return + token.Pos(len("return"))
}
func (s *SwitchStmt) End() token.Pos { return s.Body.End() }

func (*BadStmt) stmtNode()     {}
func (*AssignStmt) stmtNode()  {}
func (*BlockStmt) stmtNode()   {}
func (*BranchStmt) stmtNode()  {}
func (*CaseStmt) stmtNode()    {}
func (*DeclStmt) stmtNode()    {}
func (*DeferStmt) stmtNode()   {}
func (*EmptyStmt) stmtNode()   {}
func (*ExprStmt) stmtNode()    {}
func (*ForStmt) stmtNode()     {}
func (*IfStmt) stmtNode()      {}
func (*LabeledStmt) stmtNode() {}
func (*ReturnStmt) stmtNode()  {}
func (*SwitchStmt) stmtNode()  {}

var _ = []Node{
	&BadStmt{},
	&AssignStmt{},
	&BlockStmt{},
	&BranchStmt{},
	&CaseStmt{},
	&DeclStmt{},
	&DeferStmt{},
	&EmptyStmt{},
	&ExprStmt{},
	&ForStmt{},
	&IfStmt{},
	&LabeledStmt{},
	&ReturnStmt{},
	&SwitchStmt{},
}
