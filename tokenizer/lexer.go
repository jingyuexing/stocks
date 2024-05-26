package tokenizer

import "unicode"

func isValidChar(ch rune) bool {
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ch >= 0xff
}

func isNumberic(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isSpace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func Lexer(text string) []Token {
	runes := []rune(text)
	MAXLENGTH := len(runes)
	current := 0
	tokens := make([]Token, 0)
	nextChar := func() {
		current++
		if current > MAXLENGTH {
			current = MAXLENGTH
		}
	}
	for current < MAXLENGTH {
		ch := runes[current]

		if isNumberic(ch) {
			val := ""
			dot := 0
			for (isNumberic(ch) || ch == '.') && current < MAXLENGTH {
				if ch == '.' {
					dot++
				}
				val += string(ch)
				nextChar()
				if current < MAXLENGTH {
					ch = runes[current]
				}
			}
			switch dot {
			case 0:
				tokens = append(tokens, createToken(Integer, val))
			case 1:
				tokens = append(tokens, createToken(Float, val))
			default:
				tokens = append(tokens, createToken(Text, val))
			}
			continue
		} else if ch == '%' {
			tokens = append(tokens,
				createToken(Per, string(ch)),
			)
			nextChar()
			continue
		} else if isValidChar(ch) {
			val := ""
			for isValidChar(ch) && current < MAXLENGTH {
				val += string(ch)

				nextChar()
				if current < MAXLENGTH {
					ch = runes[current]
				}
			}
			tokens = append(tokens, createIdentifier(val))
			continue
		} else if ch == '@' {
			tokens = append(tokens, createToken(At, string(ch)))
			nextChar()
			continue
		} else if ch == '-' {
			tokens = append(tokens, createToken(Negative, string(ch)))
			nextChar()
			continue
		} else if ch == ':' {
			tokens = append(tokens, createToken(Colon, string(ch)))
			nextChar()
			ch = runes[current]
			value := ""
			for current < MAXLENGTH && ch != '\n' {
				value += string(ch)
				nextChar()
				if current < MAXLENGTH {
					ch = runes[current]
				}
			}
			tokens = append(tokens, createToken(Text, value))
			nextChar()
			continue
		} else if ch == '.' {
			if current+2 < MAXLENGTH && string(runes[current:current+3]) == "..." {
				tokens = append(tokens, createToken(TribleDot, "..."))
				current += 3
			} else {
				current++
			}
			continue
		} else if isSpace(ch) {
			for isSpace(ch) && current < MAXLENGTH {
				if ch == '\n' {
					tokens = append(tokens, createToken(Terminator, ";"))
				}
				nextChar()
				if current < MAXLENGTH {
					ch = runes[current]
				}
			}
			continue
		} else {
			current++
		}
	}
	return tokens
}
