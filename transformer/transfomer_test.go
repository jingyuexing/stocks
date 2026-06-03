package transformer_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
	"github.com/jingyuexing/stocks/transformer"
)

func TestTransfomer(t *testing.T) {
	result := transformer.Transformer(
		parser.Parser(tokenizer.Lexer("keep 1W")),
	)
	if result == nil {
		t.Fatal("expected non-nil StockContext")
	}
	_, err := result.UseKeep()
	if err != nil {
		t.Errorf("UseKeep failed: %v", err)
	}
}

func TestTransformerSelectAndValue(t *testing.T) {
	input := `
select AUL
value ork 1000
keep 1W
`
	root := parser.Parser(tokenizer.Lexer(input))

	// 调试：确认 ValueStatement 是否被正确解析
	foundValue := false
	for _, expr := range root.Expression {
		if expr.Type == ast.ValueStatement {
			foundValue = true
			t.Logf("ValueStatement found: Name=%q Value.Value=%q Value.Type=%v", expr.Name, expr.Value.Value, expr.Value.Node.Type)
		}
	}
	if !foundValue {
		t.Fatal("ValueStatement not found in AST")
	}

	ctx := transformer.Transformer(root)

	if ctx.Code != "AUL" {
		t.Errorf("expected code AUL, got %s", ctx.Code)
	}
	if ctx.GetAmount() != 1000 {
		t.Errorf("expected amount 1000, got %f", ctx.GetAmount())
	}
}

func TestStockContext_GettersSetters(t *testing.T) {
	ctx := transformer.NewStockContext()

	// GetBeginTime
	beginTime := ctx.GetBeginTime()
	if beginTime.IsZero() {
		t.Error("expected non-zero begin time")
	}

	// SetBeginPrice / GetBeginPrice
	ctx.SetBeginPrice(123.45)
	if ctx.GetBeginPrice() != 123.45 {
		t.Errorf("expected begin price 123.45, got %f", ctx.GetBeginPrice())
	}

	// SetCurrentPrice / GetCurrentPrice
	ret := ctx.SetCurrentPrice(99.99)
	if ret != 99.99 {
		t.Errorf("expected SetCurrentPrice return 99.99, got %f", ret)
	}
	if ctx.GetCurrentPrice() != 99.99 {
		t.Errorf("expected current price 99.99, got %f", ctx.GetCurrentPrice())
	}

	// GetDuration (via Transformer with keep expression)
	root := parser.Parser(tokenizer.Lexer("keep 1W"))
	ctx2 := transformer.Transformer(root)
	_ = ctx2.GetDuration()
}

func TestStockContext_StopCallback(t *testing.T) {
	ctx := transformer.NewStockContext()

	// nil callback should result in error
	ctx.StopCallback(nil)
	ok, err := ctx.UseStop()
	if err == nil {
		t.Error("expected error when stop callback is nil")
	}
	if ok {
		t.Error("expected false when stop callback is nil")
	}

	// custom callback
	ctx.StopCallback(func() bool { return true })
	ok, err = ctx.UseStop()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true from custom stop callback")
	}
}

func TestStockContext_KeepCallback(t *testing.T) {
	ctx := transformer.NewStockContext()

	// default keep (no custom callback) should not deadlock and return without error
	ok, err := ctx.UseKeep()
	if err != nil {
		t.Errorf("unexpected error for default keep: %v", err)
	}
	// default implementation compares start.After(time.Now()), usually false
	_ = ok

	// custom callback
	ctx.KeepCallback(func(duration int64) bool { return true })
	ok, err = ctx.UseKeep()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true from custom keep callback")
	}
}

func TestStockContext_SellBuyEvents(t *testing.T) {
	ctx := transformer.NewStockContext()

	sellTriggered := false
	buyTriggered := false

	ctx.Emiter.On("sell", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.StockContext); ok {
				sellTriggered = true
			}
		}
	})

	ctx.Emiter.On("buy", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.StockContext); ok {
				buyTriggered = true
			}
		}
	})

	ctx.Sell()
	if !sellTriggered {
		t.Error("sell event was not triggered")
	}

	ctx.Buy()
	if !buyTriggered {
		t.Error("buy event was not triggered")
	}
}

func TestTransformer_StopEventBinding(t *testing.T) {
	input := "select TEST\nstop\nkeep 1W"
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext")
	}
	if ctx.Code != "TEST" {
		t.Errorf("expected code TEST, got %s", ctx.Code)
	}

	stopReceived := false
	ctx.Emiter.On("stop", func(args ...any) {
		stopReceived = true
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.StockContext); !ok {
				t.Error("expected *StockContext as first argument")
			}
		}
	})
	ctx.Emiter.Emit("stop", ctx)
	if !stopReceived {
		t.Error("stop event was not received")
	}
}

func TestTransformer_EmptyAST(t *testing.T) {
	root := ast.RootNode{}
	ctx := transformer.Transformer(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext for empty AST")
	}
}

func TestTransformer_GridExpression(t *testing.T) {
	input := "grid { loss 10% { sell 100% } profit 10% { sell 50% } }"
	root := parser.Parser(tokenizer.Lexer(input))
	if len(root.Expression) == 0 {
		t.Fatal("expected AST to contain grid expression")
	}
	if root.Expression[0].Type != ast.GridExpression {
		t.Errorf("expected GridExpression, got %v", root.Expression[0].Type)
	}
	ctx := transformer.Transformer(root)
	if ctx == nil {
		t.Fatal("expected non-nil StockContext for grid AST")
	}
}

func TestStockContext_ConcurrentAccess(t *testing.T) {
	ctx := transformer.NewStockContext()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.SetCode("CODE")
			ctx.SetAmount(1000)
			ctx.SetBeginPrice(10.5)
			ctx.SetCurrentPrice(11.5)
			_ = ctx.GetAmount()
			_ = ctx.GetBeginPrice()
			_ = ctx.GetCurrentPrice()
			_ = ctx.GetBeginTime()
			_ = ctx.GetDuration()
			ctx.StopCallback(func() bool { return true })
			_, _ = ctx.UseStop()
			ctx.KeepCallback(func(duration int64) bool { return true })
			_, _ = ctx.UseKeep()
		}()
	}

	wg.Wait()
}

// ========== v2.1 新增测试 ==========

func TestTransformer_LeverageAndMargin(t *testing.T) {
	input := `
select BTC
leverage 5x
margin isolated
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

	if ctx.Code != "BTC" {
		t.Errorf("expected code BTC, got %s", ctx.Code)
	}
	if ctx.GetLeverage() != 5 {
		t.Errorf("expected leverage 5, got %f", ctx.GetLeverage())
	}
	if ctx.GetMarginMode() != "isolated" {
		t.Errorf("expected margin isolated, got %s", ctx.GetMarginMode())
	}
}

func TestTransformer_PositionAndSizing(t *testing.T) {
	input := `
position dynamic 1000 capital
sizing equal
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

	if ctx.GetPositionType() != "dynamic" {
		t.Errorf("expected position type dynamic, got %s", ctx.GetPositionType())
	}
	if ctx.GetSizingMethod() != "equal" {
		t.Errorf("expected sizing equal, got %s", ctx.GetSizingMethod())
	}
}

func TestTransformer_SessionAndRanges(t *testing.T) {
	input := `
session 09:30…16:00 EST
active 2024-01-01…2024-12-31
pause 2024-02-01…2024-02-07
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

	session := ctx.GetSession()
	if len(session) < 3 {
		t.Fatalf("expected at least 3 session params, got %d", len(session))
	}
	if session[0] != "09:30" || session[1] != "16:00" || session[2] != "EST" {
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

func TestTransformer_LongAndShort(t *testing.T) {
	input := `
long AAPL {
    leverage 2x
}
short TSLA {
    leverage 3x
}
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

	// 由于当前在同一个上下文中处理，后面的 short 会覆盖前面的 long 方向与代码
	// 此处仅验证处理不崩溃且最后一个策略的设置被记录
	if ctx.GetDirection() != "short" {
		t.Errorf("expected direction short, got %s", ctx.GetDirection())
	}
	if ctx.Code != "TSLA" {
		t.Errorf("expected code TSLA, got %s", ctx.Code)
	}
	if ctx.GetLeverage() != 3 {
		t.Errorf("expected leverage 3, got %f", ctx.GetLeverage())
	}
}

func TestTransformer_SellShortBuyCoverEvents(t *testing.T) {
	ctx := transformer.NewStockContext()

	sellShortTriggered := false
	buyCoverTriggered := false

	ctx.Emiter.On("sell_short", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.StockContext); ok {
				sellShortTriggered = true
			}
		}
	})

	ctx.Emiter.On("buy_cover", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.StockContext); ok {
				buyCoverTriggered = true
			}
		}
	})

	ctx.SellShort()
	if !sellShortTriggered {
		t.Error("sell_short event was not triggered")
	}

	ctx.BuyCover()
	if !buyCoverTriggered {
		t.Error("buy_cover event was not triggered")
	}
}

func TestTransformer_VariableStorage(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.AddVariable("amount", ast.Literal{Value: "1000", Node: ast.Node{Type: ast.IntegerLiteral}})

	val, ok := ctx.GetVariable("amount")
	if !ok {
		t.Fatal("expected variable 'amount' to exist")
	}
	if val.Value != "1000" {
		t.Errorf("expected variable value 1000, got %s", val.Value)
	}

	_, ok = ctx.GetVariable("missing")
	if ok {
		t.Error("expected variable 'missing' to not exist")
	}
}

func TestTransformer_GridWithConfig(t *testing.T) {
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
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

	if ctx.Code != "AAPL" {
		t.Errorf("expected code AAPL, got %s", ctx.Code)
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

func TestTransformer_RiskAndExecutionConfig(t *testing.T) {
	input := `
compound_profit true
skip_if_gapped
fallback reduce
slippage_tolerance 0.5%
circuit_breaker 5% in 1h
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

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

func TestTransformer_AdapterComments(t *testing.T) {
	input := `
/* @adapter: binance */
/* @adapter_mode: futures */
/* @adapter_config: { api_key: "123" } */
/* @adapter_switch: { condition: "down" } */
`
	root := parser.Parser(tokenizer.Lexer(input))
	ctx := transformer.Transformer(root)

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
