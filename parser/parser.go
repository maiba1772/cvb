package parser

import (
	"fmt"
	"strconv"
	"strings"

	"cvb-lang/ast"
	"cvb-lang/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}

	for !p.curTokenIs(lexer.TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TOKEN_IMPORT:
		return p.parseImportStatement()
	case lexer.TOKEN_IMPORTF:
		return p.parseImportFileStatement()
	case lexer.TOKEN_PRINT:
		return p.parsePrintStatement()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStatement()
	case lexer.TOKEN_FOR:
		return p.parseForStatement()
	case lexer.TOKEN_IF:
		return p.parseIfStatement()
	case lexer.TOKEN_DEF:
		return p.parseFunctionDefinition()
	case lexer.TOKEN_BREAK:
		return p.parseBreakStatement()
	case lexer.TOKEN_IDENTIFIER:
		if p.peekTokenIs(lexer.TOKEN_ARROW) {
			return p.parseVariableStatement()
		}
		return p.parseExpressionStatement()
	case lexer.TOKEN_NEWLINE:
		return nil
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Modules: []string{}}

	if !p.expectPeek(lexer.TOKEN_LT) {
		return nil
	}

	for {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
			stmt.Modules = append(stmt.Modules, p.curToken.Literal)
		} else if p.curTokenIs(lexer.TOKEN_STR) || p.curTokenIs(lexer.TOKEN_MATH) ||
			p.curTokenIs(lexer.TOKEN_FILE) || p.curTokenIs(lexer.TOKEN_RANDOM) ||
			p.curTokenIs(lexer.TOKEN_NET) || p.curTokenIs(lexer.TOKEN_SHELL) ||
			p.curTokenIs(lexer.TOKEN_LIST) || p.curTokenIs(lexer.TOKEN_INT) {
			stmt.Modules = append(stmt.Modules, p.curToken.Literal)
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier in import, got %s", p.curToken.Type))
			return nil
		}

		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		} else if p.peekTokenIs(lexer.TOKEN_GT) {
			p.nextToken()
			break
		} else {
			p.errors = append(p.errors, fmt.Sprintf("expected , or > in import, got %s", p.peekToken.Type))
			return nil
		}
	}

	return stmt
}

func (p *Parser) parseImportFileStatement() *ast.ImportFileStatement {
	stmt := &ast.ImportFileStatement{}

	p.nextToken()
	stmt.Path = p.parseExpression()

	if !p.expectPeek(lexer.TOKEN_PIPE) {
		return nil
	}

	p.nextToken()
	stmt.Name = p.parseExpression()

	if !p.expectPeek(lexer.TOKEN_DOLLAR) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_IDENTIFIER) {
		return nil
	}
	stmt.VarName = p.curToken.Literal

	return stmt
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{}

	if !p.expectPeek(lexer.TOKEN_ARROW) {
		return nil
	}

	p.nextToken()

	if p.curTokenIs(lexer.TOKEN_STR) || p.curTokenIs(lexer.TOKEN_MATH) ||
		p.curTokenIs(lexer.TOKEN_INT) || p.curTokenIs(lexer.TOKEN_LIST) ||
		p.curTokenIs(lexer.TOKEN_DIC) {
		stmt.TypeHint = p.curToken.Literal
		if !p.expectPeek(lexer.TOKEN_AMPERSAND) {
			return nil
		}
		p.nextToken()
	}

	stmt.Value = p.parseExpression()

	return stmt
}

func (p *Parser) parseVariableStatement() *ast.VariableStatement {
	stmt := &ast.VariableStatement{Name: p.curToken.Literal}

	if !p.expectPeek(lexer.TOKEN_ARROW) {
		return nil
	}

	p.nextToken()

	if p.curTokenIs(lexer.TOKEN_STR) || p.curTokenIs(lexer.TOKEN_INT) ||
		p.curTokenIs(lexer.TOKEN_LIST) || p.curTokenIs(lexer.TOKEN_DIC) {
		stmt.TypeHint = p.curToken.Literal
		if !p.expectPeek(lexer.TOKEN_AMPERSAND) {
			return nil
		}
		p.nextToken()
	}

	stmt.Value = p.parseExpression()

	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{}

	p.nextToken()

	if p.curTokenIs(lexer.TOKEN_TRUE) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_ASSIGN) {
			p.nextToken()
			stmt.Count = p.parseExpression()
		}
	} else {
		stmt.Count = p.parseExpression()
	}

	if p.peekTokenIs(lexer.TOKEN_AND) {
		p.nextToken()
		if !p.expectPeek(lexer.TOKEN_DOT) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_IF) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_COLON) {
			return nil
		}
		p.nextToken()
		stmt.Condition = p.parseExpression()
	}

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{}

	if !p.expectPeek(lexer.TOKEN_IDENTIFIER) {
		return nil
	}
	stmt.Iterator = p.curToken.Literal

	if !p.expectPeek(lexer.TOKEN_IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression()

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{}

	p.nextToken()
	stmt.Condition = p.parseExpression()

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.nextToken()
	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.TOKEN_ELSE) {
		p.nextToken()
		if !p.expectPeek(lexer.TOKEN_COLON) {
			return nil
		}
		p.nextToken()
		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Statements: []ast.Statement{}}

	if p.curTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		for !p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {
			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
		}
	} else {
		for !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) {
			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
		}
	}

	return block
}

func (p *Parser) parseFunctionDefinition() *ast.FunctionDefinition {
	stmt := &ast.FunctionDefinition{Parameters: []string{}}

	if !p.expectPeek(lexer.TOKEN_IDENTIFIER) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	if p.peekTokenIs(lexer.TOKEN_IDENTIFIER) {
		p.nextToken()
		stmt.Parameters = append(stmt.Parameters, p.curToken.Literal)
		for p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			if !p.expectPeek(lexer.TOKEN_IDENTIFIER) {
				return nil
			}
			stmt.Parameters = append(stmt.Parameters, p.curToken.Literal)
		}
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{}

	if p.peekTokenIs(lexer.TOKEN_ASSIGN) {
		p.nextToken()
		if !p.expectPeek(lexer.TOKEN_NUMBER) {
			return nil
		}
		level, _ := strconv.Atoi(p.curToken.Literal)
		stmt.Level = level
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Expression: p.parseExpression()}
	return stmt
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parsePipeExpression()
}

func (p *Parser) parsePipeExpression() ast.Expression {
	left := p.parseOrExpression()

	for p.peekTokenIs(lexer.TOKEN_PIPE) {
		p.nextToken()
		p.nextToken()
		right := p.parseOrExpression()
		left = &ast.PipeExpression{Left: left, Right: right}
	}

	return left
}

func (p *Parser) parseOrExpression() ast.Expression {
	left := p.parseAndExpression()

	for p.peekTokenIs(lexer.TOKEN_OR) || p.peekTokenIs(lexer.TOKEN_OR_OP) {
		p.nextToken()
		operator := "||"
		if p.curTokenIs(lexer.TOKEN_OR) {
			operator = "or"
		}
		p.nextToken()
		right := p.parseAndExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parseAndExpression() ast.Expression {
	left := p.parseEqualityExpression()

	for p.peekTokenIs(lexer.TOKEN_AND) || p.peekTokenIs(lexer.TOKEN_AND_OP) {
		p.nextToken()
		operator := "&&"
		if p.curTokenIs(lexer.TOKEN_AND) {
			operator = "and"
		}
		p.nextToken()
		right := p.parseEqualityExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parseEqualityExpression() ast.Expression {
	left := p.parseRelationalExpression()

	for p.peekTokenIs(lexer.TOKEN_EQ) || p.peekTokenIs(lexer.TOKEN_NE) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parseRelationalExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parseRelationalExpression() ast.Expression {
	left := p.parseAdditiveExpression()

	for p.peekTokenIs(lexer.TOKEN_LT) || p.peekTokenIs(lexer.TOKEN_GT) ||
		p.peekTokenIs(lexer.TOKEN_LE) || p.peekTokenIs(lexer.TOKEN_GE) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parseAdditiveExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parseAdditiveExpression() ast.Expression {
	left := p.parseMultiplicativeExpression()

	for p.peekTokenIs(lexer.TOKEN_PLUS) || p.peekTokenIs(lexer.TOKEN_MINUS) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parseMultiplicativeExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parseMultiplicativeExpression() ast.Expression {
	left := p.parsePrefixExpression()

	for p.peekTokenIs(lexer.TOKEN_MUL) || p.peekTokenIs(lexer.TOKEN_DIV) || p.peekTokenIs(lexer.TOKEN_MOD) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parsePrefixExpression()
		left = &ast.InfixExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	if p.curTokenIs(lexer.TOKEN_MINUS) || p.curTokenIs(lexer.TOKEN_PLUS) ||
		p.curTokenIs(lexer.TOKEN_NOT) {
		operator := p.curToken.Literal
		p.nextToken()
		return &ast.PrefixExpression{Operator: operator, Right: p.parsePrefixExpression()}
	}

	return p.parsePostfixExpression()
}

func (p *Parser) parsePostfixExpression() ast.Expression {
	left := p.parsePrimaryExpression()

	for {
		if p.peekTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken()
			left = p.parseCallExpression(left)
		} else if p.peekTokenIs(lexer.TOKEN_DOT) {
			p.nextToken()
			if !p.expectPeek(lexer.TOKEN_IDENTIFIER) {
				return nil
			}
			method := p.curToken.Literal
			if p.peekTokenIs(lexer.TOKEN_LPAREN) {
				p.nextToken()
				left = p.parseMethodCallExpression(left, method)
			} else {
				left = &ast.MethodCallExpression{Object: left, Method: method, Arguments: []ast.Expression{}}
			}
		} else if p.peekTokenIs(lexer.TOKEN_LBRACKET) {
			p.nextToken()
			p.nextToken()
			index := p.parseExpression()
			if !p.expectPeek(lexer.TOKEN_RBRACKET) {
				return nil
			}
			left = &ast.IndexExpression{Left: left, Index: index}
		} else {
			break
		}
	}

	return left
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: function, Arguments: []ast.Expression{}}

	if p.peekTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
		return exp
	}

	p.nextToken()
	exp.Arguments = append(exp.Arguments, p.parseExpression())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		exp.Arguments = append(exp.Arguments, p.parseExpression())
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseMethodCallExpression(object ast.Expression, method string) ast.Expression {
	exp := &ast.MethodCallExpression{Object: object, Method: method, Arguments: []ast.Expression{}}

	if p.peekTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
		return exp
	}

	p.nextToken()
	exp.Arguments = append(exp.Arguments, p.parseExpression())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		exp.Arguments = append(exp.Arguments, p.parseExpression())
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parsePrimaryExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.TOKEN_IDENTIFIER:
		return &ast.Identifier{Value: p.curToken.Literal}
	case lexer.TOKEN_STRING:
		return &ast.StringLiteral{Value: p.curToken.Literal}
	case lexer.TOKEN_NUMBER:
		value, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		return &ast.NumberLiteral{Value: value, Raw: p.curToken.Literal}
	case lexer.TOKEN_TRUE:
		return &ast.BooleanLiteral{Value: true}
	case lexer.TOKEN_FALSE:
		return &ast.BooleanLiteral{Value: false}
	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()
	case lexer.TOKEN_LBRACE:
		return p.parseDictLiteral()
	case lexer.TOKEN_LPAREN:
		p.nextToken()
		exp := p.parseExpression()
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
		return exp
	case lexer.TOKEN_STR, lexer.TOKEN_INT, lexer.TOKEN_LIST, lexer.TOKEN_DIC:
		typeHint := p.curToken.Literal
		if p.peekTokenIs(lexer.TOKEN_LBRACKET) {
			p.nextToken()
			return p.parseTypedListLiteral(typeHint)
		}
		return &ast.Identifier{Value: typeHint}
	default:
		return nil
	}
}

func (p *Parser) parseListLiteral() *ast.ListLiteral {
	list := &ast.ListLiteral{Elements: []ast.Expression{}}

	if p.peekTokenIs(lexer.TOKEN_RBRACKET) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list.Elements = append(list.Elements, p.parseExpression())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list.Elements = append(list.Elements, p.parseExpression())
	}

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return list
}

func (p *Parser) parseTypedListLiteral(typeHint string) *ast.ListLiteral {
	list := &ast.ListLiteral{TypeHint: typeHint, Elements: []ast.Expression{}}

	if p.peekTokenIs(lexer.TOKEN_RBRACKET) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list.Elements = append(list.Elements, p.parseExpression())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list.Elements = append(list.Elements, p.parseExpression())
	}

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return list
}

func (p *Parser) parseDictLiteral() *ast.DictLiteral {
	dict := &ast.DictLiteral{Pairs: make(map[ast.Expression]ast.Expression)}

	if p.peekTokenIs(lexer.TOKEN_RBRACE) {
		p.nextToken()
		return dict
	}

	p.nextToken()
	key := p.parseExpression()

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression()
	dict.Pairs[key] = value

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		key = p.parseExpression()
		if !p.expectPeek(lexer.TOKEN_COLON) {
			return nil
		}
		p.nextToken()
		value = p.parseExpression()
		dict.Pairs[key] = value
	}

	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}

	return dict
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func Parse(input string) (*ast.Program, []string) {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	return program, p.Errors()
}

func ParseFile(filename string) (*ast.Program, error) {
	content, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	program, errors := Parse(content)
	if len(errors) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(errors, "; "))
	}
	return program, nil
}

func readFile(filename string) (string, error) {
	data, err := readFileBytes(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readFileBytes(filename string) ([]byte, error) {
	// This will be implemented in the interpreter package
	return nil, fmt.Errorf("not implemented")
}

var (
	NotImplemented = fmt.Errorf("not implemented")
)
