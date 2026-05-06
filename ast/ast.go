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
	// stop 1W 持有止于一周 time
	// 1W 1周
	// 1D 1d
	// 1Y 年
	// 1M 月
	// 1s
	// stop 12% 持有止于加个增加到12%
	// stop 12% 持有止于加个增加到12%或减少到12% 绝对值 12%
	// stop -12% 持有止于减少到12%
	StopExpression

	SwapExpression

	RangeExpression
	// 定义 bar: name
	DefineStatement
	// Reference 引用
	// @bar
	ReferenceStatement

	// Select 选择股票代码
	// select AUL
	SelectExpression

	// Grid 网格交易
	// grid { loss 10% { sell 100% } profit 10% { sell 50% } }
	GridExpression
	// Loss 亏损分支
	LossExpression
	// Profit 盈利分支
	ProfitExpression
	// Value 资金配置
	// value: name 1000
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
