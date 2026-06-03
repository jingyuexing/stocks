package engine_test

import (
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/engine"
	"github.com/jingyuexing/stocks/transformer"
)

func TestRuleEngine_EvalRange(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// 闭区间 5...10
	r1 := &ast.RangeExpressionNode{
		Node:  ast.Node{Type: ast.RangeExpression},
		Begin: ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "5"},
		End:   ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"},
	}
	ok, _ := re.EvalCondition(ast.ExpressionNode{Node: ast.Node{Type: ast.RangeExpression}, Range: r1}, 7)
	if !ok {
		t.Error("expected 7 to be in range 5...10")
	}
	ok, _ = re.EvalCondition(ast.ExpressionNode{Node: ast.Node{Type: ast.RangeExpression}, Range: r1}, 3)
	if ok {
		t.Error("expected 3 to be outside range 5...10")
	}

	// 开放上限 10...
	r2 := &ast.RangeExpressionNode{
		Node:  ast.Node{Type: ast.RangeExpression},
		Begin: ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"},
	}
	ok, _ = re.EvalCondition(ast.ExpressionNode{Node: ast.Node{Type: ast.RangeExpression}, Range: r2}, 15)
	if !ok {
		t.Error("expected 15 to be in range 10...")
	}

	// 开放下限 ...10
	r3 := &ast.RangeExpressionNode{
		Node: ast.Node{Type: ast.RangeExpression},
		End:  ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"},
	}
	ok, _ = re.EvalCondition(ast.ExpressionNode{Node: ast.Node{Type: ast.RangeExpression}, Range: r3}, 5)
	if !ok {
		t.Error("expected 5 to be in range ...10")
	}
}

func TestRuleEngine_LossCondition(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// loss 10% -> 单点亏损阈值，profit <= -10 触发
	expr := ast.ExpressionNode{
		Node:   ast.Node{Type: ast.LossExpression},
		Params: []ast.Literal{{Node: ast.Node{Type: ast.FloatLiteral}, Value: "-10"}},
	}
	ok, _ := re.EvalCondition(expr, -12)
	if !ok {
		t.Error("expected loss -12% to trigger threshold -10%")
	}
	ok, _ = re.EvalCondition(expr, -5)
	if ok {
		t.Error("expected loss -5% NOT to trigger threshold -10%")
	}
}

func TestRuleEngine_ProfitCondition(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// profit 10% -> 单点盈利阈值，profit >= 10 触发
	expr := ast.ExpressionNode{
		Node:   ast.Node{Type: ast.ProfitExpression},
		Params: []ast.Literal{{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"}},
	}
	ok, _ := re.EvalCondition(expr, 12)
	if !ok {
		t.Error("expected profit 12% to trigger threshold 10%")
	}
	ok, _ = re.EvalCondition(expr, 5)
	if ok {
		t.Error("expected profit 5% NOT to trigger threshold 10%")
	}
}

func TestRuleEngine_ProfitRange(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// profit 5%...10%
	expr := ast.ExpressionNode{
		Node: ast.Node{Type: ast.ProfitExpression},
		Range: &ast.RangeExpressionNode{
			Node:  ast.Node{Type: ast.RangeExpression},
			Begin: ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "5"},
			End:   ast.Literal{Node: ast.Node{Type: ast.FloatLiteral}, Value: "10"},
		},
	}
	ok, _ := re.EvalCondition(expr, 7)
	if !ok {
		t.Error("expected profit 7% to be in range 5%...10%")
	}
	ok, _ = re.EvalCondition(expr, 3)
	if ok {
		t.Error("expected profit 3% to be outside range")
	}
}

func TestRuleEngine_ResolveVariable(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetCurrentPrice(150.5)
	ctx.SetAmount(10000)
	ctx.SetBeginPrice(100)

	re := engine.NewRuleEngine(ctx)

	// $price
	v, _ := re.ResolveValue(ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "price"})
	if v != 150.5 {
		t.Errorf("expected price 150.5, got %f", v)
	}

	// $amount
	v, _ = re.ResolveValue(ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "amount"})
	if v != 10000 {
		t.Errorf("expected amount 10000, got %f", v)
	}

	// $profit = (150.5 - 100) / 100 * 100 = 50.5%
	v, _ = re.ResolveValue(ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "profit"})
	if v != 50.5 {
		t.Errorf("expected profit 50.5, got %f", v)
	}
}

func TestRuleEngine_CalcProfitPercent(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// begin=0 时应返回 0
	ctx.SetBeginPrice(0)
	ctx.SetCurrentPrice(150)
	if v, _ := re.ResolveValue(ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "profit"}); v != 0 {
		t.Errorf("expected 0 profit when beginPrice=0, got %f", v)
	}

	// 正常计算
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(120)
	v, _ := re.ResolveValue(ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "profit"})
	if v != 20 {
		t.Errorf("expected profit 20, got %f", v)
	}
}

func TestRuleEngine_Comparison(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetCurrentPrice(150)
	re := engine.NewRuleEngine(ctx)

	// $price > 100
	expr := ast.ExpressionNode{
		Node:     ast.Node{Type: ast.ComparisonExpression},
		Operator: ">",
		Left:     &ast.ExpressionNode{Node: ast.Node{Type: ast.VariableExpression}, Name: "price"},
		Right:    &ast.ExpressionNode{Node: ast.Node{Type: ast.LiteralExpression}, Value: ast.Literal{Node: ast.Node{Type: ast.IntegerLiteral}, Value: "100"}},
	}
	ok, _ := re.EvalCondition(expr, 0)
	if !ok {
		t.Error("expected 150 > 100 to be true")
	}

	// $price <= 100
	expr.Operator = "<="
	ok, _ = re.EvalCondition(expr, 0)
	if ok {
		t.Error("expected 150 <= 100 to be false")
	}
}
