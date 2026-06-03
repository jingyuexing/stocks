package engine_test

import (
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/engine"
	"github.com/jingyuexing/stocks/transformer"
)

func TestExecutor_Setup(t *testing.T) {
	ctx := transformer.NewStockContext()
	ex := engine.NewExecutor(ctx, ast.RootNode{})
	ex.Setup()
	// Setup 不应 panic
}

func TestExecutor_BuildAction(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetCode("AAPL")
	ex := engine.NewExecutor(ctx, ast.RootNode{})

	buyExpr := ast.ExpressionNode{
		Node:   ast.Node{Type: ast.BuyExpression},
		Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "100", Unit: ""}},
	}
	act := ex.BuildAction(engine.ActionBuy, buyExpr)
	if act.Type != engine.ActionBuy {
		t.Errorf("expected action buy, got %v", act.Type)
	}
	if act.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", act.Symbol)
	}
	if act.Amount != 100 {
		t.Errorf("expected amount 100, got %f", act.Amount)
	}

	sellExpr := ast.ExpressionNode{
		Node:   ast.Node{Type: ast.SellExpression},
		Params: []ast.Literal{{Node: ast.Node{Type: ast.PercentLiteral}, Value: "50", Unit: "%"}},
	}
	act2 := ex.BuildAction(engine.ActionSell, sellExpr)
	if !act2.IsPercent {
		t.Error("expected percent action")
	}
	if act2.Amount != 50 {
		t.Errorf("expected amount 50%%, got %f", act2.Amount)
	}
}

func TestExecutor_ExecuteBody(t *testing.T) {
	ctx := transformer.NewStockContext()
	sellTriggered := false
	ctx.Emiter.On("sell", func(args ...any) {
		sellTriggered = true
	})

	body := []ast.ExpressionNode{
		{Node: ast.Node{Type: ast.SellExpression}, Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "100"}}},
	}
	ex := engine.NewExecutor(ctx, ast.RootNode{})
	ex.ExecuteBody(body)

	if !sellTriggered {
		t.Error("expected sell event to be triggered")
	}
}

func TestExecutor_GridCheckWithLevelBlock(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(88) // profit = -12%

	buyTriggered := false
	ctx.Emiter.On("buy", func(args ...any) {
		buyTriggered = true
	})

	astRoot := ast.RootNode{
		Expression: []ast.ExpressionNode{
			{
				Node: ast.Node{Type: ast.GridExpression},
				Range: &ast.RangeExpressionNode{
					Node:  ast.Node{Type: ast.RangeExpression},
					Begin: ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "-15"},
					End:   ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "-10"},
				},
				Body: []ast.ExpressionNode{
					{Node: ast.Node{Type: ast.BuyExpression}, Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "10"}}},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("grid_check", ctx, astRoot.Expression[0])

	if !buyTriggered {
		t.Error("expected buy to trigger when profit -12% hits level block -15%...-10%")
	}
}

func TestExecutor_StopTriggeredLoss(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(85) // profit = -15%

	sellTriggered := false
	ctx.Emiter.On("sell", func(args ...any) {
		sellTriggered = true
	})

	astRoot := ast.RootNode{
		Expression: []ast.ExpressionNode{
			{
				Node: ast.Node{Type: ast.GridExpression},
				Body: []ast.ExpressionNode{
					{
						Node:   ast.Node{Type: ast.LossExpression},
						Params: []ast.Literal{{Node: ast.Node{Type: ast.FloatLiteral}, Value: "-10"}},
						Body: []ast.ExpressionNode{
							{Node: ast.Node{Type: ast.SellExpression}, Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "100"}}},
						},
					},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)

	if !sellTriggered {
		t.Error("expected sell to trigger when loss -15% exceeds threshold -10%")
	}
}

func TestExecutor_StopTriggeredProfit(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(115) // profit = 15%

	sellTriggered := false
	ctx.Emiter.On("sell", func(args ...any) {
		sellTriggered = true
	})

	astRoot := ast.RootNode{
		Expression: []ast.ExpressionNode{
			{
				Node:   ast.Node{Type: ast.ProfitExpression},
				Params: []ast.Literal{{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"}},
				Body: []ast.ExpressionNode{
					{Node: ast.Node{Type: ast.SellExpression}, Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "50"}}},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)

	if !sellTriggered {
		t.Error("expected sell to trigger when profit 15% exceeds threshold 10%")
	}
}

func TestExecutor_LossRangeNotHit(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(95) // profit = -5%

	sellTriggered := false
	ctx.Emiter.On("sell", func(args ...any) {
		sellTriggered = true
	})

	astRoot := ast.RootNode{
		Expression: []ast.ExpressionNode{
			{
				Node:   ast.Node{Type: ast.LossExpression},
				Params: []ast.Literal{{Node: ast.Node{Type: ast.FloatLiteral}, Value: "-10"}},
				Body: []ast.ExpressionNode{
					{Node: ast.Node{Type: ast.SellExpression}, Params: []ast.Literal{{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "100"}}},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)

	if sellTriggered {
		t.Error("expected sell NOT to trigger when loss -5% does not exceed threshold -10%")
	}
}

func TestExecutor_KeepExpired(t *testing.T) {
	ctx := transformer.NewStockContext()
	astRoot := ast.RootNode{}
	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	// keep_expired 不应 panic
	ctx.Emiter.Emit("keep_expired", ctx)
}
