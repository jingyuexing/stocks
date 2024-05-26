package tokenizer

type TokenKint string

func (t TokenKint) Numberic() uint {
	switch t {
	case KeywordSell:
		return 0x1
	case KeywordBuy:
		return 0x2
	case KeywordStop:
		return 0x3
	case KeywordKeep:
		return 0x4
	case Integer:
		return 0x5
	case Float:
		return 0x6
	case Per:
		return 0x7
	case Text:
		return 0x8
	case Identifier:
		return 0x9
	case Terminator:
		return 0xa
	case Colon:
		return 0xb
	case At:
		return 0xc
	case TribleDot:
		return 0xd
	case Negative:
		return 0xe
	case Time:
		return 0xf
	default:
		return 0
	}
}

const (
	KeywordSell TokenKint = "sell"
	KeywordBuy  TokenKint = "buy"
	KeywordStop TokenKint = "stop"
	KeywordKeep TokenKint = "keep"
	Integer     TokenKint = "integer"
	Text        TokenKint = "text"
	Float       TokenKint = "float"
	Per         TokenKint = "percent"
	Terminator  TokenKint = "terminator"
	Colon       TokenKint = "colon"
	Identifier  TokenKint = "identifier"
	Negative    TokenKint = "negative"
	At          TokenKint = "at"
	TribleDot   TokenKint = "Tribledot"
	Time        TokenKint = "Time"
)

type Token struct {
	Kind  TokenKint
	Value string
}

func createToken(kind TokenKint, value string) Token {
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
