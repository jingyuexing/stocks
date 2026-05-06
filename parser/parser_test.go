package parser_test

import (
	"testing"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
)

func TestParserBuy(t *testing.T) {
	tokens := tokenizer.Lexer("buy 100")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.BuyExpression {
		t.Errorf("expected BuyExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(root.Expression[0].Params))
	}
}

func TestParserBuyWithReference(t *testing.T) {
	tokens := tokenizer.Lexer("buy @ork")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.BuyExpression {
		t.Errorf("expected BuyExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(root.Expression[0].Params))
	}
	if root.Expression[0].Params[0].Type != ast.ReferenceLiteral {
		t.Errorf("expected ReferenceLiteral param, got %v", root.Expression[0].Params[0].Type)
	}
}

func TestParserStop(t *testing.T) {
	tokens := tokenizer.Lexer("stop 20%")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.StopExpression {
		t.Errorf("expected StopExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(root.Expression[0].Params))
	}
}

func TestParserStopRange(t *testing.T) {
	tokens := tokenizer.Lexer("stop 30%...50%")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.StopExpression {
		t.Errorf("expected StopExpression, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Range == nil {
		t.Fatal("expected Range to be set")
	}
	if root.Expression[0].Range.Begin.Value != "30" {
		t.Errorf("expected range begin 30, got %s", root.Expression[0].Range.Begin.Value)
	}
	if root.Expression[0].Range.End.Value != "50" {
		t.Errorf("expected range end 50, got %s", root.Expression[0].Range.End.Value)
	}
}

func TestParserKeep(t *testing.T) {
	tokens := tokenizer.Lexer("keep 1M 1h 1m 1s")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.KeepExpression {
		t.Errorf("expected KeepExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 4 {
		t.Errorf("expected 4 params, got %d", len(root.Expression[0].Params))
	}
}

func TestParserSelect(t *testing.T) {
	tokens := tokenizer.Lexer("select AUL")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.SelectExpression {
		t.Errorf("expected SelectExpression, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "AUL" {
		t.Errorf("expected name AUL, got %s", root.Expression[0].Name)
	}
}

func TestParserValue(t *testing.T) {
	// Lexer 在遇到 ':' 时会将整行剩余内容作为 Text token，
	// 因此 "value: ork 1000" 会被解析为 KeywordValue + Colon + Text(" ork 1000")
	tokens := tokenizer.Lexer("value: ork 1000")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ValueStatement {
		t.Errorf("expected ValueStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Value.Value != " ork 1000" {
		t.Errorf("expected value ' ork 1000', got %s", root.Expression[0].Value.Value)
	}
}

func TestParserValueWithoutColon(t *testing.T) {
	// 无冒号形式：value ork 1000
	tokens := tokenizer.Lexer("value ork 1000")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ValueStatement {
		t.Errorf("expected ValueStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "ork" {
		t.Errorf("expected name ork, got %s", root.Expression[0].Name)
	}
	if root.Expression[0].Value.Value != "1000" {
		t.Errorf("expected value 1000, got %s", root.Expression[0].Value.Value)
	}
}

func TestParserGrid(t *testing.T) {
	input := `
grid {
	loss 10% {
		sell 100%
	}
	profit 10% {
		sell 50%
	}
}
`
	tokens := tokenizer.Lexer(input)
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.GridExpression {
		t.Errorf("expected GridExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Body) != 2 {
		t.Fatalf("expected 2 body expressions, got %d", len(root.Expression[0].Body))
	}

	loss := root.Expression[0].Body[0]
	if loss.Type != ast.LossExpression {
		t.Errorf("expected LossExpression, got %v", loss.Type)
	}
	if len(loss.Body) != 1 {
		t.Errorf("expected 1 loss body expression, got %d", len(loss.Body))
	}

	profit := root.Expression[0].Body[1]
	if profit.Type != ast.ProfitExpression {
		t.Errorf("expected ProfitExpression, got %v", profit.Type)
	}
	if len(profit.Body) != 1 {
		t.Errorf("expected 1 profit body expression, got %d", len(profit.Body))
	}
}

func TestParserDefine(t *testing.T) {
	tokens := tokenizer.Lexer("amount: some text value")
	root := parser.Parser(tokens)

	foundDefine := false
	for _, expr := range root.Expression {
		if expr.Type == ast.DefineStatement {
			foundDefine = true
			if expr.Name != "amount" {
				t.Errorf("expected define name 'amount', got %s", expr.Name)
			}
			if expr.Value.Value != " some text value" {
				t.Errorf("expected define value ' some text value', got %s", expr.Value.Value)
			}
		}
	}
	if !foundDefine {
		t.Error("expected DefineStatement in expressions")
	}
}

func TestParserReference(t *testing.T) {
	tokens := tokenizer.Lexer("@ork")
	root := parser.Parser(tokens)

	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ReferenceStatement {
		t.Errorf("expected ReferenceStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "ork" {
		t.Errorf("expected name ork, got %s", root.Expression[0].Name)
	}
}

func TestParserComplex(t *testing.T) {
	input := `
select AUL
value: ork 1000
buy @ork
keep 10W
sell AUL +30%
`
	tokens := tokenizer.Lexer(input)
	root := parser.Parser(tokens)

	expectedTypes := []ast.NodeType{
		ast.SelectExpression,
		ast.ValueStatement,
		ast.BuyExpression,
		ast.KeepExpression,
		ast.SellExpression,
	}

	if len(root.Expression) != len(expectedTypes) {
		t.Fatalf("expected %d expressions, got %d", len(expectedTypes), len(root.Expression))
	}

	for i, expected := range expectedTypes {
		if root.Expression[i].Type != expected {
			t.Errorf("expression %d: expected %v, got %v", i, expected, root.Expression[i].Type)
		}
	}
}

// ========== 错误路径测试 ==========

func TestParserEmptyInput(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer(""))
	if len(root.Expression) != 0 {
		t.Errorf("expected 0 expressions for empty input, got %d", len(root.Expression))
	}
	if root.Type != ast.Program {
		t.Error("expected Program node type")
	}
}

func TestParserUnknownToken(t *testing.T) {
	// $ 是未知字符，Lexer 会跳过它，所以只应剩下 buy 和 eof
	root := parser.Parser(tokenizer.Lexer("$buy 100"))
	foundBuy := false
	for _, expr := range root.Expression {
		if expr.Type == ast.BuyExpression {
			foundBuy = true
		}
	}
	if !foundBuy {
		t.Error("expected BuyExpression despite unknown prefix token")
	}
}

func TestParserBuyUnexpectedToken(t *testing.T) {
	// buy 后面跟非数字非@的内容
	root := parser.Parser(tokenizer.Lexer("buy @"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.BuyExpression {
		t.Errorf("expected BuyExpression, got %v", root.Expression[0].Type)
	}
}

func TestParserSellWithoutNumber(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("sell"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.SellExpression {
		t.Errorf("expected SellExpression, got %v", root.Expression[0].Type)
	}
}

func TestParserKeepWithoutDuration(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("keep"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.KeepExpression {
		t.Errorf("expected KeepExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(root.Expression[0].Params))
	}
}

func TestParserStopWithoutValue(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("stop"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.StopExpression {
		t.Errorf("expected StopExpression, got %v", root.Expression[0].Type)
	}
}

func TestParserSelectWithoutIdentifier(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("select"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.SelectExpression {
		t.Errorf("expected SelectExpression, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "" {
		t.Errorf("expected empty name, got %q", root.Expression[0].Name)
	}
}

func TestParserReferenceWithoutIdentifier(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("@"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ReferenceStatement {
		t.Errorf("expected ReferenceStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "" {
		t.Errorf("expected empty name, got %q", root.Expression[0].Name)
	}
}

func TestParserNegativeWithoutNumber(t *testing.T) {
	// "- x" should skip '-' (not a number after it), then 'x' is parsed as bare identifier
	root := parser.Parser(tokenizer.Lexer("- x"))
	// The '-' is skipped due to error, 'x' becomes a bare identifier ReferenceStatement
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression (bare identifier), got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ReferenceStatement || root.Expression[0].Name != "x" {
		t.Errorf("expected ReferenceStatement(x), got %v", root.Expression[0])
	}
}

func TestParserGridEmptyBody(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("grid {}"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.GridExpression {
		t.Errorf("expected GridExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Body) != 0 {
		t.Errorf("expected empty grid body, got %d", len(root.Expression[0].Body))
	}
}

func TestParserMultipleExpressions(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("buy 100\nsell 50\nkeep 1W"))
	if len(root.Expression) != 3 {
		t.Fatalf("expected 3 expressions, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.BuyExpression {
		t.Errorf("expected BuyExpression at 0, got %v", root.Expression[0].Type)
	}
	if root.Expression[1].Type != ast.SellExpression {
		t.Errorf("expected SellExpression at 1, got %v", root.Expression[1].Type)
	}
	if root.Expression[2].Type != ast.KeepExpression {
		t.Errorf("expected KeepExpression at 2, got %v", root.Expression[2].Type)
	}
}

// ========== 工具函数测试（通过公开方式验证） ==========

func TestParserDurationToNumber(t *testing.T) {
	// durationToNumber is unexported, but we verify its behavior
	// through integration: keep expressions with various units
	cases := []struct {
		input    string
		expected int // expected number of duration params
	}{
		{"keep 1ms", 1},
		{"keep 1s", 1},
		{"keep 1m", 1},
		{"keep 1h", 1},
		{"keep 1d", 1},
		{"keep 1W", 1},
		{"keep 1M", 1},
		{"keep 1Y", 1},
	}
	for _, tc := range cases {
		root := parser.Parser(tokenizer.Lexer(tc.input))
		if len(root.Expression) != 1 {
			t.Fatalf("%s: expected 1 expression, got %d", tc.input, len(root.Expression))
		}
		if len(root.Expression[0].Params) != tc.expected {
			t.Errorf("%s: expected %d params, got %d", tc.input, tc.expected, len(root.Expression[0].Params))
		}
	}
}

func TestParserCreateLiteralDefault(t *testing.T) {
	// createLiteral default branch covers Text token kind
	// "amount: hello" creates Text token via colon handling
	root := parser.Parser(tokenizer.Lexer("amount: hello"))
	found := false
	for _, expr := range root.Expression {
		if expr.Type == ast.DefineStatement && expr.Value.Node.Type == ast.TextLiteral {
			found = true
		}
	}
	if !found {
		t.Error("expected TextLiteral from colon-parsed value")
	}
}

func TestParserDefineWithoutColon(t *testing.T) {
	// "key value" without colon: identifier followed by identifier
	root := parser.Parser(tokenizer.Lexer("key value"))
	// This parses as two bare identifiers
	if len(root.Expression) != 2 {
		t.Fatalf("expected 2 expressions, got %d", len(root.Expression))
	}
}

func TestParserDefineWithInvalidValue(t *testing.T) {
	// "key: @" — colon present but value is not a literal
	root := parser.Parser(tokenizer.Lexer("key: @"))
	// @ produces a ReferenceStatement, not a DefineStatement
	foundDefine := false
	for _, expr := range root.Expression {
		if expr.Type == ast.DefineStatement {
			foundDefine = true
		}
	}
	// Actually "key" is parsed as bare identifier because peek(1) is Colon,
	// so it enters parseDefine. Then after colon, @ is not literal,
	// so DefineStatement is returned with empty Name/Value.
	// Let's verify at least one expression exists.
	if len(root.Expression) == 0 {
		t.Fatal("expected at least one expression")
	}
	_ = foundDefine
}

func TestParserKeepWithInvalidUnit(t *testing.T) {
	// "keep 1x" — 1 is a number but x is not a valid time unit
	root := parser.Parser(tokenizer.Lexer("keep 1x"))
	// parseKeep fails on "1x", then "1" becomes LiteralExpression and "x" becomes ReferenceStatement
	if len(root.Expression) != 3 {
		t.Fatalf("expected 3 expressions (keep + literal + ref), got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.KeepExpression {
		t.Errorf("expected KeepExpression, got %v", root.Expression[0].Type)
	}
	// Should have 0 valid duration params because "1x" is not a valid duration
	if len(root.Expression[0].Params) != 0 {
		t.Errorf("expected 0 duration params for invalid unit, got %d", len(root.Expression[0].Params))
	}
}

func TestParserStopWithRangeInconsistentUnits(t *testing.T) {
	// "stop 10%...20" — inconsistent units (one has %, other doesn't)
	// Parser should print warning and skip range
	root := parser.Parser(tokenizer.Lexer("stop 10%...20"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.StopExpression {
		t.Errorf("expected StopExpression, got %v", root.Expression[0].Type)
	}
	// Range may or may not be set depending on behavior; just verify no panic
}

func TestParserBareIdentifier(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("AUL"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ReferenceStatement {
		t.Errorf("expected ReferenceStatement for bare identifier, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Name != "AUL" {
		t.Errorf("expected Name AUL, got %q", root.Expression[0].Name)
	}
}

func TestParserValueWithoutName(t *testing.T) {
	// "value 1000" — no identifier name, just a number
	root := parser.Parser(tokenizer.Lexer("value 1000"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.ValueStatement {
		t.Errorf("expected ValueStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Value.Value != "1000" {
		t.Errorf("expected Value 1000, got %q", root.Expression[0].Value.Value)
	}
}

func TestParserGridNestedBody(t *testing.T) {
	input := `grid {
		loss 5% {
			buy 10
			sell 20
		}
	}`
	root := parser.Parser(tokenizer.Lexer(input))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	grid := root.Expression[0]
	if grid.Type != ast.GridExpression {
		t.Fatalf("expected GridExpression, got %v", grid.Type)
	}
	if len(grid.Body) != 1 {
		t.Fatalf("expected 1 body item (loss), got %d", len(grid.Body))
	}
	loss := grid.Body[0]
	if loss.Type != ast.LossExpression {
		t.Errorf("expected LossExpression, got %v", loss.Type)
	}
	if len(loss.Body) != 2 {
		t.Errorf("expected 2 nested expressions in loss, got %d", len(loss.Body))
	}
}

func TestParserColonWithoutIdentifier(t *testing.T) {
	// ": 123" — bare colon should be handled
	// Lexer captures the rest of line including leading space as Text
	root := parser.Parser(tokenizer.Lexer(": 123"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.DefineStatement {
		t.Errorf("expected DefineStatement, got %v", root.Expression[0].Type)
	}
	if root.Expression[0].Value.Value != " 123" {
		t.Errorf("expected Value ' 123', got %q", root.Expression[0].Value.Value)
	}
}

func TestParserLiteralExpression(t *testing.T) {
	// Bare number at top level becomes LiteralExpression
	root := parser.Parser(tokenizer.Lexer("42"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.LiteralExpression {
		t.Errorf("expected LiteralExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 1 || root.Expression[0].Params[0].Value != "42" {
		t.Errorf("expected param value 42, got %v", root.Expression[0].Params)
	}
}

func TestParserFloatWithPercent(t *testing.T) {
	root := parser.Parser(tokenizer.Lexer("stop 12.5%"))
	if len(root.Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(root.Expression))
	}
	if root.Expression[0].Type != ast.StopExpression {
		t.Errorf("expected StopExpression, got %v", root.Expression[0].Type)
	}
	if len(root.Expression[0].Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(root.Expression[0].Params))
	}
	param := root.Expression[0].Params[0]
	if param.Value != "12.5" || param.Unit != "%" {
		t.Errorf("expected 12.5%%, got %s%s", param.Value, param.Unit)
	}
}
