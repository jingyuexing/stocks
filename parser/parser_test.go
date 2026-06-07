package parser

import (
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
)

// ---------- helpers ----------
func parse(input string) *ast.ProgramNode {
	tokens := tokenizer.Lexer(input)
	return Parse(tokens)
}

func mustLen(t *testing.T, prog *ast.ProgramNode, n int) {
	t.Helper()
	if len(prog.Statements) != n {
		t.Fatalf("expected %d statements, got %d", n, len(prog.Statements))
	}
}

func mustType(t *testing.T, stmt ast.Stmt, typ ast.NodeType) {
	t.Helper()
	if stmt == nil {
		t.Fatal("expected non-nil stmt")
	}
	switch stmt.(type) {
	case *ast.StrategyStmtNode:
		if ast.StrategyStmt != typ {
			t.Fatalf("expected %v, got StrategyStmt", typ)
		}
	case *ast.FlowStmtNode:
		if ast.FlowStmt != typ {
			t.Fatalf("expected %v, got FlowStmt", typ)
		}
	case *ast.ActionStmtNode:
		if ast.ActionStmt != typ {
			t.Fatalf("expected %v, got ActionStmt", typ)
		}
	case *ast.ConfigStmtNode:
		if ast.ConfigStmt != typ {
			t.Fatalf("expected %v, got ConfigStmt", typ)
		}
	case *ast.ImportStmtNode:
		if ast.ImportStmt != typ {
			t.Fatalf("expected %v, got ImportStmt", typ)
		}
	case *ast.ExportStmtNode:
		if ast.ExportStmt != typ {
			t.Fatalf("expected %v, got ExportStmt", typ)
		}
	case *ast.UseStmtNode:
		if ast.UseStmt != typ {
			t.Fatalf("expected %v, got UseStmt", typ)
		}
	case *ast.TemplateStmtNode:
		if ast.TemplateStmt != typ {
			t.Fatalf("expected %v, got TemplateStmt", typ)
		}
	case *ast.MacroStmtNode:
		if ast.MacroStmt != typ {
			t.Fatalf("expected %v, got MacroStmt", typ)
		}
	case *ast.IfStmtNode:
		if ast.IfStmt != typ {
			t.Fatalf("expected %v, got IfStmt", typ)
		}
	case *ast.AssertStmtNode:
		if ast.AssertStmt != typ {
			t.Fatalf("expected %v, got AssertStmt", typ)
		}
	case *ast.DefineStmtNode:
		if ast.DefineStmt != typ {
			t.Fatalf("expected %v, got DefineStmt", typ)
		}
	case *ast.SelectStmtNode:
		if ast.SelectStmt != typ {
			t.Fatalf("expected %v, got SelectStmt", typ)
		}
	case *ast.AnnotationStmtNode:
		if ast.AnnotationStmt != typ {
			t.Fatalf("expected %v, got AnnotationStmt", typ)
		}
	default:
		t.Fatalf("unexpected stmt type %T", stmt)
	}
}

func mustLiteral(t *testing.T, expr ast.Expr, value string, unit string) *ast.Literal {
	t.Helper()
	lit, ok := expr.(*ast.Literal)
	if !ok {
		t.Fatalf("expected *Literal, got %T", expr)
	}
	if lit.Value != value || lit.Unit != unit {
		t.Fatalf("expected Literal{Value:%s Unit:%s}, got {Value:%s Unit:%s}", value, unit, lit.Value, lit.Unit)
	}
	return lit
}

func mustIdent(t *testing.T, expr ast.Expr, name string) {
	t.Helper()
	id, ok := expr.(*ast.IdentifierExprNode)
	if !ok {
		t.Fatalf("expected *IdentifierExprNode, got %T", expr)
	}
	if id.Name != name {
		t.Fatalf("expected identifier %s, got %s", name, id.Name)
	}
}

func mustListLiteral(t *testing.T, expr ast.Expr, expectedLen int) *ast.ListLiteralNode {
	t.Helper()
	list, ok := expr.(*ast.ListLiteralNode)
	if !ok {
		t.Fatalf("expected *ListLiteralNode, got %T", expr)
	}
	if len(list.Items) != expectedLen {
		t.Fatalf("expected ListLiteralNode with %d items, got %d", expectedLen, len(list.Items))
	}
	return list
}

// ---------- 顶层语句 ----------
func TestParseEmpty(t *testing.T) {
	prog := parse("")
	mustLen(t, prog, 0)
}

func TestParseImport(t *testing.T) {
	prog := parse(`import "path/to/file";`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.ImportStmt)
	stmt := prog.Statements[0].(*ast.ImportStmtNode)
	if stmt.Path != "path/to/file" {
		t.Fatalf("expected path 'path/to/file', got %s", stmt.Path)
	}
}

func TestParseExport(t *testing.T) {
	prog := parse(`export foo;`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.ExportStmt)
	stmt := prog.Statements[0].(*ast.ExportStmtNode)
	if stmt.Name != "foo" {
		t.Fatalf("expected name 'foo', got %s", stmt.Name)
	}
}

func TestParseUse(t *testing.T) {
	prog := parse(`use Template(100, "arg") as T;`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.UseStmt)
	stmt := prog.Statements[0].(*ast.UseStmtNode)
	if stmt.Name != "Template" {
		t.Fatalf("expected 'Template', got %s", stmt.Name)
	}
	if stmt.Alias != "T" {
		t.Fatalf("expected alias 'T', got %s", stmt.Alias)
	}
	if len(stmt.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(stmt.Args))
	}
}

func TestParseSelect(t *testing.T) {
	prog := parse(`select AAPL`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.SelectStmt)
	stmt := prog.Statements[0].(*ast.SelectStmtNode)
	if stmt.Target != "AAPL" {
		t.Fatalf("expected target 'AAPL', got %s", stmt.Target)
	}
}

func TestParseSelectNumericSymbol(t *testing.T) {
	prog := parse(`select 10023`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.SelectStmtNode)
	if stmt.Target != "10023" {
		t.Fatalf("expected target '10023', got %s", stmt.Target)
	}
}

// ---------- Strategy ----------
func TestParseGrid(t *testing.T) {
	prog := parse(`grid AAPL { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.StrategyStmt)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "grid" {
		t.Fatalf("expected kind 'grid', got %s", stmt.Kind)
	}
	if stmt.Target != "AAPL" {
		t.Fatalf("expected target 'AAPL', got %s", stmt.Target)
	}
	if len(stmt.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(stmt.Body))
	}
	mustType(t, stmt.Body[0], ast.ActionStmt)
}

func TestParseGridNumericSymbol(t *testing.T) {
	prog := parse(`grid 10023 { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.StrategyStmt)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "grid" {
		t.Fatalf("expected kind 'grid', got %s", stmt.Kind)
	}
	if stmt.Target != "10023" {
		t.Fatalf("expected target '10023', got %s", stmt.Target)
	}
	if len(stmt.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(stmt.Body))
	}
}

func TestParseLong(t *testing.T) {
	prog := parse(`long BTC { buy 1; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "long" {
		t.Fatalf("expected 'long', got %s", stmt.Kind)
	}
	if stmt.Target != "BTC" {
		t.Fatalf("expected target 'BTC', got %s", stmt.Target)
	}
}

func TestParseShort(t *testing.T) {
	prog := parse(`short ETH { sell 1; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "short" {
		t.Fatalf("expected 'short', got %s", stmt.Kind)
	}
}

func TestParseBoth(t *testing.T) {
	input := `
	both BTC {
		long {
			buy 100;
		}
		short {
			sell 100;
		}
		max_position 1000 shares;
	}
	`
	prog := parse(input)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "both" {
		t.Fatalf("expected 'both', got %s", stmt.Kind)
	}
	if stmt.Target != "BTC" {
		t.Fatalf("expected target 'BTC', got %s", stmt.Target)
	}
	// body 包含 long, short, max_position（顺序取决于源码）
	if len(stmt.Body) != 3 {
		t.Fatalf("expected 3 body stmts, got %d", len(stmt.Body))
	}
}

func TestParsePortfolio(t *testing.T) {
	input := `
	portfolio MyPort {
		grid AAPL { buy 10; }
		grid TSLA { sell 5; }
	}
	`
	prog := parse(input)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.StrategyStmtNode)
	if stmt.Kind != "portfolio" {
		t.Fatalf("expected 'portfolio', got %s", stmt.Kind)
	}
	if len(stmt.Body) != 2 {
		t.Fatalf("expected 2 body stmts, got %d", len(stmt.Body))
	}
}

// ---------- Flow ----------
func TestParseLevelProfit(t *testing.T) {
	prog := parse(`profit 10%...20% { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.FlowStmt)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "level" {
		t.Fatalf("expected 'level', got %s", stmt.Kind)
	}
	if stmt.Mode != "profit" {
		t.Fatalf("expected mode 'profit', got %s", stmt.Mode)
	}
	if stmt.Condition == nil {
		t.Fatal("expected non-nil condition")
	}
	rng, ok := stmt.Condition.(*ast.RangeExprNode)
	if !ok {
		t.Fatalf("expected RangeExprNode, got %T", stmt.Condition)
	}
	mustLiteral(t, rng.Begin, "10", "%")
	mustLiteral(t, rng.End, "20", "%")
	if len(stmt.Body) != 1 {
		t.Fatalf("expected 1 action in body, got %d", len(stmt.Body))
	}
}

func TestParseLevelLossCross(t *testing.T) {
	prog := parse(`loss -10% cross { buy 100; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Mode != "loss_cross" {
		t.Fatalf("expected mode 'loss_cross', got %s", stmt.Mode)
	}
}

func TestParseLevelPrice(t *testing.T) {
	prog := parse(`price 10%...20% { buy 50; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Mode != "price" {
		t.Fatalf("expected mode 'price', got %s", stmt.Mode)
	}
	rng := stmt.Condition.(*ast.RangeExprNode)
	mustLiteral(t, rng.Begin, "10", "%")
	mustLiteral(t, rng.End, "20", "%")
}

func TestParseTrigger(t *testing.T) {
	prog := parse(`trigger price > 100 { buy 10; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.FlowStmt)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "trigger" {
		t.Fatalf("expected 'trigger', got %s", stmt.Kind)
	}
	bin, ok := stmt.Condition.(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Condition)
	}
	if bin.Op != ">" {
		t.Fatalf("expected op '>', got %s", bin.Op)
	}
	if stmt.Frequency != "" {
		t.Fatalf("expected empty frequency, got %s", stmt.Frequency)
	}
}

func TestParseTriggerOnce(t *testing.T) {
	prog := parse(`once trigger price > 100 { buy 10; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "trigger" {
		t.Fatalf("expected 'trigger', got %s", stmt.Kind)
	}
	if stmt.Frequency != "once" {
		t.Fatalf("expected frequency 'once', got %s", stmt.Frequency)
	}
	bin, ok := stmt.Condition.(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Condition)
	}
	if bin.Op != ">" {
		t.Fatalf("expected op '>', got %s", bin.Op)
	}
}

func TestParseTriggerMonthly(t *testing.T) {
	prog := parse(`monthly trigger $signal == true { buy 100; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Frequency != "monthly" {
		t.Fatalf("expected frequency 'monthly', got %s", stmt.Frequency)
	}
}

func TestParseTriggerTwice(t *testing.T) {
	prog := parse(`twice trigger price < 50 { sell 20; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Frequency != "twice" {
		t.Fatalf("expected frequency 'twice', got %s", stmt.Frequency)
	}
}

func TestParseTriggerDaily(t *testing.T) {
	prog := parse(`daily trigger volume > 1M { buy 10; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Frequency != "daily" {
		t.Fatalf("expected frequency 'daily', got %s", stmt.Frequency)
	}
}

func TestParseUntil(t *testing.T) {
	prog := parse(`until profit >= 20% { sell 50; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "until" {
		t.Fatalf("expected 'until', got %s", stmt.Kind)
	}
}

func TestParseAtomicRollback(t *testing.T) {
	prog := parse(`atomic { buy 100; sell 50; } rollback;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "atomic" {
		t.Fatalf("expected 'atomic', got %s", stmt.Kind)
	}
	if stmt.Mode != "rollback" {
		t.Fatalf("expected mode 'rollback', got %s", stmt.Mode)
	}
	if len(stmt.Body) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(stmt.Body))
	}
}

func TestParseAtomicBestEffort(t *testing.T) {
	prog := parse(`atomic { buy 100; } best_effort;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Mode != "best_effort" {
		t.Fatalf("expected mode 'best_effort', got %s", stmt.Mode)
	}
}

func TestParseOverride(t *testing.T) {
	prog := parse(`override label1 label2 10%...20% { buy 100; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	if stmt.Kind != "override" {
		t.Fatalf("expected 'override', got %s", stmt.Kind)
	}
	if len(stmt.Labels) != 2 || stmt.Labels[0] != "label1" || stmt.Labels[1] != "label2" {
		t.Fatalf("expected labels [label1 label2], got %v", stmt.Labels)
	}
}

// ---------- Action ----------
func TestParseBuy(t *testing.T) {
	prog := parse(`buy 100;`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.ActionStmt)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.Action != "buy" {
		t.Fatalf("expected action 'buy', got %s", stmt.Action)
	}
	if len(stmt.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(stmt.Args))
	}
	mustLiteral(t, stmt.Args[0], "100", "")
}

func TestParseBuyWithUnit(t *testing.T) {
	prog := parse(`buy 100 shares;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if len(stmt.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(stmt.Args))
	}
	mustLiteral(t, stmt.Args[0], "100", "")
	mustIdent(t, stmt.Args[1], "shares")
}

func TestParseBuyMarket(t *testing.T) {
	prog := parse(`buy 100 @ market;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.PriceType != "market" {
		t.Fatalf("expected priceType 'market', got %s", stmt.PriceType)
	}
}

func TestParseBuyLimit(t *testing.T) {
	prog := parse(`buy 100 @ limit 150;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.PriceType != "limit" {
		t.Fatalf("expected priceType 'limit', got %s", stmt.PriceType)
	}
	if stmt.Price == nil {
		t.Fatal("expected non-nil price")
	}
	mustLiteral(t, stmt.Price, "150", "")
}

func TestParseSellFromOldest(t *testing.T) {
	prog := parse(`sell 50 from oldest;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.Action != "sell" {
		t.Fatalf("expected 'sell', got %s", stmt.Action)
	}
	if stmt.Modifier != "from_oldest" {
		t.Fatalf("expected modifier 'from_oldest', got %s", stmt.Modifier)
	}
}

func TestParseSellRemaining(t *testing.T) {
	prog := parse(`sell remaining;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.Modifier != "remaining" {
		t.Fatalf("expected modifier 'remaining', got %s", stmt.Modifier)
	}
}

func TestParseSellShort(t *testing.T) {
	prog := parse(`sell_short 100;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.Action != "sell_short" {
		t.Fatalf("expected 'sell_short', got %s", stmt.Action)
	}
}

func TestParseBuyCover(t *testing.T) {
	prog := parse(`buy_cover 50;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if stmt.Action != "buy_cover" {
		t.Fatalf("expected 'buy_cover', got %s", stmt.Action)
	}
}

// ---------- Config ----------
func TestParseKeep(t *testing.T) {
	prog := parse(`keep 1d 2h;`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.ConfigStmt)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if stmt.Key != "keep" {
		t.Fatalf("expected key 'keep', got %s", stmt.Key)
	}
	if len(stmt.Params) != 1 {
		t.Fatalf("expected 1 param (ListLiteralNode), got %d", len(stmt.Params))
	}
	list := mustListLiteral(t, stmt.Params[0], 2)
	mustLiteral(t, list.Items[0], "1", "d")
	mustLiteral(t, list.Items[1], "2", "h")
}

func TestParseMaxPosition(t *testing.T) {
	prog := parse(`max_position 1000 shares;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if stmt.Key != "max_position" {
		t.Fatalf("expected key 'max_position', got %s", stmt.Key)
	}
	if len(stmt.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(stmt.Params))
	}
	mustLiteral(t, stmt.Params[0], "1000", "")
	mustIdent(t, stmt.Params[1], "shares")
}

func TestParseStopLossRange(t *testing.T) {
	prog := parse(`stop_loss -10%...-5%;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if stmt.Key != "stop_loss" {
		t.Fatalf("expected key 'stop_loss', got %s", stmt.Key)
	}
	if stmt.Range == nil {
		t.Fatal("expected non-nil Range")
	}
	mustLiteral(t, stmt.Range.Begin, "-10", "%")
	mustLiteral(t, stmt.Range.End, "-5", "%")
}

func TestParseStopLossSingle(t *testing.T) {
	prog := parse(`stop_loss -10%;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if len(stmt.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(stmt.Params))
	}
	mustLiteral(t, stmt.Params[0], "-10", "%")
}

func TestParseCircuitBreaker(t *testing.T) {
	prog := parse(`circuit_breaker 5% in 1d;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if stmt.Key != "circuit_breaker" {
		t.Fatalf("expected key 'circuit_breaker', got %s", stmt.Key)
	}
	if len(stmt.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(stmt.Params))
	}
	mustLiteral(t, stmt.Params[0], "5", "%")
	mustIdent(t, stmt.Params[1], "in")
	list := mustListLiteral(t, stmt.Params[2], 1)
	mustLiteral(t, list.Items[0], "1", "d")
}

func TestParseLeverage(t *testing.T) {
	prog := parse(`leverage 5x;`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ConfigStmtNode)
	if stmt.Key != "leverage" {
		t.Fatalf("expected key 'leverage', got %s", stmt.Key)
	}
	mustLiteral(t, stmt.Params[0], "5", "")
	mustIdent(t, stmt.Params[1], "x")
}

// ---------- 表达式 ----------
func TestParseLiteralExpression(t *testing.T) {
	// 在 action 中测试表达式解析
	prog := parse(`buy 100 + 50;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	bin, ok := stmt.Args[0].(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Args[0])
	}
	if bin.Op != "+" {
		t.Fatalf("expected op '+', got %s", bin.Op)
	}
	mustLiteral(t, bin.Left, "100", "")
	mustLiteral(t, bin.Right, "50", "")
}

func TestParseUnaryNegative(t *testing.T) {
	prog := parse(`buy -100;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	mustLiteral(t, stmt.Args[0], "-100", "")
}

func TestParseBinaryMinus(t *testing.T) {
	prog := parse(`buy 100 - 30;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	bin, ok := stmt.Args[0].(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Args[0])
	}
	if bin.Op != "-" {
		t.Fatalf("expected op '-', got %s", bin.Op)
	}
}

func TestParseVariable(t *testing.T) {
	prog := parse(`buy $amount;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	vr, ok := stmt.Args[0].(*ast.VariableExprNode)
	if !ok {
		t.Fatalf("expected VariableExprNode, got %T", stmt.Args[0])
	}
	if vr.Name != "amount" {
		t.Fatalf("expected name 'amount', got %s", vr.Name)
	}
}

func TestParseReference(t *testing.T) {
	// @ref 在表达式中作为 reference
	prog := parse(`trigger @ref > 100 { buy 10; }`)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	bin, ok := stmt.Condition.(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Condition)
	}
	ref, ok := bin.Left.(*ast.ReferenceExprNode)
	if !ok {
		t.Fatalf("expected ReferenceExprNode, got %T", bin.Left)
	}
	if ref.Name != "ref" {
		t.Fatalf("expected name 'ref', got %s", ref.Name)
	}
}

func TestParseMacroExpand(t *testing.T) {
	prog := parse(`buy ${macro};`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	m, ok := stmt.Args[0].(*ast.MacroExpandExprNode)
	if !ok {
		t.Fatalf("expected MacroExpandExprNode, got %T", stmt.Args[0])
	}
	if m.Name != "macro" {
		t.Fatalf("expected name 'macro', got %s", m.Name)
	}
}

func TestParseComparisonExpr(t *testing.T) {
	prog := parse(`trigger $price == 100 { buy 10; }`)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	bin, ok := stmt.Condition.(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Condition)
	}
	if bin.Op != "==" {
		t.Fatalf("expected op '==', got %s", bin.Op)
	}
}

func TestParseLogicalExpr(t *testing.T) {
	prog := parse(`trigger a > 1 && b < 2 { buy 10; }`)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	bin, ok := stmt.Condition.(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Condition)
	}
	if bin.Op != "&&" {
		t.Fatalf("expected op '&&', got %s", bin.Op)
	}
}

// ---------- 其他语句 ----------
func TestParseIf(t *testing.T) {
	prog := parse(`if $var > 10 { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.IfStmt)
	stmt := prog.Statements[0].(*ast.IfStmtNode)
	if stmt.Condition == nil {
		t.Fatal("expected non-nil condition")
	}
	if len(stmt.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(stmt.Body))
	}
}

func TestParseAssert(t *testing.T) {
	prog := parse(`assert price > 0;`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.AssertStmt)
	stmt := prog.Statements[0].(*ast.AssertStmtNode)
	if stmt.Condition == nil {
		t.Fatal("expected non-nil condition")
	}
}

func TestParseDefine(t *testing.T) {
	prog := parse(`myVar: 100`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.DefineStmt)
	stmt := prog.Statements[0].(*ast.DefineStmtNode)
	if stmt.Name != "myVar" {
		t.Fatalf("expected name 'myVar', got %s", stmt.Name)
	}
	mustLiteral(t, stmt.Value, "100", "")
}

func TestParseAnnotation(t *testing.T) {
	prog := parse(`/* @adapter: binance */`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.AnnotationStmt)
	stmt := prog.Statements[0].(*ast.AnnotationStmtNode)
	if stmt.Key != "adapter" {
		t.Fatalf("expected key 'adapter', got %s", stmt.Key)
	}
	if stmt.Value != "binance" {
		t.Fatalf("expected value 'binance', got %s", stmt.Value)
	}
}

func TestParseTemplate(t *testing.T) {
	prog := parse(`template T(a: int, b: float = 1.5) extends Base { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.TemplateStmt)
	stmt := prog.Statements[0].(*ast.TemplateStmtNode)
	if stmt.Name != "T" {
		t.Fatalf("expected name 'T', got %s", stmt.Name)
	}
	if stmt.Extends != "Base" {
		t.Fatalf("expected extends 'Base', got %s", stmt.Extends)
	}
	if len(stmt.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(stmt.Params))
	}
}

func TestParseMacro(t *testing.T) {
	prog := parse(`macro M { buy 100; }`)
	mustLen(t, prog, 1)
	mustType(t, prog.Statements[0], ast.MacroStmt)
	stmt := prog.Statements[0].(*ast.MacroStmtNode)
	if stmt.Name != "M" {
		t.Fatalf("expected name 'M', got %s", stmt.Name)
	}
}

// ---------- 复杂嵌套 ----------
func TestParseComplexGrid(t *testing.T) {
	input := `
	grid AAPL {
		keep 1d;
		profit 10%...20% {
			buy 100;
		}
		loss -10% {
			sell 50;
		}
		max_position 1000 shares;
		trigger price > 200 {
			buy 10 @ market;
		}
		atomic { buy 100; } rollback;
	}
	`
	prog := parse(input)
	mustLen(t, prog, 1)
	grid := prog.Statements[0].(*ast.StrategyStmtNode)
	if len(grid.Body) != 6 {
		t.Fatalf("expected 6 body stmts, got %d", len(grid.Body))
	}
}

func TestParseGridWithGlobalAndRiskConfig(t *testing.T) {
	input := `
	grid AAPL {
		position fixed 100;
		sizing equal;
		stop_loss -10%...-5%;
		circuit_breaker 5% in 1d;
	}
	`
	prog := parse(input)
	mustLen(t, prog, 1)
	grid := prog.Statements[0].(*ast.StrategyStmtNode)
	if len(grid.Body) != 4 {
		t.Fatalf("expected 4 body stmts, got %d", len(grid.Body))
	}
}

// ---------- 边界情况 ----------
func TestParseUnknownToken(t *testing.T) {
	// 使用不合法的 token（如 standalone @）会报错但不 panic
	prog := parse(`@`)
	mustLen(t, prog, 0)
}

func TestParseBareIdentifier(t *testing.T) {
	prog := parse(`AAPL`)
	mustLen(t, prog, 0)
}

func TestParseOpenRangeLower(t *testing.T) {
	prog := parse(`profit ...20% { buy 100; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	rng := stmt.Condition.(*ast.RangeExprNode)
	if rng.Begin != nil {
		t.Fatal("expected nil begin")
	}
	mustLiteral(t, rng.End, "20", "%")
}

func TestParseOpenRangeUpper(t *testing.T) {
	prog := parse(`profit 10%... { buy 100; }`)
	mustLen(t, prog, 1)
	stmt := prog.Statements[0].(*ast.FlowStmtNode)
	rng := stmt.Condition.(*ast.RangeExprNode)
	mustLiteral(t, rng.Begin, "10", "%")
	if rng.End != nil {
		t.Fatal("expected nil end")
	}
}

func TestParseNegativePercent(t *testing.T) {
	prog := parse(`buy -5%;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	mustLiteral(t, stmt.Args[0], "-5", "%")
}

func TestParseArithExprInAction(t *testing.T) {
	prog := parse(`buy 100 * 2 + 50;`)
	stmt := prog.Statements[0].(*ast.ActionStmtNode)
	if len(stmt.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(stmt.Args))
	}
	// 100 * 2 + 50 => ((100 * 2) + 50)
	bin, ok := stmt.Args[0].(*ast.BinaryExprNode)
	if !ok {
		t.Fatalf("expected BinaryExprNode, got %T", stmt.Args[0])
	}
	if bin.Op != "+" {
		t.Fatalf("expected top-level op '+', got %s", bin.Op)
	}
}
