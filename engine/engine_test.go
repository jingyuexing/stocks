package engine_test

import (
	"sync"
	"testing"

	"github.com/jingyuexing/stocks/compiler"
	"github.com/jingyuexing/stocks/engine"
)

func TestEngine_Lifecycle(t *testing.T) {
	e := engine.New()
	if e.State() != engine.StateCreated {
		t.Errorf("expected initial state created, got %s", e.State())
	}

	bundle, err := compiler.Compile(`select AAPL`)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	// Load
	if err := e.Load(bundle); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if e.State() != engine.StateLoaded {
		t.Errorf("expected state loaded, got %s", e.State())
	}
	if e.Bundle() == nil {
		t.Error("expected bundle to be set")
	}
	if e.Context() == nil {
		t.Error("expected context to be accessible")
	}
	if e.Scheduler() == nil {
		t.Error("expected scheduler to be initialized")
	}

	// Start
	if err := e.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if e.State() != engine.StateRunning {
		t.Errorf("expected state running, got %s", e.State())
	}

	// Pause
	if err := e.Pause(); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	if e.State() != engine.StatePaused {
		t.Errorf("expected state paused, got %s", e.State())
	}

	// Start again from paused
	if err := e.Start(); err != nil {
		t.Fatalf("restart from paused failed: %v", err)
	}
	if e.State() != engine.StateRunning {
		t.Errorf("expected state running after resume, got %s", e.State())
	}

	// Stop
	if err := e.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if e.State() != engine.StateStopped {
		t.Errorf("expected state stopped, got %s", e.State())
	}

	// Unload
	if err := e.Unload(); err != nil {
		t.Fatalf("unload failed: %v", err)
	}
	if e.State() != engine.StateCreated {
		t.Errorf("expected state created after unload, got %s", e.State())
	}
	if e.Bundle() != nil {
		t.Error("expected bundle to be nil after unload")
	}
}

func TestEngine_AutoStart(t *testing.T) {
	e := engine.New(engine.WithAutoStart(true))
	bundle, err := compiler.Compile(`keep 1W`)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if err := e.Load(bundle); err != nil {
		t.Fatalf("load with autostart failed: %v", err)
	}
	if e.State() != engine.StateRunning {
		t.Errorf("expected state running with autostart, got %s", e.State())
	}
	_ = e.Stop()
	_ = e.Unload()
}

func TestEngine_CompileAndLoad(t *testing.T) {
	e := engine.New()
	if err := e.CompileAndLoad(`select TSLA`); err != nil {
		t.Fatalf("compile and load failed: %v", err)
	}
	if e.State() != engine.StateLoaded {
		t.Errorf("expected state loaded, got %s", e.State())
	}
	if e.Context().GetCode() != "TSLA" {
		t.Errorf("expected code TSLA, got %s", e.Context().GetCode())
	}
}

func TestEngine_InvalidTransitions(t *testing.T) {
	e := engine.New()

	// 未加载时不能 Start
	if err := e.Start(); err == nil {
		t.Error("expected error when starting without load")
	}

	// 未加载时不能 Pause
	if err := e.Pause(); err == nil {
		t.Error("expected error when pausing without load")
	}

	// 未加载时不能 Stop
	if err := e.Stop(); err == nil {
		t.Error("expected error when stopping without load")
	}
}

func TestEngine_DoubleLoad(t *testing.T) {
	e := engine.New()
	bundle1, _ := compiler.Compile(`select A`)
	bundle2, _ := compiler.Compile(`select B`)

	if err := e.Load(bundle1); err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	// 当前状态为 Loaded，再次 Load 应该报错（非 Created/Stopped）
	if err := e.Load(bundle2); err == nil {
		t.Error("expected error when double-loading without unload")
	}
}

func TestEngine_StateQueries(t *testing.T) {
	states := []engine.State{
		engine.StateCreated,
		engine.StateLoaded,
		engine.StateRunning,
		engine.StatePaused,
		engine.StateStopped,
	}
	for _, s := range states {
		_ = s.IsActive()
		_ = s.CanStart()
		_ = s.CanPause()
		_ = s.CanStop()
	}
}

func TestEngine_ContextWithoutBundle(t *testing.T) {
	e := engine.New()
	if e.Context() != nil {
		t.Error("expected nil context when no bundle loaded")
	}
}

func TestEngine_StopFromPaused(t *testing.T) {
	e := engine.New()
	bundle, _ := compiler.Compile(`select A`)
	_ = e.Load(bundle)
	_ = e.Start()
	_ = e.Pause()
	if e.State() != engine.StatePaused {
		t.Fatalf("expected paused state")
	}
	if err := e.Stop(); err != nil {
		t.Fatalf("stop from paused failed: %v", err)
	}
	if e.State() != engine.StateStopped {
		t.Errorf("expected stopped, got %s", e.State())
	}
}

func TestEngine_ReloadAfterStop(t *testing.T) {
	e := engine.New()
	bundle, _ := compiler.Compile(`select A`)
	_ = e.Load(bundle)
	_ = e.Start()
	_ = e.Stop()

	bundle2, _ := compiler.Compile(`select B`)
	if err := e.Load(bundle2); err != nil {
		t.Fatalf("reload after stop failed: %v", err)
	}
	if e.State() != engine.StateLoaded {
		t.Errorf("expected loaded after reload, got %s", e.State())
	}
}

func TestEngine_WithOptions(t *testing.T) {
	e := engine.New(
		engine.WithPriceCheckInterval("*/1 * * * * *"),
		engine.WithStrictMode(true),
		engine.WithAutoStart(false),
	)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	// 选项已内部持有，通过行为间接验证
	bundle, _ := compiler.Compile(`select A`)
	_ = e.Load(bundle)
	if e.State() != engine.StateLoaded {
		t.Error("autostart should be false")
	}
}

func TestEngine_ConcurrentStateAccess(t *testing.T) {
	e := engine.New()
	bundle, _ := compiler.Compile(`select A`)
	_ = e.Load(bundle)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = e.State()
		}()
		go func() {
			defer wg.Done()
			_ = e.Context()
		}()
		go func() {
			defer wg.Done()
			_ = e.Bundle()
		}()
	}
	wg.Wait()
}

func TestEngine_WithAdapterAnnotations(t *testing.T) {
	source := `
/* @adapter: backtest */
/* @adapter_mode: spot */
select BTC
`
	eng := engine.New()
	if err := eng.CompileAndLoad(source); err != nil {
		t.Fatalf("compile and load failed: %v", err)
	}
	if eng.AdapterManager() == nil {
		t.Fatal("expected adapter manager to be initialized")
	}
	if !eng.AdapterManager().IsLoaded() {
		t.Fatal("expected adapter manager to be loaded")
	}
	if eng.RiskGuard() == nil {
		t.Fatal("expected risk guard to be initialized")
	}
	_ = eng.Start()
	_ = eng.Stop()
	_ = eng.Unload()
}

func TestEngine_AdapterStrictMode(t *testing.T) {
	source := `/* @adapter: nonexistent_gateway */`
	eng := engine.New(engine.WithStrictMode(true))
	if err := eng.CompileAndLoad(source); err == nil {
		t.Fatal("expected error in strict mode for unknown gateway")
	}
}

func TestEngine_AdapterNonStrictMode(t *testing.T) {
	source := `/* @adapter: nonexistent_gateway */`
	eng := engine.New(engine.WithStrictMode(false))
	if err := eng.CompileAndLoad(source); err != nil {
		t.Fatalf("expected load to succeed in non-strict mode, got: %v", err)
	}
	if eng.AdapterManager() != nil {
		t.Error("expected adapter manager to be nil when load failed")
	}
	_ = eng.Unload()
}
