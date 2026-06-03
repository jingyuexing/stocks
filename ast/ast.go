package ast

import (
	"fmt"
	"strconv"
)

type NodeType uint

const (
	// the start of program
	Program NodeType = iota + 1

	FloatLiteral
	IntegerLiteral
	TextLiteral
	DurationLiteral
	PercentLiteral
	ReferenceLiteral
	LiteralExpression

	// 买入
	BuyExpression
	// 卖出
	SellExpression
	// 持有
	KeepExpression
	// 止于某时或某值
	StopExpression

	SwapExpression

	RangeExpression
	// 定义 bar: name
	DefineStatement
	// Reference 引用
	ReferenceStatement
	// Variable 变量（如 $profit, $price）
	VariableExpression

	// Select 选择股票代码
	SelectExpression

	// Grid 网格交易
	GridExpression
	// Loss 亏损分支
	LossExpression
	// Profit 盈利分支
	ProfitExpression
	// Value 资金配置
	ValueStatement

	// a year
	YearUnit
	// a month
	MonthUnit
	// day
	DayUnit
	///
	WeekUnit
	///  hour
	HourUnit
	// min
	MinuteUnit
	// second
	SecondUnit
	// milion second
	MillisecondUnit

	// ========== v2.1 新增 ==========
	// 策略声明
	LongExpression
	ShortExpression
	BothExpression
	PortfolioExpression
	TemplateExpression
	UseExpression
	ImportExpression
	ExportExpression
	MacroExpression
	ExtendsExpression
	ParamExpression
	ConditionalExpression
	AssertExpression

	// 范围与优先级
	CrossExpression
	PriorityExpression
	OverrideExpression

	// 时间控制
	HoldMaxExpression
	CooldownExpression
	SessionExpression
	ActiveExpression
	PauseExpression

	// 交易动作
	SellShortExpression
	BuyCoverExpression

	// 头寸管理
	PositionExpression
	SizingExpression
	BasePositionExpression
	PyramidStepExpression
	MaxPyramidLayersExpression
	FixedFractionExpression
	VolatilityTargetExpression
	AtrPeriodExpression
	RiskPerTradeExpression
	RiskPerGridExpression
	MaxDrawdownExpression
	PositionDecayExpression
	GrossExposureExpression
	NetExposureExpression
	BetaNeutralExpression
	RebalanceExpression

	// 杠杆与保证金
	LeverageExpression
	MarginExpression
	HedgeExpression
	FundingPriorityExpression
	MaxShortExpression
	BorrowRateLimitExpression

	// 风控
	MaxPositionExpression
	StopLossExpression
	SlippageToleranceExpression
	PartialFillExpression
	CircuitBreakerExpression

	// 执行偏好
	CompoundProfitExpression
	SkipIfGappedExpression
	FallbackExpression

	// 新字面量
	StringLiteral
	BooleanLiteral
	SymbolLiteral
	TimePointLiteral
	DateLiteral
	DateRangeLiteral
	SessionRangeLiteral

	// 表达式
	ArithExpression
	ComparisonExpression
	LogicalExpression
	MacroExpandExpression

	// 组合配置
	PortfolioHeatExpression
	CorrelationLimitExpression

	// 文档注释 (通用)
	AnnotationExpression

	// unknow node
	Unknown
)

func (t NodeType) String() string {
	switch t {
	case FloatLiteral:
		return "FloatLiteral"
	case IntegerLiteral:
		return "IntegerLiteral"
	case TextLiteral:
		return "TextLiteral"
	case BuyExpression:
		return "BuyExpression"
	case SellExpression:
		return "SellExpression"
	case KeepExpression:
		return "KeepExpression"
	case StopExpression:
		return "StopExpression"
	case RangeExpression:
		return "RangeExpression"
	case DefineStatement:
		return "DefineStatement"
	case ReferenceStatement:
		return "ReferenceStatement"
	case VariableExpression:
		return "VariableExpression"
	case LiteralExpression:
		return "LiteralExpression"
	case SelectExpression:
		return "SelectExpression"
	case GridExpression:
		return "GridExpression"
	case LossExpression:
		return "LossExpression"
	case ProfitExpression:
		return "ProfitExpression"
	case ValueStatement:
		return "ValueStatement"
	case YearUnit:
		return "YearUnit"
	case MonthUnit:
		return "MonthUnit"
	case DayUnit:
		return "DayUnit"
	case HourUnit:
		return "HourUnit"
	case MinuteUnit:
		return "MinuteUnit"
	case SecondUnit:
		return "SecondUnit"
	case DurationLiteral:
		return "DurationLiteral"
	case MillisecondUnit:
		return "MillisecondUnit"
	case WeekUnit:
		return "WeekUnit"
	// v2.1
	case LongExpression:
		return "LongExpression"
	case ShortExpression:
		return "ShortExpression"
	case BothExpression:
		return "BothExpression"
	case PortfolioExpression:
		return "PortfolioExpression"
	case TemplateExpression:
		return "TemplateExpression"
	case UseExpression:
		return "UseExpression"
	case ImportExpression:
		return "ImportExpression"
	case ExportExpression:
		return "ExportExpression"
	case MacroExpression:
		return "MacroExpression"
	case ExtendsExpression:
		return "ExtendsExpression"
	case ParamExpression:
		return "ParamExpression"
	case ConditionalExpression:
		return "ConditionalExpression"
	case AssertExpression:
		return "AssertExpression"
	case CrossExpression:
		return "CrossExpression"
	case PriorityExpression:
		return "PriorityExpression"
	case OverrideExpression:
		return "OverrideExpression"
	case HoldMaxExpression:
		return "HoldMaxExpression"
	case CooldownExpression:
		return "CooldownExpression"
	case SessionExpression:
		return "SessionExpression"
	case ActiveExpression:
		return "ActiveExpression"
	case PauseExpression:
		return "PauseExpression"
	case SellShortExpression:
		return "SellShortExpression"
	case BuyCoverExpression:
		return "BuyCoverExpression"
	case PositionExpression:
		return "PositionExpression"
	case SizingExpression:
		return "SizingExpression"
	case BasePositionExpression:
		return "BasePositionExpression"
	case PyramidStepExpression:
		return "PyramidStepExpression"
	case MaxPyramidLayersExpression:
		return "MaxPyramidLayersExpression"
	case FixedFractionExpression:
		return "FixedFractionExpression"
	case VolatilityTargetExpression:
		return "VolatilityTargetExpression"
	case AtrPeriodExpression:
		return "AtrPeriodExpression"
	case RiskPerTradeExpression:
		return "RiskPerTradeExpression"
	case RiskPerGridExpression:
		return "RiskPerGridExpression"
	case MaxDrawdownExpression:
		return "MaxDrawdownExpression"
	case PositionDecayExpression:
		return "PositionDecayExpression"
	case GrossExposureExpression:
		return "GrossExposureExpression"
	case NetExposureExpression:
		return "NetExposureExpression"
	case BetaNeutralExpression:
		return "BetaNeutralExpression"
	case RebalanceExpression:
		return "RebalanceExpression"
	case LeverageExpression:
		return "LeverageExpression"
	case MarginExpression:
		return "MarginExpression"
	case HedgeExpression:
		return "HedgeExpression"
	case FundingPriorityExpression:
		return "FundingPriorityExpression"
	case MaxShortExpression:
		return "MaxShortExpression"
	case BorrowRateLimitExpression:
		return "BorrowRateLimitExpression"
	case MaxPositionExpression:
		return "MaxPositionExpression"
	case StopLossExpression:
		return "StopLossExpression"
	case SlippageToleranceExpression:
		return "SlippageToleranceExpression"
	case PartialFillExpression:
		return "PartialFillExpression"
	case CircuitBreakerExpression:
		return "CircuitBreakerExpression"
	case CompoundProfitExpression:
		return "CompoundProfitExpression"
	case SkipIfGappedExpression:
		return "SkipIfGappedExpression"
	case FallbackExpression:
		return "FallbackExpression"
	case StringLiteral:
		return "StringLiteral"
	case BooleanLiteral:
		return "BooleanLiteral"
	case SymbolLiteral:
		return "SymbolLiteral"
	case TimePointLiteral:
		return "TimePointLiteral"
	case DateLiteral:
		return "DateLiteral"
	case DateRangeLiteral:
		return "DateRangeLiteral"
	case SessionRangeLiteral:
		return "SessionRangeLiteral"
	case ArithExpression:
		return "ArithExpression"
	case ComparisonExpression:
		return "ComparisonExpression"
	case LogicalExpression:
		return "LogicalExpression"
	case MacroExpandExpression:
		return "MacroExpandExpression"
	case PortfolioHeatExpression:
		return "PortfolioHeatExpression"
	case CorrelationLimitExpression:
		return "CorrelationLimitExpression"
	case AnnotationExpression:
		return "AnnotationExpression"
	default:
		return "unknown"
	}
}

type Node struct {
	Type NodeType
}

type Literal struct {
	Node
	Value string
	Unit  string
}

type RangeExpressionNode struct {
	Node
	Begin Literal
	End   Literal
}

type ExpressionNode struct {
	Node
	Name string
	// 关键字后面的数据视为参数
	Params []Literal
	// 范围
	Range *RangeExpressionNode

	// value 定义的值
	Value Literal

	// Body 用于块级表达式（如 grid、loss、profit）
	Body []ExpressionNode

	// v2.1 扩展字段
	// Alias 用于 use ... as alias
	Alias string
	// Operator 用于算术/比较/逻辑表达式
	Operator string
	// Left / Right 用于二元表达式（复用 Body 也可，但独立字段更清晰）
	Left  *ExpressionNode
	Right *ExpressionNode
}

type RootNode struct {
	Node
	Expression []ExpressionNode
}

func (literal Literal) String() string {
	switch literal.Type {
	case FloatLiteral:
		if literal.Unit == "%" {
			value, _ := strconv.ParseFloat(literal.Value, 64)
			return fmt.Sprintf("%f", value/100)
		}
		return literal.Value
	case IntegerLiteral:
		return literal.Value
	default:
		return ""
	}
}

// AsFloat 将字面量解析为 float64
func (literal Literal) AsFloat() float64 {
	v, _ := strconv.ParseFloat(literal.Value, 64)
	return v
}

// AsInt 将字面量解析为 int
func (literal Literal) AsInt() int {
	v, _ := strconv.Atoi(literal.Value)
	return v
}

// AsBool 将字面量解析为 bool
func (literal Literal) AsBool() bool {
	return literal.Value == "true"
}
