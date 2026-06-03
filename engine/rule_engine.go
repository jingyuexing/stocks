package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/transformer"
)

// RuleEngine 运行时规则求值引擎
type RuleEngine struct {
	ctx *transformer.StockContext
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(ctx *transformer.StockContext) *RuleEngine {
	return &RuleEngine{ctx: ctx}
}

// EvalCondition 对 AST 条件表达式求值
// metric 通常为当前 profit 或 price，由调用方根据策略语义传入
func (re *RuleEngine) EvalCondition(expr ast.ExpressionNode, metric float64) (bool, error) {
	switch expr.Type {
	case ast.GridExpression:
		// LevelBlock 以 GridExpression 形式存在，带 Range
		return re.evalRangeCondition(expr.Range, metric)
	case ast.LossExpression:
		return re.evalLossCondition(expr, metric)
	case ast.ProfitExpression:
		return re.evalProfitCondition(expr, metric)
	case ast.RangeExpression:
		// RangeExpression 通常通过 Range 字段承载，直接通过 ExpressionNode 处理此处留空
		return false, nil
	case ast.ComparisonExpression:
		return re.evalComparison(expr)
	case ast.LogicalExpression:
		return re.evalLogical(expr)
	default:
		return false, fmt.Errorf("unsupported condition type: %v", expr.Type)
	}
}

// evalRangeCondition 纯数学范围判断
func (re *RuleEngine) evalRangeCondition(r *ast.RangeExpressionNode, value float64) (bool, error) {
	if r == nil {
		return false, nil
	}
	hasBegin := r.Begin.Value != ""
	hasEnd := r.End.Value != ""
	if !hasBegin && !hasEnd {
		return false, nil
	}

	begin := r.Begin.AsFloat()
	end := r.End.AsFloat()

	if hasBegin && hasEnd {
		return value >= begin && value <= end, nil
	}
	if hasBegin && !hasEnd {
		// 开放上限，默认语义为 value >= begin
		return value >= begin, nil
	}
	// !hasBegin && hasEnd
	return value <= end, nil
}

// evalLossCondition loss 语义：profit <= threshold（亏损超过阈值）
func (re *RuleEngine) evalLossCondition(expr ast.ExpressionNode, profit float64) (bool, error) {
	if expr.Range != nil {
		return re.evalRangeCondition(expr.Range, profit)
	}
	if len(expr.Params) > 0 {
		threshold := expr.Params[0].AsFloat()
		return profit <= threshold, nil
	}
	return false, nil
}

// evalProfitCondition profit 语义：profit >= threshold（盈利超过阈值）
func (re *RuleEngine) evalProfitCondition(expr ast.ExpressionNode, profit float64) (bool, error) {
	if expr.Range != nil {
		return re.evalRangeCondition(expr.Range, profit)
	}
	if len(expr.Params) > 0 {
		threshold := expr.Params[0].AsFloat()
		return profit >= threshold, nil
	}
	return false, nil
}

// evalComparison 比较表达式求值（如 $price > 150）
func (re *RuleEngine) evalComparison(expr ast.ExpressionNode) (bool, error) {
	if expr.Left == nil || expr.Right == nil {
		return false, errors.New("comparison missing left or right operand")
	}
	leftVal, err := re.ResolveValue(*expr.Left)
	if err != nil {
		return false, err
	}
	rightVal, err := re.ResolveValue(*expr.Right)
	if err != nil {
		return false, err
	}

	switch expr.Operator {
	case "==":
		return leftVal == rightVal, nil
	case "!=":
		return leftVal != rightVal, nil
	case "<":
		return leftVal < rightVal, nil
	case ">":
		return leftVal > rightVal, nil
	case "<=":
		return leftVal <= rightVal, nil
	case ">=":
		return leftVal >= rightVal, nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", expr.Operator)
	}
}

// evalLogical 逻辑表达式求值（当前仅支持 && 和 ||）
func (re *RuleEngine) evalLogical(expr ast.ExpressionNode) (bool, error) {
	if expr.Left == nil || expr.Right == nil {
		return false, errors.New("logical expression missing operand")
	}
	leftVal, err := re.EvalCondition(*expr.Left, 0) // metric 在子比较表达式中通过变量解析获得
	if err != nil {
		return false, err
	}
	rightVal, err := re.EvalCondition(*expr.Right, 0)
	if err != nil {
		return false, err
	}

	switch expr.Operator {
	case "&&":
		return leftVal && rightVal, nil
	case "||":
		return leftVal || rightVal, nil
	default:
		return false, fmt.Errorf("unknown logical operator: %s", expr.Operator)
	}
}

// ResolveValue 将 AST 表达式解析为 float64
func (re *RuleEngine) ResolveValue(expr ast.ExpressionNode) (float64, error) {
	switch expr.Type {
	case ast.LiteralExpression:
		return expr.Value.AsFloat(), nil
	case ast.VariableExpression:
		return re.resolveVariable(expr.Name), nil
	case ast.IntegerLiteral, ast.FloatLiteral, ast.PercentLiteral:
		return expr.Value.AsFloat(), nil
	default:
		return 0, fmt.Errorf("cannot resolve value for type: %v", expr.Type)
	}
}

// resolveVariable 解析内置运行时变量
func (re *RuleEngine) resolveVariable(name string) float64 {
	if re.ctx == nil {
		return 0
	}
	switch name {
	case "price":
		return re.ctx.GetCurrentPrice()
	case "amount":
		return re.ctx.GetAmount()
	case "profit":
		return re.calcProfitPercent()
	case "high", "low", "open", "close", "volume":
		// 周期数据，暂由运行时外部注入（后续扩展）
		return 0
	case "position":
		// 当前持仓数量，后续在 StockContext 中扩展
		return 0
	case "cost":
		return re.ctx.GetBeginPrice()
	case "grid_level":
		return 0
	case "time":
		return float64(time.Now().Unix())
	default:
		// 尝试从 Variables 字典读取
		if val, ok := re.ctx.GetVariable(name); ok {
			return val.AsFloat()
		}
		return 0
	}
}

// calcProfitPercent 计算当前盈亏百分比
func (re *RuleEngine) calcProfitPercent() float64 {
	begin := re.ctx.GetBeginPrice()
	current := re.ctx.GetCurrentPrice()
	if begin <= 0 || current <= 0 {
		return 0
	}
	return (current - begin) / begin * 100
}
