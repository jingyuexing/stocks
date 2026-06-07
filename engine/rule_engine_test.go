package engine_test

import (
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/engine"
	"github.com/jingyuexing/stocks/transformer"
)

func TestRuleEngine_ResolveVariable(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(150.5)
	ctx.SetAmount(10000)

	// 注册内置变量函数（避免在回调中调用 GetCurrentPrice/GetAmount 造成递归）
	ctx.Vars.Register("price", func() float64 { return 150.5 })
	ctx.Vars.Register("amount", func() float64 { return 10000 })
	ctx.Vars.Register("profit", func() float64 { return 50.5 })
	// v2.2 实时价格变量
	ctx.Vars.Register("high", func() float64 { return 160 })
	ctx.Vars.Register("low", func() float64 { return 140 })
	ctx.Vars.Register("volume", func() float64 { return 1000000 })

	re := engine.NewRuleEngine(ctx)

	// $price
	v, _ := re.ResolveValue(&ast.VariableExprNode{Name: "price"})
	if v != 150.5 {
		t.Errorf("expected price 150.5, got %f", v)
	}

	// $amount
	v, _ = re.ResolveValue(&ast.VariableExprNode{Name: "amount"})
	if v != 10000 {
		t.Errorf("expected amount 10000, got %f", v)
	}

	// $profit = (150.5 - 100) / 100 * 100 = 50.5%
	v, _ = re.ResolveValue(&ast.VariableExprNode{Name: "profit"})
	if v != 50.5 {
		t.Errorf("expected profit 50.5, got %f", v)
	}

	// $high
	v, _ = re.ResolveValue(&ast.VariableExprNode{Name: "high"})
	if v != 160 {
		t.Errorf("expected high 160, got %f", v)
	}

	// $volume
	v, _ = re.ResolveValue(&ast.VariableExprNode{Name: "volume"})
	if v != 1000000 {
		t.Errorf("expected volume 1000000, got %f", v)
	}
}

func TestRuleEngine_Comparison(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.Vars.Register("price", func() float64 { return 150 })
	re := engine.NewRuleEngine(ctx)

	// $price > 100
	expr := &ast.BinaryExprNode{
		Op:    ">",
		Left:  &ast.VariableExprNode{Name: "price"},
		Right: &ast.Literal{Value: "100"},
	}
	ok, _ := re.EvalCondition(expr)
	if !ok {
		t.Error("expected 150 > 100 to be true")
	}

	// $price <= 100
	expr.Op = "<="
	ok, _ = re.EvalCondition(expr)
	if ok {
		t.Error("expected 150 <= 100 to be false")
	}
}

func TestRuleEngine_Logical(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.Vars.Register("a", func() float64 { return 10 })
	ctx.Vars.Register("b", func() float64 { return 20 })
	re := engine.NewRuleEngine(ctx)

	// $a > 5 && $b < 30
	expr := &ast.BinaryExprNode{
		Op: "&&",
		Left: &ast.BinaryExprNode{
			Op:    ">",
			Left:  &ast.VariableExprNode{Name: "a"},
			Right: &ast.Literal{Value: "5"},
		},
		Right: &ast.BinaryExprNode{
			Op:    "<",
			Left:  &ast.VariableExprNode{Name: "b"},
			Right: &ast.Literal{Value: "30"},
		},
	}
	ok, _ := re.EvalCondition(expr)
	if !ok {
		t.Error("expected (10 > 5 && 20 < 30) to be true")
	}
}

func TestRuleEngine_Arithmetic(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// 100 * 2 + 50 = 250
	expr := &ast.BinaryExprNode{
		Op: "+",
		Left: &ast.BinaryExprNode{
			Op:    "*",
			Left:  &ast.Literal{Value: "100"},
			Right: &ast.Literal{Value: "2"},
		},
		Right: &ast.Literal{Value: "50"},
	}
	v, _ := re.ResolveValue(expr)
	if v != 250 {
		t.Errorf("expected 250, got %f", v)
	}
}

func TestRuleEngine_UserVariable(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.Vars.SetUserVar("myVar", transformer.LiteralValue{Value: "42"})
	re := engine.NewRuleEngine(ctx)

	v, _ := re.ResolveValue(&ast.VariableExprNode{Name: "myVar"})
	if v != 42 {
		t.Errorf("expected 42, got %f", v)
	}
}

func TestRuleEngine_EvalRange(t *testing.T) {
	ctx := transformer.NewStockContext()
	re := engine.NewRuleEngine(ctx)

	// Range -15 ... -10, metric -12 => hit
	r := &ast.RangeExprNode{
		Begin: &ast.Literal{Value: "-15"},
		End:   &ast.Literal{Value: "-10"},
	}
	if !re.EvalRange(r, -12) {
		t.Error("expected -12 to be in range -15...-10")
	}

	// Range -15 ... -10, metric -5 => miss
	if re.EvalRange(r, -5) {
		t.Error("expected -5 to NOT be in range -15...-10")
	}

	// Open-ended begin: ...10, metric 5 => hit
	r2 := &ast.RangeExprNode{
		End: &ast.Literal{Value: "10"},
	}
	if !re.EvalRange(r2, 5) {
		t.Error("expected 5 to be in open-ended range ...10")
	}

	// Open-ended begin: -10..., metric -5 => hit
	r3 := &ast.RangeExprNode{
		Begin: &ast.Literal{Value: "-10"},
	}
	if !re.EvalRange(r3, -5) {
		t.Error("expected -5 to be in open-ended range -10...")
	}
	if re.EvalRange(r3, -15) {
		t.Error("expected -15 to NOT be in open-ended range -10...")
	}
}
