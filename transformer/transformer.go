package transformer

import (
	"strconv"

	"github.com/jingyuexing/go-utils"
	"github.com/jingyuexing/stocks/ast"
)

// NewStockContextFromAST 根据 AST 创建新的执行上下文
func NewStockContextFromAST() *StockContext {
	ctx := NewStockContext()
	return ctx
}

func duration(params []ast.Literal) utils.DateTime {
	date := utils.NewDateTime()
	for _, param := range params {
		if param.Node.Type == ast.DurationLiteral {
			val, _ := strconv.Atoi(param.Value)
			date = date.Add(val, utils.AddUnits(param.Unit))
		}
	}
	return date
}

func Transformer(root ast.RootNode) *StockContext {
	stocksContext := NewStockContextFromAST()

	keep := utils.Filter(root.Expression, func(exp ast.ExpressionNode) bool {
		return exp.Type == ast.KeepExpression
	})
	if len(keep) != 0 {
		stocksContext.SetDuration(duration(keep[0].Params))
	}

	for _, exp := range root.Expression {
		switch exp.Type {
		case ast.SelectExpression:
			stocksContext.SetCode(exp.Name)
		case ast.ValueStatement:
			if exp.Value.Value != "" {
				val, _ := strconv.ParseFloat(exp.Value.Value, 64)
				stocksContext.SetAmount(val)
			}
		case ast.StopExpression:
			stocksContext.Emiter.On("stop", func(args ...any) {
				currentValue, ok := args[0].(*StockContext)
				if !ok {
					return
				}
				isKeep, _ := currentValue.UseKeep()
				if isKeep {
					// [datetime] 当前 止盈/止损 于 {currentPrice} 总计 盈利 /亏损 {cha}, 总盈利率/ 亏损率 为 {}%

				}
			})
		case ast.BuyExpression:
			stocksContext.Buy()
		case ast.SellExpression:
			stocksContext.Sell()
		case ast.KeepExpression:
			stocksContext.Emiter.Emit("stop", stocksContext)
		default:
			continue
		}
	}
	return stocksContext
}
