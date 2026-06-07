package parser

import (
	"fmt"
	"strings"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
)

// ---------- 优先级 ----------
const (
	precLowest = iota
	precOr
	precAnd
	precEq
	precRel
	precAdd
	precMul
	precPrefix
	precCall
)

// ---------- Parser ----------
type Parser struct {
	tokens []tokenizer.Token
	pos    int
	length int
	errors []string
}

// NewParser 创建解析器
func NewParser(tokens []tokenizer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		length: len(tokens),
	}
}

// Parse 包级入口函数
func Parse(tokens []tokenizer.Token) *ast.ProgramNode {
	return NewParser(tokens).ParseProgram()
}

// ---------- 底层辅助 ----------
func (p *Parser) next() {
	p.pos++
	if p.pos >= p.length {
		p.pos = p.length - 1
	}
}

func (p *Parser) peek(offset int) tokenizer.Token {
	idx := p.pos + offset
	if idx >= 0 && idx < p.length {
		return p.tokens[idx]
	}
	return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
}

func (p *Parser) cur() tokenizer.Token {
	if p.pos < p.length {
		return p.tokens[p.pos]
	}
	return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
}

func (p *Parser) expect(kind tokenizer.TokenKind) bool {
	return p.cur().Kind == kind
}

func (p *Parser) consume(kind tokenizer.TokenKind) bool {
	if p.cur().Kind == kind {
		p.next()
		return true
	}
	return false
}

func (p *Parser) skipTerminators() {
	for p.pos < p.length {
		k := p.cur().Kind
		if k == tokenizer.Terminator || k == tokenizer.Semicolon {
			p.next()
		} else {
			break
		}
	}
}

func (p *Parser) isAtEnd() bool {
	return p.cur().Kind == tokenizer.EOF
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.errors = append(p.errors, fmt.Sprintf(format, args...))
}

func (p *Parser) precedence() int {
	switch p.cur().Kind {
	case tokenizer.Or:
		return precOr
	case tokenizer.And:
		return precAnd
	case tokenizer.EqualEqual, tokenizer.NotEqual:
		return precEq
	case tokenizer.LessThan, tokenizer.GreaterThan, tokenizer.LessEqual, tokenizer.GreaterEqual:
		return precRel
	case tokenizer.Plus, tokenizer.Negative:
		return precAdd
	case tokenizer.Star, tokenizer.Slash:
		return precMul
	default:
		return precLowest
	}
}

// ---------- 表达式解析（Pratt Parser） ----------
type prefixParselet func(*Parser) ast.Expr
type infixParselet func(*Parser, ast.Expr) ast.Expr

var prefixParselets map[tokenizer.TokenKind]prefixParselet
var infixParselets map[tokenizer.TokenKind]infixParselet

func init() {
	prefixParselets = map[tokenizer.TokenKind]prefixParselet{
		tokenizer.Integer:      parseLiteralPrefix,
		tokenizer.Float:        parseLiteralPrefix,
		tokenizer.String:       parseLiteralPrefix,
		tokenizer.Identifier:   parseIdentifierPrefix,
		tokenizer.KeywordTrue:  parseIdentifierPrefix,
		tokenizer.KeywordFalse: parseIdentifierPrefix,
		tokenizer.Negative:     parsePrefixNegative,
		tokenizer.Dollar:       parseVariablePrefix,
		tokenizer.At:           parseReferencePrefix,
		tokenizer.LeftParen:    parseGroupedPrefix,
	}

	infixParselets = map[tokenizer.TokenKind]infixParselet{
		tokenizer.Plus:         parseInfixLeft(precAdd),
		tokenizer.Negative:     parseInfixLeft(precAdd),
		tokenizer.Star:         parseInfixLeft(precMul),
		tokenizer.Slash:        parseInfixLeft(precMul),
		tokenizer.EqualEqual:   parseInfixLeft(precEq),
		tokenizer.NotEqual:     parseInfixLeft(precEq),
		tokenizer.LessThan:     parseInfixLeft(precRel),
		tokenizer.GreaterThan:  parseInfixLeft(precRel),
		tokenizer.LessEqual:    parseInfixLeft(precRel),
		tokenizer.GreaterEqual: parseInfixLeft(precRel),
		tokenizer.And:          parseInfixLeft(precAnd),
		tokenizer.Or:           parseInfixLeft(precOr),
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	prefix := prefixParselets[p.cur().Kind]
	if prefix == nil {
		// 未注册前缀解析函数的关键字，回退为标识符
		if p.cur().Kind.IsKeyword() {
			prefix = parseIdentifierPrefix
		} else {
			p.errorf("unexpected token in expression: %v (%s)", p.cur().Kind, p.cur().Value)
			return nil
		}
	}
	left := prefix(p)

	for precedence < p.precedence() {
		infix := infixParselets[p.cur().Kind]
		if infix == nil {
			break
		}
		left = infix(p, left)
	}
	return left
}

func parseLiteralPrefix(p *Parser) ast.Expr {
	return p.parseLiteral()
}

func parseIdentifierPrefix(p *Parser) ast.Expr {
	lit := &ast.IdentifierExprNode{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.IdentifierExpr}},
		Name:     p.cur().Value,
	}
	p.next()
	return lit
}

func parsePrefixNegative(p *Parser) ast.Expr {
	p.next() // consume -
	expr := p.parseExpression(precPrefix)
	if lit, ok := expr.(*ast.Literal); ok {
		lit.Value = "-" + lit.Value
		return lit
	}
	return &ast.UnaryExprNode{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.UnaryExpr}},
		Op:       "-",
		Expr:     expr,
	}
}

func parseVariablePrefix(p *Parser) ast.Expr {
	p.next() // consume $
	if p.cur().Kind == tokenizer.LeftBraces {
		// ${name} 宏展开
		p.next() // consume {
		name := ""
		if p.cur().Kind == tokenizer.Identifier || p.cur().Kind.IsKeyword() {
			name = p.cur().Value
			p.next()
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
		return &ast.MacroExpandExprNode{
			BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.MacroExpandExpr}},
			Name:     name,
		}
	}
	name := ""
	if p.cur().Kind == tokenizer.Identifier || p.cur().Kind.IsKeyword() {
		name = p.cur().Value
		p.next()
	}
	return &ast.VariableExprNode{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.VariableExpr}},
		Name:     name,
	}
}

func parseReferencePrefix(p *Parser) ast.Expr {
	p.next() // consume @
	name := ""
	if p.cur().Kind == tokenizer.Identifier {
		name = p.cur().Value
		p.next()
	}
	return &ast.ReferenceExprNode{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.ReferenceExpr}},
		Name:     name,
	}
}

func parseGroupedPrefix(p *Parser) ast.Expr {
	p.next() // consume (
	expr := p.parseExpression(precLowest)
	if p.cur().Kind == tokenizer.RightParen {
		p.next()
	}
	return expr
}

func parseInfixLeft(precedence int) infixParselet {
	return func(p *Parser, left ast.Expr) ast.Expr {
		op := p.cur().Value
		p.next()
		right := p.parseExpression(precedence)
		return &ast.BinaryExprNode{
			BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.BinaryExpr}},
			Op:       op,
			Left:     left,
			Right:    right,
		}
	}
}

// ---------- 字面量 / 范围 ----------
func (p *Parser) parseLiteral() *ast.Literal {
	neg := false
	if p.cur().Kind == tokenizer.Negative {
		neg = true
		p.next()
	}

	tok := p.cur()
	if !tokenizer.TokenIsNumber(tok) && tok.Kind != tokenizer.String && tok.Kind != tokenizer.Time && tok.Kind != tokenizer.Date {
		if neg {
			p.errorf("expected number or string after '-'")
		}
		return nil
	}

	lit := &ast.Literal{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.LiteralExpr}},
		Value:    tok.Value,
	}

	// 根据 token 类型设置 LiteralType，防止后续转换丢失类型信息
	switch tok.Kind {
	case tokenizer.Integer:
		lit.LiteralType = ast.Integer
	case tokenizer.Float:
		lit.LiteralType = ast.Float
	case tokenizer.String:
		lit.LiteralType = ast.Text
	case tokenizer.Time:
		lit.LiteralType = ast.Time
	case tokenizer.Date:
		lit.LiteralType = ast.Date
	}

	if neg {
		lit.Value = "-" + lit.Value
	}
	p.next()

	// 百分比
	if p.cur().Kind == tokenizer.Per {
		lit.Unit = "%"
		p.next()
	} else if p.cur().Kind == tokenizer.Identifier {
		if tokenizer.IsTimeUnit(p.cur().Value) {
			lit.Unit = p.cur().Value
			p.next()
		} else if tokenizer.IsAmountUnit(p.cur().Value) {
			lit.Unit = p.cur().Value
			p.next()
		}
	}

	return lit
}

func (p *Parser) parseLiteralOrIdentifier() ast.Expr {
	switch p.cur().Kind {
	case tokenizer.Integer, tokenizer.Float, tokenizer.String, tokenizer.Negative, tokenizer.Time, tokenizer.Date:
		return p.parseLiteral()
	case tokenizer.Identifier, tokenizer.KeywordTrue, tokenizer.KeywordFalse:
		lit := &ast.IdentifierExprNode{
			BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.IdentifierExpr}},
			Name:     p.cur().Value,
		}
		p.next()
		return lit
	default:
		p.next()
		return nil
	}
}

func (p *Parser) parseRangeExpr() *ast.RangeExprNode {
	saved := p.pos
	r := &ast.RangeExprNode{
		BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.RangeExpr}},
	}

	// open-ended begin: ...end
	if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		p.next()
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			r.End = p.parseLiteral()
			return r
		}
		p.pos = saved
		return nil
	}

	// 必须有 begin
	if !tokenizer.TokenIsNumber(p.cur()) && p.cur().Kind != tokenizer.Negative {
		return nil
	}

	r.Begin = p.parseLiteral()
	if r.Begin == nil {
		p.pos = saved
		return nil
	}

	if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		p.next()
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			r.End = p.parseLiteral()
		}
		return r
	}

	p.pos = saved
	return nil
}

// ---------- 程序 / Block ----------

// ParseProgram 主解析入口
func (p *Parser) ParseProgram() *ast.ProgramNode {
	prog := &ast.ProgramNode{
		Node:       ast.Node{Type: ast.Program},
		Statements: []ast.Stmt{},
	}
	for !p.isAtEnd() {
		p.skipTerminators()
		if p.isAtEnd() {
			break
		}
		stmt := p.parseStmt()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}
	return prog
}

func (p *Parser) parseBlock(stopAt ...tokenizer.TokenKind) []ast.Stmt {
	stmts := []ast.Stmt{}
	for !p.isAtEnd() {
		p.skipTerminators()
		if p.isAtEnd() {
			break
		}
		// 检查停止条件
		for _, kind := range stopAt {
			if p.cur().Kind == kind {
				return stmts
			}
		}
		stmt := p.parseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}

// ---------- 语句解析（表驱动） ----------
var stmtParsers map[tokenizer.TokenKind]func(*Parser) ast.Stmt

func init() {
	stmtParsers = map[tokenizer.TokenKind]func(*Parser) ast.Stmt{
		tokenizer.KeywordGrid:       parseStrategyStmt,
		tokenizer.KeywordLong:       parseStrategyStmt,
		tokenizer.KeywordShort:      parseStrategyStmt,
		tokenizer.KeywordBoth:       parseStrategyStmt,
		tokenizer.KeywordPortfolio:  parseStrategyStmt,
		tokenizer.KeywordArbitrage:  parseStrategyStmt,
		tokenizer.KeywordImport:     parseImportStmt,
		tokenizer.KeywordExport:     parseExportStmt,
		tokenizer.KeywordUse:        parseUseStmt,
		tokenizer.KeywordTemplate:   parseTemplateStmt,
		tokenizer.KeywordMacro:      parseMacroStmt,
		tokenizer.KeywordIf:         parseIfStmt,
		tokenizer.KeywordAssert:     parseAssertStmt,
		tokenizer.KeywordTrigger:    parseFlowStmt,
		tokenizer.KeywordUntil:      parseFlowStmt,
		tokenizer.KeywordAtomic:     parseFlowStmt,
		tokenizer.KeywordOnce:       parseFlowStmt,
		tokenizer.KeywordTwice:      parseFlowStmt,
		tokenizer.KeywordDaily:      parseFlowStmt,
		tokenizer.KeywordWeekly:     parseFlowStmt,
		tokenizer.KeywordMonthly:    parseFlowStmt,
		tokenizer.KeywordYearly:     parseFlowStmt,
		tokenizer.KeywordHourly:     parseFlowStmt,
		tokenizer.KeywordBuy:        parseActionStmt,
		tokenizer.KeywordSell:       parseActionStmt,
		tokenizer.KeywordSellShort:  parseActionStmt,
		tokenizer.KeywordBuyCover:   parseActionStmt,
		tokenizer.KeywordSelect:     parseSelectStmt,
		tokenizer.AnnotationComment: parseAnnotationStmt,
	}
}

func (p *Parser) parseStmt() ast.Stmt {
	p.skipTerminators()
	if p.isAtEnd() {
		return nil
	}

	tok := p.cur()

	// 1. 表驱动解析
	if parseFn, ok := stmtParsers[tok.Kind]; ok {
		return parseFn(p)
	}

	// 2. Level / Override
	switch tok.Kind {
	case tokenizer.KeywordProfit, tokenizer.KeywordLoss, tokenizer.KeywordPrice:
		kind := tok.Value
		p.next()
		return parseLevelStmt(p, kind)
	case tokenizer.KeywordOverride:
		return parseOverrideStmt(p)
	}

	// 3. 配置关键字
	if isConfigKeyword(tok.Kind) {
		return parseConfigStmt(p)
	}

	// 4. 其他
	switch tok.Kind {
	case tokenizer.Identifier:
		if p.peek(1).Kind == tokenizer.Colon {
			return parseDefineStmt(p)
		}
		p.errorf("bare identifier not allowed as statement: %s", tok.Value)
		p.next()
		return nil
	case tokenizer.At, tokenizer.Dollar, tokenizer.Integer, tokenizer.Float, tokenizer.String, tokenizer.Negative, tokenizer.LeftParen:
		p.errorf("bare expression not allowed as statement")
		p.parseExpression(precLowest)
		return nil
	default:
		p.errorf("unexpected token: %v (%s)", tok.Kind, tok.Value)
		p.next()
		return nil
	}
}

func isConfigKeyword(kind tokenizer.TokenKind) bool {
	switch kind {
	case tokenizer.KeywordKeep, tokenizer.KeywordHoldMax, tokenizer.KeywordCooldown,
		tokenizer.KeywordSession, tokenizer.KeywordActive, tokenizer.KeywordPause,
		tokenizer.KeywordPartialFill, tokenizer.KeywordCompoundProfit, tokenizer.KeywordSkipIfGapped,
		tokenizer.KeywordFallback, tokenizer.KeywordPosition, tokenizer.KeywordSizing,
		tokenizer.KeywordBasePosition, tokenizer.KeywordPyramidStep, tokenizer.KeywordMaxPyramidLayers,
		tokenizer.KeywordFixedFraction, tokenizer.KeywordVolatilityTarget, tokenizer.KeywordAtrPeriod,
		tokenizer.KeywordRiskPerTrade, tokenizer.KeywordRiskPerGrid, tokenizer.KeywordMaxDrawdown,
		tokenizer.KeywordPositionDecay, tokenizer.KeywordGrossExposure, tokenizer.KeywordNetExposure,
		tokenizer.KeywordBetaNeutral, tokenizer.KeywordRebalance, tokenizer.KeywordLeverage,
		tokenizer.KeywordMargin, tokenizer.KeywordHedge, tokenizer.KeywordFundingPriority,
		tokenizer.KeywordMaxShort, tokenizer.KeywordBorrowRateLimit, tokenizer.KeywordMaxPosition,
		tokenizer.KeywordStopLoss, tokenizer.KeywordSlippageTolerance, tokenizer.KeywordCircuitBreaker,
		tokenizer.KeywordLeg, tokenizer.KeywordSide, tokenizer.KeywordAuto, tokenizer.KeywordRatio,
		tokenizer.KeywordTimeout, tokenizer.KeywordSlippage, tokenizer.KeywordFillMode,
		tokenizer.KeywordIoc, tokenizer.KeywordFok, tokenizer.KeywordGtd, tokenizer.KeywordPostOnly,
		tokenizer.KeywordBps, tokenizer.KeywordTif, tokenizer.KeywordSpot, tokenizer.KeywordFutures,
		tokenizer.KeywordPerp, tokenizer.KeywordBetween, tokenizer.KeywordOn, tokenizer.KeywordTag,
		tokenizer.KeywordSpread, tokenizer.KeywordBasis, tokenizer.KeywordFunding, tokenizer.KeywordLatency,
		tokenizer.KeywordDepth:
		return true
	}
	return false
}

// ---------- 各语句具体解析 ----------

func parseImportStmt(p *Parser) ast.Stmt {
	p.next()
	path := ""
	if p.cur().Kind == tokenizer.String {
		path = p.cur().Value
		p.next()
	}
	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)
	return &ast.ImportStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.ImportStmt}},
		Path:     path,
	}
}

func parseExportStmt(p *Parser) ast.Stmt {
	p.next()
	name := ""
	if p.cur().Kind == tokenizer.Identifier {
		name = p.cur().Value
		p.next()
	}
	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)
	return &ast.ExportStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.ExportStmt}},
		Name:     name,
	}
}

func parseUseStmt(p *Parser) ast.Stmt {
	p.next()
	name := ""
	if p.cur().Kind == tokenizer.Identifier {
		name = p.cur().Value
		p.next()
	}

	args := []ast.Expr{}
	if p.cur().Kind == tokenizer.LeftParen {
		p.next()
		for !p.isAtEnd() && p.cur().Kind != tokenizer.RightParen {
			if p.cur().Kind == tokenizer.Comma {
				p.next()
				continue
			}
			if arg := p.parseLiteralOrIdentifier(); arg != nil {
				args = append(args, arg)
			} else {
				p.next()
			}
			if p.cur().Kind == tokenizer.Comma {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.RightParen {
			p.next()
		}
	}

	alias := ""
	if p.cur().Kind == tokenizer.KeywordAs {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			alias = p.cur().Value
			p.next()
		}
	}

	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)

	return &ast.UseStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.UseStmt}},
		Name:     name,
		Args:     args,
		Alias:    alias,
	}
}

func parseSelectStmt(p *Parser) ast.Stmt {
	p.next()
	target := ""
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) {
		target = p.cur().Value
		p.next()
	}
	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)
	return &ast.SelectStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.SelectStmt}},
		Target:   target,
	}
}

func parseStrategyStmt(p *Parser) ast.Stmt {
	kind := p.cur().Value
	p.next()

	target := ""
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) {
		target = p.cur().Value
		p.next()
	}

	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}

	return &ast.StrategyStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.StrategyStmt}},
		Kind:     kind,
		Target:   target,
		Body:     body,
	}
}

func parseTemplateStmt(p *Parser) ast.Stmt {
	p.next()
	name := ""
	if p.cur().Kind == tokenizer.Identifier {
		name = p.cur().Value
		p.next()
	}

	params := []ast.ParamDef{}
	if p.cur().Kind == tokenizer.LeftParen {
		p.next()
		for !p.isAtEnd() && p.cur().Kind != tokenizer.RightParen {
			if p.cur().Kind == tokenizer.Comma {
				p.next()
				continue
			}
			param := ast.ParamDef{}
			if p.cur().Kind == tokenizer.Identifier {
				param.Name = p.cur().Value
				p.next()
			}
			if p.cur().Kind == tokenizer.Colon {
				p.next()
				if p.cur().Kind == tokenizer.Identifier {
					param.Type = p.cur().Value
					p.next()
				}
			}
			if p.cur().Kind == tokenizer.Assign {
				p.next()
				param.Default = p.parseLiteralOrIdentifier()
			}
			params = append(params, param)
			if p.cur().Kind == tokenizer.Comma {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.RightParen {
			p.next()
		}
	}

	extends := ""
	if p.cur().Kind == tokenizer.KeywordExtends {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			extends = p.cur().Value
			p.next()
		}
	}

	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}

	return &ast.TemplateStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.TemplateStmt}},
		Name:     name,
		Params:   params,
		Extends:  extends,
		Body:     body,
	}
}

func parseMacroStmt(p *Parser) ast.Stmt {
	p.next()
	name := ""
	if p.cur().Kind == tokenizer.Identifier {
		name = p.cur().Value
		p.next()
	}

	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}

	return &ast.MacroStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.MacroStmt}},
		Name:     name,
		Body:     body,
	}
}

func parseIfStmt(p *Parser) ast.Stmt {
	p.next()
	cond := p.parseExpression(precLowest)
	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return &ast.IfStmtNode{
		BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.IfStmt}},
		Condition: cond,
		Body:      body,
	}
}

func parseAssertStmt(p *Parser) ast.Stmt {
	p.next()
	cond := p.parseExpression(precLowest)
	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)
	return &ast.AssertStmtNode{
		BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.AssertStmt}},
		Condition: cond,
	}
}

func parseFlowStmt(p *Parser) ast.Stmt {
	// 可选的频次修饰符（如 once trigger ...）
	frequency := ""
	if isFrequencyKeyword(p.cur().Kind) {
		frequency = p.cur().Value
		p.next()
	}

	kind := p.cur().Value
	p.next()

	switch kind {
	case "trigger", "until":
		cond := p.parseExpression(precLowest)
		body := []ast.Stmt{}
		if p.expect(tokenizer.LeftBraces) {
			p.next()
			body = p.parseBlock(tokenizer.RightBraces)
			if p.cur().Kind == tokenizer.RightBraces {
				p.next()
			}
		}
		return &ast.FlowStmtNode{
			BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.FlowStmt}},
			Kind:      kind,
			Condition: cond,
			Body:      body,
			Frequency: frequency,
		}
	case "atomic":
		body := []ast.Stmt{}
		if p.expect(tokenizer.LeftBraces) {
			p.next()
			body = p.parseBlock(tokenizer.RightBraces)
			if p.cur().Kind == tokenizer.RightBraces {
				p.next()
			}
		}
		mode := ""
		if p.cur().Kind == tokenizer.KeywordRollback {
			mode = "rollback"
			p.next()
		} else if p.cur().Kind == tokenizer.KeywordBestEffort {
			mode = "best_effort"
			p.next()
		}
		p.consume(tokenizer.Semicolon)
		p.consume(tokenizer.Terminator)
		return &ast.FlowStmtNode{
			BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.FlowStmt}},
			Kind:     "atomic",
			Body:     body,
			Mode:     mode,
		}
	default:
		p.errorf("unknown flow kind: %s", kind)
		return nil
	}
}

func isFrequencyKeyword(kind tokenizer.TokenKind) bool {
	switch kind {
	case tokenizer.KeywordOnce, tokenizer.KeywordTwice, tokenizer.KeywordDaily,
		tokenizer.KeywordWeekly, tokenizer.KeywordMonthly, tokenizer.KeywordYearly,
		tokenizer.KeywordHourly:
		return true
	}
	return false
}

func parseLevelStmt(p *Parser, levelKind string) ast.Stmt {
	var cond ast.Expr

	if r := p.parseRangeExpr(); r != nil {
		cond = r
	} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
		cond = p.parseLiteral()
	}

	mode := levelKind
	if p.cur().Kind == tokenizer.KeywordCross {
		mode = levelKind + "_cross"
		p.next()
	}

	if p.cur().Kind == tokenizer.KeywordPriority {
		p.next()
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			p.parseLiteral() // consume priority value
		}
	}

	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}

	return &ast.FlowStmtNode{
		BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.FlowStmt}},
		Kind:      "level",
		Condition: cond,
		Body:      body,
		Mode:      mode,
	}
}

func parseOverrideStmt(p *Parser) ast.Stmt {
	p.next()

	labels := []string{}
	for p.cur().Kind == tokenizer.Identifier {
		labels = append(labels, p.cur().Value)
		p.next()
	}

	var cond ast.Expr
	if r := p.parseRangeExpr(); r != nil {
		cond = r
	}

	body := []ast.Stmt{}
	if p.expect(tokenizer.LeftBraces) {
		p.next()
		body = p.parseBlock(tokenizer.RightBraces)
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}

	return &ast.FlowStmtNode{
		BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.FlowStmt}},
		Kind:      "override",
		Condition: cond,
		Body:      body,
		Labels:    labels,
	}
}

func isActionTerminator(tok tokenizer.Token) bool {
	switch tok.Kind {
	case tokenizer.Terminator, tokenizer.Semicolon, tokenizer.EOF, tokenizer.RightBraces, tokenizer.At:
		return true
	case tokenizer.Identifier, tokenizer.KeywordRemaining:
		if tok.Value == "from" || tok.Value == "remaining" {
			return true
		}
	}
	return false
}

func parseActionStmt(p *Parser) ast.Stmt {
	action := p.cur().Value
	p.next()

	args := []ast.Expr{}

	// 解析 amount 表达式
	if !p.isAtEnd() && !isActionTerminator(p.cur()) {
		expr := p.parseExpression(precLowest)
		if expr != nil {
			args = append(args, expr)
		}
	}

	// 解析后续标识符参数（如 unit）
	for !p.isAtEnd() && !isActionTerminator(p.cur()) {
		if p.cur().Kind == tokenizer.Identifier {
			args = append(args, &ast.IdentifierExprNode{
				BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.IdentifierExpr}},
				Name:     p.cur().Value,
			})
			p.next()
		} else {
			break
		}
	}

	priceType := ""
	var price ast.Expr
	if p.cur().Kind == tokenizer.At {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			if p.cur().Value == "market" {
				priceType = "market"
				p.next()
			} else if p.cur().Value == "limit" {
				priceType = "limit"
				p.next()
				if !p.isAtEnd() && !isActionTerminator(p.cur()) {
					price = p.parseExpression(precLowest)
				}
			}
		}
	}

	modifier := ""
	if p.cur().Value == "from" {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			modifier = "from_" + p.cur().Value
			p.next()
		}
	} else if p.cur().Value == "remaining" {
		modifier = "remaining"
		p.next()
	}

	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)

	return &ast.ActionStmtNode{
		BaseStmt:  ast.BaseStmt{Node: ast.Node{Type: ast.ActionStmt}},
		Action:    action,
		Args:      args,
		Price:     price,
		PriceType: priceType,
		Modifier:  modifier,
	}
}

// isDurationKey 判断配置项是否支持 duration 序列（如 keep 1d 2h 3min）
func isDurationKey(key string) bool {
	switch key {
	case "keep", "hold_max", "cooldown", "position_decay", "circuit_breaker":
		return true
	default:
		return false
	}
}

func parseConfigStmt(p *Parser) ast.Stmt {
	key := p.cur().Value
	p.next()

	params := []ast.Expr{}
	var rng *ast.RangeExprNode
	durationKey := isDurationKey(key)

	for !p.isAtEnd() {
		tok := p.cur()
		if tok.Kind == tokenizer.Semicolon || tok.Kind == tokenizer.Terminator || tok.Kind == tokenizer.EOF || tok.Kind == tokenizer.RightBraces {
			break
		}

		// 尝试范围表达式
		if r := p.parseRangeExpr(); r != nil {
			rng = r
			continue
		}

		// 尝试字面量（数字、负号、时间、日期）
		if tokenizer.TokenIsNumber(tok) || tok.Kind == tokenizer.Negative || tok.Kind == tokenizer.Time || tok.Kind == tokenizer.Date {
			lit := p.parseLiteral()
			if lit == nil {
				continue
			}
			// duration 配置项：连续的数字+时间单位 聚合成 ListLiteralNode
			if durationKey && lit.Unit != "" && tokenizer.IsTimeUnit(lit.Unit) {
				items := []ast.Expr{lit}
				for !p.isAtEnd() {
					t := p.cur()
					if t.Kind == tokenizer.Semicolon || t.Kind == tokenizer.Terminator || t.Kind == tokenizer.EOF || t.Kind == tokenizer.RightBraces {
						break
					}
					if !tokenizer.TokenIsNumber(t) && t.Kind != tokenizer.Negative {
						break
					}
					nextLit := p.parseLiteral()
					if nextLit == nil || nextLit.Unit == "" || !tokenizer.IsTimeUnit(nextLit.Unit) {
						break
					}
					items = append(items, nextLit)
				}
				params = append(params, &ast.ListLiteralNode{
					BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.ListExpr}},
					Items:    items,
				})
				continue
			}
			params = append(params, lit)
			continue
		}

		// 标识符/关键字作为值
		if tok.Kind == tokenizer.Identifier || tok.Kind.IsKeyword() {
			if tok.Kind == tokenizer.KeywordRollback || tok.Kind == tokenizer.KeywordBestEffort {
				break
			}
			params = append(params, &ast.IdentifierExprNode{
				BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.IdentifierExpr}},
				Name:     tok.Value,
			})
			p.next()
			continue
		}

		// 字符串
		if tok.Kind == tokenizer.String {
			params = append(params, &ast.Literal{
				BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.LiteralExpr}},
				Value:    tok.Value,
			})
			p.next()
			continue
		}

		// 变量
		if tok.Kind == tokenizer.Dollar {
			params = append(params, parseVariablePrefix(p))
			continue
		}

		// 其他 token 跳过
		p.next()
	}

	p.consume(tokenizer.Semicolon)
	p.consume(tokenizer.Terminator)

	return &ast.ConfigStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.ConfigStmt}},
		Key:      key,
		Params:   params,
		Range:    rng,
	}
}

func parseAnnotationStmt(p *Parser) ast.Stmt {
	parts := strings.SplitN(p.cur().Value, "|", 2)
	p.next()

	key := ""
	val := ""
	if len(parts) == 2 {
		key = parts[0]
		val = parts[1]
	} else {
		val = parts[0]
	}

	return &ast.AnnotationStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.AnnotationStmt}},
		Key:      key,
		Value:    val,
	}
}

func parseDefineStmt(p *Parser) ast.Stmt {
	name := p.cur().Value
	p.next() // consume identifier
	p.next() // consume :

	var value ast.Expr
	if tokenizer.IsLiteral(p.cur()) || p.cur().Kind == tokenizer.Identifier {
		if tokenizer.IsLiteral(p.cur()) {
			value = p.parseLiteral()
		} else {
			value = &ast.IdentifierExprNode{
				BaseExpr: ast.BaseExpr{Node: ast.Node{Type: ast.IdentifierExpr}},
				Name:     p.cur().Value,
			}
			p.next()
		}
	} else if p.cur().Kind == tokenizer.Dollar {
		value = parseVariablePrefix(p)
	} else if p.cur().Kind == tokenizer.At {
		value = parseReferencePrefix(p)
	} else {
		p.errorf("expected value after ':' in define statement")
	}

	return &ast.DefineStmtNode{
		BaseStmt: ast.BaseStmt{Node: ast.Node{Type: ast.DefineStmt}},
		Name:     name,
		Value:    value,
	}
}
