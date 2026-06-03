package tokenizer

type TokenKind string

func (t TokenKind) Numberic() uint {
	switch t {
	case KeywordSell:
		return 0x1
	case KeywordBuy:
		return 0x2
	case KeywordStop:
		return 0x3
	case KeywordKeep:
		return 0x4
	case KeywordSelect:
		return 0x5
	case KeywordValue:
		return 0x6
	case KeywordGrid:
		return 0x7
	case KeywordLoss:
		return 0x8
	case KeywordProfit:
		return 0x9
	case Integer:
		return 0xa
	case Float:
		return 0xb
	case Per:
		return 0xc
	case Text:
		return 0xd
	case Identifier:
		return 0xe
	case Terminator:
		return 0xf
	case Colon:
		return 0x10
	case At:
		return 0x11
	case TribleDot:
		return 0x12
	case Negative:
		return 0x13
	case Time:
		return 0x14
	case LeftBraces:
		return 0x15
	case RightBraces:
		return 0x16
	case Plus:
		return 0x17
	case EOF:
		return 0x18
	case KeywordLong:
		return 0x19
	case KeywordShort:
		return 0x1a
	case KeywordBoth:
		return 0x1b
	case KeywordPortfolio:
		return 0x1c
	case KeywordTemplate:
		return 0x1d
	case KeywordUse:
		return 0x1e
	case KeywordExtends:
		return 0x1f
	case KeywordAs:
		return 0x20
	case KeywordImport:
		return 0x21
	case KeywordExport:
		return 0x22
	case KeywordMacro:
		return 0x23
	case KeywordCross:
		return 0x24
	case KeywordOverride:
		return 0x25
	case KeywordPriority:
		return 0x26
	case KeywordHoldMax:
		return 0x27
	case KeywordCooldown:
		return 0x28
	case KeywordSession:
		return 0x29
	case KeywordActive:
		return 0x2a
	case KeywordPause:
		return 0x2b
	case KeywordSellShort:
		return 0x2c
	case KeywordBuyCover:
		return 0x2d
	case KeywordPosition:
		return 0x2e
	case KeywordSizing:
		return 0x2f
	case KeywordBasePosition:
		return 0x30
	case KeywordPyramidStep:
		return 0x31
	case KeywordMaxPyramidLayers:
		return 0x32
	case KeywordFixedFraction:
		return 0x33
	case KeywordVolatilityTarget:
		return 0x34
	case KeywordAtrPeriod:
		return 0x35
	case KeywordRiskPerTrade:
		return 0x36
	case KeywordRiskPerGrid:
		return 0x37
	case KeywordMaxDrawdown:
		return 0x38
	case KeywordPositionDecay:
		return 0x39
	case KeywordGrossExposure:
		return 0x3a
	case KeywordNetExposure:
		return 0x3b
	case KeywordBetaNeutral:
		return 0x3c
	case KeywordRebalance:
		return 0x3d
	case KeywordLeverage:
		return 0x3e
	case KeywordMargin:
		return 0x3f
	case KeywordHedge:
		return 0x40
	case KeywordFundingPriority:
		return 0x41
	case KeywordMaxShort:
		return 0x42
	case KeywordBorrowRateLimit:
		return 0x43
	case KeywordMaxPosition:
		return 0x44
	case KeywordStopLoss:
		return 0x45
	case KeywordSlippageTolerance:
		return 0x46
	case KeywordPartialFill:
		return 0x47
	case KeywordCircuitBreaker:
		return 0x48
	case KeywordCompoundProfit:
		return 0x49
	case KeywordSkipIfGapped:
		return 0x4a
	case KeywordFallback:
		return 0x4b
	case KeywordParam:
		return 0x4c
	case KeywordIf:
		return 0x4d
	case KeywordAssert:
		return 0x4e
	case KeywordTrue:
		return 0x4f
	case KeywordFalse:
		return 0x50
	case String:
		return 0x51
	case Comma:
		return 0x52
	case Semicolon:
		return 0x53
	case LeftParen:
		return 0x54
	case RightParen:
		return 0x55
	case LeftBracket:
		return 0x56
	case RightBracket:
		return 0x57
	case Assign:
		return 0x58
	case Star:
		return 0x59
	case Slash:
		return 0x5a
	case LessThan:
		return 0x5b
	case GreaterThan:
		return 0x5c
	case EqualEqual:
		return 0x5d
	case NotEqual:
		return 0x5e
	case LessEqual:
		return 0x5f
	case GreaterEqual:
		return 0x60
	case And:
		return 0x61
	case Or:
		return 0x62
	case Range:
		return 0x63
	case Dot:
		return 0x64
	case Dollar:
		return 0x65
	default:
		return 0
	}
}

func (t TokenKind) IsKeyword() bool {
	switch t {
	case KeywordBuy, KeywordSell, KeywordStop, KeywordKeep,
		KeywordSelect, KeywordValue, KeywordGrid, KeywordLoss, KeywordProfit,
		KeywordLong, KeywordShort, KeywordBoth, KeywordPortfolio,
		KeywordTemplate, KeywordUse, KeywordExtends, KeywordAs, KeywordImport, KeywordExport, KeywordMacro,
		KeywordCross, KeywordOverride, KeywordPriority,
		KeywordHoldMax, KeywordCooldown, KeywordSession, KeywordActive, KeywordPause,
		KeywordSellShort, KeywordBuyCover,
		KeywordPosition, KeywordSizing, KeywordBasePosition, KeywordPyramidStep, KeywordMaxPyramidLayers,
		KeywordFixedFraction, KeywordVolatilityTarget, KeywordAtrPeriod, KeywordRiskPerTrade, KeywordRiskPerGrid,
		KeywordMaxDrawdown, KeywordPositionDecay, KeywordGrossExposure, KeywordNetExposure, KeywordBetaNeutral, KeywordRebalance,
		KeywordLeverage, KeywordMargin, KeywordHedge, KeywordFundingPriority, KeywordMaxShort, KeywordBorrowRateLimit,
		KeywordMaxPosition, KeywordStopLoss, KeywordSlippageTolerance, KeywordPartialFill, KeywordCircuitBreaker,
		KeywordCompoundProfit, KeywordSkipIfGapped, KeywordFallback,
		KeywordParam, KeywordIf, KeywordAssert,
		KeywordTrue, KeywordFalse:
		return true
	default:
		return false
	}
}

const (
	KeywordSell   TokenKind = "sell"
	KeywordBuy    TokenKind = "buy"
	KeywordStop   TokenKind = "stop"
	KeywordKeep   TokenKind = "keep"
	KeywordLoss   TokenKind = "loss"
	KeywordProfit TokenKind = "profit"
	KeywordGrid   TokenKind = "grid"
	KeywordSelect TokenKind = "select"
	KeywordValue  TokenKind = "value"
	Integer       TokenKind = "integer"
	Text          TokenKind = "text"
	Float         TokenKind = "float"
	Per           TokenKind = "percent"
	Terminator    TokenKind = "terminator"
	Colon         TokenKind = "colon"
	Identifier    TokenKind = "identifier"
	Negative      TokenKind = "negative"
	At            TokenKind = "at"
	TribleDot     TokenKind = "Tribledot"
	Time          TokenKind = "Time"
	LeftBraces    TokenKind = "LeftBraces"
	RightBraces   TokenKind = "RightBraces"
	Plus          TokenKind = "Plus"
	EOF           TokenKind = "EOF"

	// v2.1 策略声明
	KeywordLong      TokenKind = "long"
	KeywordShort     TokenKind = "short"
	KeywordBoth      TokenKind = "both"
	KeywordPortfolio TokenKind = "portfolio"
	KeywordTemplate  TokenKind = "template"
	KeywordUse       TokenKind = "use"
	KeywordExtends   TokenKind = "extends"
	KeywordAs        TokenKind = "as"
	KeywordImport    TokenKind = "import"
	KeywordExport    TokenKind = "export"
	KeywordMacro     TokenKind = "macro"

	// v2.1 范围与优先级
	KeywordCross    TokenKind = "cross"
	KeywordOverride TokenKind = "override"
	KeywordPriority TokenKind = "priority"

	// v2.1 时间控制
	KeywordHoldMax  TokenKind = "hold_max"
	KeywordCooldown TokenKind = "cooldown"
	KeywordSession  TokenKind = "session"
	KeywordActive   TokenKind = "active"
	KeywordPause    TokenKind = "pause"

	// v2.1 交易动作
	KeywordSellShort TokenKind = "sell_short"
	KeywordBuyCover  TokenKind = "buy_cover"

	// v2.1 头寸管理
	KeywordPosition         TokenKind = "position"
	KeywordSizing           TokenKind = "sizing"
	KeywordBasePosition     TokenKind = "base_position"
	KeywordPyramidStep      TokenKind = "pyramid_step"
	KeywordMaxPyramidLayers TokenKind = "max_pyramid_layers"
	KeywordFixedFraction    TokenKind = "fixed_fraction"
	KeywordVolatilityTarget TokenKind = "volatility_target"
	KeywordAtrPeriod        TokenKind = "atr_period"
	KeywordRiskPerTrade     TokenKind = "risk_per_trade"
	KeywordRiskPerGrid      TokenKind = "risk_per_grid"
	KeywordMaxDrawdown      TokenKind = "max_drawdown"
	KeywordPositionDecay    TokenKind = "position_decay"
	KeywordGrossExposure    TokenKind = "gross_exposure"
	KeywordNetExposure      TokenKind = "net_exposure"
	KeywordBetaNeutral      TokenKind = "beta_neutral"
	KeywordRebalance        TokenKind = "rebalance"

	// v2.1 杠杆与保证金
	KeywordLeverage        TokenKind = "leverage"
	KeywordMargin          TokenKind = "margin"
	KeywordHedge           TokenKind = "hedge"
	KeywordFundingPriority TokenKind = "funding_priority"
	KeywordMaxShort        TokenKind = "max_short"
	KeywordBorrowRateLimit TokenKind = "borrow_rate_limit"

	// v2.1 风控
	KeywordMaxPosition       TokenKind = "max_position"
	KeywordStopLoss          TokenKind = "stop_loss"
	KeywordSlippageTolerance TokenKind = "slippage_tolerance"
	KeywordPartialFill       TokenKind = "partial_fill"
	KeywordCircuitBreaker    TokenKind = "circuit_breaker"

	// v2.1 执行偏好
	KeywordCompoundProfit TokenKind = "compound_profit"
	KeywordSkipIfGapped   TokenKind = "skip_if_gapped"
	KeywordFallback       TokenKind = "fallback"

	// v2.1 策略复用与编译
	KeywordParam  TokenKind = "param"
	KeywordIf     TokenKind = "if"
	KeywordAssert TokenKind = "assert"

	// v2.1 字面量
	KeywordTrue  TokenKind = "true"
	KeywordFalse TokenKind = "false"

	// v2.1 符号
	String       TokenKind = "string"
	Comma        TokenKind = "comma"
	Semicolon    TokenKind = "semicolon"
	LeftParen    TokenKind = "left_paren"
	RightParen   TokenKind = "right_paren"
	LeftBracket  TokenKind = "left_bracket"
	RightBracket TokenKind = "right_bracket"
	Assign       TokenKind = "assign"
	Star         TokenKind = "star"
	Slash        TokenKind = "slash"
	LessThan     TokenKind = "less_than"
	GreaterThan  TokenKind = "greater_than"
	EqualEqual   TokenKind = "equal_equal"
	NotEqual     TokenKind = "not_equal"
	LessEqual    TokenKind = "less_equal"
	GreaterEqual TokenKind = "greater_equal"
	And          TokenKind = "and"
	Or           TokenKind = "or"
	Range        TokenKind = "range"
	Dot          TokenKind = "dot"
	Dollar       TokenKind = "dollar"

	// 通用文档注释 (/* @name: value */)
	AnnotationComment TokenKind = "annotation_comment"
)

type Token struct {
	Kind  TokenKind
	Value string
}

func createToken(kind TokenKind, value string) Token {
	return Token{
		Kind:  kind,
		Value: value,
	}
}

func createIdentifier(val string) Token {
	var token Token
	switch val {
	case "buy":
		token = createToken(KeywordBuy, val)
	case "sell":
		token = createToken(KeywordSell, val)
	case "stop":
		token = createToken(KeywordStop, val)
	case "keep":
		token = createToken(KeywordKeep, val)
	case "grid":
		token = createToken(KeywordGrid, val)
	case "loss":
		token = createToken(KeywordLoss, val)
	case "profit":
		token = createToken(KeywordProfit, val)
	case "select":
		token = createToken(KeywordSelect, val)
	case "value":
		token = createToken(KeywordValue, val)
	case "long":
		token = createToken(KeywordLong, val)
	case "short":
		token = createToken(KeywordShort, val)
	case "both":
		token = createToken(KeywordBoth, val)
	case "portfolio":
		token = createToken(KeywordPortfolio, val)
	case "template":
		token = createToken(KeywordTemplate, val)
	case "use":
		token = createToken(KeywordUse, val)
	case "extends":
		token = createToken(KeywordExtends, val)
	case "as":
		token = createToken(KeywordAs, val)
	case "import":
		token = createToken(KeywordImport, val)
	case "export":
		token = createToken(KeywordExport, val)
	case "macro":
		token = createToken(KeywordMacro, val)
	case "cross":
		token = createToken(KeywordCross, val)
	case "override":
		token = createToken(KeywordOverride, val)
	case "priority":
		token = createToken(KeywordPriority, val)
	case "hold_max":
		token = createToken(KeywordHoldMax, val)
	case "cooldown":
		token = createToken(KeywordCooldown, val)
	case "session":
		token = createToken(KeywordSession, val)
	case "active":
		token = createToken(KeywordActive, val)
	case "pause":
		token = createToken(KeywordPause, val)
	case "sell_short":
		token = createToken(KeywordSellShort, val)
	case "buy_cover":
		token = createToken(KeywordBuyCover, val)
	case "position":
		token = createToken(KeywordPosition, val)
	case "sizing":
		token = createToken(KeywordSizing, val)
	case "base_position":
		token = createToken(KeywordBasePosition, val)
	case "pyramid_step":
		token = createToken(KeywordPyramidStep, val)
	case "max_pyramid_layers":
		token = createToken(KeywordMaxPyramidLayers, val)
	case "fixed_fraction":
		token = createToken(KeywordFixedFraction, val)
	case "volatility_target":
		token = createToken(KeywordVolatilityTarget, val)
	case "atr_period":
		token = createToken(KeywordAtrPeriod, val)
	case "risk_per_trade":
		token = createToken(KeywordRiskPerTrade, val)
	case "risk_per_grid":
		token = createToken(KeywordRiskPerGrid, val)
	case "max_drawdown":
		token = createToken(KeywordMaxDrawdown, val)
	case "position_decay":
		token = createToken(KeywordPositionDecay, val)
	case "gross_exposure":
		token = createToken(KeywordGrossExposure, val)
	case "net_exposure":
		token = createToken(KeywordNetExposure, val)
	case "beta_neutral":
		token = createToken(KeywordBetaNeutral, val)
	case "rebalance":
		token = createToken(KeywordRebalance, val)
	case "leverage":
		token = createToken(KeywordLeverage, val)
	case "margin":
		token = createToken(KeywordMargin, val)
	case "hedge":
		token = createToken(KeywordHedge, val)
	case "funding_priority":
		token = createToken(KeywordFundingPriority, val)
	case "max_short":
		token = createToken(KeywordMaxShort, val)
	case "borrow_rate_limit":
		token = createToken(KeywordBorrowRateLimit, val)
	case "max_position":
		token = createToken(KeywordMaxPosition, val)
	case "stop_loss":
		token = createToken(KeywordStopLoss, val)
	case "slippage_tolerance":
		token = createToken(KeywordSlippageTolerance, val)
	case "partial_fill":
		token = createToken(KeywordPartialFill, val)
	case "circuit_breaker":
		token = createToken(KeywordCircuitBreaker, val)
	case "compound_profit":
		token = createToken(KeywordCompoundProfit, val)
	case "skip_if_gapped":
		token = createToken(KeywordSkipIfGapped, val)
	case "fallback":
		token = createToken(KeywordFallback, val)
	case "param":
		token = createToken(KeywordParam, val)
	case "if":
		token = createToken(KeywordIf, val)
	case "assert":
		token = createToken(KeywordAssert, val)
	case "true":
		token = createToken(KeywordTrue, val)
	case "false":
		token = createToken(KeywordFalse, val)
	default:
		token = createToken(Identifier, val)
	}
	return token
}

func TokenIsNumber(token Token) bool {
	switch token.Kind {
	case Integer, Float, Per:
		return true
	default:
		return false
	}
}

// 是否是时间单位
func IsTimeUnit(val string) bool {
	if val != "" {
		switch val {
		case "Y", "M", "W", "d", "h", "H", "m", "min", "s", "ms":
			// 年 月 周 天 时 分 秒 毫秒
			return true
		default:
			return false
		}
	}
	return false
}

// 是否是金钱单位token
func IsAmountUnit(val string) bool {
	if val != "" {
		switch val {
		case "K", "k", "M", "m", "B", "b":
			return true
		default:
			return false
		}
	}
	return false
}

// 是否是时间单位Token
func IsTimeUnitToken(token Token) bool {
	return IsTimeUnit(token.Value)
}

func IsLiteral(val Token) bool {
	switch val.Kind {
	case Float, Integer, Text, String:
		return true
	default:
		return false
	}
}

// IsSizingMethod 判断是否为头寸规模算法关键字
func IsSizingMethod(val string) bool {
	switch val {
	case "equal", "pyramid", "pyramid_inverse", "martingale", "anti_martingale",
		"kelly_half", "fixed_fractional", "fixed_ratio", "volatility_target":
		return true
	default:
		return false
	}
}

// IsMarginMode 判断是否为保证金模式
func IsMarginMode(val string) bool {
	switch val {
	case "isolated", "cross":
		return true
	default:
		return false
	}
}

// IsRebalancePeriod 判断是否为再平衡周期
func IsRebalancePeriod(val string) bool {
	switch val {
	case "daily", "weekly", "monthly":
		return true
	default:
		return false
	}
}

// IsTimezone 判断是否为常用时区标识
func IsTimezone(val string) bool {
	switch val {
	case "EST", "CST", "PST", "UTC":
		return true
	default:
		return false
	}
}
