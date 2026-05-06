package tokenizer

import (
	"testing"
)

func TestTokenKindNumberic(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want uint
	}{
		{KeywordSell, 0x1},
		{KeywordBuy, 0x2},
		{KeywordStop, 0x3},
		{KeywordKeep, 0x4},
		{KeywordSelect, 0x5},
		{KeywordValue, 0x6},
		{KeywordGrid, 0x7},
		{KeywordLoss, 0x8},
		{KeywordProfit, 0x9},
		{Integer, 0xa},
		{Float, 0xb},
		{Per, 0xc},
		{Text, 0xd},
		{Identifier, 0xe},
		{Terminator, 0xf},
		{Colon, 0x10},
		{At, 0x11},
		{TribleDot, 0x12},
		{Negative, 0x13},
		{Time, 0x14},
		{LeftBraces, 0x15},
		{RightBraces, 0x16},
		{Plus, 0x17},
		{EOF, 0x18},
		{TokenKind("unknown"), 0},
	}
	for _, tc := range tests {
		got := tc.kind.Numberic()
		if got != tc.want {
			t.Errorf("%s.Numberic() = %d, want %d", tc.kind, got, tc.want)
		}
	}
}

func TestTokenKindIsKeyword(t *testing.T) {
	keywords := []TokenKind{
		KeywordBuy, KeywordSell, KeywordStop, KeywordKeep,
		KeywordSelect, KeywordValue, KeywordGrid, KeywordLoss, KeywordProfit,
	}
	for _, k := range keywords {
		if !k.IsKeyword() {
			t.Errorf("%s should be keyword", k)
		}
	}

	nonKeywords := []TokenKind{Integer, Float, Identifier, Text, At, Colon}
	for _, k := range nonKeywords {
		if k.IsKeyword() {
			t.Errorf("%s should not be keyword", k)
		}
	}
}

func TestTokenIsNumber(t *testing.T) {
	if !TokenIsNumber(Token{Kind: Integer, Value: "1"}) {
		t.Error("Integer should be number")
	}
	if !TokenIsNumber(Token{Kind: Float, Value: "1.5"}) {
		t.Error("Float should be number")
	}
	if !TokenIsNumber(Token{Kind: Per, Value: "%"}) {
		t.Error("Per should be number")
	}
	if TokenIsNumber(Token{Kind: Identifier, Value: "x"}) {
		t.Error("Identifier should not be number")
	}
	if TokenIsNumber(Token{Kind: Text, Value: "hello"}) {
		t.Error("Text should not be number")
	}
}

func TestIsTimeUnit(t *testing.T) {
	valid := []string{"Y", "M", "W", "d", "h", "H", "m", "s", "ms"}
	for _, u := range valid {
		if !IsTimeUnit(u) {
			t.Errorf("IsTimeUnit(%q) should be true", u)
		}
	}
	invalid := []string{"", "x", "days", "hour"}
	for _, u := range invalid {
		if IsTimeUnit(u) {
			t.Errorf("IsTimeUnit(%q) should be false", u)
		}
	}
}

func TestIsAmountUnit(t *testing.T) {
	valid := []string{"K", "k", "M", "m", "B", "b"}
	for _, u := range valid {
		if !IsAmountUnit(u) {
			t.Errorf("IsAmountUnit(%q) should be true", u)
		}
	}
	invalid := []string{"", "x", "W", "g"}
	for _, u := range invalid {
		if IsAmountUnit(u) {
			t.Errorf("IsAmountUnit(%q) should be false", u)
		}
	}
}

func TestIsTimeUnitToken(t *testing.T) {
	if !IsTimeUnitToken(Token{Kind: Identifier, Value: "W"}) {
		t.Error("Token with Value W should be time unit")
	}
	if IsTimeUnitToken(Token{Kind: Identifier, Value: "X"}) {
		t.Error("Token with Value X should not be time unit")
	}
}

func TestIsLiteral(t *testing.T) {
	if !IsLiteral(Token{Kind: Integer}) {
		t.Error("Integer is literal")
	}
	if !IsLiteral(Token{Kind: Float}) {
		t.Error("Float is literal")
	}
	if !IsLiteral(Token{Kind: Text}) {
		t.Error("Text is literal")
	}
	if IsLiteral(Token{Kind: Identifier}) {
		t.Error("Identifier is not literal")
	}
	if IsLiteral(Token{Kind: KeywordBuy}) {
		t.Error("KeywordBuy is not literal")
	}
}

func TestCreateIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenKind
	}{
		{"buy", KeywordBuy},
		{"sell", KeywordSell},
		{"stop", KeywordStop},
		{"keep", KeywordKeep},
		{"grid", KeywordGrid},
		{"loss", KeywordLoss},
		{"profit", KeywordProfit},
		{"select", KeywordSelect},
		{"value", KeywordValue},
		{"myVar", Identifier},
		{"AUL", Identifier},
	}
	for _, tc := range tests {
		tok := createIdentifier(tc.input)
		if tok.Kind != tc.expected {
			t.Errorf("createIdentifier(%q).Kind = %v, want %v", tc.input, tok.Kind, tc.expected)
		}
		if tok.Value != tc.input {
			t.Errorf("createIdentifier(%q).Value = %q, want %q", tc.input, tok.Value, tc.input)
		}
	}
}

func TestCreateToken(t *testing.T) {
	tok := createToken(KeywordBuy, "buy")
	if tok.Kind != KeywordBuy || tok.Value != "buy" {
		t.Error("createToken failed")
	}
}
