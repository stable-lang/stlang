package parser

import (
	"github.com/stable-lang/stlang/ast"
	"github.com/stable-lang/stlang/token"
)

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.Ident {
		name = p.lit
		p.next()
	} else {
		p.expect(token.Ident)
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

// types

func (p *parser) parseType() ast.Expr {
	if typ := p.tryIdentOrType(); typ != nil {
		return typ
	}

	pos := p.pos
	p.errorExpected(pos, "type")
	p.advance(exprEnd)
	return &ast.BadExpr{
		From: pos,
		To:   p.pos,
	}
}

func (p *parser) tryIdentOrType() ast.Expr {
	switch p.tok {
	case token.Any, token.Bool, token.Void:
		p.next()
		return &ast.Ident{
			NamePos: p.pos,
			Name:    p.tok.String(),
		}
	case token.Ident:
		return p.parseTypeName(nil)
	default:
		return nil // no type found
	}
}

func (p *parser) parseTypeName(ident *ast.Ident) ast.Expr {
	if ident == nil {
		ident = p.parseIdent()
	}

	if p.tok == token.Period {
		// ident is a package name
		p.next()
		sel := p.parseIdent()
		return &ast.SelectorExpr{X: ident, Sel: sel}
	}
	return ident
}
