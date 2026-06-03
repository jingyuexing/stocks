package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
)

func createExpression(kind ast.NodeType) ast.ExpressionNode {
	return ast.ExpressionNode{
		Node:   ast.Node{Type: kind},
		Params: make([]ast.Literal, 0),
		Body:   make([]ast.ExpressionNode, 0),
	}
}

func createRange() ast.RangeExpressionNode {
	return ast.RangeExpressionNode{Node: ast.Node{Type: ast.RangeExpression}}
}

func createRoot() ast.RootNode {
	return ast.RootNode{
		Node:       ast.Node{Type: ast.Program},
		Expression: make([]ast.ExpressionNode, 0),
	}
}

func createLiteral(token tokenizer.Token) ast.Literal {
	lit := ast.Literal{Value: token.Value}
	switch token.Kind {
	case tokenizer.Float:
		lit.Node.Type = ast.FloatLiteral
	case tokenizer.Integer:
		lit.Node.Type = ast.IntegerLiteral
	case tokenizer.String:
		lit.Node.Type = ast.StringLiteral
	default:
		lit.Node.Type = ast.TextLiteral
	}
	return lit
}

func durationToNumber(literal ast.Literal) time.Duration {
	if literal.Node.Type != ast.DurationLiteral {
		return 0
	}
	var d time.Duration
	switch literal.Unit {
	case "ms":
		d = time.Millisecond
	case "s":
		d = time.Second
	case "m", "min":
		d = time.Minute
	case "h", "H":
		d = time.Hour
	case "d":
		d = time.Hour * 24
	case "W":
		d = time.Hour * 24 * 7
	case "M":
		d = time.Hour * 24 * 30
	case "Y":
		d = time.Hour * 24 * 365
	}
	v, _ := strconv.ParseInt(literal.Value, 10, 64)
	return time.Duration(v * int64(d))
}

// parserImpl 解析器
type parserImpl struct {
	tokens  []tokenizer.Token
	current int
	length  int
}

// NewParser 创建解析器
func NewParser(tokens []tokenizer.Token) *parserImpl {
	return &parserImpl{tokens: tokens, current: 0, length: len(tokens)}
}

func (p *parserImpl) next() {
	p.current++
	if p.current >= p.length {
		p.current = p.length - 1
	}
}

func (p *parserImpl) peek(offset int) tokenizer.Token {
	idx := p.current + offset
	if idx >= 0 && idx < p.length {
		return p.tokens[idx]
	}
	return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
}

func (p *parserImpl) cur() tokenizer.Token {
	if p.current < p.length {
		return p.tokens[p.current]
	}
	return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
}

func (p *parserImpl) skipTerminators() {
	for p.current < p.length {
		k := p.cur().Kind
		if k == tokenizer.Terminator || k == tokenizer.Semicolon {
			p.next()
		} else {
			break
		}
	}
}

func (p *parserImpl) isGlobalConfigToken(kind tokenizer.TokenKind) bool {
	switch kind {
	case tokenizer.KeywordKeep, tokenizer.KeywordHoldMax, tokenizer.KeywordCooldown,
		tokenizer.KeywordSession, tokenizer.KeywordActive, tokenizer.KeywordPause,
		tokenizer.KeywordPartialFill, tokenizer.KeywordCompoundProfit, tokenizer.KeywordSkipIfGapped, tokenizer.KeywordFallback,
		tokenizer.KeywordPosition, tokenizer.KeywordSizing, tokenizer.KeywordBasePosition,
		tokenizer.KeywordPyramidStep, tokenizer.KeywordMaxPyramidLayers, tokenizer.KeywordFixedFraction,
		tokenizer.KeywordVolatilityTarget, tokenizer.KeywordAtrPeriod, tokenizer.KeywordRiskPerTrade,
		tokenizer.KeywordRiskPerGrid, tokenizer.KeywordMaxDrawdown, tokenizer.KeywordPositionDecay,
		tokenizer.KeywordGrossExposure, tokenizer.KeywordNetExposure, tokenizer.KeywordBetaNeutral,
		tokenizer.KeywordRebalance, tokenizer.KeywordLeverage, tokenizer.KeywordMargin,
		tokenizer.KeywordHedge, tokenizer.KeywordFundingPriority, tokenizer.KeywordMaxShort,
		tokenizer.KeywordBorrowRateLimit:
		return true
	}
	return false
}

func (p *parserImpl) isRiskConfigToken(kind tokenizer.TokenKind) bool {
	switch kind {
	case tokenizer.KeywordMaxPosition, tokenizer.KeywordStopLoss, tokenizer.KeywordSlippageTolerance, tokenizer.KeywordCircuitBreaker:
		return true
	}
	return false
}

func (p *parserImpl) parseLiteral() ast.Literal {
	literal := ast.Literal{}
	neg := false
	if p.cur().Kind == tokenizer.Negative {
		neg = true
		p.next()
	}
	tok := p.cur()
	if !tokenizer.TokenIsNumber(tok) {
		if neg {
			fmt.Println("Expected number after '-'")
		}
		return literal
	}
	literal = createLiteral(tok)
	if neg {
		literal.Value = "-" + literal.Value
	}
	if p.current < p.length-1 {
		nextTok := p.tokens[p.current+1]
		if tokenizer.IsTimeUnitToken(nextTok) {
			literal.Node.Type = ast.DurationLiteral
			literal.Unit = nextTok.Value
			p.next()
		} else if nextTok.Kind == tokenizer.Per {
			literal.Node.Type = ast.PercentLiteral
			literal.Unit = nextTok.Value
			p.next()
		}
	}
	p.next()
	return literal
}

func (p *parserImpl) parseDuration() ast.Literal {
	dur := ast.Literal{}
	neg := false
	if p.cur().Kind == tokenizer.Negative {
		neg = true
		p.next()
	}
	if !tokenizer.TokenIsNumber(p.cur()) {
		return dur
	}
	if p.current+1 >= p.length {
		return dur
	}
	nextTok := p.tokens[p.current+1]
	if !tokenizer.IsTimeUnitToken(nextTok) {
		if neg {
			fmt.Printf("invalid time unit %s\n", nextTok.Value)
		}
		return dur
	}
	dur = createLiteral(p.cur())
	if neg {
		dur.Value = "-" + dur.Value
	}
	dur.Node.Type = ast.DurationLiteral
	dur.Unit = nextTok.Value
	p.next()
	p.next()
	return dur
}

func (p *parserImpl) parseRange() *ast.RangeExpressionNode {
	saved := p.current
	if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		p.next()
		r := createRange()
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			r.End = p.parseLiteral()
		}
		return &r
	}
	if !tokenizer.TokenIsNumber(p.cur()) && p.cur().Kind != tokenizer.Negative {
		return nil
	}
	begin := p.parseLiteral()
	if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		p.next()
		r := createRange()
		r.Begin = begin
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			r.End = p.parseLiteral()
		}
		return &r
	}
	p.current = saved
	return nil
}

func (p *parserImpl) parseDate() ast.Literal {
	lit := ast.Literal{}
	if p.cur().Kind != tokenizer.Integer {
		return lit
	}
	year := p.cur().Value
	p.next()
	if p.cur().Kind != tokenizer.Negative {
		return lit
	}
	p.next()
	if p.cur().Kind != tokenizer.Integer {
		return lit
	}
	month := p.cur().Value
	p.next()
	if p.cur().Kind != tokenizer.Negative {
		return lit
	}
	p.next()
	if p.cur().Kind != tokenizer.Integer {
		return lit
	}
	day := p.cur().Value
	p.next()
	lit.Node.Type = ast.DateLiteral
	lit.Value = fmt.Sprintf("%s-%s-%s", year, month, day)
	return lit
}

func (p *parserImpl) parseActionStmt() ast.ExpressionNode {
	var expr ast.ExpressionNode
	switch p.cur().Kind {
	case tokenizer.KeywordBuy:
		expr = createExpression(ast.BuyExpression)
	case tokenizer.KeywordSell:
		expr = createExpression(ast.SellExpression)
	case tokenizer.KeywordSellShort:
		expr = createExpression(ast.SellShortExpression)
	case tokenizer.KeywordBuyCover:
		expr = createExpression(ast.BuyCoverExpression)
	default:
		return expr
	}
	p.next()
	for p.current < p.length {
		tok := p.cur()
		if tok.Kind == tokenizer.Terminator || tok.Kind == tokenizer.Semicolon ||
			tok.Kind == tokenizer.EOF || tok.Kind == tokenizer.RightBraces || tok.Kind == tokenizer.At {
			break
		}
		if tok.Kind == tokenizer.Identifier && (tok.Value == "from" || tok.Value == "remaining") {
			break
		}
		if tokenizer.TokenIsNumber(tok) || tok.Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		} else if tok.Kind == tokenizer.Identifier {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: tok.Value})
			p.next()
			if p.cur().Kind == tokenizer.Plus {
				p.next()
			}
		} else if tok.Kind == tokenizer.Dollar {
			varExpr := p.parseVariable()
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.ReferenceLiteral}, Value: varExpr.Name})
		} else if tok.Kind == tokenizer.String {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.StringLiteral}, Value: tok.Value})
			p.next()
		} else {
			p.next()
		}
	}
	if p.cur().Kind == tokenizer.At {
		p.next()
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "market" {
			expr.Name = "market"
			p.next()
		} else if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "limit" {
			expr.Name = "limit"
			p.next()
			if p.cur().Kind == tokenizer.Plus {
				p.next()
			}
			if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
				expr.Value = p.parseLiteral()
			}
		}
	}
	if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "from" {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: "from_" + p.cur().Value})
			p.next()
		}
	} else if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "remaining" {
		expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: "remaining"})
		p.next()
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseGlobalConfig() ast.ExpressionNode {
	kindMap := map[tokenizer.TokenKind]ast.NodeType{
		tokenizer.KeywordKeep: ast.KeepExpression, tokenizer.KeywordHoldMax: ast.HoldMaxExpression,
		tokenizer.KeywordCooldown: ast.CooldownExpression, tokenizer.KeywordSession: ast.SessionExpression,
		tokenizer.KeywordActive: ast.ActiveExpression, tokenizer.KeywordPause: ast.PauseExpression,
		tokenizer.KeywordPartialFill: ast.PartialFillExpression, tokenizer.KeywordCompoundProfit: ast.CompoundProfitExpression,
		tokenizer.KeywordSkipIfGapped: ast.SkipIfGappedExpression, tokenizer.KeywordFallback: ast.FallbackExpression,
		tokenizer.KeywordPosition: ast.PositionExpression, tokenizer.KeywordSizing: ast.SizingExpression,
		tokenizer.KeywordBasePosition: ast.BasePositionExpression, tokenizer.KeywordPyramidStep: ast.PyramidStepExpression,
		tokenizer.KeywordMaxPyramidLayers: ast.MaxPyramidLayersExpression, tokenizer.KeywordFixedFraction: ast.FixedFractionExpression,
		tokenizer.KeywordVolatilityTarget: ast.VolatilityTargetExpression, tokenizer.KeywordAtrPeriod: ast.AtrPeriodExpression,
		tokenizer.KeywordRiskPerTrade: ast.RiskPerTradeExpression, tokenizer.KeywordRiskPerGrid: ast.RiskPerGridExpression,
		tokenizer.KeywordMaxDrawdown: ast.MaxDrawdownExpression, tokenizer.KeywordPositionDecay: ast.PositionDecayExpression,
		tokenizer.KeywordGrossExposure: ast.GrossExposureExpression, tokenizer.KeywordNetExposure: ast.NetExposureExpression,
		tokenizer.KeywordBetaNeutral: ast.BetaNeutralExpression, tokenizer.KeywordRebalance: ast.RebalanceExpression,
		tokenizer.KeywordLeverage: ast.LeverageExpression, tokenizer.KeywordMargin: ast.MarginExpression,
		tokenizer.KeywordHedge: ast.HedgeExpression, tokenizer.KeywordFundingPriority: ast.FundingPriorityExpression,
		tokenizer.KeywordMaxShort: ast.MaxShortExpression, tokenizer.KeywordBorrowRateLimit: ast.BorrowRateLimitExpression,
	}
	nodeType, ok := kindMap[p.cur().Kind]
	if !ok {
		return ast.ExpressionNode{}
	}
	expr := createExpression(nodeType)
	p.next()
	switch nodeType {
	case ast.KeepExpression, ast.HoldMaxExpression:
		for p.current < p.length {
			k := p.cur().Kind
			if k == tokenizer.Terminator || k == tokenizer.Semicolon || k == tokenizer.EOF || k == tokenizer.RightBraces {
				break
			}
			dur := p.parseDuration()
			if dur.Node.Type == ast.DurationLiteral {
				expr.Params = append(expr.Params, dur)
			} else {
				break
			}
		}
	case ast.CooldownExpression:
		dur := p.parseDuration()
		if dur.Node.Type == ast.DurationLiteral {
			expr.Params = append(expr.Params, dur)
		}
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "per" {
			p.next()
			if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "level" {
				expr.Name = "per_level"
				p.next()
			}
		}
	case ast.SessionExpression:
		for p.current < p.length {
			if !tokenizer.TokenIsNumber(p.cur()) {
				break
			}
			hour := p.cur().Value
			p.next()
			if p.cur().Kind == tokenizer.Colon {
				p.next()
			}
			if !tokenizer.TokenIsNumber(p.cur()) {
				break
			}
			minute := p.cur().Value
			p.next()
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TimePointLiteral}, Value: hour + ":" + minute})
			if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot || p.cur().Kind == tokenizer.Negative {
				if p.cur().Kind == tokenizer.Negative {
					p.next()
				} else {
					p.next()
				}
			} else {
				break
			}
		}
		if p.cur().Kind == tokenizer.Identifier && tokenizer.IsTimezone(p.cur().Value) {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
	case ast.ActiveExpression, ast.PauseExpression:
		date1 := p.parseDate()
		if date1.Node.Type == ast.DateLiteral {
			expr.Params = append(expr.Params, date1)
		}
		if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
			p.next()
			date2 := p.parseDate()
			if date2.Node.Type == ast.DateLiteral {
				expr.Params = append(expr.Params, date2)
			}
		}
	case ast.PartialFillExpression, ast.CompoundProfitExpression, ast.FallbackExpression, ast.BetaNeutralExpression, ast.MarginExpression, ast.FundingPriorityExpression:
		if p.cur().Kind == tokenizer.Identifier || p.cur().Kind == tokenizer.KeywordTrue || p.cur().Kind == tokenizer.KeywordFalse {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
	case ast.SkipIfGappedExpression:
	case ast.SizingExpression:
		if p.cur().Kind == tokenizer.Identifier && tokenizer.IsSizingMethod(p.cur().Value) {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
	case ast.PositionExpression:
		if p.cur().Kind == tokenizer.Identifier && (p.cur().Value == "fixed" || p.cur().Value == "dynamic") {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
	case ast.RebalanceExpression:
		if p.cur().Kind == tokenizer.Identifier && tokenizer.IsRebalancePeriod(p.cur().Value) {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "at" {
			p.next()
			if tokenizer.TokenIsNumber(p.cur()) {
				hour := p.cur().Value
				p.next()
				if p.cur().Kind == tokenizer.Colon {
					p.next()
				}
				if tokenizer.TokenIsNumber(p.cur()) {
					minute := p.cur().Value
					p.next()
					expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TimePointLiteral}, Value: hour + ":" + minute})
				}
			}
		}
	case ast.BasePositionExpression, ast.MaxDrawdownExpression, ast.BorrowRateLimitExpression,
		ast.FixedFractionExpression, ast.VolatilityTargetExpression,
		ast.GrossExposureExpression, ast.NetExposureExpression,
		ast.RiskPerTradeExpression, ast.RiskPerGridExpression, ast.PositionDecayExpression,
		ast.SlippageToleranceExpression:
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if nodeType == ast.RiskPerTradeExpression || nodeType == ast.RiskPerGridExpression || nodeType == ast.GrossExposureExpression || nodeType == ast.NetExposureExpression {
			if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "capital" {
				expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: "capital"})
				p.next()
			}
		}
		if nodeType == ast.PositionDecayExpression {
			if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "per" {
				p.next()
				dur := p.parseDuration()
				if dur.Node.Type == ast.DurationLiteral {
					expr.Params = append(expr.Params, dur)
				}
			}
		}
	case ast.PyramidStepExpression:
		if tokenizer.TokenIsNumber(p.cur()) {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "x" {
			p.next()
		}
	case ast.MaxPyramidLayersExpression, ast.AtrPeriodExpression, ast.MaxShortExpression, ast.LeverageExpression:
		if tokenizer.TokenIsNumber(p.cur()) {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if nodeType == ast.LeverageExpression && p.cur().Kind == tokenizer.Identifier && p.cur().Value == "x" {
			p.next()
		}
		if nodeType == ast.MaxShortExpression && p.cur().Kind == tokenizer.Identifier && p.cur().Value == "shares" {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: "shares"})
			p.next()
		}
	case ast.HedgeExpression:
		if tokenizer.TokenIsNumber(p.cur()) {
			left := p.cur().Value
			p.next()
			if p.cur().Kind == tokenizer.Colon {
				p.next()
			}
			if tokenizer.TokenIsNumber(p.cur()) {
				right := p.cur().Value
				p.next()
				expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: left + ":" + right})
			}
		}
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseRiskConfig() ast.ExpressionNode {
	kindMap := map[tokenizer.TokenKind]ast.NodeType{
		tokenizer.KeywordMaxPosition: ast.MaxPositionExpression, tokenizer.KeywordStopLoss: ast.StopLossExpression,
		tokenizer.KeywordSlippageTolerance: ast.SlippageToleranceExpression, tokenizer.KeywordCircuitBreaker: ast.CircuitBreakerExpression,
	}
	nodeType, ok := kindMap[p.cur().Kind]
	if !ok {
		return ast.ExpressionNode{}
	}
	expr := createExpression(nodeType)
	p.next()
	switch nodeType {
	case ast.MaxPositionExpression:
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if p.cur().Kind == tokenizer.Identifier {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
	case ast.StopLossExpression:
		r := p.parseRange()
		if r != nil {
			expr.Range = r
		} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			// 单点值，如 stop_loss -10%
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "on" {
			p.next()
			if p.cur().Kind == tokenizer.Identifier {
				expr.Name = "on_" + p.cur().Value
				p.next()
			}
		}
	case ast.SlippageToleranceExpression:
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
	case ast.CircuitBreakerExpression:
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
		if p.cur().Kind == tokenizer.Identifier && p.cur().Value == "in" {
			p.next()
			dur := p.parseDuration()
			if dur.Node.Type == ast.DurationLiteral {
				expr.Params = append(expr.Params, dur)
			}
		}
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseLevelBlock() ast.ExpressionNode {
	if p.cur().Kind == tokenizer.KeywordOverride {
		expr := createExpression(ast.OverrideExpression)
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.Range && p.cur().Kind != tokenizer.TribleDot && p.cur().Kind != tokenizer.LeftBraces && p.cur().Kind != tokenizer.EOF {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value})
			p.next()
		}
		if p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
			r := p.parseRange()
			if r != nil {
				expr.Range = r
			}
		}
		if p.cur().Kind == tokenizer.LeftBraces {
			p.next()
			expr.Body = p.parseGridBody()
			if p.cur().Kind == tokenizer.RightBraces {
				p.next()
			}
		}
		return expr
	}
	expr := createExpression(ast.GridExpression)
	r := p.parseRange()
	if r != nil {
		expr.Range = r
	}
	if p.cur().Kind == tokenizer.KeywordCross {
		expr.Name = "cross"
		p.next()
	}
	if p.cur().Kind == tokenizer.KeywordPriority {
		p.next()
		if tokenizer.TokenIsNumber(p.cur()) {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
				break
			}
			if p.cur().Kind == tokenizer.KeywordBuy || p.cur().Kind == tokenizer.KeywordSell ||
				p.cur().Kind == tokenizer.KeywordSellShort || p.cur().Kind == tokenizer.KeywordBuyCover {
				expr.Body = append(expr.Body, p.parseActionStmt())
			} else {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseGridBody() []ast.ExpressionNode {
	body := make([]ast.ExpressionNode, 0)
	for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
		p.skipTerminators()
		if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
			break
		}
		tok := p.cur()
		switch tok.Kind {
		case tokenizer.KeywordLoss:
			body = append(body, p.parseLoss())
		case tokenizer.KeywordProfit:
			body = append(body, p.parseProfit())
		case tokenizer.KeywordBuy, tokenizer.KeywordSell, tokenizer.KeywordSellShort, tokenizer.KeywordBuyCover:
			body = append(body, p.parseActionStmt())
		case tokenizer.KeywordKeep, tokenizer.KeywordHoldMax, tokenizer.KeywordCooldown,
			tokenizer.KeywordSession, tokenizer.KeywordActive, tokenizer.KeywordPause,
			tokenizer.KeywordPartialFill, tokenizer.KeywordCompoundProfit, tokenizer.KeywordSkipIfGapped, tokenizer.KeywordFallback,
			tokenizer.KeywordPosition, tokenizer.KeywordSizing, tokenizer.KeywordBasePosition,
			tokenizer.KeywordPyramidStep, tokenizer.KeywordMaxPyramidLayers, tokenizer.KeywordFixedFraction,
			tokenizer.KeywordVolatilityTarget, tokenizer.KeywordAtrPeriod, tokenizer.KeywordRiskPerTrade,
			tokenizer.KeywordRiskPerGrid, tokenizer.KeywordMaxDrawdown, tokenizer.KeywordPositionDecay,
			tokenizer.KeywordGrossExposure, tokenizer.KeywordNetExposure, tokenizer.KeywordBetaNeutral,
			tokenizer.KeywordRebalance, tokenizer.KeywordLeverage, tokenizer.KeywordMargin,
			tokenizer.KeywordHedge, tokenizer.KeywordFundingPriority, tokenizer.KeywordMaxShort,
			tokenizer.KeywordBorrowRateLimit:
			body = append(body, p.parseGlobalConfig())
		case tokenizer.KeywordMaxPosition, tokenizer.KeywordStopLoss, tokenizer.KeywordSlippageTolerance, tokenizer.KeywordCircuitBreaker:
			body = append(body, p.parseRiskConfig())
		case tokenizer.Negative, tokenizer.Integer, tokenizer.Float, tokenizer.Range, tokenizer.TribleDot:
			body = append(body, p.parseLevelBlock())
		case tokenizer.KeywordOverride:
			body = append(body, p.parseLevelBlock())
		case tokenizer.Dollar:
			body = append(body, p.parseVariable())
		case tokenizer.KeywordCross, tokenizer.KeywordPriority:
			fmt.Printf("Unexpected token in grid body: %v\n", tok)
			p.next()
		case tokenizer.Terminator, tokenizer.Semicolon:
			p.next()
		default:
			fmt.Printf("Unexpected token in grid body: %v\n", tok)
			p.next()
		}
	}
	return body
}

func (p *parserImpl) parseLoss() ast.ExpressionNode {
	expr := createExpression(ast.LossExpression)
	p.next()
	if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative ||
		p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		r := p.parseRange()
		if r != nil {
			expr.Range = r
		} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseProfit() ast.ExpressionNode {
	expr := createExpression(ast.ProfitExpression)
	p.next()
	if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative ||
		p.cur().Kind == tokenizer.Range || p.cur().Kind == tokenizer.TribleDot {
		r := p.parseRange()
		if r != nil {
			expr.Range = r
		} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Params = append(expr.Params, p.parseLiteral())
		}
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseGrid() ast.ExpressionNode {
	expr := createExpression(ast.GridExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	} else {
		fmt.Println("Expected '{' after 'grid'")
	}
	return expr
}

func (p *parserImpl) parseLong() ast.ExpressionNode {
	expr := createExpression(ast.LongExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseShort() ast.ExpressionNode {
	expr := createExpression(ast.ShortExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseBoth() ast.ExpressionNode {
	expr := createExpression(ast.BothExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.KeywordLong && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.KeywordLong {
				break
			}
			if p.isGlobalConfigToken(p.cur().Kind) {
				expr.Body = append(expr.Body, p.parseGlobalConfig())
			} else {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.KeywordLong {
			expr.Body = append(expr.Body, p.parseLong())
		}
		if p.cur().Kind == tokenizer.KeywordShort {
			expr.Body = append(expr.Body, p.parseShort())
		}
		for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
				break
			}
			if p.isGlobalConfigToken(p.cur().Kind) {
				expr.Body = append(expr.Body, p.parseGlobalConfig())
			} else if p.isRiskConfigToken(p.cur().Kind) {
				expr.Body = append(expr.Body, p.parseRiskConfig())
			} else {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parsePortfolio() ast.ExpressionNode {
	expr := createExpression(ast.PortfolioExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
				break
			}
			tok := p.cur()
			switch tok.Kind {
			case tokenizer.KeywordGrid:
				expr.Body = append(expr.Body, p.parseGrid())
			case tokenizer.KeywordLong:
				expr.Body = append(expr.Body, p.parseLong())
			case tokenizer.KeywordShort:
				expr.Body = append(expr.Body, p.parseShort())
			case tokenizer.KeywordBoth:
				expr.Body = append(expr.Body, p.parseBoth())
			case tokenizer.KeywordPortfolio:
				expr.Body = append(expr.Body, p.parsePortfolio())
			case tokenizer.KeywordUse:
				expr.Body = append(expr.Body, p.parseUse())
			case tokenizer.KeywordTemplate:
				expr.Body = append(expr.Body, p.parseTemplate())
			default:
				if p.isGlobalConfigToken(tok.Kind) {
					expr.Body = append(expr.Body, p.parseGlobalConfig())
				} else if p.isRiskConfigToken(tok.Kind) {
					expr.Body = append(expr.Body, p.parseRiskConfig())
				} else {
					p.next()
				}
			}
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseTemplate() ast.ExpressionNode {
	expr := createExpression(ast.TemplateExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftParen {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightParen && p.cur().Kind != tokenizer.EOF {
			if p.cur().Kind == tokenizer.Comma {
				p.next()
				continue
			}
			param := createExpression(ast.ParamExpression)
			if p.cur().Kind == tokenizer.Identifier {
				param.Name = p.cur().Value
				p.next()
			}
			if p.cur().Kind == tokenizer.Colon {
				p.next()
				if p.cur().Kind == tokenizer.Identifier {
					param.Value = ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value}
					p.next()
				}
			}
			if p.cur().Kind == tokenizer.Assign {
				p.next()
				if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String || p.cur().Kind == tokenizer.Identifier {
					param.Value = createLiteral(p.cur())
					p.next()
				}
			}
			expr.Body = append(expr.Body, param)
			if p.cur().Kind == tokenizer.Comma {
				p.next()
			}
		}
		if p.cur().Kind == tokenizer.RightParen {
			p.next()
		}
	}
	if p.cur().Kind == tokenizer.KeywordExtends {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			extendsExpr := createExpression(ast.ExtendsExpression)
			extendsExpr.Name = p.cur().Value
			expr.Body = append([]ast.ExpressionNode{extendsExpr}, expr.Body...)
			p.next()
		}
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
				break
			}
			tok := p.cur()
			switch tok.Kind {
			case tokenizer.KeywordGrid:
				expr.Body = append(expr.Body, p.parseGrid())
			case tokenizer.KeywordLong:
				expr.Body = append(expr.Body, p.parseLong())
			case tokenizer.KeywordShort:
				expr.Body = append(expr.Body, p.parseShort())
			case tokenizer.KeywordBoth:
				expr.Body = append(expr.Body, p.parseBoth())
			case tokenizer.KeywordPortfolio:
				expr.Body = append(expr.Body, p.parsePortfolio())
			case tokenizer.KeywordIf:
				expr.Body = append(expr.Body, p.parseConditional())
			case tokenizer.KeywordAssert:
				expr.Body = append(expr.Body, p.parseAssert())
			default:
				if p.isGlobalConfigToken(tok.Kind) {
					expr.Body = append(expr.Body, p.parseGlobalConfig())
				} else {
					p.next()
				}
			}
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseUse() ast.ExpressionNode {
	expr := createExpression(ast.UseExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftParen {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightParen && p.cur().Kind != tokenizer.EOF {
			if p.cur().Kind == tokenizer.Comma {
				p.next()
				continue
			}
			if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String || p.cur().Kind == tokenizer.Identifier {
				expr.Params = append(expr.Params, createLiteral(p.cur()))
				p.next()
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
	if p.cur().Kind == tokenizer.KeywordAs {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			expr.Alias = p.cur().Value
			p.next()
		}
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseImport() ast.ExpressionNode {
	expr := createExpression(ast.ImportExpression)
	p.next()
	if p.cur().Kind == tokenizer.String {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseExport() ast.ExpressionNode {
	expr := createExpression(ast.ExportExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseMacro() ast.ExpressionNode {
	expr := createExpression(ast.MacroExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		expr.Body = p.parseGridBody()
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseConditional() ast.ExpressionNode {
	expr := createExpression(ast.ConditionalExpression)
	p.next()
	left := ast.ExpressionNode{}
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String {
		left = ast.ExpressionNode{Node: ast.Node{Type: ast.LiteralExpression}, Value: createLiteral(p.cur())}
		p.next()
	} else if p.cur().Kind == tokenizer.Dollar {
		left = p.parseVariable()
	}
	op := ""
	if p.cur().Kind == tokenizer.EqualEqual || p.cur().Kind == tokenizer.NotEqual ||
		p.cur().Kind == tokenizer.LessThan || p.cur().Kind == tokenizer.GreaterThan ||
		p.cur().Kind == tokenizer.LessEqual || p.cur().Kind == tokenizer.GreaterEqual {
		op = p.cur().Value
		p.next()
	}
	right := ast.ExpressionNode{}
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String {
		right = ast.ExpressionNode{Node: ast.Node{Type: ast.LiteralExpression}, Value: createLiteral(p.cur())}
		p.next()
	} else if p.cur().Kind == tokenizer.Dollar {
		right = p.parseVariable()
	}
	comp := createExpression(ast.ComparisonExpression)
	comp.Operator = op
	comp.Left = &left
	comp.Right = &right
	expr.Body = append(expr.Body, comp)
	if p.cur().Kind == tokenizer.LeftBraces {
		p.next()
		for p.current < p.length && p.cur().Kind != tokenizer.RightBraces && p.cur().Kind != tokenizer.EOF {
			p.skipTerminators()
			if p.cur().Kind == tokenizer.RightBraces || p.cur().Kind == tokenizer.EOF {
				break
			}
			tok := p.cur()
			switch tok.Kind {
			case tokenizer.KeywordGrid:
				expr.Body = append(expr.Body, p.parseGrid())
			case tokenizer.KeywordLong:
				expr.Body = append(expr.Body, p.parseLong())
			case tokenizer.KeywordShort:
				expr.Body = append(expr.Body, p.parseShort())
			case tokenizer.KeywordBoth:
				expr.Body = append(expr.Body, p.parseBoth())
			case tokenizer.KeywordPortfolio:
				expr.Body = append(expr.Body, p.parsePortfolio())
			default:
				if p.isGlobalConfigToken(tok.Kind) {
					expr.Body = append(expr.Body, p.parseGlobalConfig())
				} else {
					p.next()
				}
			}
		}
		if p.cur().Kind == tokenizer.RightBraces {
			p.next()
		}
	}
	return expr
}

func (p *parserImpl) parseAssert() ast.ExpressionNode {
	expr := createExpression(ast.AssertExpression)
	p.next()
	left := ast.ExpressionNode{}
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String {
		left = ast.ExpressionNode{Node: ast.Node{Type: ast.LiteralExpression}, Value: createLiteral(p.cur())}
		p.next()
	} else if p.cur().Kind == tokenizer.Dollar {
		left = p.parseVariable()
	}
	op := ""
	if p.cur().Kind == tokenizer.EqualEqual || p.cur().Kind == tokenizer.NotEqual ||
		p.cur().Kind == tokenizer.LessThan || p.cur().Kind == tokenizer.GreaterThan ||
		p.cur().Kind == tokenizer.LessEqual || p.cur().Kind == tokenizer.GreaterEqual {
		op = p.cur().Value
		p.next()
	}
	right := ast.ExpressionNode{}
	if p.cur().Kind == tokenizer.Identifier || tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.String {
		right = ast.ExpressionNode{Node: ast.Node{Type: ast.LiteralExpression}, Value: createLiteral(p.cur())}
		p.next()
	} else if p.cur().Kind == tokenizer.Dollar {
		right = p.parseVariable()
	}
	comp := createExpression(ast.ComparisonExpression)
	comp.Operator = op
	comp.Left = &left
	comp.Right = &right
	expr.Body = append(expr.Body, comp)
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseBuy() ast.ExpressionNode {
	expr := createExpression(ast.BuyExpression)
	p.next()
	if p.cur().Kind == tokenizer.At {
		p.next()
		if p.cur().Kind == tokenizer.Identifier {
			expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.ReferenceLiteral}, Value: p.cur().Value})
			p.next()
		} else {
			fmt.Println("Unexpected token after @ in buy expression")
		}
	} else if p.cur().Kind == tokenizer.Dollar {
		varExpr := p.parseVariable()
		expr.Params = append(expr.Params, ast.Literal{Node: ast.Node{Type: ast.ReferenceLiteral}, Value: varExpr.Name})
	} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
		expr.Params = append(expr.Params, p.parseLiteral())
	} else {
		fmt.Println("Unexpected token in buy expression")
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseSell() ast.ExpressionNode {
	expr := createExpression(ast.SellExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	}
	if p.cur().Kind == tokenizer.Plus {
		p.next()
	}
	if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
		expr.Params = append(expr.Params, p.parseLiteral())
	} else {
		fmt.Println("Unexpected token in sell expression")
	}
	if p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon {
		p.next()
	}
	return expr
}

func (p *parserImpl) parseKeep() ast.ExpressionNode {
	expr := createExpression(ast.KeepExpression)
	p.next()
	for p.current < p.length {
		k := p.cur().Kind
		if k == tokenizer.Terminator || k == tokenizer.Semicolon || k == tokenizer.EOF {
			break
		}
		dur := p.parseDuration()
		if dur.Node.Type != ast.DurationLiteral {
			break
		}
		expr.Params = append(expr.Params, dur)
	}
	return expr
}

func (p *parserImpl) parseStop() ast.ExpressionNode {
	expr := createExpression(ast.StopExpression)
	cache := make([]ast.Literal, 0)
	p.next()
	for p.current < p.length {
		k := p.cur().Kind
		if k == tokenizer.Terminator || k == tokenizer.Semicolon || k == tokenizer.EOF {
			break
		}
		switch k {
		case tokenizer.Float, tokenizer.Integer:
			cache = append(cache, p.parseLiteral())
			if p.current >= p.length || p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon || p.cur().Kind == tokenizer.EOF {
				expr.Params = append(expr.Params, cache...)
				cache = nil
			}
		case tokenizer.TribleDot, tokenizer.Range:
			r := createRange()
			if len(cache) >= 1 {
				r.Begin = cache[len(cache)-1]
				cache = cache[:len(cache)-1]
			}
			p.next() // consume …/...
			if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
				r.End = p.parseLiteral()
			}
			if r.Begin.Node.Type != 0 && r.End.Node.Type != 0 && r.Begin.Unit != r.End.Unit {
				fmt.Printf("Inconsistent units: %s -> %s\n", r.Begin.Unit, r.End.Unit)
				continue
			}
			expr.Range = &r
		case tokenizer.Negative:
			if tokenizer.TokenIsNumber(p.peek(1)) {
				cache = append(cache, p.parseLiteral())
				if p.current >= p.length || p.cur().Kind == tokenizer.Terminator || p.cur().Kind == tokenizer.Semicolon || p.cur().Kind == tokenizer.EOF {
					expr.Params = append(expr.Params, cache...)
					cache = nil
				}
			} else {
				if len(cache) != 0 {
					expr.Params = append(expr.Params, cache...)
					cache = nil
				}
				p.next()
			}
		default:
			if len(cache) != 0 {
				expr.Params = append(expr.Params, cache...)
				cache = nil
			}
			p.next()
		}
	}
	if len(cache) != 0 {
		expr.Params = append(expr.Params, cache...)
	}
	return expr
}

func (p *parserImpl) parseSelect() ast.ExpressionNode {
	expr := createExpression(ast.SelectExpression)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	} else {
		fmt.Println("Expected identifier after 'select'")
	}
	return expr
}

func (p *parserImpl) parseValue() ast.ExpressionNode {
	expr := createExpression(ast.ValueStatement)
	p.next()
	if p.cur().Kind == tokenizer.Colon {
		p.next()
	}
	if p.cur().Kind == tokenizer.Text {
		expr.Value = createLiteral(p.cur())
		p.next()
	} else if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
		if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
			expr.Value = p.parseLiteral()
		}
	} else if tokenizer.TokenIsNumber(p.cur()) || p.cur().Kind == tokenizer.Negative {
		expr.Value = p.parseLiteral()
	}
	return expr
}

func (p *parserImpl) parseDefine() ast.ExpressionNode {
	expr := createExpression(ast.DefineStatement)
	name := p.cur()
	p.next()
	if p.cur().Kind != tokenizer.Colon {
		return expr
	}
	p.next()
	if p.cur().Kind == tokenizer.Identifier || tokenizer.IsLiteral(p.cur()) {
		expr.Name = name.Value
		if tokenizer.IsLiteral(p.cur()) {
			expr.Value = createLiteral(p.cur())
		} else {
			expr.Value = ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value}
		}
		p.next()
	}
	return expr
}

func (p *parserImpl) parseReference() ast.ExpressionNode {
	expr := createExpression(ast.ReferenceStatement)
	p.next()
	if p.cur().Kind == tokenizer.Identifier {
		expr.Name = p.cur().Value
		p.next()
	} else {
		fmt.Println("Invalid reference, expected identifier after @")
	}
	return expr
}

func (p *parserImpl) parseVariable() ast.ExpressionNode {
	expr := createExpression(ast.VariableExpression)
	p.next() // consume $
	if p.cur().Kind == tokenizer.Identifier || p.cur().Kind.IsKeyword() {
		expr.Name = p.cur().Value
		p.next()
	} else {
		fmt.Println("Invalid variable, expected identifier after $")
	}
	return expr
}

func (p *parserImpl) parseAnnotationComment() ast.ExpressionNode {
	expr := createExpression(ast.AnnotationExpression)
	parts := strings.SplitN(p.cur().Value, "|", 2)
	if len(parts) == 2 {
		expr.Name = parts[0]
		expr.Value = ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: parts[1]}
	} else {
		expr.Value = ast.Literal{Node: ast.Node{Type: ast.TextLiteral}, Value: p.cur().Value}
	}
	p.next()
	return expr
}

// Parse 主解析入口
func (p *parserImpl) Parse() ast.RootNode {
	root := createRoot()
	for p.current < p.length {
		p.skipTerminators()
		tok := p.cur()
		if tok.Kind == tokenizer.EOF {
			break
		}
		var expr ast.ExpressionNode
		switch tok.Kind {
		case tokenizer.KeywordBuy:
			expr = p.parseBuy()
		case tokenizer.KeywordSell:
			expr = p.parseSell()
		case tokenizer.KeywordKeep:
			expr = p.parseKeep()
		case tokenizer.KeywordStop:
			expr = p.parseStop()
		case tokenizer.KeywordSelect:
			expr = p.parseSelect()
		case tokenizer.KeywordValue:
			expr = p.parseValue()
		case tokenizer.KeywordGrid:
			expr = p.parseGrid()
		case tokenizer.KeywordLong:
			expr = p.parseLong()
		case tokenizer.KeywordShort:
			expr = p.parseShort()
		case tokenizer.KeywordBoth:
			expr = p.parseBoth()
		case tokenizer.KeywordPortfolio:
			expr = p.parsePortfolio()
		case tokenizer.KeywordTemplate:
			expr = p.parseTemplate()
		case tokenizer.KeywordUse:
			expr = p.parseUse()
		case tokenizer.KeywordImport:
			expr = p.parseImport()
		case tokenizer.KeywordExport:
			expr = p.parseExport()
		case tokenizer.KeywordMacro:
			expr = p.parseMacro()
		case tokenizer.KeywordIf:
			expr = p.parseConditional()
		case tokenizer.KeywordAssert:
			expr = p.parseAssert()
		case tokenizer.KeywordHoldMax:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordCooldown:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordSession:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordActive:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordPause:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordPartialFill:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordCompoundProfit:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordSkipIfGapped:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordFallback:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordPosition:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordSizing:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordBasePosition:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordPyramidStep:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordMaxPyramidLayers:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordFixedFraction:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordVolatilityTarget:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordAtrPeriod:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordRiskPerTrade:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordRiskPerGrid:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordMaxDrawdown:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordPositionDecay:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordGrossExposure:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordNetExposure:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordBetaNeutral:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordRebalance:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordLeverage:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordMargin:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordHedge:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordFundingPriority:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordMaxShort:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordBorrowRateLimit:
			expr = p.parseGlobalConfig()
		case tokenizer.KeywordMaxPosition:
			expr = p.parseRiskConfig()
		case tokenizer.KeywordStopLoss:
			expr = p.parseRiskConfig()
		case tokenizer.KeywordSlippageTolerance:
			expr = p.parseRiskConfig()
		case tokenizer.KeywordCircuitBreaker:
			expr = p.parseRiskConfig()
		case tokenizer.KeywordSellShort:
			expr = p.parseActionStmt()
		case tokenizer.KeywordBuyCover:
			expr = p.parseActionStmt()
		case tokenizer.AnnotationComment:
			expr = p.parseAnnotationComment()
		case tokenizer.Identifier:
			if p.peek(1).Kind == tokenizer.Colon {
				expr = p.parseDefine()
			} else {
				expr = createExpression(ast.ReferenceStatement)
				expr.Name = tok.Value
				p.next()
			}
		case tokenizer.At:
			expr = p.parseReference()
		case tokenizer.Dollar:
			expr = p.parseVariable()
		case tokenizer.Float, tokenizer.Integer:
			expr = createExpression(ast.LiteralExpression)
			expr.Params = append(expr.Params, p.parseLiteral())
		case tokenizer.Colon:
			p.next()
			if tokenizer.IsLiteral(p.cur()) {
				expr = createExpression(ast.DefineStatement)
				expr.Value = createLiteral(p.cur())
				p.next()
			}
		case tokenizer.Negative:
			if !tokenizer.TokenIsNumber(p.peek(1)) {
				fmt.Println("Unexpected token after '-'")
				p.next()
				continue
			}
			p.next()
			lit := p.parseLiteral()
			lit.Value = "-" + lit.Value
			expr = createExpression(ast.LiteralExpression)
			expr.Params = append(expr.Params, lit)
		default:
			fmt.Printf("Unknown token kind: %v, value: %s\n", tok.Kind, tok.Value)
			p.next()
			continue
		}
		root.Expression = append(root.Expression, expr)
	}
	return root
}

// Parser 包级入口函数
func Parser(tokens []tokenizer.Token) ast.RootNode {
	return NewParser(tokens).Parse()
}
