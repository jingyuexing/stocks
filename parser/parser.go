package parser

import (
	"fmt"
	"strconv"
	"time"

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
		Body:   make([]ast.ExpressionNode, 0),
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
		literal.Node.Type = ast.FloatLiteral
	case tokenizer.Integer:
		literal.Node.Type = ast.IntegerLiteral
	default:
		literal.Node.Type = ast.TextLiteral
	}
	return literal
}

func durationToNumber(literal ast.Literal) time.Duration {
	if literal.Node.Type != ast.DurationLiteral {
		return 0
	}
	var d time.Duration = 0
	switch literal.Unit {
	case "ms":
		d = time.Millisecond
	case "s":
		d = time.Second
	case "m":
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
	val, _ := strconv.ParseInt(literal.Value, 10, 64)
	return time.Duration(val * int64(d))
}

// Parser 将 token 序列解析为 AST
func Parser(tokens []tokenizer.Token) ast.RootNode {
	current := 0
	MAXLENGTH := len(tokens)
	root := createRoot()

	next := func() {
		current++
		if current >= MAXLENGTH {
			current = MAXLENGTH - 1
		}
	}

	peek := func(offset int) tokenizer.Token {
		idx := current + offset
		if idx >= 0 && idx < MAXLENGTH {
			return tokens[idx]
		}
		return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
	}

	getCurrentToken := func() tokenizer.Token {
		if current < MAXLENGTH {
			return tokens[current]
		}
		return tokenizer.Token{Kind: tokenizer.EOF, Value: ""}
	}

	// 跳过终结符和空白
	skipTerminators := func() {
		for current < MAXLENGTH && getCurrentToken().Kind == tokenizer.Terminator {
			next()
		}
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
				literal.Node.Type = ast.DurationLiteral
				literal.Unit = nextToken.Value
				next()
			} else if nextToken.Kind == tokenizer.Per {
				literal.Node.Type = ast.PercentLiteral
				literal.Unit = nextToken.Value
				next()
			}
		}

		next()
		return literal
	}

	parseDuration := func() ast.Literal {
		duration := ast.Literal{}
		if !tokenizer.TokenIsNumber(getCurrentToken()) {
			return duration
		}
		if current+1 >= MAXLENGTH {
			return duration
		}
		nextToken := tokens[current+1]
		if !tokenizer.IsTimeUnitToken(nextToken) {
			fmt.Printf("the unit %s not in (Y, M, W, d, h, m, s, ms)\n", nextToken.Value)
			return duration
		}
		duration = createLiteral(getCurrentToken())
		duration.Node.Type = ast.DurationLiteral
		duration.Unit = nextToken.Value
		next()
		next()
		return duration
	}

	parseDefine := func() ast.ExpressionNode {
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

	parseReference := func() ast.ExpressionNode {
		expr := createExpression(ast.ReferenceStatement)
		next() // consume @
		if getCurrentToken().Kind == tokenizer.Identifier {
			expr.Name = getCurrentToken().Value
			next()
		} else {
			fmt.Println("Invalid reference, expected identifier after @")
		}
		return expr
	}

	parseBuy := func() ast.ExpressionNode {
		expr := createExpression(ast.BuyExpression)
		next() // consume buy
		currentToken := getCurrentToken()
		if currentToken.Kind == tokenizer.At {
			next()
			if getCurrentToken().Kind == tokenizer.Identifier {
				ref := ast.Literal{Node: ast.Node{Type: ast.ReferenceLiteral}, Value: getCurrentToken().Value}
				expr.Params = append(expr.Params, ref)
				next()
			} else {
				fmt.Println("Unexpected token after @ in buy expression")
			}
		} else if tokenizer.TokenIsNumber(currentToken) {
			expr.Params = append(expr.Params, parseLiteral())
		} else {
			fmt.Println("Unexpected token in buy expression")
		}
		return expr
	}

	parseSell := func() ast.ExpressionNode {
		expr := createExpression(ast.SellExpression)
		next() // consume sell

		// 支持 sell AUL +30% 的语法：可选的标识符和加号
		if getCurrentToken().Kind == tokenizer.Identifier {
			expr.Name = getCurrentToken().Value
			next()
		}
		if getCurrentToken().Kind == tokenizer.Plus {
			next()
		}
		if tokenizer.TokenIsNumber(getCurrentToken()) {
			expr.Params = append(expr.Params, parseLiteral())
		} else {
			fmt.Println("Unexpected token in sell expression")
		}
		return expr
	}

	parseKeep := func() ast.ExpressionNode {
		expr := createExpression(ast.KeepExpression)
		next() // consume keep
		for current < MAXLENGTH && getCurrentToken().Kind != tokenizer.Terminator && getCurrentToken().Kind != tokenizer.EOF {
			duration := parseDuration()
			if duration.Node.Type != ast.DurationLiteral {
				break
			}
			expr.Params = append(expr.Params, duration)
		}
		return expr
	}

	parseStop := func() ast.ExpressionNode {
		expr := createExpression(ast.StopExpression)
		literalCache := make([]ast.Literal, 0)
		next() // consume stop
		for current < MAXLENGTH && getCurrentToken().Kind != tokenizer.Terminator && getCurrentToken().Kind != tokenizer.EOF {
			currentToken := getCurrentToken()
			switch currentToken.Kind {
			case tokenizer.Float, tokenizer.Integer:
				literalCache = append(literalCache, parseLiteral())
				if current >= MAXLENGTH || getCurrentToken().Kind == tokenizer.Terminator || getCurrentToken().Kind == tokenizer.EOF {
					expr.Params = append(expr.Params, literalCache...)
					literalCache = nil
				}
			case tokenizer.TribleDot:
				if len(literalCache) < 1 {
					next()
					continue
				}
				rangeExp := createRange()
				rangeExp.Begin = literalCache[len(literalCache)-1]
				literalCache = literalCache[:len(literalCache)-1]
				next() // consume ...
				rangeExp.End = parseLiteral()
				if rangeExp.Begin.Unit != rangeExp.End.Unit {
					fmt.Printf("Inconsistent units: %s -> %s\n", rangeExp.Begin.Unit, rangeExp.End.Unit)
					continue
				}
				expr.Range = &rangeExp
			default:
				if len(literalCache) != 0 {
					expr.Params = append(expr.Params, literalCache...)
					literalCache = nil
				}
				next()
			}
		}
		if len(literalCache) != 0 {
			expr.Params = append(expr.Params, literalCache...)
		}
		return expr
	}

	parseSelect := func() ast.ExpressionNode {
		expr := createExpression(ast.SelectExpression)
		next() // consume select
		if getCurrentToken().Kind == tokenizer.Identifier {
			expr.Name = getCurrentToken().Value
			next()
		} else {
			fmt.Println("Expected identifier after 'select'")
		}
		return expr
	}

	parseValue := func() ast.ExpressionNode {
		expr := createExpression(ast.ValueStatement)
		next() // consume value
		if getCurrentToken().Kind == tokenizer.Colon {
			next() // consume :
		}
		// 支持三种形式：
		// 1. value: ork 1000  -> Lexer 会将 " ork 1000" 作为 Text
		// 2. value ork 1000  -> 普通 token 序列
		// 3. value 1000       -> 纯数字
		if getCurrentToken().Kind == tokenizer.Text {
			expr.Value = createLiteral(getCurrentToken())
			next()
		} else if getCurrentToken().Kind == tokenizer.Identifier {
			expr.Name = getCurrentToken().Value
			next()
			if tokenizer.TokenIsNumber(getCurrentToken()) {
				expr.Value = parseLiteral()
			}
		} else if tokenizer.TokenIsNumber(getCurrentToken()) {
			expr.Value = parseLiteral()
		}
		return expr
	}

	var parseGridBody func() []ast.ExpressionNode
	var parseLoss func() ast.ExpressionNode
	var parseProfit func() ast.ExpressionNode

	parseGridBody = func() []ast.ExpressionNode {
		body := make([]ast.ExpressionNode, 0)
		for current < MAXLENGTH && getCurrentToken().Kind != tokenizer.RightBraces && getCurrentToken().Kind != tokenizer.EOF {
			skipTerminators()
			// skipTerminators 之后可能遇到 RightBraces，需要重新检查
			if getCurrentToken().Kind == tokenizer.RightBraces || getCurrentToken().Kind == tokenizer.EOF {
				break
			}
			currentToken := getCurrentToken()
			switch currentToken.Kind {
			case tokenizer.KeywordLoss:
				body = append(body, parseLoss())
			case tokenizer.KeywordProfit:
				body = append(body, parseProfit())
			case tokenizer.KeywordSell:
				body = append(body, parseSell())
			case tokenizer.KeywordBuy:
				body = append(body, parseBuy())
			case tokenizer.Terminator:
				next()
			default:
				fmt.Printf("Unexpected token in grid body: %v\n", currentToken)
				next()
			}
		}
		return body
	}

	parseLoss = func() ast.ExpressionNode {
		expr := createExpression(ast.LossExpression)
		next() // consume loss
		if tokenizer.TokenIsNumber(getCurrentToken()) {
			expr.Params = append(expr.Params, parseLiteral())
		}
		if getCurrentToken().Kind == tokenizer.LeftBraces {
			next()
			expr.Body = parseGridBody()
			if getCurrentToken().Kind == tokenizer.RightBraces {
				next()
			}
		}
		return expr
	}

	parseProfit = func() ast.ExpressionNode {
		expr := createExpression(ast.ProfitExpression)
		next() // consume profit
		if tokenizer.TokenIsNumber(getCurrentToken()) {
			expr.Params = append(expr.Params, parseLiteral())
		}
		if getCurrentToken().Kind == tokenizer.LeftBraces {
			next()
			expr.Body = parseGridBody()
			if getCurrentToken().Kind == tokenizer.RightBraces {
				next()
			}
		}
		return expr
	}

	parseGrid := func() ast.ExpressionNode {
		expr := createExpression(ast.GridExpression)
		next() // consume grid
		if getCurrentToken().Kind == tokenizer.LeftBraces {
			next()
			expr.Body = parseGridBody()
			if getCurrentToken().Kind == tokenizer.RightBraces {
				next()
			}
		} else {
			fmt.Println("Expected '{' after 'grid'")
		}
		return expr
	}

	// 主循环
	for current < MAXLENGTH {
		skipTerminators()
		currentToken := getCurrentToken()

		if currentToken.Kind == tokenizer.EOF {
			break
		}

		var expression ast.ExpressionNode

		switch currentToken.Kind {
		case tokenizer.KeywordBuy:
			expression = parseBuy()
		case tokenizer.KeywordSell:
			expression = parseSell()
		case tokenizer.KeywordKeep:
			expression = parseKeep()
		case tokenizer.KeywordStop:
			expression = parseStop()
		case tokenizer.KeywordSelect:
			expression = parseSelect()
		case tokenizer.KeywordValue:
			expression = parseValue()
		case tokenizer.KeywordGrid:
			expression = parseGrid()
		case tokenizer.Identifier:
			// identifier : value 视为定义
			if peek(1).Kind == tokenizer.Colon {
				expression = parseDefine()
			} else {
				// 单独标识符视为引用或忽略
				expression = createExpression(ast.ReferenceStatement)
				expression.Name = currentToken.Value
				next()
			}
		case tokenizer.At:
			expression = parseReference()
		case tokenizer.Float, tokenizer.Integer:
			expression = createExpression(ast.LiteralExpression)
			expression.Params = append(expression.Params, parseLiteral())
		case tokenizer.Colon:
			// 孤立的 colon，尝试解析定义（无标识符前缀）
			next()
			if tokenizer.IsLiteral(getCurrentToken()) {
				expression = createExpression(ast.DefineStatement)
				expression.Value = createLiteral(getCurrentToken())
				next()
			}
		case tokenizer.Negative:
			nextToken := peek(1)
			if !tokenizer.TokenIsNumber(nextToken) {
				fmt.Println("Unexpected token after '-'")
				next()
				continue
			}
			// 消费 '-'，将后续数字解析为负数 literal
			next() // skip '-'
			literal := parseLiteral()
			literal.Value = "-" + literal.Value
			expression = createExpression(ast.LiteralExpression)
			expression.Params = append(expression.Params, literal)
		default:
			fmt.Printf("Unknown token kind: %v, value: %s\n", currentToken.Kind, currentToken.Value)
			next()
			continue
		}

		root.Expression = append(root.Expression, expression)
	}

	return root
}
