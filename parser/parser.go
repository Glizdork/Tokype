package parser

import (
	"Tokype/ast"
	"Tokype/lexer"
	"Tokype/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	ASSIGN_PRIO
	SUM_PRIO
	PRODUCT_PRIO
	CALL_PRIO
	INDEX_PRIO
	PREFIX_PRIO
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:      ASSIGN_PRIO,
	token.PLUS:        SUM_PRIO,
	token.MINUS:       SUM_PRIO,
	token.ASTERISK:    PRODUCT_PRIO,
	token.SLASH:       PRODUCT_PRIO,
	token.MODULO:      PRODUCT_PRIO,
	token.LPAREN:      CALL_PRIO,
	token.DOTLBRACKET: INDEX_PRIO,
	token.EQ:          SUM_PRIO,
	token.NOT_EQ:      SUM_PRIO,
	token.LT:          SUM_PRIO,
	token.GT:          SUM_PRIO,
	token.LTE:         SUM_PRIO,
	token.GTE:         SUM_PRIO,
	token.DOT:         CALL_PRIO,
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d, column %d: expected next token to be %s, got %s instead",
		p.peekToken.Line,
		p.peekToken.Column,
		t,
		p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Node{}}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		if p.curToken.Type != token.EOF {
			p.nextToken()
		}
	}
	return program
}

func (p *Parser) parseStatement() ast.Node {
	switch p.curToken.Type {
	case token.FUNCT:
		return p.parseFunctionStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IDENT:
		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStatement()
		}
		if p.peekTokenIs(token.DOTLBRACKET) {
			return p.parseAssignIndexStatement()
		}
	case token.INCLUDE:
		return p.parseIncludeStatement()
	}
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseAssignStatement() ast.Node {
	stmt := &ast.AssignStatement{Name: &ast.Identifier{Value: p.curToken.Literal}}
	p.nextToken()
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseAssignIndexStatement() ast.Node {
	varName := p.curToken.Literal
	p.nextToken()
	p.nextToken()

	index := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)

	return &ast.AssignIndexStatement{
		Left: &ast.IndexExpression{
			Left:  &ast.Identifier{Value: varName},
			Index: index,
		},
		Value:  value,
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
	}
}

func (p *Parser) parseFunctionStatement() ast.Node {
	p.nextToken()

	stmt := &ast.FunctionStatement{Name: &ast.Identifier{Value: p.curToken.Literal}}

	p.nextToken()
	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.COLON) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if !p.curTokenIs(token.END) {
		p.peekError(token.END)
		return nil
	}

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	idents := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return idents
	}

	p.nextToken()
	idents = append(idents, &ast.Identifier{Value: p.curToken.Literal})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		idents = append(idents, &ast.Identifier{Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return idents
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Statements: []ast.Node{}}

	for !p.curTokenIs(token.END) && !p.curTokenIs(token.ELIF) &&
		!p.curTokenIs(token.ELSE) && !p.curTokenIs(token.EOF) {

		if p.curTokenIs(token.IF) {
			stmt := p.parseIfStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
			continue
		}

		if p.curTokenIs(token.FOR) {
			stmt := p.parseForStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
			continue
		}

		if p.curTokenIs(token.WHILE) {
			stmt := p.parseWhileStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		if !p.curTokenIs(token.END) && !p.curTokenIs(token.ELIF) &&
			!p.curTokenIs(token.ELSE) && !p.curTokenIs(token.EOF) {
			p.nextToken()
		}
	}

	return block
}

func (p *Parser) parseIfStatement() ast.Node {
	p.nextToken()

	condition := p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	consequence := p.parseBlockStatement()

	var alternatives []*ast.ElseIfExpression
	var elseBlock *ast.BlockStatement

	for p.curTokenIs(token.ELIF) || p.curTokenIs(token.ELSE) {
		if p.curTokenIs(token.ELIF) {
			p.nextToken()

			elifCondition := p.parseExpression(LOWEST)

			if !p.expectPeek(token.COLON) {
				return nil
			}

			p.nextToken()
			elifBody := p.parseBlockStatement()

			alternatives = append(alternatives, &ast.ElseIfExpression{
				Condition:   elifCondition,
				Consequence: elifBody,
			})
		} else if p.curTokenIs(token.ELSE) {

			if !p.peekTokenIs(token.COLON) {
				p.peekError(token.COLON)
				return nil
			}

			p.nextToken()

			p.nextToken()
			elseBlock = p.parseBlockStatement()
			break
		}
	}

	if !p.curTokenIs(token.END) {
		p.peekError(token.END)
		return nil
	}

	return &ast.IfExpression{
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternatives,
		Else:        elseBlock,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	var leftExp ast.Expression

	switch p.curToken.Type {
	case token.IDENT:
		leftExp = &ast.Identifier{
			Value:  p.curToken.Literal,
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
		}
	case token.INT:
		value, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
		leftExp = &ast.IntegerLiteral{Value: value}
	case token.FLOAT:
		value, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		leftExp = &ast.FloatLiteral{Value: value}
	case token.MINUS:
		line := p.curToken.Line
		column := p.curToken.Column
		p.nextToken()
		right := p.parseExpression(PREFIX_PRIO)
		leftExp = &ast.PrefixExpression{
			Operator: "-",
			Right:    right,
			Line:     line,
			Column:   column,
		}
	case token.STRING:
		leftExp = &ast.StringLiteral{Value: p.curToken.Literal}
	case token.TRUE:
		leftExp = &ast.BooleanLiteral{Value: true}
	case token.FALSE:
		leftExp = &ast.BooleanLiteral{Value: false}
	case token.NULL:
		leftExp = &ast.NullLiteral{}
	case token.OPENLIST:
		leftExp = p.parseListLiteral()
	case token.OPENMAP:
		leftExp = p.parseMapLiteral()
	case token.LPAREN:
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		leftExp = exp
	default:
		return nil
	}

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case token.LPAREN:
			p.nextToken()
			leftExp = p.parseCallExpression(leftExp)
		case token.DOTLBRACKET:
			p.nextToken()
			leftExp = p.parseIndexExpression(leftExp)
		case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.MODULO,
			token.EQ, token.NOT_EQ, token.LT, token.GT, token.LTE, token.GTE:
			p.nextToken()
			leftExp = p.parseInfixExpression(leftExp)
		case token.DOT:
			p.nextToken()
			p.nextToken()
			right := p.parseExpression(PREFIX_PRIO)
			if ident, ok := right.(*ast.Identifier); ok {
				leftExp = &ast.DotExpression{
					Left:  leftExp,
					Right: ident,
				}
			} else {
				p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: expected identifier after '.'",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
		default:
			return leftExp
		}
	}
	return leftExp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
		Line:     p.curToken.Line,
		Column:   p.curToken.Column,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return p.foldConstants(expression)
}

func (p *Parser) foldConstants(expr ast.Expression) ast.Expression {
	infix, ok := expr.(*ast.InfixExpression)
	if !ok {
		return expr
	}

	leftInt, leftIsInt := infix.Left.(*ast.IntegerLiteral)
	rightInt, rightIsInt := infix.Right.(*ast.IntegerLiteral)

	if leftIsInt && rightIsInt {
		var result int64
		switch infix.Operator {
		case "+":
			result = leftInt.Value + rightInt.Value
		case "-":
			result = leftInt.Value - rightInt.Value
		case "*":
			result = leftInt.Value * rightInt.Value
		case "/":
			if rightInt.Value == 0 {
				return expr
			}
			if leftInt.Value%rightInt.Value == 0 {
				return &ast.IntegerLiteral{Value: leftInt.Value / rightInt.Value}
			}
			return &ast.FloatLiteral{Value: float64(leftInt.Value) / float64(rightInt.Value)}
		default:
			return expr
		}
		return &ast.IntegerLiteral{Value: result}
	}

	leftFloat, leftIsFloat := infix.Left.(*ast.FloatLiteral)
	rightFloat, rightIsFloat := infix.Right.(*ast.FloatLiteral)

	if leftIsFloat && rightIsFloat {
		var result float64
		switch infix.Operator {
		case "+":
			result = leftFloat.Value + rightFloat.Value
		case "-":
			result = leftFloat.Value - rightFloat.Value
		case "*":
			result = leftFloat.Value * rightFloat.Value
		case "/":
			if rightFloat.Value != 0 {
				result = leftFloat.Value / rightFloat.Value
			} else {
				return expr
			}
		default:
			return expr
		}
		return &ast.FloatLiteral{Value: result}
	}

	return expr
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Function:  function,
		Arguments: []ast.Expression{},
		Line:      p.curToken.Line,
		Column:    p.curToken.Column,
	}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return exp
	}

	p.nextToken()
	exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	p.nextToken()

	start := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.DOTDOT) {
		p.nextToken()
		var end ast.Expression
		if !p.peekTokenIs(token.RBRACKET) {
			p.nextToken()
			end = p.parseExpression(LOWEST)
		}
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return &ast.SliceExpression{
			Left:   left,
			Start:  start,
			End:    end,
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
		}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return &ast.IndexExpression{
		Left:  left,
		Index: start,
	}
}

func (p *Parser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{Elements: []ast.Expression{}}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list.Elements = append(list.Elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list.Elements = append(list.Elements, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return list
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLiteral{Pairs: make(map[ast.Expression]ast.Expression)}

	if p.peekTokenIs(token.CLOSEMAP) {
		p.nextToken()
		return mapLit
	}

	p.nextToken()
	key := p.parseExpression(LOWEST)

	if !p.expectPeek(token.MAPEQ) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	mapLit.Pairs[key] = value

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.MAPEQ) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		mapLit.Pairs[key] = value
	}

	if !p.expectPeek(token.CLOSEMAP) {
		return nil
	}

	return mapLit
}

func (p *Parser) parseForStatement() ast.Node {
	p.nextToken()
	varName := p.curToken.Literal

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	startVal := p.parseExpression(LOWEST)

	if !p.expectPeek(token.COMMA) {
		return nil
	}

	p.nextToken()
	endVal := p.parseExpression(LOWEST)

	var stepVal ast.Expression
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		stepVal = p.parseExpression(LOWEST)
	}

	init := &ast.AssignStatement{
		Name:  &ast.Identifier{Value: varName},
		Value: startVal,
	}

	condition := &ast.InfixExpression{
		Left:     &ast.Identifier{Value: varName},
		Operator: "<=",
		Right:    endVal,
	}

	step := ast.Expression(&ast.IntegerLiteral{Value: 1})
	if stepVal != nil {
		step = stepVal
	}

	update := &ast.AssignStatement{
		Name: &ast.Identifier{Value: varName},
		Value: &ast.InfixExpression{
			Left:     &ast.Identifier{Value: varName},
			Operator: "+",
			Right:    step,
		},
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	body := p.parseBlockStatement()

	if !p.curTokenIs(token.END) {
		p.peekError(token.END)
		return nil
	}

	return &ast.ForStatement{
		Initialization: init,
		Condition:      condition,
		Update:         update,
		Body:           body,
		Step:           stepVal,
	}
}

func (p *Parser) parseWhileStatement() ast.Node {
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if !p.expectPeek(token.COLON) {
		return nil
	}
	p.nextToken()
	body := p.parseBlockStatement()

	if !p.curTokenIs(token.END) {
		p.peekError(token.END)
		return nil
	}
	return &ast.WhileStatement{
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) parseReturnStatement() ast.Node {
	stmt := &ast.ReturnStatement{}

	if p.peekTokenIs(token.EOF) || p.peekTokenIs(token.END) {
		return stmt
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseIncludeStatement() ast.Node {
	p.nextToken()

	if !p.curTokenIs(token.STRING) {
		p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: expected string after 'include'",
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	path := p.curToken.Literal
	namespace := ""

	if p.peekTokenIs(token.AS) {
		p.nextToken()
		p.nextToken()

		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("line %d, column %d: expected identifier after 'as'",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		namespace = p.curToken.Literal
	}

	return &ast.IncludeStatement{
		Path:      path,
		Namespace: namespace,
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
