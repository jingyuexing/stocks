package transformer

import (
	"time"
)

// RangeValue 表示范围值，替代对 ast.RangeExprNode 的直接依赖
type RangeValue struct {
	Begin *LiteralValue
	End   *LiteralValue
}

// StrategyConfig 策略配置
type StrategyConfig struct {
	Code      string // 股票代码
	Direction string // grid / long / short / both / portfolio
}

// RiskConfig 风控配置
type RiskConfig struct {
	MaxPosition       float64
	StopLossRange     *RangeValue
	SlippageTolerance float64
	CircuitBreaker    float64
	CircuitBreakerIn  time.Duration
	MaxShort          float64
	BorrowRateLimit   float64
}

// PositionConfig 头寸管理配置
type PositionConfig struct {
	Type             string // fixed / dynamic
	SizingMethod     string
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
	PositionDecayPer time.Duration
	GrossExposure    float64
	NetExposure      float64
	BetaNeutral      bool
	Rebalance        string
	RebalanceAt      string
	MaxPosition      float64
}

// LeverageConfig 杠杆与保证金配置
type LeverageConfig struct {
	Leverage        float64
	MarginMode      string
	Hedge           string
	FundingPriority string
}

// ExecutionConfig 执行偏好配置
type ExecutionConfig struct {
	PartialFill    string
	CompoundProfit bool
	SkipIfGapped   bool
	Fallback       string
}

// TimeConfig 时间控制配置
type TimeConfig struct {
	KeepDuration     time.Duration
	HoldMax          time.Duration
	Cooldown         time.Duration
	CooldownPerLevel bool
	Session          []string
	ActiveRange      []string
	PauseRange       []string
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	Adapter       string
	AdapterMode   string
	AdapterConfig string
	AdapterSwitch string
}
