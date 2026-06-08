package transformer_test

import (
	"sync"
	"testing"

	"github.com/jingyuexing/stocks/transformer"
)

func TestStockContext_GettersSetters(t *testing.T) {
	ctx := transformer.NewStockContext()

	ctx.SetBeginPrice(123.45)
	if ctx.GetBeginPrice() != 123.45 {
		t.Errorf("expected begin price 123.45, got %f", ctx.GetBeginPrice())
	}

	ctx.SetCurrentPrice(99.99)
	if ctx.GetCurrentPrice() != 99.99 {
		t.Errorf("expected current price 99.99, got %f", ctx.GetCurrentPrice())
	}

	ctx.SetAmount(1000)
	if ctx.GetAmount() != 1000 {
		t.Errorf("expected amount 1000, got %f", ctx.GetAmount())
	}

	ctx.SetCode("TEST")
	if ctx.GetCode() != "TEST" {
		t.Errorf("expected code TEST, got %s", ctx.GetCode())
	}

	ctx.SetDirection("long")
	if ctx.GetDirection() != "long" {
		t.Errorf("expected direction long, got %s", ctx.GetDirection())
	}
}

func TestStockContext_StopCallback(t *testing.T) {
	ctx := transformer.NewStockContext()

	ctx.StopCallback(nil)
	ok, err := ctx.UseStop()
	if err == nil {
		t.Error("expected error when stop callback is nil")
	}
	if ok {
		t.Error("expected false when stop callback is nil")
	}

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

	ok, err := ctx.UseKeep()
	if err != nil {
		t.Errorf("unexpected error for default keep: %v", err)
	}
	_ = ok

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
			if _, ok := args[0].(*transformer.Event); ok {
				sellTriggered = true
			}
		}
	})

	ctx.Emiter.On("buy", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.Event); ok {
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

func TestStockContext_SellShortBuyCoverEvents(t *testing.T) {
	ctx := transformer.NewStockContext()

	sellShortTriggered := false
	buyCoverTriggered := false

	ctx.Emiter.On("sell_short", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.Event); ok {
				sellShortTriggered = true
			}
		}
	})

	ctx.Emiter.On("buy_cover", func(args ...any) {
		if len(args) > 0 {
			if _, ok := args[0].(*transformer.Event); ok {
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
			_ = ctx.GetDuration()
			ctx.StopCallback(func() bool { return true })
			_, _ = ctx.UseStop()
			ctx.KeepCallback(func(duration int64) bool { return true })
			_, _ = ctx.UseKeep()
		}()
	}

	wg.Wait()
}

func TestStockContext_EventFields(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetCode("AAPL")
	ctx.SetDirection("long")
	ctx.SetCurrentPrice(150.5)
	ctx.Vars.Register("volume", func(ctx *transformer.StockContext) float64 { return 10000 })

	evt := ctx.NewEvent("test")
	if evt.Name != "test" {
		t.Errorf("expected event name test, got %s", evt.Name)
	}
	if evt.Code != "AAPL" {
		t.Errorf("expected code AAPL, got %s", evt.Code)
	}
	if evt.Direction != "long" {
		t.Errorf("expected direction long, got %s", evt.Direction)
	}
	if evt.Price != 150.5 {
		t.Errorf("expected price 150.5, got %f", evt.Price)
	}
	if evt.Volume != 10000 {
		t.Errorf("expected volume 10000, got %f", evt.Volume)
	}
	if evt.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestTransformer_VariableStorage(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.Vars.SetUserVar("amount", transformer.LiteralValue{Value: "1000"})

	val, ok := ctx.Vars.UserVars["amount"]
	if !ok {
		t.Fatal("expected variable 'amount' to exist")
	}
	if val.Value != "1000" {
		t.Errorf("expected variable value 1000, got %s", val.Value)
	}

	_, ok = ctx.Vars.UserVars["missing"]
	if ok {
		t.Error("expected variable 'missing' to not exist")
	}
}
