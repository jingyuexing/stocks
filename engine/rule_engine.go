package engine

import (
	"errors"
	"fmt"
	"math"

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

// EvalCondition 对 AST 条件表达式求值（支持 binary / logical 表达式）
func (re *RuleEngine) EvalCondition(expr ast.Expr) (bool, error) {
	switch e := expr.(type) {
	case *ast.BinaryExprNode:
		if e.Op == "&&" || e.Op == "||" {
			return re.evalLogical(e)
		}
		return re.evalComparison(e)
	default:
		return false, fmt.Errorf("unsupported condition type: %T", expr)
	}
}

// EvalRange 对范围表达式求值，metric 为当前指标值（如盈亏百分比）
func (re *RuleEngine) EvalRange(r *ast.RangeExprNode, metric float64) bool {
	if r == nil {
		return false
	}
	begin := math.Inf(-1)
	end := math.Inf(1)
	if r.Begin != nil {
		begin = r.Begin.AsFloat()
	}
	if r.End != nil {
		end = r.End.AsFloat()
	}
	return metric >= begin && metric <= end
}

// evalComparison 比较表达式求值（如 $price > 150）
func (re *RuleEngine) evalComparison(expr *ast.BinaryExprNode) (bool, error) {
	if expr.Left == nil || expr.Right == nil {
		return false, errors.New("comparison missing left or right operand")
	}
	leftVal, err := re.ResolveValue(expr.Left)
	if err != nil {
		return false, err
	}
	rightVal, err := re.ResolveValue(expr.Right)
	if err != nil {
		return false, err
	}

	switch expr.Op {
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
		return false, fmt.Errorf("unknown comparison operator: %s", expr.Op)
	}
}

// evalLogical 逻辑表达式求值（支持 && 和 ||）
func (re *RuleEngine) evalLogical(expr *ast.BinaryExprNode) (bool, error) {
	if expr.Left == nil || expr.Right == nil {
		return false, errors.New("logical expression missing operand")
	}
	leftVal, err := re.EvalCondition(expr.Left)
	if err != nil {
		return false, err
	}
	rightVal, err := re.EvalCondition(expr.Right)
	if err != nil {
		return false, err
	}

	switch expr.Op {
	case "&&":
		return leftVal && rightVal, nil
	case "||":
		return leftVal || rightVal, nil
	default:
		return false, fmt.Errorf("unknown logical operator: %s", expr.Op)
	}
}

// ResolveValue 将 AST 表达式解析为 float64
func (re *RuleEngine) ResolveValue(expr ast.Expr) (float64, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		return e.AsFloat(), nil
	case *ast.IdentifierExprNode:
		// 标识符在数值上下文中视为 0
		return 0, nil
	case *ast.VariableExprNode:
		return re.resolveVariable(e.Name), nil
	case *ast.BinaryExprNode:
		return re.evalArithmetic(e)
	case *ast.UnaryExprNode:
		if e.Op == "-" {
			val, err := re.ResolveValue(e.Expr)
			return -val, err
		}
		return 0, fmt.Errorf("unsupported unary op: %s", e.Op)
	default:
		return 0, fmt.Errorf("cannot resolve value for type: %T", expr)
	}
}

// evalArithmetic 对算术表达式求值
func (re *RuleEngine) evalArithmetic(expr *ast.BinaryExprNode) (float64, error) {
	if expr.Left == nil || expr.Right == nil {
		return 0, errors.New("arithmetic expression missing operand")
	}
	leftVal, err := re.ResolveValue(expr.Left)
	if err != nil {
		return 0, err
	}
	rightVal, err := re.ResolveValue(expr.Right)
	if err != nil {
		return 0, err
	}

	switch expr.Op {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		if rightVal == 0 {
			return 0, errors.New("division by zero")
		}
		return leftVal / rightVal, nil
	default:
		return 0, fmt.Errorf("unknown arithmetic operator: %s", expr.Op)
	}
}

// resolveVariable 解析内置运行时变量
func (re *RuleEngine) resolveVariable(name string) float64 {
	if re.ctx == nil || re.ctx.Vars == nil {
		return 0
	}
	return re.ctx.Vars.Resolve(name)
}
