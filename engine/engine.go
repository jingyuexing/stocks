package engine

import (
	"fmt"
	"sync"

	"github.com/jingyuexing/stocks/adapter"
	"github.com/jingyuexing/stocks/compiler"
	"github.com/jingyuexing/stocks/gateway"
	"github.com/jingyuexing/stocks/scheduler"
	"github.com/jingyuexing/stocks/transformer"
)

// Engine 是策略执行引擎，负责管理策略生命周期与调度
type Engine struct {
	mu     sync.RWMutex
	state  State
	opts   Options
	bundle *compiler.StrategyBundle

	// Scheduler 定时任务调度器
	scheduler *scheduler.Scheduler

	// 执行器
	executor *Executor

	// 风控拦截器
	riskGuard *RiskGuard

	// 适配器管理器
	adapterMgr *adapter.Manager
}

// New 创建新的 Engine
func New(opts ...Option) *Engine {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Engine{
		state: StateCreated,
		opts:  o,
	}
}

// Load 加载编译后的 StrategyBundle
func (e *Engine) Load(bundle *compiler.StrategyBundle) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateCreated && e.state != StateStopped {
		return fmt.Errorf("cannot load bundle in state %s", e.state)
	}
	if bundle == nil {
		return fmt.Errorf("bundle is nil")
	}
	if bundle.Context == nil {
		return fmt.Errorf("bundle context is nil")
	}

	e.bundle = bundle
	e.scheduler = scheduler.NewScheduler(bundle.Context)

	// 初始化风控拦截器
	e.riskGuard = NewRiskGuard(bundle.Context)

	// 若配置了适配器，初始化 AdapterManager
	if !bundle.Adapter.IsEmpty() && e.opts.EnableAdapterManager {
		mgr := adapter.NewManager()
		if err := mgr.Load(bundle.Adapter); err != nil {
			// 非严格模式下仅记录，不阻断
			fmt.Printf("[engine] adapter load warning: %v\n", err)
			if e.opts.StrictMode {
				return err
			}
		} else {
			e.adapterMgr = mgr
		}
	}

	e.state = StateLoaded

	if e.opts.AutoStart {
		return e.startLocked()
	}
	return nil
}

// CompileAndLoad 从 DSL 源码直接编译并加载
func (e *Engine) CompileAndLoad(source string) error {
	bundle, err := compiler.Compile(source)
	if err != nil {
		return fmt.Errorf("compile failed: %w", err)
	}
	return e.Load(bundle)
}

// Start 启动引擎（调度器开始工作）
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.startLocked()
}

func (e *Engine) startLocked() error {
	if !e.state.CanStart() {
		return fmt.Errorf("cannot start engine in state %s", e.state)
	}
	if e.scheduler == nil {
		return fmt.Errorf("scheduler not initialized")
	}

	// 启动调度器
	e.scheduler.Start()

	// 根据 AST 注册默认定时任务
	if err := e.scheduler.ScheduleFromAST(&e.bundle.AST, e.opts.PriceCheckInterval); err != nil {
		return fmt.Errorf("schedule from AST failed: %w", err)
	}

	// 若已加载网关，注入行情数据提供者到 Context
	if e.adapterMgr != nil && e.adapterMgr.IsLoaded() {
		e.bundle.Context.Vars.Provider = &gatewayProvider{gw: e.adapterMgr.Primary()}
		e.bundle.Context.Vars.Symbol = e.bundle.Context.GetCode()
	}

	// 初始化 Executor 并订阅事件
	e.executor = NewExecutor(e.bundle.Context, &e.bundle.AST)
	if e.adapterMgr != nil {
		e.executor.SetAdapterManager(e.adapterMgr)
	}
	e.executor.Setup()

	// 通知网关引擎启动
	if e.adapterMgr != nil && e.adapterMgr.IsLoaded() {
		_ = e.adapterMgr.Primary().OnEngineStart()
	}

	e.state = StateRunning
	return nil
}

// Pause 暂停引擎（调度器停止触发新任务，但保留状态）
func (e *Engine) Pause() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.state.CanPause() {
		return fmt.Errorf("cannot pause engine in state %s", e.state)
	}
	if e.scheduler != nil {
		e.scheduler.Stop()
	}
	e.state = StatePaused
	return nil
}

// Stop 停止引擎（完全停止调度器，释放资源）
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.state.CanStop() {
		return fmt.Errorf("cannot stop engine in state %s", e.state)
	}
	if e.scheduler != nil {
		e.scheduler.Stop()
	}

	// 通知网关引擎停止
	if e.adapterMgr != nil && e.adapterMgr.IsLoaded() {
		_ = e.adapterMgr.Primary().OnEngineStop()
	}

	e.state = StateStopped
	return nil
}

// Unload 卸载当前策略，重置引擎到 Created 状态
func (e *Engine) Unload() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.IsActive() {
		if e.scheduler != nil {
			e.scheduler.Stop()
		}
	}
	e.bundle = nil
	e.scheduler = nil
	e.executor = nil
	e.riskGuard = nil

	if e.adapterMgr != nil {
		_ = e.adapterMgr.Close()
		e.adapterMgr = nil
	}

	e.state = StateCreated
	return nil
}

// State 返回当前引擎状态（线程安全）
func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// Bundle 返回当前加载的策略包（线程安全）
func (e *Engine) Bundle() *compiler.StrategyBundle {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.bundle
}

// Context 返回运行时上下文（便捷方法）
func (e *Engine) Context() *transformer.StockContext {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.bundle == nil {
		return nil
	}
	return e.bundle.Context
}

// Scheduler 返回内部调度器（用于高级自定义任务注册）
func (e *Engine) Scheduler() *scheduler.Scheduler {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.scheduler
}

// RiskGuard 返回风控拦截器
func (e *Engine) RiskGuard() *RiskGuard {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.riskGuard
}

// AdapterManager 返回适配器管理器
func (e *Engine) AdapterManager() *adapter.Manager {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adapterMgr
}

// ---------- gateway -> MarketDataProvider 适配 ----------

type gatewayProvider struct {
	gw gateway.TradingGateway
}

func (p *gatewayProvider) GetPrice(symbol string) float64 {
	if p.gw == nil {
		return 0
	}
	price, err := p.gw.GetMarketPrice(symbol)
	if err != nil {
		return 0
	}
	return price
}

func (p *gatewayProvider) GetVolume(symbol string) float64 {
	// TODO: gateway 接口扩展 GetVolume 后接入
	return 0
}

func (p *gatewayProvider) GetHigh(symbol string) float64 {
	// TODO: gateway 接口扩展 GetHigh 后接入
	return 0
}

func (p *gatewayProvider) GetLow(symbol string) float64 {
	// TODO: gateway 接口扩展 GetLow 后接入
	return 0
}

func (p *gatewayProvider) GetOpen(symbol string) float64 {
	// TODO: gateway 接口扩展 GetVolume 后接入
	return 0
}

func (p *gatewayProvider) GetClose(symbol string) float64 {
	// TODO: gateway 接口扩展 GetVolume 后接入
	return 0
}

func (p *gatewayProvider) GetCost(symbol string) float64 {
	// TODO: gateway 接口扩展 GetCost 后接入
	return 0
}

func (p *gatewayProvider) GetProfit(symbol string) float64 {
	// TODO: gateway 接口扩展 GetProfit 后接入
	return 0
}
