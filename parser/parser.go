package parser

import (
	"fmt"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
)

func createExpression(kind ast.NodeType) ast.ExpressionNode {
	return ast.ExpressionNode{
		Node: ast.Node{
			Type: kind,
		},
		Params: make([]ast.Literal, 0),
		Name:   "",
		Range:  nil,
	}
}

func createRange() ast.RangeExpressionNode {
	return ast.RangeExpressionNode{
		Node: ast.Node{
			Type: ast.RangeExpression,
		},
	}
}

func createRoot() ast.RootNode {
	return ast.RootNode{
		Node: ast.Node{
			Type: ast.Program,
		},
		Expression: make([]ast.ExpressionNode, 0),
	}
}

func createLiteral(token tokenizer.Token) ast.Literal {
	literal := ast.Literal{}
	literal.Value = token.Value
	switch token.Kind {
	case tokenizer.Float:
		literal.Type = ast.FloatLiteral
	case tokenizer.Integer:
		literal.Type = ast.IntegerLiteral
	default:
		literal.Type = ast.TextLiteral
	}
	return literal
}

func Parser(tokens []tokenizer.Token) ast.RootNode {

	current := 0
	MAXLENGTH := len(tokens)
	root := createRoot()
	next := func() {
		current++
		if current >= MAXLENGTH {
			current = MAXLENGTH // 确保 current 不会超出边界
		}
	}
	getCurrentToken := func() tokenizer.Token {
		token := tokenizer.Token{}
		if current < MAXLENGTH {
			token = tokens[current]
		}
		return token
	}
	parseDefined := func() ast.ExpressionNode {
		definedExpression := createExpression(ast.DefineStatement)
		identifier := getCurrentToken()
		next()

		if getCurrentToken().Kind != tokenizer.Colon {
			return definedExpression
		}
		next()

		if !tokenizer.IsLiteral(getCurrentToken()) {
			return definedExpression
		}
		definedExpression.Name = identifier.Value
		definedExpression.Value = createLiteral(getCurrentToken())
		next()
		return definedExpression
	}
	parseDuration := func() ast.Literal {
		duration := ast.Literal{}
		if !tokenizer.TokenIsNumber(getCurrentToken()) {
			return duration
		}
		nextToken := tokens[current+1]
		if !tokenizer.IsTimeUnitToken(nextToken) {
			fmt.Printf("the unit %s not in (Y, M, W, d, h, m, s, ms)", nextToken.Value)
			return duration
		}
		duration = createLiteral(getCurrentToken())
		duration.Type = ast.DurationLiteral
		duration.Unit = nextToken.Value
		next()
		next()
		return duration
	}
	parseLiteral := func() ast.Literal {
		literal := ast.Literal{}
		currentToken := getCurrentToken()
		var nextToken tokenizer.Token

		if !tokenizer.TokenIsNumber(currentToken) {
			return literal
		}
		literal = createLiteral(currentToken)

		if current < (MAXLENGTH - 1) {
			nextToken = tokens[current+1]

			if tokenizer.IsTimeUnitToken(nextToken) {
				literal.Type = ast.DurationLiteral
				literal.Unit = nextToken.Value
				next()
			} else if nextToken.Kind == tokenizer.Per {
				literal.Type = ast.PercentLiteral
				literal.Unit = nextToken.Value
				next()
			}
		}

		next()
		return literal
	}

	for current < MAXLENGTH {
		currentToken := getCurrentToken()
		cache := make([]tokenizer.Token, 0)
		var expression ast.ExpressionNode
		if currentToken.Kind.Numberic() <= 4 {
			switch currentToken.Kind {
			case tokenizer.KeywordBuy:
				expression = createExpression(ast.BuyExpression)
				expression.Name = currentToken.Value
				next()
				currentToken = getCurrentToken()
				if !tokenizer.TokenIsNumber(currentToken) {
					print("unepect token")
					continue
				}
				expression.Params = append(expression.Params, parseLiteral())
				next()
			case tokenizer.KeywordKeep:
				expression = createExpression(ast.KeepExpression)
				next()
				for current < MAXLENGTH && getCurrentToken().Kind != tokenizer.Terminator {
					duration := parseDuration()
					if duration.Type != ast.DurationLiteral {
						continue
					}
					expression.Params = append(expression.Params, duration)
				}
			case tokenizer.KeywordSell:
				expression = createExpression(ast.SellExpression)
				next()
				if !tokenizer.TokenIsNumber(getCurrentToken()) {
					fmt.Println("Unexpected token")
					continue
				}
				expression.Params = append(expression.Params, parseLiteral())
			case tokenizer.KeywordStop:
				expression = createExpression(ast.StopExpression)
				literalCache := make([]ast.Literal, 0)
				next()
				currentToken = getCurrentToken()
				var rangeExp ast.RangeExpressionNode
				for currentToken.Kind != tokenizer.Terminator && current < MAXLENGTH {
					switch currentToken.Kind {
					case tokenizer.Float, tokenizer.Integer:
						literalCache = append(literalCache, parseLiteral())
						if !(current < MAXLENGTH) {
							expression.Params = append(expression.Params, literalCache...)
						}
						continue
					case tokenizer.TribleDot:
						if len(literalCache) < 1 {
							next()
							continue
						}
						rangeExp = createRange()
						rangeExp.Begin = literalCache[len(literalCache)-1]
						next()
						rangeExp.End = parseLiteral()
						if rangeExp.Begin.Unit != rangeExp.End.Unit {
							fmt.Printf("Inconsistent units: %s -> %s", rangeExp.Begin.Unit, rangeExp.End.Unit)
							next()
							continue
						}
						expression.Range = &rangeExp
						literalCache = nil // Clear the cache after using it
					default:
						if len(literalCache) != 0 {
							expression.Params = append(expression.Params, literalCache...)
						}
						next()
						continue
					}
				}
			}
			root.Expression = append(root.Expression, expression)
			next()
		}

		switch currentToken.Kind {
		case tokenizer.Identifier:
			cache = append(cache, currentToken)
			next()
		case tokenizer.Float, tokenizer.Integer:
			numberic := createLiteral(currentToken)
			next()

			if current >= MAXLENGTH {
				fmt.Println("Unexpected end of tokens after number")
				break
			}
			currentToken = tokens[current]
			if tokenizer.IsTimeUnitToken(currentToken) {
				if numberic.Type != ast.IntegerLiteral {
					print("日期数不支持小数")
					continue
				}
				current++
			}
			if currentToken.Kind == tokenizer.Per {
				numberic.Unit = currentToken.Value
				current++
			}
			expression.Params = append(expression.Params, numberic)
		case tokenizer.At:
			expression = createExpression(ast.ReferenceStatement)
			next()
			if getCurrentToken().Kind == tokenizer.Identifier {
				expression.Name = getCurrentToken().Value
			} else {
				fmt.Println("Invalid reference, expected identifier after @")
				next()
				continue
			}
		case tokenizer.Colon:
			definedExpression := parseDefined()
			root.Expression = append(root.Expression, definedExpression)
			continue
		case tokenizer.Negative:
			nextToken := tokens[current+1]
			if !tokenizer.TokenIsNumber(nextToken) {
				fmt.Println("Unexpected token")
				next()
				break
			}
			literal := createLiteral(currentToken)
			literal.Unit = nextToken.Value
			expression.Params = append(expression.Params, literal)
		case tokenizer.Terminator:
			next()
			continue
		}
		// root.Expression = make([]ast.ExpressionNode, expression)
		current++
	}
	return root
}
