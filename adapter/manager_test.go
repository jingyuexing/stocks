package adapter_test

import (
	"testing"

	"github.com/jingyuexing/stocks/adapter"
	"github.com/jingyuexing/stocks/compiler"
	"github.com/jingyuexing/stocks/gateway"
)

func TestManager_Load(t *testing.T) {
	mgr := adapter.NewManager()
	cfg := compiler.AdapterConfig{
		Primary: "backtest",
		Mode:    "spot",
		Params:  `{"initial_cash": 500000}`,
	}
	if err := mgr.Load(cfg); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !mgr.IsLoaded() {
		t.Error("expected manager to be loaded")
	}
	if mgr.Primary() == nil {
		t.Fatal("expected primary gateway to be set")
	}
}

func TestManager_LoadUnknownGateway(t *testing.T) {
	mgr := adapter.NewManager()
	cfg := compiler.AdapterConfig{
		Primary: "unknown_exchange",
	}
	if err := mgr.Load(cfg); err == nil {
		t.Fatal("expected error for unknown gateway")
	}
}

func TestManager_LoadEmptyPrimary(t *testing.T) {
	mgr := adapter.NewManager()
	cfg := compiler.AdapterConfig{}
	if err := mgr.Load(cfg); err == nil {
		t.Fatal("expected error when primary is empty")
	}
}

func TestManager_Close(t *testing.T) {
	mgr := adapter.NewManager()
	_ = mgr.Load(compiler.AdapterConfig{
		Primary: "backtest",
	})
	if err := mgr.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if mgr.IsLoaded() {
		t.Error("expected manager not loaded after close")
	}
}

func TestManager_Fallback(t *testing.T) {
	mgr := adapter.NewManager()
	cfg := compiler.AdapterConfig{
		Primary:  "backtest",
		Fallback: "backtest",
	}
	if err := mgr.Load(cfg); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if mgr.Fallback() == nil {
		t.Error("expected fallback gateway to be set")
	}
}

func TestManager_GatewayRegistry(t *testing.T) {
	known := gateway.KnownGateways()
	found := false
	for _, name := range known {
		if name == "backtest" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'backtest' to be in known gateways")
	}
}
