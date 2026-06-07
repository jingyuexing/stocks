package transformer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jingyuexing/go-utils"
)

type LiteralType int

const (
	Text LiteralType = iota
	Float
	Integer
	Boolean
	Time
	Date
	Duration
	Percentage
	List
	Enum
)

// LiteralValue 表示一个字面量值，完全隔离 ast 模块
type LiteralValue struct {
	Type  LiteralType
	Value string
	Unit  string
}

func (l LiteralValue) String() string {
	switch l.Type {
	case Percentage:
		if l.Value == "" {
			return "0"
		}
		v, _ := strconv.ParseFloat(l.Value, 64)
		return fmt.Sprintf("%f", v/100)
	case Time, Date, Duration, Enum, Text, List:
		if l.Unit != "" {
			return l.Value + l.Unit
		}
		return l.Value
	default:
		return l.Value
	}
}

func (l LiteralValue) AutoConvert() any {
	switch l.Type {
	case Float:
		return l.AsFloat()
	case Integer:
		return l.AsInt()
	case Text, Enum:
		return l.String()
	case Boolean:
		return l.AsBool()
	case Time:
		return l.AsTime()
	case Date:
		return l.AsDate()
	case Duration:
		return l.AsDuration()
	case Percentage:
		return l.AsPercentage()
	case List:
		return l.AsList()
	default:
		return l.Value
	}
}

// IsNumeric 判断当前字面量是否为数值类型
func (l LiteralValue) IsNumeric() bool {
	switch l.Type {
	case Float, Integer, Percentage:
		return true
	default:
		// 对未设置类型的文本值尝试判断
		if l.Type == Text && l.Value != "" {
			if _, err := strconv.ParseFloat(l.Value, 64); err == nil {
				return true
			}
		}
		return false
	}
}

func (l LiteralValue) AsFloat() float64 {
	if l.Value == "" {
		return 0
	}
	// Percentage 的原始数值（如 10% -> 10）
	v, _ := strconv.ParseFloat(l.Value, 64)
	return v
}

func (l LiteralValue) AsInt() int {
	if l.Value == "" {
		return 0
	}
	v, _ := strconv.Atoi(l.Value)
	return v
}

func (l LiteralValue) AsBool() bool {
	v := strings.ToLower(l.Value)
	return v == "true" || v == "1" || v == "yes" || v == "on" || v == "enable"
}

// AsPercentage 返回百分比的小数值（如 10% -> 0.1）
func (l LiteralValue) AsPercentage() float64 {
	if l.Value == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(l.Value, 64)
	if l.Unit == "%" || l.Type == Percentage {
		return v / 100
	}
	return v
}

// AsTime 将 Time 类型解析为 time.Time（当天的时间点）
func (l LiteralValue) AsTime() time.Time {
	if l.Value == "" {
		return time.Time{}
	}
	layouts := []string{"15:04", "15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, l.Value); err == nil {
			return t
		}
	}
	return time.Time{}
}

// AsDate 将 Date 类型解析为 time.Time
func (l LiteralValue) AsDate() time.Time {
	if l.Value == "" {
		return time.Time{}
	}
	layouts := []string{"2006-01-02", "2006/01/02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, l.Value); err == nil {
			return t
		}
	}
	return time.Time{}
}

// AsDuration 将 Duration 类型解析为 time.Duration
func (l LiteralValue) AsDuration() time.Duration {
	if l.Value == "" {
		return 0
	}
	val, err := strconv.ParseFloat(l.Value, 64)
	if err != nil {
		// 尝试直接按 Go duration 格式解析（如 "1h30m"）
		if d, err := time.ParseDuration(l.Value + l.Unit); err == nil {
			return d
		}
		return 0
	}
	switch l.Unit {
	case "ns":
		return time.Duration(val)
	case "us", "μs":
		return time.Duration(val * float64(time.Microsecond))
	case "ms":
		return time.Duration(val * float64(time.Millisecond))
	case "s":
		return time.Duration(val * float64(time.Second))
	case "m", "min":
		return time.Duration(val * float64(time.Minute))
	case "h", "H":
		return time.Duration(val * float64(time.Hour))
	case "d":
		return time.Duration(val * float64(time.Hour) * 24)
	case "W", "w":
		return time.Duration(val * float64(time.Hour) * 24 * 7)
	case "M":
		return time.Duration(val * float64(time.Hour) * 24 * 30)
	case "Y":
		return time.Duration(val * float64(time.Hour) * 24 * 365)
	default:
		if d, err := time.ParseDuration(l.Value + l.Unit); err == nil {
			return d
		}
		return time.Duration(val)
	}
}

// AsEnum 返回枚举字符串
func (l LiteralValue) AsEnum() string {
	return l.Value
}

// AsList 解析列表字面量（逗号分隔的简易实现）
func (l LiteralValue) AsList() []LiteralValue {
	if l.Value == "" {
		return nil
	}
	parts := strings.Split(l.Value, ",")
	result := make([]LiteralValue, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, LiteralValue{Type: Text, Value: p})
		}
	}
	return result
}

// IsZero 检查是否为零值
func (l LiteralValue) IsZero() bool {
	return l.Value == "" && l.Unit == "" && l.Type == Text
}

// Event 表示运行时事件，包含上下文状态与市场数据
type Event struct {
	Name      string
	Timestamp time.Time
	Price     float64
	Volume    float64
	Code      string
	Direction string
}

// MarketDataProvider 行情数据提供者接口
// Engine 在启动时将 gateway.TradingGateway 包装为此接口注入 StockContext
type MarketDataProvider interface {
	GetPrice(symbol string) float64
	GetVolume(symbol string) float64
	GetHigh(symbol string) float64
	GetLow(symbol string) float64
	GetClose(symbol string) float64
	GetOpen(symbol string) float64
}

// VariableResolver 变量解析器 —— 内置变量优先通过 MarketDataProvider 实时查询
type VariableResolver struct {
	mu sync.RWMutex
	// 手动注册的内置变量（优先级最高，用于测试或覆盖）
	Builtins map[string]func() float64
	// 用户自定义变量（define / value 语句）
	UserVars map[string]LiteralValue
	// 行情数据提供者（运行时自动查询 gateway）
	Provider MarketDataProvider
	// 当前标的代码，用于 Provider 查询
	Symbol string
}

func NewVariableResolver() *VariableResolver {
	return &VariableResolver{
		Builtins: make(map[string]func() float64),
		UserVars: make(map[string]LiteralValue),
	}
}

// Register 手动注册内置变量（优先级高于 Provider）
func (vr *VariableResolver) Register(name string, fn func() float64) {
	vr.mu.Lock()
	defer vr.mu.Unlock()
	vr.Builtins[name] = fn
}

// Resolve 解析变量值，查询优先级：Builtins > Provider > UserVars
func (vr *VariableResolver) Resolve(name string) float64 {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	// 1. 手动注册优先（允许覆盖或测试）
	if fn, ok := vr.Builtins[name]; ok {
		return fn()
	}

	// 2. 通过 Provider 实时查询行情
	if vr.Provider != nil && vr.Symbol != "" {
		switch name {
		case "price":
			return vr.Provider.GetPrice(vr.Symbol)
		case "volume":
			return vr.Provider.GetVolume(vr.Symbol)
		case "high":
			return vr.Provider.GetHigh(vr.Symbol)
		case "low":
			return vr.Provider.GetLow(vr.Symbol)
		case "close":
			return vr.Provider.GetClose(vr.Symbol)
		case "open":
			return vr.Provider.GetOpen(vr.Symbol)
		}
	}

	// 3. 用户自定义变量
	if lit, ok := vr.UserVars[name]; ok {
		return lit.AsFloat()
	}
	return 0
}

func (vr *VariableResolver) SetUserVar(name string, lit LiteralValue) {
	vr.mu.Lock()
	defer vr.mu.Unlock()
	vr.UserVars[name] = lit
}

// StockContext 运行时上下文 —— 由多个独立配置结构体组合而成
type StockContext struct {
	mu sync.RWMutex

	// 基础状态
	Start        time.Time
	BeginPrice   float64
	CurrentPrice float64 // 本地缓存价格（无 Provider 时的 fallback）
	Amount       float64
	Emiter       *utils.EventEmit

	// 组合配置（按领域拆分）
	Strategy  StrategyConfig
	Risk      RiskConfig
	Position  PositionConfig
	Leverage  LeverageConfig
	Execution ExecutionConfig
	Time      TimeConfig
	Adapter   AdapterConfig

	// 变量解析器（内置变量通过函数实时查询）
	Vars *VariableResolver

	// 事件回调
	stop func() bool
	keep func(duration int64) bool
}

// NewStockContext 创建新的执行上下文
func NewStockContext() *StockContext {
	ctx := &StockContext{
		Start:  time.Now(),
		Emiter: utils.NewEventEmit(),
		Vars:   NewVariableResolver(),
	}
	return ctx
}

// NewEvent 创建一个包含当前上下文状态的事件
func (s *StockContext) NewEvent(name string) *Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	price := s.Vars.Resolve("price")
	if price == 0 {
		price = s.CurrentPrice
	}
	return &Event{
		Name:      name,
		Timestamp: time.Now(),
		Price:     price,
		Volume:    s.Vars.Resolve("volume"),
		Code:      s.Strategy.Code,
		Direction: s.Strategy.Direction,
	}
}

// EmitEvent 发射一个结构化事件
func (s *StockContext) EmitEvent(name string) {
	s.Emiter.Emit(name, s.NewEvent(name))
}

// ---------- 基础状态 Getter/Setter ----------

func (s *StockContext) SetCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if code != "" {
		s.Strategy.Code = code
	}
}

func (s *StockContext) GetCode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Strategy.Code
}

func (s *StockContext) SetDirection(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Strategy.Direction = dir
}

func (s *StockContext) GetDirection() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Strategy.Direction
}

func (s *StockContext) SetBeginPrice(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BeginPrice = val
}

func (s *StockContext) GetBeginPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BeginPrice
}

func (s *StockContext) SetAmount(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Amount = val
}

func (s *StockContext) GetAmount() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Amount
}

func (s *StockContext) SetCurrentPrice(val float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentPrice = val
	return val
}

func (s *StockContext) GetCurrentPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// 优先通过 Provider 实时查询，无 Provider 时回退到本地缓存
	if s.Vars.Provider != nil && s.Vars.Symbol != "" {
		if price := s.Vars.Provider.GetPrice(s.Vars.Symbol); price > 0 {
			return price
		}
	}
	return s.CurrentPrice
}

// ---------- 交易动作 ----------

func (s *StockContext) Buy()       { s.EmitEvent("buy") }
func (s *StockContext) Sell()      { s.EmitEvent("sell") }
func (s *StockContext) SellShort() { s.EmitEvent("sell_short") }
func (s *StockContext) BuyCover()  { s.EmitEvent("buy_cover") }

// ---------- 配置领域 Getter/Setter ----------

func (s *StockContext) SetMaxPosition(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.MaxPosition = val
	s.Position.MaxPosition = val
}

func (s *StockContext) GetMaxPosition() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.MaxPosition
}

func (s *StockContext) SetStopLossRange(r *RangeValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.StopLossRange = r
}

func (s *StockContext) GetStopLossRange() *RangeValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.StopLossRange
}

func (s *StockContext) SetLeverage(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Leverage.Leverage = val
}

func (s *StockContext) GetLeverage() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Leverage.Leverage
}

func (s *StockContext) SetMarginMode(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Leverage.MarginMode = val
}

func (s *StockContext) GetMarginMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Leverage.MarginMode
}

func (s *StockContext) SetPositionType(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.Type = val
}

func (s *StockContext) GetPositionType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.Type
}

func (s *StockContext) SetSizingMethod(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.SizingMethod = val
}

func (s *StockContext) GetSizingMethod() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.SizingMethod
}

func (s *StockContext) SetCompoundProfit(val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Execution.CompoundProfit = val
}

func (s *StockContext) GetCompoundProfit() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Execution.CompoundProfit
}

func (s *StockContext) SetSkipIfGapped(val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Execution.SkipIfGapped = val
}

func (s *StockContext) GetSkipIfGapped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Execution.SkipIfGapped
}

func (s *StockContext) SetFallback(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Execution.Fallback = val
}

func (s *StockContext) GetFallback() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Execution.Fallback
}

func (s *StockContext) SetDuration(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.KeepDuration = d
}

func (s *StockContext) GetDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.KeepDuration
}

func (s *StockContext) SetHoldMax(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.HoldMax = d
}

func (s *StockContext) GetHoldMax() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.HoldMax
}

func (s *StockContext) SetCooldown(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.Cooldown = d
}

func (s *StockContext) GetCooldown() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.Cooldown
}

func (s *StockContext) SetCooldownPerLevel(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.CooldownPerLevel = v
}

func (s *StockContext) GetCooldownPerLevel() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.CooldownPerLevel
}

func (s *StockContext) SetSession(vals []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.Session = vals
}

func (s *StockContext) GetSession() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.Session
}

func (s *StockContext) SetActiveRange(vals []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.ActiveRange = vals
}

func (s *StockContext) GetActiveRange() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.ActiveRange
}

func (s *StockContext) SetPauseRange(vals []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Time.PauseRange = vals
}

func (s *StockContext) GetPauseRange() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Time.PauseRange
}

func (s *StockContext) SetBasePosition(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.BasePosition = val
}

func (s *StockContext) GetBasePosition() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.BasePosition
}

func (s *StockContext) SetPyramidStep(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.PyramidStep = val
}

func (s *StockContext) GetPyramidStep() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.PyramidStep
}

func (s *StockContext) SetMaxPyramidLayers(val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.MaxPyramidLayers = val
}

func (s *StockContext) GetMaxPyramidLayers() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.MaxPyramidLayers
}

func (s *StockContext) SetFixedFraction(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.FixedFraction = val
}

func (s *StockContext) GetFixedFraction() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.FixedFraction
}

func (s *StockContext) SetVolatilityTarget(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.VolatilityTarget = val
}

func (s *StockContext) GetVolatilityTarget() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.VolatilityTarget
}

func (s *StockContext) SetAtrPeriod(val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.AtrPeriod = val
}

func (s *StockContext) GetAtrPeriod() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.AtrPeriod
}

func (s *StockContext) SetRiskPerTrade(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.RiskPerTrade = val
}

func (s *StockContext) GetRiskPerTrade() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.RiskPerTrade
}

func (s *StockContext) SetRiskPerGrid(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.RiskPerGrid = val
}

func (s *StockContext) GetRiskPerGrid() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.RiskPerGrid
}

func (s *StockContext) SetMaxDrawdown(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.MaxDrawdown = val
}

func (s *StockContext) GetMaxDrawdown() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.MaxDrawdown
}

func (s *StockContext) SetPositionDecay(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.PositionDecay = val
}

func (s *StockContext) GetPositionDecay() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.PositionDecay
}

func (s *StockContext) SetPositionDecayPer(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.PositionDecayPer = d
}

func (s *StockContext) GetPositionDecayPer() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.PositionDecayPer
}

func (s *StockContext) SetGrossExposure(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.GrossExposure = val
}

func (s *StockContext) GetGrossExposure() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.GrossExposure
}

func (s *StockContext) SetNetExposure(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.NetExposure = val
}

func (s *StockContext) GetNetExposure() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.NetExposure
}

func (s *StockContext) SetBetaNeutral(val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.BetaNeutral = val
}

func (s *StockContext) GetBetaNeutral() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.BetaNeutral
}

func (s *StockContext) SetRebalance(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.Rebalance = val
}

func (s *StockContext) GetRebalance() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.Rebalance
}

func (s *StockContext) SetRebalanceAt(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Position.RebalanceAt = val
}

func (s *StockContext) GetRebalanceAt() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Position.RebalanceAt
}

func (s *StockContext) SetMaxShort(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.MaxShort = val
}

func (s *StockContext) GetMaxShort() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.MaxShort
}

func (s *StockContext) SetBorrowRateLimit(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.BorrowRateLimit = val
}

func (s *StockContext) GetBorrowRateLimit() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.BorrowRateLimit
}

func (s *StockContext) SetSlippageTolerance(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.SlippageTolerance = val
}

func (s *StockContext) GetSlippageTolerance() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.SlippageTolerance
}

func (s *StockContext) SetCircuitBreaker(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.CircuitBreaker = val
}

func (s *StockContext) GetCircuitBreaker() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.CircuitBreaker
}

func (s *StockContext) SetCircuitBreakerIn(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Risk.CircuitBreakerIn = d
}

func (s *StockContext) GetCircuitBreakerIn() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Risk.CircuitBreakerIn
}

func (s *StockContext) SetHedge(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Leverage.Hedge = val
}

func (s *StockContext) GetHedge() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Leverage.Hedge
}

func (s *StockContext) SetFundingPriority(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Leverage.FundingPriority = val
}

func (s *StockContext) GetFundingPriority() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Leverage.FundingPriority
}

func (s *StockContext) SetAdapter(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Adapter.Adapter = val
}

func (s *StockContext) GetAdapter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Adapter.Adapter
}

func (s *StockContext) SetAdapterMode(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Adapter.AdapterMode = val
}

func (s *StockContext) GetAdapterMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Adapter.AdapterMode
}

func (s *StockContext) SetAdapterConfig(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Adapter.AdapterConfig = val
}

func (s *StockContext) GetAdapterConfig() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Adapter.AdapterConfig
}

func (s *StockContext) SetAdapterSwitch(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Adapter.AdapterSwitch = val
}

func (s *StockContext) GetAdapterSwitch() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Adapter.AdapterSwitch
}

// ---------- 事件回调 ----------

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
			return s.Start.After(time.Now())
		}
		s.keep = defaultKeep
	}
	return s.keep(0), nil
}
