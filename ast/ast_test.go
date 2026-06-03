package ast

import (
	"testing"
)

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nt       NodeType
		expected string
	}{
		{Program, "unknown"},
		{FloatLiteral, "FloatLiteral"},
		{IntegerLiteral, "IntegerLiteral"},
		{TextLiteral, "TextLiteral"},
		{DurationLiteral, "DurationLiteral"},
		{PercentLiteral, "unknown"},
		{ReferenceLiteral, "unknown"},
		{LiteralExpression, "LiteralExpression"},
		{BuyExpression, "BuyExpression"},
		{SellExpression, "SellExpression"},
		{KeepExpression, "KeepExpression"},
		{StopExpression, "StopExpression"},
		{RangeExpression, "RangeExpression"},
		{DefineStatement, "DefineStatement"},
		{ReferenceStatement, "ReferenceStatement"},
		{SelectExpression, "SelectExpression"},
		{GridExpression, "GridExpression"},
		{LossExpression, "LossExpression"},
		{ProfitExpression, "ProfitExpression"},
		{ValueStatement, "ValueStatement"},
		{YearUnit, "YearUnit"},
		{MonthUnit, "MonthUnit"},
		{DayUnit, "DayUnit"},
		{WeekUnit, "WeekUnit"},
		{HourUnit, "HourUnit"},
		{MinuteUnit, "MinuteUnit"},
		{SecondUnit, "SecondUnit"},
		{MillisecondUnit, "MillisecondUnit"},
		{Unknown, "unknown"},
		{NodeType(9999), "unknown"},
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
			literal:  Literal{Node: Node{Type: FloatLiteral}, Value: "3.14"},
			expected: "3.14",
		},
		{
			name:     "float percent",
			literal:  Literal{Node: Node{Type: FloatLiteral}, Value: "12.5", Unit: "%"},
			expected: "0.125000",
		},
		{
			name:     "integer",
			literal:  Literal{Node: Node{Type: IntegerLiteral}, Value: "42"},
			expected: "42",
		},
		{
			name:     "text returns empty",
			literal:  Literal{Node: Node{Type: TextLiteral}, Value: "hello"},
			expected: "",
		},
		{
			name:     "default returns empty",
			literal:  Literal{Node: Node{Type: Unknown}, Value: "x"},
			expected: "",
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

func TestExpressionNodeBody(t *testing.T) {
	expr := ExpressionNode{
		Node: Node{Type: GridExpression},
		Body: []ExpressionNode{
			{Node: Node{Type: LossExpression}},
			{Node: Node{Type: ProfitExpression}},
		},
	}
	if len(expr.Body) != 2 {
		t.Fatalf("expected 2 body items, got %d", len(expr.Body))
	}
	if expr.Body[0].Type != LossExpression {
		t.Error("expected LossExpression in body[0]")
	}
	if expr.Body[1].Type != ProfitExpression {
		t.Error("expected ProfitExpression in body[1]")
	}
}

func TestRangeExpressionNode(t *testing.T) {
	node := RangeExpressionNode{
		Node:  Node{Type: RangeExpression},
		Begin: Literal{Value: "10", Unit: "%"},
		End:   Literal{Value: "20", Unit: "%"},
	}
	if node.Begin.Value != "10" || node.End.Value != "20" {
		t.Error("RangeExpressionNode values mismatch")
	}
}
