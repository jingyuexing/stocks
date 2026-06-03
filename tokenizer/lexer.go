package tokenizer

import (
	"strings"
	"unicode"
)

func isValidChar(ch rune) bool {
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ch >= 0xff || ch == '_'
}

func isNumberic(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isSpace(ch rune) bool {
	return unicode.IsSpace(ch)
}

func isOperatorChar(ch rune) bool {
	return ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '&' || ch == '|'
}

// cleanCommentContent 去除块注释中每行开头的 * 和空白字符
func cleanCommentContent(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
		if strings.HasPrefix(lines[i], "*") {
			lines[i] = strings.TrimSpace(lines[i][1:])
		}
	}
	return strings.Join(lines, "\n")
}

// parseAnnotationComment 解析块注释内容，如果是 /* @name: value */ 格式则返回 token，否则返回零值
func parseAnnotationComment(raw string) (string, string, bool) {
	cleaned := cleanCommentContent(raw)
	cleaned = strings.TrimSpace(cleaned)

	if !strings.HasPrefix(cleaned, "@") {
		return "", "", false
	}
	// 去掉前缀 @
	body := cleaned[1:]
	// 查找第一个冒号分隔符
	colonIdx := strings.Index(body, ":")
	if colonIdx < 0 {
		return "", "", false
	}
	name := strings.TrimSpace(body[:colonIdx])
	value := strings.TrimSpace(body[colonIdx+1:])
	return name, value, name != ""
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

	peekChar := func(offset int) rune {
		idx := current + offset
		if idx >= 0 && idx < MAXLENGTH {
			return runes[idx]
		}
		return '\x00'
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
			tokens = append(tokens, createToken(Per, string(ch)))
			nextChar()
			continue
		} else if ch == '+' {
			tokens = append(tokens, createToken(Plus, string(ch)))
			nextChar()
			continue
		} else if ch == '*' {
			// 注意：块注释内部的 * 已经在块注释逻辑中处理，
			// 这里遇到的 * 只有在不是块注释结束符时才作为独立 token
			if peekChar(1) == '/' {
				// 块注释结束符，不应该单独出现在这里
				nextChar()
				nextChar()
				continue
			}
			tokens = append(tokens, createToken(Star, string(ch)))
			nextChar()
			continue
		} else if ch == ',' {
			tokens = append(tokens, createToken(Comma, string(ch)))
			nextChar()
			continue
		} else if ch == ';' {
			tokens = append(tokens, createToken(Semicolon, string(ch)))
			nextChar()
			continue
		} else if ch == '(' {
			tokens = append(tokens, createToken(LeftParen, string(ch)))
			nextChar()
			continue
		} else if ch == ')' {
			tokens = append(tokens, createToken(RightParen, string(ch)))
			nextChar()
			continue
		} else if ch == '[' {
			tokens = append(tokens, createToken(LeftBracket, string(ch)))
			nextChar()
			continue
		} else if ch == ']' {
			tokens = append(tokens, createToken(RightBracket, string(ch)))
			nextChar()
			continue
		} else if ch == '@' {
			tokens = append(tokens, createToken(At, string(ch)))
			nextChar()
			continue
		} else if ch == '$' {
			tokens = append(tokens, createToken(Dollar, string(ch)))
			nextChar()
			continue
		} else if ch == '{' {
			tokens = append(tokens, createToken(LeftBraces, string(ch)))
			nextChar()
			continue
		} else if ch == '}' {
			tokens = append(tokens, createToken(RightBraces, string(ch)))
			nextChar()
			continue
		} else if ch == '-' {
			tokens = append(tokens, createToken(Negative, string(ch)))
			nextChar()
			continue
		} else if ch == ':' {
			tokens = append(tokens, createToken(Colon, string(ch)))
			nextChar()
			continue
		} else if ch == '=' {
			if peekChar(1) == '=' {
				tokens = append(tokens, createToken(EqualEqual, "=="))
				current += 2
			} else {
				tokens = append(tokens, createToken(Assign, string(ch)))
				nextChar()
			}
			continue
		} else if ch == '!' {
			if peekChar(1) == '=' {
				tokens = append(tokens, createToken(NotEqual, "!="))
				current += 2
			} else {
				// 单独的 ! 暂不支持，跳过
				nextChar()
			}
			continue
		} else if ch == '<' {
			if peekChar(1) == '=' {
				tokens = append(tokens, createToken(LessEqual, "<="))
				current += 2
			} else {
				tokens = append(tokens, createToken(LessThan, string(ch)))
				nextChar()
			}
			continue
		} else if ch == '>' {
			if peekChar(1) == '=' {
				tokens = append(tokens, createToken(GreaterEqual, ">="))
				current += 2
			} else {
				tokens = append(tokens, createToken(GreaterThan, string(ch)))
				nextChar()
			}
			continue
		} else if ch == '&' {
			if peekChar(1) == '&' {
				tokens = append(tokens, createToken(And, "&&"))
				current += 2
			} else {
				nextChar()
			}
			continue
		} else if ch == '|' {
			if peekChar(1) == '|' {
				tokens = append(tokens, createToken(Or, "||"))
				current += 2
			} else {
				nextChar()
			}
			continue
		} else if ch == '"' {
			// 字符串解析
			val := ""
			nextChar() // consume opening quote
			for current < MAXLENGTH && runes[current] != '"' {
				val += string(runes[current])
				nextChar()
			}
			if current < MAXLENGTH && runes[current] == '"' {
				nextChar() // consume closing quote
			}
			tokens = append(tokens, createToken(String, val))
			continue
		} else if ch == '/' {
			if peekChar(1) == '/' {
				// 行注释，跳过到行尾
				for current < MAXLENGTH && runes[current] != '\n' {
					nextChar()
				}
				continue
			} else if peekChar(1) == '*' {
				// 块注释 /* ... */
				nextChar() // consume /
				nextChar() // consume *
				start := current
				for current < MAXLENGTH {
					if runes[current] == '*' && peekChar(1) == '/' {
						nextChar()
						nextChar()
						break
					}
					nextChar()
				}
				raw := string(runes[start : current-2])
				if name, value, ok := parseAnnotationComment(raw); ok {
					// token value 存储为 "name|value"，由 parser 拆分
					tokens = append(tokens, createToken(AnnotationComment, name+"|"+value))
				}
				continue
			} else {
				tokens = append(tokens, createToken(Slash, string(ch)))
				nextChar()
				continue
			}
		} else if ch == '\u2026' {
			// 水平省略号 …
			tokens = append(tokens, createToken(Range, string(ch)))
			nextChar()
			continue
		} else if ch == '.' {
			if current+2 < MAXLENGTH && string(runes[current:current+3]) == "..." {
				tokens = append(tokens, createToken(Range, "..."))
				current += 3
			} else {
				tokens = append(tokens, createToken(Dot, string(ch)))
				nextChar()
			}
			continue
		} else if isValidChar(ch) {
			val := ""
			for (isValidChar(ch) || isNumberic(ch) || ch == '-') && current < MAXLENGTH {
				val += string(ch)
				nextChar()
				if current < MAXLENGTH {
					ch = runes[current]
				}
			}
			tokens = append(tokens, createIdentifier(val))
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
	tokens = append(tokens, createToken(EOF, ""))
	return tokens
}
