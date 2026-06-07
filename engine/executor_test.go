package engine_test

import (
	"testing"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/engine"
	"github.com/jingyuexing/stocks/transformer"
)

func TestExecutor_Setup(t *testing.T) {
	ctx := transformer.NewStockContext()
	ex := engine.NewExecutor(ctx, &ast.ProgramNode{})
	ex.Setup()
	// Setup 不应 panic
}

func TestExecutor_BuildAction(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetCode("AAPL")
	ex := engine.NewExecutor(ctx, &ast.ProgramNode{})

	buyExpr := &ast.ActionStmtNode{
		Action: "buy",
		Args:   []ast.Expr{&ast.Literal{Value: "100"}},
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

	sellExpr := &ast.ActionStmtNode{
		Action: "sell",
		Args:   []ast.Expr{&ast.Literal{Value: "50", Unit: "%"}},
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

	body := []ast.Stmt{
		&ast.ActionStmtNode{
			Action: "sell",
			Args:   []ast.Expr{&ast.Literal{Value: "100"}},
		},
	}
	ex := engine.NewExecutor(ctx, &ast.ProgramNode{})
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

	astRoot := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.StrategyStmtNode{
				Kind: "grid",
				Body: []ast.Stmt{
					&ast.FlowStmtNode{
						Kind: "level",
						Condition: &ast.RangeExprNode{
							Begin: &ast.Literal{Value: "-15"},
							End:   &ast.Literal{Value: "-10"},
						},
						Body: []ast.Stmt{
							&ast.ActionStmtNode{
								Action: "buy",
								Args:   []ast.Expr{&ast.Literal{Value: "10"}},
							},
						},
					},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("grid_check", ctx)
	time.Sleep(100 * time.Millisecond)

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

	astRoot := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.StrategyStmtNode{
				Kind: "grid",
				Body: []ast.Stmt{
					&ast.FlowStmtNode{
						Kind:      "level",
						Mode:      "loss",
						Condition: &ast.Literal{Value: "-10"},
						Body: []ast.Stmt{
							&ast.ActionStmtNode{
								Action: "sell",
								Args:   []ast.Expr{&ast.Literal{Value: "100"}},
							},
						},
					},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)
	time.Sleep(100 * time.Millisecond)

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

	astRoot := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.FlowStmtNode{
				Kind:      "level",
				Mode:      "profit",
				Condition: &ast.Literal{Value: "10"},
				Body: []ast.Stmt{
					&ast.ActionStmtNode{
						Action: "sell",
						Args:   []ast.Expr{&ast.Literal{Value: "50"}},
					},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)
	time.Sleep(100 * time.Millisecond)

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

	astRoot := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.FlowStmtNode{
				Kind:      "level",
				Mode:      "loss",
				Condition: &ast.Literal{Value: "-10"},
				Body: []ast.Stmt{
					&ast.ActionStmtNode{
						Action: "sell",
						Args:   []ast.Expr{&ast.Literal{Value: "100"}},
					},
				},
			},
		},
	}

	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	ctx.Emiter.Emit("stop_triggered", ctx)
	time.Sleep(100 * time.Millisecond)

	if sellTriggered {
		t.Error("expected sell NOT to trigger when loss -5% does not exceed threshold -10%")
	}
}

func TestExecutor_KeepExpired(t *testing.T) {
	ctx := transformer.NewStockContext()
	astRoot := &ast.ProgramNode{}
	ex := engine.NewExecutor(ctx, astRoot)
	ex.Setup()
	// keep_expired 不应 panic
	ctx.Emiter.Emit("keep_expired", ctx)
}
