package transformer

import (
	"errors"
	"sync"
	"time"

	"github.com/jingyuexing/go-utils"
	"github.com/jingyuexing/stocks/ast"
)

// StockContext 存储策略执行过程中的状态与事件
type StockContext struct {
	mu sync.RWMutex

	Code string // 股票代码

	// 任务起始时间
	start time.Time // 起始时间
	// 起始价格
	beginPrice float64 // 起始价格
	// 持续时长
	KeepDuration time.Duration
	// 钱包资金
	amount float64 // 资金
	// 当前价格
	currentPrice float64 //当前价格
	// 事件发射器
	Emiter *utils.EventEmit

	stop     func() bool
	keep     func(duration int64) bool
	duration utils.DateTime

	// ========== v2.1 新增字段 ==========
	// 策略方向
	Direction string
	// 杠杆倍数
	Leverage float64
	// 保证金模式
	MarginMode string
	// 头寸类型 fixed/dynamic
	PositionType string
	// 头寸规模算法
	SizingMethod string
	// 最大持仓
	MaxPosition float64
	// 止损范围
	StopLossRange *ast.RangeExpressionNode
	// 交易时段
	Session []string
	// 有效日期范围
	ActiveRange []string
	// 暂停日期范围
	PauseRange []string

	// 时间控制
	HoldMax          utils.DateTime
	Cooldown         utils.DateTime
	CooldownPerLevel bool

	// 执行偏好
	PartialFill    string
	CompoundProfit bool
	SkipIfGapped   bool
	Fallback       string

	// 头寸管理
	BasePosition     float64
	PyramidStep      float64
	MaxPyramidLayers int
	FixedFraction    float64
	VolatilityTarget float64
	AtrPeriod        int
	RiskPerTrade     float64
	RiskPerGrid      float64
	MaxDrawdown      float64
	PositionDecay    float64
	PositionDecayPer utils.DateTime
	GrossExposure    float64
	NetExposure      float64
	BetaNeutral      bool
	Rebalance        string
	RebalanceAt      string

	// 杠杆与保证金
	Hedge           string
	FundingPriority string
	MaxShort        float64
	BorrowRateLimit float64

	// 风控
	SlippageTolerance float64
	CircuitBreaker    float64
	CircuitBreakerIn  utils.DateTime

	// 组合配置
	PortfolioHeat    float64
	CorrelationLimit float64

	// 运行时变量存储
	Variables map[string]ast.Literal

	// Adapter 配置
	Adapter       string
	AdapterMode   string
	AdapterConfig string
	AdapterSwitch string
}

// NewStockContext 创建新的执行上下文
func NewStockContext() *StockContext {
	return &StockContext{
		start:     time.Now(),
		Emiter:    utils.NewEventEmit(),
		Variables: make(map[string]ast.Literal),
	}
}

func (s *StockContext) GetBeginTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.start
}

func (s *StockContext) SetCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if code != "" {
		s.Code = code
	}
}

func (s *StockContext) GetBeginPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.beginPrice
}

func (s *StockContext) SetBeginPrice(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.beginPrice = val
}

func (s *StockContext) GetAmount() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.amount
}

func (s *StockContext) SetAmount(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.amount = val
}

func (s *StockContext) GetCurrentPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentPrice
}

func (s *StockContext) SetCurrentPrice(val float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentPrice = val
	return s.currentPrice
}

func (s *StockContext) StopCallback(callback func() bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if callback != nil {
		s.stop = callback
	}
}

func (s *StockContext) KeepCallback(callback func(duration int64) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if callback != nil {
		s.keep = callback
	}
}

func (s *StockContext) SetDuration(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.duration = dt
}

func (s *StockContext) GetDuration() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.duration
}

func (s *StockContext) UseStop() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.stop == nil {
		return false, errors.New(`the callback "stop" is nil`)
	}
	return s.stop(), nil
}

func (s *StockContext) UseKeep() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.keep == nil {
		defaultKeep := func(duration int64) bool {
			// 注意：不在闭包内调用加锁方法，避免死锁
			return s.start.After(time.Now())
		}
		s.keep = defaultKeep
	}
	return s.keep(0), nil
}

func (s *StockContext) Sell() {
	s.Emiter.Emit("sell", s)
}

func (s *StockContext) Buy() {
	s.Emiter.Emit("buy", s)
}

func (s *StockContext) SellShort() {
	s.Emiter.Emit("sell_short", s)
}

func (s *StockContext) BuyCover() {
	s.Emiter.Emit("buy_cover", s)
}

// ========== v2.1 Getters & Setters ==========

func (s *StockContext) SetDirection(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Direction = dir
}

func (s *StockContext) GetDirection() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Direction
}

func (s *StockContext) SetLeverage(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Leverage = val
}

func (s *StockContext) GetLeverage() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Leverage
}

func (s *StockContext) SetMarginMode(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MarginMode = val
}

func (s *StockContext) GetMarginMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MarginMode
}

func (s *StockContext) SetPositionType(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PositionType = val
}

func (s *StockContext) GetPositionType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PositionType
}

func (s *StockContext) SetSizingMethod(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SizingMethod = val
}

func (s *StockContext) GetSizingMethod() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SizingMethod
}

func (s *StockContext) SetMaxPosition(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MaxPosition = val
}

func (s *StockContext) GetMaxPosition() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxPosition
}

func (s *StockContext) SetStopLossRange(r *ast.RangeExpressionNode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StopLossRange = r
}

func (s *StockContext) GetStopLossRange() *ast.RangeExpressionNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.StopLossRange
}

func (s *StockContext) SetSession(params []ast.Literal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Session = nil
	for _, p := range params {
		s.Session = append(s.Session, p.Value)
	}
}

func (s *StockContext) GetSession() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Session
}

func (s *StockContext) SetActiveRange(params []ast.Literal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveRange = nil
	for _, p := range params {
		s.ActiveRange = append(s.ActiveRange, p.Value)
	}
}

func (s *StockContext) GetActiveRange() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ActiveRange
}

func (s *StockContext) SetPauseRange(params []ast.Literal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PauseRange = nil
	for _, p := range params {
		s.PauseRange = append(s.PauseRange, p.Value)
	}
}

func (s *StockContext) GetPauseRange() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PauseRange
}

func (s *StockContext) SetHoldMax(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HoldMax = dt
}

func (s *StockContext) GetHoldMax() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.HoldMax
}

func (s *StockContext) SetCooldown(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Cooldown = dt
}

func (s *StockContext) GetCooldown() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Cooldown
}

func (s *StockContext) SetCooldownPerLevel(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CooldownPerLevel = v
}

func (s *StockContext) GetCooldownPerLevel() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CooldownPerLevel
}

func (s *StockContext) SetPartialFill(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PartialFill = val
}

func (s *StockContext) GetPartialFill() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PartialFill
}

func (s *StockContext) SetCompoundProfit(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CompoundProfit = v
}

func (s *StockContext) GetCompoundProfit() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CompoundProfit
}

func (s *StockContext) SetSkipIfGapped(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SkipIfGapped = v
}

func (s *StockContext) GetSkipIfGapped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SkipIfGapped
}

func (s *StockContext) SetFallback(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Fallback = val
}

func (s *StockContext) GetFallback() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Fallback
}

func (s *StockContext) SetBasePosition(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BasePosition = val
}

func (s *StockContext) GetBasePosition() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BasePosition
}

func (s *StockContext) SetPyramidStep(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PyramidStep = val
}

func (s *StockContext) GetPyramidStep() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PyramidStep
}

func (s *StockContext) SetMaxPyramidLayers(val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MaxPyramidLayers = val
}

func (s *StockContext) GetMaxPyramidLayers() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxPyramidLayers
}

func (s *StockContext) SetFixedFraction(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FixedFraction = val
}

func (s *StockContext) GetFixedFraction() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FixedFraction
}

func (s *StockContext) SetVolatilityTarget(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VolatilityTarget = val
}

func (s *StockContext) GetVolatilityTarget() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.VolatilityTarget
}

func (s *StockContext) SetAtrPeriod(val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AtrPeriod = val
}

func (s *StockContext) GetAtrPeriod() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AtrPeriod
}

func (s *StockContext) SetRiskPerTrade(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RiskPerTrade = val
}

func (s *StockContext) GetRiskPerTrade() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RiskPerTrade
}

func (s *StockContext) SetRiskPerGrid(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RiskPerGrid = val
}

func (s *StockContext) GetRiskPerGrid() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RiskPerGrid
}

func (s *StockContext) SetMaxDrawdown(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MaxDrawdown = val
}

func (s *StockContext) GetMaxDrawdown() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxDrawdown
}

func (s *StockContext) SetPositionDecay(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PositionDecay = val
}

func (s *StockContext) GetPositionDecay() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PositionDecay
}

func (s *StockContext) SetPositionDecayPer(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PositionDecayPer = dt
}

func (s *StockContext) GetPositionDecayPer() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PositionDecayPer
}

func (s *StockContext) SetGrossExposure(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GrossExposure = val
}

func (s *StockContext) GetGrossExposure() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.GrossExposure
}

func (s *StockContext) SetNetExposure(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NetExposure = val
}

func (s *StockContext) GetNetExposure() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.NetExposure
}

func (s *StockContext) SetBetaNeutral(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BetaNeutral = v
}

func (s *StockContext) GetBetaNeutral() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BetaNeutral
}

func (s *StockContext) SetRebalance(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Rebalance = val
}

func (s *StockContext) GetRebalance() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Rebalance
}

func (s *StockContext) SetRebalanceAt(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RebalanceAt = val
}

func (s *StockContext) GetRebalanceAt() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RebalanceAt
}

func (s *StockContext) SetHedge(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hedge = val
}

func (s *StockContext) GetHedge() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Hedge
}

func (s *StockContext) SetFundingPriority(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FundingPriority = val
}

func (s *StockContext) GetFundingPriority() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FundingPriority
}

func (s *StockContext) SetMaxShort(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MaxShort = val
}

func (s *StockContext) GetMaxShort() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxShort
}

func (s *StockContext) SetBorrowRateLimit(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BorrowRateLimit = val
}

func (s *StockContext) GetBorrowRateLimit() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BorrowRateLimit
}

func (s *StockContext) SetSlippageTolerance(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SlippageTolerance = val
}

func (s *StockContext) GetSlippageTolerance() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SlippageTolerance
}

func (s *StockContext) SetCircuitBreaker(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CircuitBreaker = val
}

func (s *StockContext) GetCircuitBreaker() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CircuitBreaker
}

func (s *StockContext) SetCircuitBreakerIn(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CircuitBreakerIn = dt
}

func (s *StockContext) GetCircuitBreakerIn() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CircuitBreakerIn
}

func (s *StockContext) SetPortfolioHeat(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PortfolioHeat = val
}

func (s *StockContext) GetPortfolioHeat() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PortfolioHeat
}

func (s *StockContext) SetCorrelationLimit(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CorrelationLimit = val
}

func (s *StockContext) GetCorrelationLimit() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CorrelationLimit
}

// AddVariable 添加运行时变量
func (s *StockContext) AddVariable(name string, val ast.Literal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Variables == nil {
		s.Variables = make(map[string]ast.Literal)
	}
	s.Variables[name] = val
}

// GetVariable 获取运行时变量
func (s *StockContext) GetVariable(name string) (ast.Literal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Variables[name]
	return v, ok
}

func (s *StockContext) SetAdapter(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Adapter = val
}

func (s *StockContext) GetAdapter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Adapter
}

func (s *StockContext) SetAdapterMode(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdapterMode = val
}

func (s *StockContext) GetAdapterMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AdapterMode
}

func (s *StockContext) SetAdapterConfig(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdapterConfig = val
}

func (s *StockContext) GetAdapterConfig() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AdapterConfig
}

func (s *StockContext) SetAdapterSwitch(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdapterSwitch = val
}

func (s *StockContext) GetAdapterSwitch() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AdapterSwitch
}
