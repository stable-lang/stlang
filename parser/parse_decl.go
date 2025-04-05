package parser

import (
	"github.com/stable-lang/stlang/ast"
	"github.com/stable-lang/stlang/token"
)

func (p *parser) parsePackageDecl() (*ast.CommentGroup, token.Pos, *ast.Ident) {
	doc := p.leadComment
	pos := p.expect(token.Package)

	ident := p.parseIdent()
	switch ident.Name {
	case "_":
		p.error(p.pos, "invalid package name _")
	case "builtin", "init", "internal", "vendor":
		p.error(p.pos, "package name '%s' is reserved", ident.Name)
	}
	p.expectSemi()

	return doc, pos, ident
}

func (p *parser) parseDecl(sync map[token.Token]bool) ast.Decl {
	switch p.tok {
	case token.Const:
		return p.parseConstDecl()
	case token.Func:
		return p.parseFuncDecl()
	case token.Import:
		return p.parseImportDecl()
	case token.Struct:
		return p.parseStructDecl()
	case token.Typedef:
		return p.parseTypedefDecl()
	case token.Var:
		return p.parseVarDecl()
	default:
		pos := p.pos
		p.errorExpected(pos, "declaration")
		p.advance(sync)
		return &ast.BadDecl{From: pos, To: p.pos}
	}
}

func (p *parser) parseConstDecl() *ast.ConstDecl {
	doc := p.leadComment
	p.expect(token.Const)

	name := p.parseIdent()
	typ := p.tryIdentOrType()
	p.expect(token.Assign)
	value := p.parseIdent() // TODO(oleg): must be an expression.

	comment := p.expectSemi()

	return &ast.ConstDecl{
		Doc:     doc,
		Name:    name,
		Type:    typ,
		Value:   value,
		Comment: comment,
	}
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	doc := p.leadComment
	pos := p.expect(token.Func)

	var recv *ast.Ident
	if p.tok == token.LeftParen {
		p.next()
		recv = p.parseIdent()
		p.expect(token.RightParen)
	}

	name := p.parseIdent()
	params := p.parseParameters()
	results := p.parseResult()

	var body *ast.BlockStmt
	if p.tok == token.Semicolon {
		p.next()
		if p.tok == token.LeftBrace {
			p.error(p.pos, "unexpected semicolon or newline before {")
			body = p.parseBlockStmt()
			p.expectSemi()
		} else {
			p.expect(token.LeftBrace)
		}
	} else {
		body = p.parseBlockStmt()
		p.expectSemi()
	}

	return &ast.FuncDecl{
		Doc:  doc,
		Recv: recv,
		Name: name,
		Type: &ast.FuncType{
			Func:    pos,
			Params:  params,
			Results: results,
		},
		Body: body,
	}
}

func (p *parser) parseImportDecl() *ast.ImportDecl {
	doc := p.leadComment
	pos := p.expect(token.Import)

	var ident *ast.Ident
	switch p.tok {
	case token.Ident:
		ident = p.parseIdent()
	case token.Period:
		ident = &ast.Ident{NamePos: p.pos, Name: "."}
		p.next()
	}

	var path string
	switch {
	case p.tok == token.String:
		path = p.lit
		p.next()
	case p.tok.IsLiteral():
		p.error(pos, "import path must be a string")
		p.next()
	default:
		p.error(pos, "missing import path")
		p.advance(exprEnd)
	}
	comment := p.expectSemi()

	return &ast.ImportDecl{
		Doc:  doc,
		Name: ident,
		Path: &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.String,
			Value:    path,
		},
		Comment: comment,
	}
}

func (p *parser) parseStructDecl() *ast.StructDecl {
	doc := p.leadComment
	p.expect(token.Struct)
	name := p.parseIdent()

	leftBrace := p.expect(token.LeftBrace)
	var list []*ast.Field
	for p.tok == token.Ident {
		list = append(list, p.parseFieldDecl())
	}
	rightBrace := p.expect(token.RightBrace)

	p.expectSemi()

	return &ast.StructDecl{
		Doc:  doc,
		Name: name,
		Fields: &ast.FieldList{
			Opening: leftBrace,
			List:    list,
			Closing: rightBrace,
		},
	}
}

func (p *parser) parseTypedefDecl() *ast.TypedefDecl {
	doc := p.leadComment
	p.expect(token.Typedef)

	name := p.parseIdent()

	var assignPos token.Pos
	if p.tok == token.Assign { // type alias
		assignPos = p.pos
		p.next()
	}
	typ := p.parseType()

	comment := p.expectSemi()

	return &ast.TypedefDecl{
		Doc:     doc,
		Name:    name,
		Assign:  assignPos,
		Type:    typ,
		Comment: comment,
	}
}

func (p *parser) parseVarDecl() *ast.VarDecl {
	doc := p.leadComment
	p.expect(token.Var)

	name := p.parseIdent()
	typ := p.tryIdentOrType()
	p.expect(token.Assign)
	value := p.parseIdent() // TODO(oleg): must be an expression.

	comment := p.expectSemi()

	return &ast.VarDecl{
		Doc:     doc,
		Name:    name,
		Type:    typ,
		Value:   value,
		Comment: comment,
	}
}

func (p *parser) parseFieldDecl() *ast.Field {
	doc := p.leadComment

	var names []*ast.Ident
	var typ ast.Expr

	if p.tok == token.Ident {
		name := p.parseIdent()

		names = []*ast.Ident{name}
		for p.tok == token.Comma {
			p.next()
			names = append(names, p.parseIdent())
		}
		typ = p.parseType()
	} else {
		pos := p.pos
		p.errorExpected(pos, "field name")
		p.advance(exprEnd)
		typ = &ast.BadExpr{From: pos, To: p.pos}
	}

	comment := p.expectSemi()

	return &ast.Field{
		Doc:     doc,
		Names:   names,
		Type:    typ,
		Comment: comment,
	}
}

func (p *parser) parseParameters() (params *ast.FieldList) {
	leftParen := p.expect(token.LeftParen)

	var fields []*ast.Field
	if p.tok != token.RightParen {
		fields = p.parseParameterList()
	}

	rightParen := p.expect(token.RightParen)

	return &ast.FieldList{
		Opening: leftParen,
		List:    fields,
		Closing: rightParen,
	}
}

func (p *parser) parseParameterList() []*ast.Field {
	var params []*ast.Field
	for p.tok != token.RightParen {
		p.next()
	}
	return params
}

func (p *parser) parseResult() *ast.FieldList {
	if p.tok == token.LeftParen {
		return p.parseParameters()
	}

	if typ := p.tryIdentOrType(); typ != nil {
		return &ast.FieldList{
			List: []*ast.Field{{Type: typ}},
		}
	}
	return nil
}
