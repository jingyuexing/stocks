package engine_test

import (
	"testing"

	"github.com/jingyuexing/stocks/engine"
	"github.com/jingyuexing/stocks/transformer"
)

func TestRiskGuard_CheckPass(t *testing.T) {
	ctx := transformer.NewStockContext()
	rg := engine.NewRiskGuard(ctx)

	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 100}
	if err := rg.Check(act); err != nil {
		t.Fatalf("expected check to pass, got: %v", err)
	}
}

func TestRiskGuard_CircuitBreakerLoss(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(90) // profit = -10%
	ctx.SetCircuitBreaker(5)

	rg := engine.NewRiskGuard(ctx)
	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 10}
	if err := rg.Check(act); err == nil {
		t.Fatal("expected circuit breaker to trigger on loss")
	}
}

func TestRiskGuard_CircuitBreakerProfit(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(115) // profit = 15%
	ctx.SetCircuitBreaker(10)

	rg := engine.NewRiskGuard(ctx)
	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 10}
	if err := rg.Check(act); err == nil {
		t.Fatal("expected circuit breaker to trigger on profit")
	}
}

func TestRiskGuard_MaxPosition(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetMaxPosition(500)

	rg := engine.NewRiskGuard(ctx)
	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 1000}
	if err := rg.Check(act); err == nil {
		t.Fatal("expected max position to block")
	}
}

func TestRiskGuard_MaxPositionUnderLimit(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetMaxPosition(500)

	rg := engine.NewRiskGuard(ctx)
	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 300}
	if err := rg.Check(act); err != nil {
		t.Fatalf("expected check to pass, got: %v", err)
	}
}

func TestRiskGuard_NilContext(t *testing.T) {
	rg := engine.NewRiskGuard(nil)
	act := engine.Action{Type: engine.ActionBuy, Symbol: "AAPL", Amount: 100}
	if err := rg.Check(act); err == nil {
		t.Fatal("expected error when context is nil")
	}
}

func TestRiskGuard_SellNotBlockedByMaxPosition(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetMaxPosition(500)

	rg := engine.NewRiskGuard(ctx)
	act := engine.Action{Type: engine.ActionSell, Symbol: "AAPL", Amount: 1000}
	if err := rg.Check(act); err != nil {
		t.Fatalf("expected sell not blocked by max_position, got: %v", err)
	}
}

func TestRiskGuard_SessionActive(t *testing.T) {
	ctx := transformer.NewStockContext()
	rg := engine.NewRiskGuard(ctx)
	if !rg.IsSessionActive() {
		t.Error("expected session active when no session configured")
	}
}

func TestRiskGuard_NotPaused(t *testing.T) {
	ctx := transformer.NewStockContext()
	rg := engine.NewRiskGuard(ctx)
	if rg.IsPaused() {
		t.Error("expected not paused when no pause range configured")
	}
}
