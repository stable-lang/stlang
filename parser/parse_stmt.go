package parser

import (
	"github.com/stable-lang/stlang/ast"
	"github.com/stable-lang/stlang/token"
)

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	leftBrace := p.expect(token.LeftBrace)
	list := p.parseStmtList()
	rightBrace := p.expect(token.RightBrace)

	return &ast.BlockStmt{
		LeftBrace:  leftBrace,
		List:       list,
		RightBrace: rightBrace,
	}
}

func (p *parser) parseStmtList() []ast.Stmt {
	var list []ast.Stmt
	return list
}
