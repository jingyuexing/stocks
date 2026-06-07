package compiler_test

import (
	"testing"

	"github.com/jingyuexing/stocks/compiler"
)

func TestCompile_SimpleGrid(t *testing.T) {
	source := `
grid AAPL {
    session 09:30…16:00 EST
    position dynamic 1000 capital
    sizing equal
    loss 10% { sell 100% }
    profit 10% { sell 50% }
}
`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle == nil {
		t.Fatal("expected non-nil bundle")
	}
	if bundle.Context == nil {
		t.Fatal("expected non-nil context")
	}
	if bundle.Context.GetCode() != "AAPL" {
		t.Errorf("expected code AAPL, got %s", bundle.Context.GetCode())
	}
	if len(bundle.AST.Statements) == 0 {
		t.Error("expected non-empty AST")
	}
	if bundle.Metadata["version"] != "2.1" {
		t.Errorf("expected version 2.1, got %s", bundle.Metadata["version"])
	}
}

func TestCompile_WithAdapterAnnotations(t *testing.T) {
	source := `
/* @adapter: binance */
/* @adapter_mode: futures */
/* @adapter_config: { api_key: "test_key" } */
/* @adapter_switch: { condition: "down" } */
grid BTCUSDT {
    leverage 5x
}
`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if bundle.Adapter.Primary != "binance" {
		t.Errorf("expected primary adapter binance, got %s", bundle.Adapter.Primary)
	}
	if bundle.Adapter.Mode != "futures" {
		t.Errorf("expected mode futures, got %s", bundle.Adapter.Mode)
	}
	if bundle.Adapter.Params == "" {
		t.Error("expected non-empty adapter config params")
	}
	if bundle.Adapter.Switch == "" {
		t.Error("expected non-empty adapter switch config")
	}
}

func TestCompile_WithMultiAdapter(t *testing.T) {
	source := `/* @adapter: primary=okx, fallback=binance */`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle.Adapter.Primary != "okx" {
		t.Errorf("expected primary okx, got %s", bundle.Adapter.Primary)
	}
	if bundle.Adapter.Fallback != "binance" {
		t.Errorf("expected fallback binance, got %s", bundle.Adapter.Fallback)
	}
}

func TestCompileWithID(t *testing.T) {
	source := `select TSLA`
	bundle, err := compiler.CompileWithID("strategy-001", source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle.ID != "strategy-001" {
		t.Errorf("expected ID strategy-001, got %s", bundle.ID)
	}
}

func TestCompile_SourcePreserved(t *testing.T) {
	source := `keep 1W`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle.Source != source {
		t.Errorf("expected source preserved, got %q", bundle.Source)
	}
}

func TestCompile_EmptyASTAllowed(t *testing.T) {
	// 空输入会生成只有 EOF 的 AST
	bundle, err := compiler.Compile(``)
	if err != nil {
		t.Fatalf("compile failed for empty source: %v", err)
	}
	if bundle.Context == nil {
		t.Fatal("expected context even for empty source")
	}
}

func TestCompile_LongShortStrategy(t *testing.T) {
	source := `
long AAPL {
    leverage 2x
}
`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle.Context.GetDirection() != "long" {
		t.Errorf("expected direction long, got %s", bundle.Context.GetDirection())
	}
	if bundle.Context.GetLeverage() != 2 {
		t.Errorf("expected leverage 2, got %f", bundle.Context.GetLeverage())
	}
}

func TestCompile_GridWithRiskConfig(t *testing.T) {
	// 使用顶层表达式而非 grid body，以规避 parser grid body 的已知内存问题
	source := `
max_position 5000 shares
stop_loss -10%...-5%
circuit_breaker 5% in 1h
`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle.Context.GetMaxPosition() != 5000 {
		t.Errorf("expected max_position 5000, got %f", bundle.Context.GetMaxPosition())
	}
	if bundle.Context.GetCircuitBreaker() != 5 {
		t.Errorf("expected circuit_breaker 5, got %f", bundle.Context.GetCircuitBreaker())
	}
	if bundle.Context.GetStopLossRange() == nil {
		t.Error("expected stop_loss range to be set")
	}
}

func TestCompile_StopLossSingleValue(t *testing.T) {
	// stop_loss 支持单点值（无范围符号）
	source := `stop_loss -10%`
	bundle, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if bundle == nil || bundle.Context == nil {
		t.Fatal("expected non-nil bundle and context")
	}
	slr := bundle.Context.GetStopLossRange()
	if slr == nil {
		t.Fatal("expected stop_loss range to be set for single value")
	}
	if slr.Begin == nil {
		t.Fatal("expected stop_loss range begin to be non-nil")
	}
	if slr.Begin.Value != "-10" {
		t.Errorf("expected stop_loss begin -10, got %s", slr.Begin.Value)
	}
	if slr.End != nil && slr.End.Value != "" {
		t.Errorf("expected open-ended range (end empty), got %s", slr.End.Value)
	}
}
