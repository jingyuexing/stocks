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
	default:
		return 0
	}
}

func (t TokenKind) IsKeyword() bool {
	switch t {
	case KeywordBuy, KeywordSell, KeywordStop, KeywordKeep,
		KeywordSelect, KeywordValue, KeywordGrid, KeywordLoss, KeywordProfit:
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
		case "Y", "M", "W", "d", "h", "H", "m", "s", "ms":
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
	case Float, Integer, Text:
		return true
	default:
		return false
	}
}
