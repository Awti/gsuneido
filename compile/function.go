package compile

import . "github.com/apmckinlay/gsuneido/lexer"

func ParseFunction(src string) Ast {
	p := newParser(src)
	return p.function()
}

func (p *parser) function() Ast {
	it := p.Item
	p.match(FUNCTION)
	p.match(L_PAREN)
	p.match(R_PAREN)
	body := p.compound()
	return ast(it, ast(Item{Token: PARAMS}), body)
}

func (p *parser) compound() Ast {
	p.match(L_CURLY)
	stmts := p.statements()
	p.match(R_CURLY)
	return stmts
}

func (p *parser) statements() Ast {
	list := []Ast{}
	for p.Token != R_CURLY {
		if p.matchIf(NEWLINE) || p.matchIf(SEMICOLON) {
			continue
		}
		stmt := p.statement()
		//fmt.Println("stmt:", stmt)
		list = append(list, stmt)
	}
	return ast(code, list...)
}

var code = Item{Token: STATEMENTS}

func (p *parser) statement() Ast {
	if p.Token == NEWLINE {
		p.nextSkipNL()
	}
	if p.Token == L_CURLY {
		return p.compound()
	}
	if p.matchIf(SEMICOLON) {
		return ast(Item{})
	}
	// TODO other statement types
	switch p.Keyword {
	case RETURN:
		return p.returnStmt()
	case IF:
		return p.ifStmt()
	case SWITCH:
		return p.switchStmt()
	case FOREVER:
		return p.foreverStmt()
	case WHILE:
		return p.whileStmt()
	case DO:
		return p.dowhileStmt()
	case FOR:
		return p.forStmt()
	case THROW:
		return p.throwStmt()
	case BREAK, CONTINUE:
		it := p.Item
		p.next()
		p.matchIf(SEMICOLON)
		return ast(it)
	default:
		return p.exprStmt()
	}
}

func (p *parser) ifStmt() Ast {
	it, expr := p.ctrlExpr()
	t := p.statement()
	if p.Keyword == ELSE {
		p.nextSkipNL()
		f := p.statement()
		return ast(it, expr, t, f)
	}
	return ast(it, expr, t)
}

func (p *parser) switchStmt() Ast {
	it := p.Item
	p.nextSkipNL()
	var expr Ast
	if p.Token == L_CURLY {
		expr = ast(Item{Token: TRUE})
	} else {
		expr = p.exprAst()
		if p.Token == NEWLINE {
			p.nextSkipNL()
		}
	}
	p.nextSkipNL()
	var cases []Ast
	for p.matchIf(CASE) {
		cases = p.switchCase(cases)
	}
	result := ast(it, expr, ast2("cases", cases...))
	if p.matchIf(DEFAULT) {
		result.Children = append(result.Children, p.switchBody())
	}
	p.match(R_CURLY)
	return result
}

func (p *parser) switchCase(cases []Ast) []Ast {
	var values []Ast
	for {
		values = append(values, p.exprAst())
		if !p.matchIf(COMMA) {
			break
		}
	}
	body := p.switchBody()
	c := ast(Item{Token: CASE}, ast2("vals", values...), body)
	return append(cases, c)
}

func (p *parser) switchBody() Ast {
	p.match(COLON)
	var stmts []Ast
	for p.Token != R_CURLY && p.Keyword != CASE && p.Keyword != DEFAULT {
		stmts = append(stmts, p.statement())
	}
	return ast(code, stmts...)
}

func (p *parser) foreverStmt() Ast {
	it := p.Item
	p.match(FOREVER)
	body := p.statement()
	return ast(it, body)
}

func (p *parser) whileStmt() Ast {
	it, expr := p.ctrlExpr()
	body := p.statement()
	return ast(it, expr, body)
}

func (p *parser) dowhileStmt() Ast {
	it := p.Item
	p.match(DO)
	body := p.statement()
	p.match(WHILE)
	expr := p.exprAst()
	if p.Token == NEWLINE {
		p.nextSkipNL()
	}
	return ast(it, body, expr)
}

func (p *parser) forStmt() Ast {
	it := p.Item
	forIn := p.isForIn()
	p.match(FOR)
	if forIn {
		return p.forIn(it)
	} else {
		return p.forClassic(it)
	}
}

func (p *parser) isForIn() bool {
	i := 0
	//fmt.Println("isForIn", p.lxr.Ahead(i), p.lxr.Ahead(i+1), p.lxr.Ahead(i+2))
	for ; skip(p.lxr.Ahead(i)); i++ {
	}
	//fmt.Println("isForIn2", p.lxr.Ahead(i), p.lxr.Ahead(i+1), p.lxr.Ahead(i+2))
	if p.lxr.Ahead(i).Token == L_PAREN {
		i++
	}
	for ; skip(p.lxr.Ahead(i)); i++ {
	}
	if p.lxr.Ahead(i).Token != IDENTIFIER {
		return false
	}
	for i++; skip(p.lxr.Ahead(i)); i++ {
	}
	//fmt.Println("SHOULD BE IN", p.lxr.Ahead(i))
	return p.lxr.Ahead(i).Keyword == IN
}

func skip(it Item) bool {
	return it.Token == WHITESPACE || it.Token == NEWLINE || it.Token == COMMENT
}

func (p *parser) forIn(it Item) Ast {
	it.Token = FOR_IN
	parens := p.matchIf(L_PAREN)
	id := p.Text
	p.match(IDENTIFIER)
	p.matchSkipNL(IN)
	if !parens {
		defer func(prev int) { p.nest = prev }(p.nest)
		p.nest = 0
	}
	expr := p.exprAst()
	if parens {
		p.match(R_PAREN)
	} else {
		p.matchIf(NEWLINE)
	}
	body := p.statement()
	return ast(it, ast2(id), expr, body)
}

func (p *parser) forClassic(it Item) Ast {
	p.match(L_PAREN)
	init := p.optExprList(SEMICOLON)
	p.match(SEMICOLON)
	cond := p.exprAst()
	p.match(SEMICOLON)
	incr := p.optExprList(R_PAREN)
	p.match(R_PAREN)
	body := p.statement()
	return ast(it, init, cond, incr, body)
}

func (p *parser) optExprList(after Token) Ast {
	ast := ast2("exprs")
	if p.Token != after {
		for {
			ast.Children = append(ast.Children, p.exprAst())
			if p.Token != COMMA {
				break
			}
			p.next()
		}
	}
	return ast
}

// used by if and while
func (p *parser) ctrlExpr() (Item, Ast) {
	it := p.Item
	p.nextSkipNL()
	expr := p.exprAst()
	if p.Token == NEWLINE {
		p.nextSkipNL()
	}
	return it, expr
}

func (p *parser) returnStmt() Ast {
	item := p.Item
	p.matchKeepNL(RETURN)
	if p.matchIf(NEWLINE) || p.matchIf(SEMICOLON) || p.Token == R_CURLY {
		return ast(item)
	}
	return ast(item, p.exprStmt())
}

func (p *parser) exprStmt() Ast {
	result := p.exprAst()
	for p.Token == SEMICOLON || p.Token == NEWLINE {
		p.next()
	}
	return result
}

func (p *parser) throwStmt() Ast {
	item := p.Item
	p.matchSkipNL(THROW)
	return ast(item, p.exprStmt())
}

func (p *parser) exprAst() Ast {
	return expression(p, astBuilder).(Ast)
}
