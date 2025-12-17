package parser

import (
	"fmt"
	"io"
	"strconv"

	"github.com/donmahallem/go_gcode/lexer"
	"github.com/donmahallem/go_gcode/nodes"
	"github.com/donmahallem/go_gcode/token"
)

const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LBRACKET: INDEX,
	token.AND:      AND,
	token.OR:       OR,
}

type (
	prefixParseFn func() nodes.Expression
	infixParseFn  func(nodes.Expression) nodes.Expression
)

type Parser struct {
	l         lexer.TokenSource
	curToken  token.Token
	peekToken token.Token
	errors    []ParseError

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l lexer.TokenSource) *Parser {
	p := &Parser{
		l:      l,
		errors: []ParseError{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.LookupIdent("ident"), p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *nodes.GroupNode {
	program := &nodes.GroupNode{Nodes: []nodes.Node{}}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Nodes = append(program.Nodes, stmt)
		}
	}

	return program
}

func (p *Parser) synchronize() {
	p.l.Reset()
	p.nextToken()
	for p.curToken.Type != token.EOF {
		switch p.curToken.Type {
		case token.TEXT, token.COMMENT, token.LBRACE, token.LBRACKET:
			return
		}
		p.nextToken()
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

func (p *Parser) parseExpression(precedence int) nodes.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.RBRACKET) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, ParseError{
		Message: msg,
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
	})
}

func (p *Parser) parseIdentifier() nodes.Expression {
	return &nodes.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() nodes.Expression {
	lit := &nodes.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, ParseError{Message: msg, Line: p.curToken.Line, Column: p.curToken.Column})
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() nodes.Expression {
	lit := &nodes.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, ParseError{Message: msg, Line: p.curToken.Line, Column: p.curToken.Column})
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() nodes.Expression {
	return &nodes.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() nodes.Expression {
	return &nodes.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) parsePrefixExpression() nodes.Expression {
	expression := &nodes.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left nodes.Expression) nodes.Expression {
	expression := &nodes.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() nodes.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIndexExpression(left nodes.Expression) nodes.Expression {
	exp := &nodes.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) ParseNext() (nodes.Node, error) {
	if p.curToken.Type == token.EOF {
		return nil, io.EOF
	}
	stmt := p.parseStatement()
	return stmt, nil
}

func (p *Parser) ParseStream(w io.Writer, env map[string]any) error {
	for {
		node, err := p.ParseNext()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if node == nil {
			continue
		}
		if e, ok := node.(nodes.Emitter); ok {
			if err := e.Emit(w, env); err != nil {
				return err
			}
		} else {
			val, err := node.Evaluate(env)
			if err != nil {
				return err
			}
			if _, err := io.WriteString(w, fmt.Sprintf("%v", val)); err != nil {
				return err
			}
		}
	}
}

func (p *Parser) parseStatement() nodes.Node {
	switch p.curToken.Type {
	case token.TEXT:
		// Fallback for raw text if any remains
		node := &nodes.TextNode{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		return node
	case token.IDENT:
		return p.parseInstruction()
	case token.NEWLINE:
		p.nextToken()
		return nil
	case token.COMMENT:
		node := &nodes.CommentNode{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		return node
	case token.LBRACE:
		if p.peekToken.Type == token.IF {
			return p.parseConditional()
		}
		return p.parseInterpolation(token.RBRACE)
	case token.LBRACKET:
		return p.parseInterpolation(token.RBRACKET)
	default:
		p.errors = append(p.errors, ParseError{
			Message: fmt.Sprintf("unexpected token %s", p.curToken.Type),
			Line:    p.curToken.Line,
			Column:  p.curToken.Column,
		})
		p.nextToken() // Advance to avoid infinite loop
		return nil
	}
}

func (p *Parser) parseInstruction() nodes.Node {
	instr := &nodes.InstructionNode{
		Token:      p.curToken,
		Command:    p.curToken.Literal,
		Parameters: []*nodes.ParameterNode{},
	}

	p.nextToken() // consume Command

	for !p.curTokenIs(token.NEWLINE) && !p.curTokenIs(token.EOF) && !p.curTokenIs(token.COMMENT) {
		param := p.parseParameter()
		if param != nil {
			instr.Parameters = append(instr.Parameters, param)
		} else {
			// If we couldn't parse a parameter but we aren't at EOL, we might be stuck.
			// Consume token to avoid infinite loop
			p.nextToken()
		}
	}

	// If we stopped at NEWLINE, consume it
	if p.curTokenIs(token.NEWLINE) {
		p.nextToken()
	}
	// If we stopped at COMMENT, leave it for next parseStatement call

	return instr
}

func (p *Parser) parseParameter() *nodes.ParameterNode {
	// Parameter can be:
	// 1. Key Value (X10, X-10, X10.5, X{val})
	// 2. Standalone Flag (judge_flag)

	if !p.curTokenIs(token.IDENT) {
		// Unexpected token in parameter list
		return nil
	}

	param := &nodes.ParameterNode{
		Key: p.curToken.Literal,
	}
	p.nextToken() // consume Key

	// Check if next token is a value start
	if p.curTokenIs(token.LBRACE) {
		// Interpolation: X{val}
		// We need to parse the interpolation
		// parseInterpolation expects current token to be LBRACE
		// It returns a Node (InterpolationNode)
		// But parseInterpolation consumes tokens.
		// Let's use p.parseInterpolation(token.RBRACE)
		// But parseInterpolation returns nodes.Node, we need to cast or store as Node
		param.Value = p.parseInterpolation(token.RBRACE)
		// parseInterpolation consumes the closing brace and calls nextToken.
		// So we are good.
	}

	return param
}

func (p *Parser) parseConditional() nodes.Node {
	// Current token is LBRACE, peek is IF
	// We want to capture the IF token for the Condition, and maybe LBRACE for ConditionalNode?
	// Or just use IF for both.

	p.nextToken() // move to IF
	ifToken := p.curToken

	condNode := &nodes.ConditionalNode{
		Token:      ifToken,
		Conditions: []nodes.Condition{},
	}

	// Parse first if
	condition := p.parseConditionExpression()
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	// Parse block
	p.nextToken() // move past RBRACE
	body := p.parseBlock()

	condNode.Conditions = append(condNode.Conditions, nodes.Condition{
		Token:      ifToken,
		Expression: condition,
		Body:       body,
	})

	// Handle elsif
	for p.curTokenIs(token.LBRACE) && p.peekTokenIs(token.ELSIF) {
		p.nextToken() // move to ELSIF
		elsifToken := p.curToken

		condition := p.parseConditionExpression()
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		p.nextToken() // move past RBRACE
		body := p.parseBlock()
		condNode.Conditions = append(condNode.Conditions, nodes.Condition{
			Token: elsifToken, Expression: condition,
			Body: body})
	}

	// Handle else
	if p.curTokenIs(token.LBRACE) && p.peekTokenIs(token.ELSE) {
		p.nextToken() // move to ELSE

		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		p.nextToken() // move past RBRACE
		elseBody := p.parseBlock()
		condNode.Else = elseBody
	}

	// Expect ENDIF
	if p.curTokenIs(token.LBRACE) && p.peekTokenIs(token.ENDIF) {
		p.nextToken() // move to ENDIF
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		// expectPeek consumes RBRACE. curToken is RBRACE.
		// We need to advance past it.
		p.nextToken()
	} else {
		p.errors = append(p.errors, ParseError{
			Message: fmt.Sprintf("expected {endif}, got %s %s", p.curToken.Literal, p.peekToken.Literal),
			Line:    p.curToken.Line,
			Column:  p.curToken.Column,
		})
	}

	return condNode
}

func (p *Parser) parseBlock() *nodes.GroupNode {
	block := &nodes.GroupNode{Nodes: []nodes.Node{}}

	for p.curToken.Type != token.EOF {
		// Check for end of block keywords
		if p.curToken.Type == token.LBRACE {
			if p.peekToken.Type == token.ELSIF ||
				p.peekToken.Type == token.ELSE ||
				p.peekToken.Type == token.ENDIF {
				return block
			}
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Nodes = append(block.Nodes, stmt)
		}
	}
	return block
}

func (p *Parser) parseConditionExpression() nodes.Expression {
	p.nextToken() // Move from Keyword to first token of expression
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseInterpolation(endToken token.TokenType) nodes.Node {
	// We are at LBRACE or LBRACKET
	startToken := p.curToken
	p.nextToken() // Move to content

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(endToken) {
		return nil
	}
	// expectPeek consumes endToken. curToken is endToken.
	// We need to advance past it.
	p.nextToken()

	return &nodes.InterpolationNode{Token: startToken, Expression: exp}
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
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, ParseError{
		Message: msg,
		Line:    p.peekToken.Line,
		Column:  p.peekToken.Column,
	})
}
