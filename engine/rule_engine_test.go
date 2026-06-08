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
	ctx.Vars.Register("price", func(ctx *transformer.StockContext) float64 { return 150.5 })
	ctx.Vars.Register("amount", func(ctx *transformer.StockContext) float64 { return 10000 })
	ctx.Vars.Register("profit", func(ctx *transformer.StockContext) float64 { return 50.5 })
	// v2.2 实时价格变量
	ctx.Vars.Register("high", func(ctx *transformer.StockContext) float64 { return 160 })
	ctx.Vars.Register("low", func(ctx *transformer.StockContext) float64 { return 140 })
	ctx.Vars.Register("volume", func(ctx *transformer.StockContext) float64 { return 1000000 })

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

// TestRuleEngine_BuiltinVariables 验证 NewStockContext 默认注册的内置变量
func TestRuleEngine_BuiltinVariables(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(150)
	ctx.SetVolume(10)
	ctx.SetAmount(5000)
	ctx.SetGridLevel(3)

	re := engine.NewRuleEngine(ctx)

	cases := []struct {
		name      string
		expected  float64
		tolerance float64
	}{
		{"profit", 500, 0.001},      // (150 - 100) * 10 = 500
		{"profit_per", 50, 0.001},   // (150 - 100) / 100 * 100 = 50%
		{"cost", 100, 0.001},        // 持仓成本价 = 建仓价格
		{"amount", 5000, 0.001},     // 可用资金
		{"position", 10, 0.001},     // 持仓数量
		{"begin_price", 100, 0.001}, // 建仓价格
		{"grid_level", 3, 0.001},    // 网格层级
		{"price", 150, 0.001},       // 当前价格（本地缓存）
	}

	for _, c := range cases {
		v, err := re.ResolveValue(&ast.VariableExprNode{Name: c.name})
		if err != nil {
			t.Errorf("$%s resolve error: %v", c.name, err)
			continue
		}
		if v < c.expected-c.tolerance || v > c.expected+c.tolerance {
			t.Errorf("$%s expected %f, got %f", c.name, c.expected, v)
		}
	}

	// $time 是当前时间戳，验证非零且合理
	v, _ := re.ResolveValue(&ast.VariableExprNode{Name: "time"})
	if v == 0 {
		t.Error("$time expected non-zero timestamp")
	}
}

// TestRuleEngine_VariableInCondition 验证 DSL 中 $variable 在条件表达式中求值
func TestRuleEngine_VariableInCondition(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.SetBeginPrice(100)
	ctx.SetCurrentPrice(150)
	ctx.SetVolume(10)

	re := engine.NewRuleEngine(ctx)

	// $profit > 400  => 500 > 400 => true
	expr := &ast.BinaryExprNode{
		Op:    ">",
		Left:  &ast.VariableExprNode{Name: "profit"},
		Right: &ast.Literal{Value: "400"},
	}
	ok, err := re.EvalCondition(expr)
	if err != nil {
		t.Fatalf("eval condition error: %v", err)
	}
	if !ok {
		t.Errorf("expected $profit > 400 to be true, got false")
	}

	// $profit_per > 60  => 50 > 60 => false
	expr2 := &ast.BinaryExprNode{
		Op:    ">",
		Left:  &ast.VariableExprNode{Name: "profit_per"},
		Right: &ast.Literal{Value: "60"},
	}
	ok2, _ := re.EvalCondition(expr2)
	if ok2 {
		t.Errorf("expected $profit_per > 60 to be false, got true")
	}

	// $price > 120  => 150 > 120 => true
	expr3 := &ast.BinaryExprNode{
		Op:    ">",
		Left:  &ast.VariableExprNode{Name: "price"},
		Right: &ast.Literal{Value: "120"},
	}
	ok3, _ := re.EvalCondition(expr3)
	if !ok3 {
		t.Errorf("expected $price > 120 to be true, got false")
	}
}

func TestRuleEngine_Comparison(t *testing.T) {
	ctx := transformer.NewStockContext()
	ctx.Vars.Register("price", func(ctx *transformer.StockContext) float64 { return 150 })
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
	ctx.Vars.Register("a", func(ctx *transformer.StockContext) float64 { return 10 })
	ctx.Vars.Register("b", func(ctx *transformer.StockContext) float64 { return 20 })
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
