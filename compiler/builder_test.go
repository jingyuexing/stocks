package compiler

import (
	"strings"
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
)

func TestBuilder(t *testing.T) {
	result := BuildContext(
		parser.Parse(tokenizer.Lexer("keep 1W")),
	)
	if result == nil {
		t.Fatal("expected non-nil StockContext")
	}
	_, err := result.UseKeep()
	if err != nil {
		t.Errorf("UseKeep failed: %v", err)
	}
}

func TestBuilderSelectAndValue(t *testing.T) {
	input := `
select AUL
ork: 1000
keep 1W
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetCode() != "AUL" {
		t.Errorf("expected code AUL, got %s", ctx.GetCode())
	}
	if ctx.GetAmount() != 0 {
		lit, ok := ctx.Vars.UserVars["ork"]
		if !ok || lit.Value != "1000" {
			t.Errorf("expected variable ork=1000, got %v", lit)
		}
	}
}

func TestBuilder_StopEventBinding(t *testing.T) {
	input := "select TEST\nkeep 1W"
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext")
	}
	if ctx.GetCode() != "TEST" {
		t.Errorf("expected code TEST, got %s", ctx.GetCode())
	}

	stopReceived := false
	ctx.Emiter.On("stop", func(args ...any) {
		stopReceived = true
	})
	ctx.Emiter.Emit("stop", ctx)
	if !stopReceived {
		t.Error("stop event was not received")
	}
}

func TestBuilder_EmptyAST(t *testing.T) {
	root := &ast.ProgramNode{}
	ctx := BuildContext(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext for empty AST")
	}
}

func TestBuilder_GridExpression(t *testing.T) {
	input := "grid AAPL { loss 10% { sell 100% } profit 10% { sell 50% } }"
	root := parser.Parse(tokenizer.Lexer(input))
	if len(root.Statements) == 0 {
		t.Fatal("expected AST to contain grid expression")
	}
	if _, ok := root.Statements[0].(*ast.StrategyStmtNode); !ok {
		t.Errorf("expected StrategyStmtNode, got %T", root.Statements[0])
	}
	ctx := BuildContext(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext for grid AST")
	}
}

// ========== v2.1 新增测试 ==========

func TestBuilder_LeverageAndMargin(t *testing.T) {
	input := `
select BTC
leverage 5x
margin isolated
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetCode() != "BTC" {
		t.Errorf("expected code BTC, got %s", ctx.GetCode())
	}
	if ctx.GetLeverage() != 5 {
		t.Errorf("expected leverage 5, got %f", ctx.GetLeverage())
	}
	if ctx.GetMarginMode() != "isolated" {
		t.Errorf("expected margin isolated, got %s", ctx.GetMarginMode())
	}
}

func TestBuilder_PositionAndSizing(t *testing.T) {
	input := `
position dynamic 1000 capital
sizing equal
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetPositionType() != "dynamic" {
		t.Errorf("expected position type dynamic, got %s", ctx.GetPositionType())
	}
	if ctx.GetSizingMethod() != "equal" {
		t.Errorf("expected sizing equal, got %s", ctx.GetSizingMethod())
	}
}

func TestBuilder_SessionAndRanges(t *testing.T) {
	input := `
session 09:30 16:00 EST
active 2024-01-01 2024-12-31
pause 2024-02-01 2024-02-07
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	session := ctx.GetSession()
	if len(session) != 3 || session[0] != "09:30" || session[1] != "16:00" || session[2] != "EST" {
		t.Errorf("unexpected session values: %v", session)
	}

	active := ctx.GetActiveRange()
	if len(active) != 2 || active[0] != "2024-01-01" || active[1] != "2024-12-31" {
		t.Errorf("unexpected active range: %v", active)
	}

	pause := ctx.GetPauseRange()
	if len(pause) != 2 || pause[0] != "2024-02-01" || pause[1] != "2024-02-07" {
		t.Errorf("unexpected pause range: %v", pause)
	}
}

func TestBuilder_LongAndShort(t *testing.T) {
	input := `
long AAPL {
    leverage 2x
}
short TSLA {
    leverage 3x
}
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetDirection() != "short" {
		t.Errorf("expected direction short, got %s", ctx.GetDirection())
	}
	if ctx.GetCode() != "TSLA" {
		t.Errorf("expected code TSLA, got %s", ctx.GetCode())
	}
	if ctx.GetLeverage() != 3 {
		t.Errorf("expected leverage 3, got %f", ctx.GetLeverage())
	}
}

func TestBuilder_GridWithConfig(t *testing.T) {
	input := `
grid AAPL {
    session 09:30…16:00 EST
    position dynamic 1000 capital
    sizing equal
    max_position 5000 shares
    stop_loss -10%
    leverage 2x
}
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetCode() != "AAPL" {
		t.Errorf("expected code AAPL, got %s", ctx.GetCode())
	}
	if ctx.GetSizingMethod() != "equal" {
		t.Errorf("expected sizing equal, got %s", ctx.GetSizingMethod())
	}
	if ctx.GetLeverage() != 2 {
		t.Errorf("expected leverage 2, got %f", ctx.GetLeverage())
	}
	if ctx.GetPositionType() != "dynamic" {
		t.Errorf("expected position dynamic, got %s", ctx.GetPositionType())
	}
	if ctx.GetMaxPosition() != 5000 {
		t.Errorf("expected max_position 5000, got %f", ctx.GetMaxPosition())
	}
	if ctx.GetStopLossRange() == nil {
		t.Error("expected stop_loss range to be set")
	}

	session := ctx.GetSession()
	if len(session) < 3 {
		t.Fatalf("expected at least 3 session params, got %d", len(session))
	}
}

func TestBuilder_RiskAndExecutionConfig(t *testing.T) {
	input := `
compound_profit true
skip_if_gapped
fallback reduce
slippage_tolerance 0.5%
circuit_breaker 5% in 1h
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if !ctx.GetCompoundProfit() {
		t.Error("expected compound_profit true")
	}
	if !ctx.GetSkipIfGapped() {
		t.Error("expected skip_if_gapped true")
	}
	if ctx.GetFallback() != "reduce" {
		t.Errorf("expected fallback reduce, got %s", ctx.GetFallback())
	}
	if ctx.GetSlippageTolerance() != 0.5 {
		t.Errorf("expected slippage_tolerance 0.5, got %f", ctx.GetSlippageTolerance())
	}
	if ctx.GetCircuitBreaker() != 5 {
		t.Errorf("expected circuit_breaker 5, got %f", ctx.GetCircuitBreaker())
	}
}

func TestBuilder_AdapterComments(t *testing.T) {
	input := `
/* @adapter: binance */
/* @adapter_mode: futures */
/* @adapter_config: { api_key: "123" } */
/* @adapter_switch: { condition: "down" } */
`
	root := parser.Parse(tokenizer.Lexer(input))
	ctx := BuildContext(root)

	if ctx.GetAdapter() != "binance" {
		t.Errorf("expected adapter binance, got %s", ctx.GetAdapter())
	}
	if ctx.GetAdapterMode() != "futures" {
		t.Errorf("expected adapter_mode futures, got %s", ctx.GetAdapterMode())
	}
	if !strings.Contains(ctx.GetAdapterConfig(), "api_key") {
		t.Errorf("expected adapter_config to contain api_key, got %s", ctx.GetAdapterConfig())
	}
	if !strings.Contains(ctx.GetAdapterSwitch(), "down") {
		t.Errorf("expected adapter_switch to contain down, got %s", ctx.GetAdapterSwitch())
	}
}
