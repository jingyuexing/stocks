package tokenizer_test

import (
	"testing"

	"github.com/jingyuexing/stocks/tokenizer"
)

func TestTokenizer(t *testing.T) {
	tokens := tokenizer.Lexer(
		`
amount: some text value
`)
	if len(tokens) != 5 { // identifier, colon, text, terminator, eof
		t.Errorf("not pass tokenizer, expected 5 tokens, got %d", len(tokens))
	}

	tokens2 := tokenizer.Lexer(`buy 20%`)
	if len(tokens2) != 4 { // keywordBuy, integer, per, eof
		t.Errorf("lexer not pass, expected 4 tokens, got %d", len(tokens2))
	}

	tokens3 := tokenizer.Lexer("stop 30%...50%")
	if len(tokens3) != 7 { // keywordStop, integer, per, tribleDot, integer, per, eof
		t.Errorf("lexer is not pass, expected 7 tokens, got %d", len(tokens3))
	}
	tokens4 := tokenizer.Lexer("keep 1H 1m 1s")
	if len(tokens4) != 8 { // keywordKeep, integer, unit, integer, unit, integer, unit, eof
		t.Errorf("lexer is not pass, expected 8 tokens, got %d", len(tokens4))
	}
}

func TestLexerEmptyInput(t *testing.T) {
	tokens := tokenizer.Lexer("")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token for empty input, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.EOF {
		t.Errorf("expected EOF for empty input, got %v", tokens[0].Kind)
	}
}

func TestLexerOnlyWhitespace(t *testing.T) {
	tokens := tokenizer.Lexer("   \n\t  ")
	// \n produces Terminator, plus EOF = 2 tokens
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens for whitespace-only input (terminator + eof), got %d", len(tokens))
	}
	if tokens[1].Kind != tokenizer.EOF {
		t.Errorf("expected EOF as last token, got %v", tokens[1].Kind)
	}
}

func TestLexerSpecialCharacters(t *testing.T) {
	input := `@ - { } +`
	tokens := tokenizer.Lexer(input)
	expected := []tokenizer.TokenKind{
		tokenizer.At,
		tokenizer.Negative,
		tokenizer.LeftBraces,
		tokenizer.RightBraces,
		tokenizer.Plus,
		tokenizer.EOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, exp := range expected {
		if tokens[i].Kind != exp {
			t.Errorf("token[%d]: expected %v, got %v", i, exp, tokens[i].Kind)
		}
	}
}

func TestLexerMultipleDots(t *testing.T) {
	// "1...2" has 3 dots in the numeric scanner, so becomes a Text token
	tokens := tokenizer.Lexer("1...2")
	if len(tokens) != 2 { // text + eof
		t.Fatalf("expected 2 tokens (text + eof), got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Text {
		t.Errorf("expected Text for 1...2, got %v", tokens[0].Kind)
	}
}

func TestLexerMultipleDecimalPoints(t *testing.T) {
	// "1.2.3" has two dots so becomes Text
	tokens := tokenizer.Lexer("1.2.3")
	foundText := false
	for _, tok := range tokens {
		if tok.Kind == tokenizer.Text {
			foundText = true
		}
	}
	if !foundText {
		t.Error("expected a Text token for 1.2.3")
	}
}

func TestLexerNegativeNumber(t *testing.T) {
	tokens := tokenizer.Lexer("-10")
	if len(tokens) != 3 { // negative, integer, eof
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Negative {
		t.Errorf("expected Negative, got %v", tokens[0].Kind)
	}
	if tokens[1].Kind != tokenizer.Integer || tokens[1].Value != "10" {
		t.Errorf("expected Integer(10), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
}

func TestLexerMixedIdentifiersAndNumbers(t *testing.T) {
	// Lexer splits letters and digits: AUL123 -> Identifier("AUL") + Integer("123")
	tokens := tokenizer.Lexer("select AUL123")
	if len(tokens) != 4 { // keywordSelect, identifier("AUL"), integer("123"), eof
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.KeywordSelect {
		t.Errorf("expected KeywordSelect, got %v", tokens[0].Kind)
	}
	if tokens[1].Kind != tokenizer.Identifier || tokens[1].Value != "AUL" {
		t.Errorf("expected Identifier(AUL), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
	if tokens[2].Kind != tokenizer.Integer || tokens[2].Value != "123" {
		t.Errorf("expected Integer(123), got %v(%s)", tokens[2].Kind, tokens[2].Value)
	}
}

func TestLexerUnknownCharacters(t *testing.T) {
	// Characters like $, #, & should be skipped
	tokens := tokenizer.Lexer("buy $100")
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	if tokens[0].Kind != tokenizer.KeywordBuy {
		t.Errorf("expected KeywordBuy, got %v", tokens[0].Kind)
	}
}

func TestLexerUnicodeIdentifier(t *testing.T) {
	// Characters >= 0xff are valid identifier chars
	tokens := tokenizer.Lexer("股票 AUL")
	if len(tokens) != 3 { // identifier, identifier, eof
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Identifier {
		t.Errorf("expected Identifier for unicode, got %v", tokens[0].Kind)
	}
}

func TestLexerAtReference(t *testing.T) {
	tokens := tokenizer.Lexer("@ork")
	if len(tokens) != 3 { // at, identifier, eof
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.At {
		t.Errorf("expected At, got %v", tokens[0].Kind)
	}
	if tokens[1].Kind != tokenizer.Identifier || tokens[1].Value != "ork" {
		t.Errorf("expected Identifier(ork), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
}

func TestLexerColonText(t *testing.T) {
	tokens := tokenizer.Lexer("key: value here")
	if len(tokens) != 4 { // identifier, colon, text, eof
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Identifier || tokens[0].Value != "key" {
		t.Errorf("expected Identifier(key), got %v(%s)", tokens[0].Kind, tokens[0].Value)
	}
	if tokens[1].Kind != tokenizer.Colon {
		t.Errorf("expected Colon, got %v", tokens[1].Kind)
	}
	if tokens[2].Kind != tokenizer.Text || tokens[2].Value != " value here" {
		t.Errorf("expected Text(' value here'), got %v(%s)", tokens[2].Kind, tokens[2].Value)
	}
}

func TestLexerMultipleNewlines(t *testing.T) {
	// Each \n produces a Terminator; there are 3 newlines between buy and sell
	tokens := tokenizer.Lexer("buy 100\n\n\nsell 50")
	terminatorCount := 0
	for _, tok := range tokens {
		if tok.Kind == tokenizer.Terminator {
			terminatorCount++
		}
	}
	if terminatorCount != 3 {
		t.Errorf("expected 3 terminators, got %d", terminatorCount)
	}
}
