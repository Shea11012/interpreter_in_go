package parser

import (
	"fmt"
	"github.com/Shea11012/interpreter_in_go/ast"
	"github.com/Shea11012/interpreter_in_go/lexer"
	"github.com/Shea11012/interpreter_in_go/token"
	"strconv"
)

// 顺序表示执行优先级，由大到小
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x or !x
	CALL        // myFunction(x)
	INDEX       // array[index]
)

var precedences = map[token.Type]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(expression ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer // 解析器

	errors    []string    // 存放解析错误的所有信息
	curToken  token.Token // 当前token
	peekToken token.Token // 下一个token

	prefixParseFns map[token.Type]prefixParseFn // 前缀token解析
	infixParseFns  map[token.Type]infixParseFn  // 中缀token解析
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[token.Type]prefixParseFn),
		infixParseFns:  make(map[token.Type]infixParseFn),
	}

	// 前缀表达式解析
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBool)
	p.registerPrefix(token.FALSE, p.parseBool)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// 中缀表达式解析
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// 读取两次，初始化curToken和peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken 获取下一个token和下下一个token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram 解析token，生成statements
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement 根据token类型进行解析
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement 解析let类型的token，生成letStatement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// 跳过 =
	p.nextToken()

	// 解析 = 后面的表达式
	stmt.Value = p.parseExpression(LOWEST)

	if fl,ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fl.Name = stmt.Name.Value
	}

	// 跳过 ;
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// curTokenIs 与当前token类型比较
func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

// curTokenIs 与下一个token类型比较
func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// expectPeek 当前token类型与给定类型是否相等，相等则读取下一个token，不等则写入errors中
func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

// peekError 格式化下一个类型的错误信息
func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s,got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// parseReturnStatement 将类型为return的token，解析为ReturnStatement
func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{
		Token: p.curToken,
	}

	// 跳过 return
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	// 下一个是 ; , 则获取 ; 的下一个token
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// registerPrefix 注册前缀表达式
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix 注册中缀表达式
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// parseExpressionStatement 解析条件表达式
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 解析表达式，传入一个最低优先级
	stmt.Expression = p.parseExpression(LOWEST)

	// 下一个token等于分号则继续读取
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression 解析表达式
// 中缀表达式会进行递归调用，直至解析到分号或者(传入符号优先级变低且下一个token的优先级等于或小于传入优先级)
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 根据当前token类型，获取注册好的前缀表达式函数地址
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// 执行前缀解析
	leftExp := prefix()

	// 下一个token不是分号且优先级不能大于下一个token优先级，如 a - b，变量优先级会小于符号优先级
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// 根据下一个token的类型，找到对应的中缀表达式函数
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// 读取下一个token
		p.nextToken()
		// 将解析好的前缀表达式(中缀表达式的左半部分)当做变量传入，继续进行解析且覆盖leftExp变量
		leftExp = infix(leftExp)
	}
	// 此时生成的是一个由多个expression组成的节点
	return leftExp
}

// parseIdentifier 解析变量，属于前缀表达式
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseIntegerLiteral 解析整数，属于前缀表达式
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{
		Token: p.curToken,
	}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// noPrefixParseFnError 格式化遇到未注册的token类型前缀表达式函数错误
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// parsePrefixExpression 解析前缀表达式
func (p *Parser) parsePrefixExpression() ast.Expression {
	// 前缀表达式，当前的token一定是一个操作符，所以先将操作符初始化进前缀表达式
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	// 读取操作符后面的变量
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// peekPrecedence 获取下一个token.Type的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence 获取当前token.Type优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// parseInfixExpression 中缀表达式解析
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// 将左半部分的表达式初始化
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// 获取当前token类型的优先级
	precedence := p.curPrecedence()
	// 读取下一个token
	p.nextToken()
	// 根据上一个token的优先级，继续进行表达式解析
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseBool 解析布尔值
func (p *Parser) parseBool() ast.Expression {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.curTokenIs(token.TRUE),
	}
}

// parseGroupedExpression 解析组表达式
func (p *Parser) parseGroupedExpression() ast.Expression {
	// 读取（ token的下一个token
	p.nextToken()
	// 解析出括号内的表达式
	exp := p.parseExpression(LOWEST)

	// 如果是 ) 读取下一个token
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression 解析if语句表达式
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{
		Token: p.curToken,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	// 解析if括号内的表达式
	expression.Condition = p.parseExpression(LOWEST)

	// 如果是 ) 读取下一个token
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 如果是 { 读取下一个token
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	// 解析完if的第一个代码块，下一个是否是 else token
	if p.peekTokenIs(token.ELSE) {
		// 跳过 else token
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// 解析 else 代码块
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseBlockStatement 解析代码块
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}

	// 跳过 { token
	p.nextToken()

	// 当前token是 } 或 eof 则退出循环
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseFunctionLiteral 解析函数
func (p *Parser) parseFunctionLiteral() ast.Expression {
	fnLiteral := &ast.FunctionLiteral{
		Token: p.curToken,
	}

	// 是 ( , 则跳过 (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fnLiteral.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fnLiteral.Body = p.parseBlockStatement()

	return fnLiteral
}

// parseFunctionParameters 解析函数参数
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	params := make([]*ast.Identifier, 0)

	// 如果是下一个),则跳过 )
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	// 跳过 (，获取第一个参数
	p.nextToken()

	ident := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	params = append(params, ident)

	// 下一个是,则表示还有参数
	for p.peekTokenIs(token.COMMA) {
		// 将第一个参数略过
		p.nextToken()
		// 将逗号略过
		p.nextToken()
		ident := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		params = append(params, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// parseCallExpression 解析函数调用语句
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}
	exp.Arguments = p.parseExpressionList(token.RPAREN)

	return exp
}

// parseCallArguments 解析函数调用参数
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// 下一个 )，表示无参数且跳过 )
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return nil
	}
	// 跳过 (
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	args = append(args, exp)

	for p.peekTokenIs(token.COMMA) {
		// 跳过参数
		p.nextToken()
		// 跳过逗号
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// parseStringLiteral 解析字符串
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseArrayLiteral 解析数组
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	// 跳过 [
	p.nextToken()
	// 解析第一个值
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		// 跳过当前值
		p.nextToken()
		// 跳过 ,
		p.nextToken()

		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseIndexExpression 解析数组索引表达式
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	// 跳过 [
	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseHashLiteral 解析哈希表达式
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{
		Token: p.curToken,
		Pairs: make(map[ast.Expression]ast.Expression),
	}

	// 下一个token不是 }
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}

		// 跳过 :
		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		// 判断hash边界，且跳过 ,
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}
