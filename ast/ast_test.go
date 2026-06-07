package ast

import (
	"testing"
)

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nt       NodeType
		expected string
	}{
		{Program, "Program"},
		{StrategyStmt, "StrategyStmt"},
		{BlockStmt, "BlockStmt"},
		{FlowStmt, "FlowStmt"},
		{ActionStmt, "ActionStmt"},
		{ConfigStmt, "ConfigStmt"},
		{ImportStmt, "ImportStmt"},
		{ExportStmt, "ExportStmt"},
		{UseStmt, "UseStmt"},
		{TemplateStmt, "TemplateStmt"},
		{MacroStmt, "MacroStmt"},
		{IfStmt, "IfStmt"},
		{AssertStmt, "AssertStmt"},
		{AnnotationStmt, "AnnotationStmt"},
		{DefineStmt, "DefineStmt"},
		{SelectStmt, "SelectStmt"},
		{LiteralExpr, "LiteralExpr"},
		{BinaryExpr, "BinaryExpr"},
		{UnaryExpr, "UnaryExpr"},
		{RangeExpr, "RangeExpr"},
		{VariableExpr, "VariableExpr"},
		{ReferenceExpr, "ReferenceExpr"},
		{MacroExpandExpr, "MacroExpandExpr"},
		{CallExpr, "CallExpr"},
		{IdentifierExpr, "IdentifierExpr"},
		{Unknown, "Unknown"},
		{NodeType(9999), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.nt.String()
		if got != tc.expected {
			t.Errorf("NodeType(%d).String() = %q, want %q", tc.nt, got, tc.expected)
		}
	}
}

func TestLiteralString(t *testing.T) {
	tests := []struct {
		name     string
		literal  Literal
		expected string
	}{
		{
			name:     "float without unit",
			literal:  Literal{Value: "3.14"},
			expected: "3.14",
		},
		{
			name:     "float percent",
			literal:  Literal{Value: "12.5", Unit: "%"},
			expected: "0.125000",
		},
		{
			name:     "integer",
			literal:  Literal{Value: "42"},
			expected: "42",
		},
		{
			name:     "text",
			literal:  Literal{Value: "hello"},
			expected: "hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.literal.String()
			if got != tc.expected {
				t.Errorf("Literal.String() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestLiteralAsFloat(t *testing.T) {
	l := Literal{Value: "3.14"}
	if got := l.AsFloat(); got != 3.14 {
		t.Errorf("AsFloat() = %v, want 3.14", got)
	}
	// invalid parse
	l2 := Literal{Value: "abc"}
	if got := l2.AsFloat(); got != 0 {
		t.Errorf("AsFloat() invalid = %v, want 0", got)
	}
}

func TestLiteralAsInt(t *testing.T) {
	l := Literal{Value: "42"}
	if got := l.AsInt(); got != 42 {
		t.Errorf("AsInt() = %v, want 42", got)
	}
	// invalid parse
	l2 := Literal{Value: "xyz"}
	if got := l2.AsInt(); got != 0 {
		t.Errorf("AsInt() invalid = %v, want 0", got)
	}
}

func TestLiteralAsBool(t *testing.T) {
	l := Literal{Value: "true"}
	if !l.AsBool() {
		t.Error("AsBool() = false, want true")
	}
	l2 := Literal{Value: "false"}
	if l2.AsBool() {
		t.Error("AsBool() = true, want false")
	}
}

func TestProgramNode(t *testing.T) {
	prog := ProgramNode{
		Statements: []Stmt{
			&StrategyStmtNode{Kind: "grid"},
			&ActionStmtNode{Action: "buy"},
		},
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
	if s, ok := prog.Statements[0].(*StrategyStmtNode); !ok || s.Kind != "grid" {
		t.Error("expected StrategyStmtNode with kind grid")
	}
	if a, ok := prog.Statements[1].(*ActionStmtNode); !ok || a.Action != "buy" {
		t.Error("expected ActionStmtNode with action buy")
	}
}

func TestRangeExprNode(t *testing.T) {
	node := RangeExprNode{
		Begin: &Literal{Value: "10", Unit: "%"},
		End:   &Literal{Value: "20", Unit: "%"},
	}
	if node.Begin.Value != "10" || node.End.Value != "20" {
		t.Error("RangeExprNode values mismatch")
	}
}

func TestStrategyStmtNode(t *testing.T) {
	stmt := StrategyStmtNode{
		Kind:   "grid",
		Target: "AAPL",
		Body: []Stmt{
			&ActionStmtNode{Action: "buy"},
		},
	}
	if stmt.Kind != "grid" {
		t.Errorf("expected kind grid, got %s", stmt.Kind)
	}
	if stmt.Target != "AAPL" {
		t.Errorf("expected target AAPL, got %s", stmt.Target)
	}
	if len(stmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(stmt.Body))
	}
}

func TestFlowStmtNode(t *testing.T) {
	stmt := FlowStmtNode{
		Kind:      "level",
		Condition: &RangeExprNode{Begin: &Literal{Value: "-10"}, End: &Literal{Value: "-5"}},
		Body: []Stmt{
			&ActionStmtNode{Action: "sell"},
		},
	}
	if stmt.Kind != "level" {
		t.Errorf("expected kind level, got %s", stmt.Kind)
	}
	if len(stmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(stmt.Body))
	}
}

func TestBinaryExprNode(t *testing.T) {
	expr := BinaryExprNode{
		Op:    ">",
		Left:  &VariableExprNode{Name: "price"},
		Right: &Literal{Value: "100"},
	}
	if expr.Op != ">" {
		t.Errorf("expected op >, got %s", expr.Op)
	}
	if _, ok := expr.Left.(*VariableExprNode); !ok {
		t.Error("expected left to be VariableExprNode")
	}
	if _, ok := expr.Right.(*Literal); !ok {
		t.Error("expected right to be Literal")
	}
}

func TestUnaryExprNode(t *testing.T) {
	expr := UnaryExprNode{
		Op:   "-",
		Expr: &Literal{Value: "10"},
	}
	if expr.Op != "-" {
		t.Errorf("expected op -, got %s", expr.Op)
	}
}
