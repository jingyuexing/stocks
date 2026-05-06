package transformer_test

import (
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
