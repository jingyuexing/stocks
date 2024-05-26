package transformer

import (
	"strconv"
	"time"

	"github.com/jingyuexing/go-utils"
	"github.com/jingyuexing/stocks/ast"
)

var stocksContext = func() *StockContext {
	return &StockContext{
		start:  time.Now(),
		Emiter: utils.NewEventEmit(),
	}
}()

func duration(params []ast.Literal) utils.DateTime {
	date := utils.NewDateTime()
	for _, pram := range params {
		if pram.Type == ast.DurationLiteral {
			val, _ := strconv.Atoi(pram.Value)
			date = date.Add(val, utils.AddUnits(pram.Unit))
		}
	}
	return date
}

func Transformer(root ast.RootNode) *StockContext {
	keep := utils.Filter(root.Expression, func(exp ast.ExpressionNode) bool {
		return exp.Type == ast.KeepExpression
	})
	if len(keep) != 0 {
		stocksContext.SetDuration(duration(keep[0].Params))

	}
	for _, exp := range root.Expression {
		switch exp.Type {
		case ast.StopExpression:
			stocksContext.Emiter.On("stop", func(args ...any) {
				currentValue, ok := args[0].(StockContext)
				if !ok {
					return
				}
				isKeep,_:= currentValue.UseKeep()
				if isKeep {
					// [datetime] 当前 止盈/止损 于 {currentPrice} 总计 盈利 /亏损 {cha}, 总盈利率/ 亏损率 为 {}%
					
				}
			})
		case ast.BuyExpression:
			stocksContext.Buy()
		case ast.SellExpression:
			//
			stocksContext.Sell()
		case ast.KeepExpression:
			stocksContext.Emiter.Emit("stop", stocksContext)
		default:
			continue
		}
	}
	return stocksContext
}
