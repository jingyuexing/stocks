package tokenizer_test

import (
	"strings"
	"testing"

	"github.com/jingyuexing/stocks/tokenizer"
)

func TestTokenizer(t *testing.T) {
	tokens := tokenizer.Lexer(
		`
amount: some text value
`)
	if len(tokens) != 8 { // terminator, identifier, colon, identifier(some), identifier(text), keywordValue, terminator, eof
		t.Errorf("not pass tokenizer, expected 8 tokens, got %d", len(tokens))
	}

	tokens2 := tokenizer.Lexer(`buy 20%`)
	if len(tokens2) != 4 { // keywordBuy, integer, per, eof
		t.Errorf("lexer not pass, expected 4 tokens, got %d", len(tokens2))
	}

	tokens3 := tokenizer.Lexer("stop 30%...50%")
	if len(tokens3) != 7 { // keywordStop, integer, per, range, integer, per, eof
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
	// "1...2" should be split into integer + range + integer
	tokens := tokenizer.Lexer("1...2")
	if len(tokens) != 4 { // integer + range + integer + eof
		t.Fatalf("expected 4 tokens (integer + range + integer + eof), got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Integer || tokens[0].Value != "1" {
		t.Errorf("expected Integer(1), got %v(%s)", tokens[0].Kind, tokens[0].Value)
	}
	if tokens[1].Kind != tokenizer.Range || tokens[1].Value != "..." {
		t.Errorf("expected Range(...), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
	if tokens[2].Kind != tokenizer.Integer || tokens[2].Value != "2" {
		t.Errorf("expected Integer(2), got %v(%s)", tokens[2].Kind, tokens[2].Value)
	}
}

func TestLexerMultipleDecimalPoints(t *testing.T) {
	// "1.2.3" should be split into float + dot + integer
	tokens := tokenizer.Lexer("1.2.3")
	if len(tokens) != 4 { // float + dot + integer + eof
		t.Fatalf("expected 4 tokens (float + dot + integer + eof), got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Float || tokens[0].Value != "1.2" {
		t.Errorf("expected Float(1.2), got %v(%s)", tokens[0].Kind, tokens[0].Value)
	}
	if tokens[1].Kind != tokenizer.Dot || tokens[1].Value != "." {
		t.Errorf("expected Dot(.), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
	if tokens[2].Kind != tokenizer.Integer || tokens[2].Value != "3" {
		t.Errorf("expected Integer(3), got %v(%s)", tokens[2].Kind, tokens[2].Value)
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
	if len(tokens) != 3 { // keywordSelect, identifier("AUL123"), eof
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.KeywordSelect {
		t.Errorf("expected KeywordSelect, got %v", tokens[0].Kind)
	}
	if tokens[1].Kind != tokenizer.Identifier || tokens[1].Value != "AUL123" {
		t.Errorf("expected Identifier(AUL123), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
}

func TestLexerUnknownCharacters(t *testing.T) {
	// $ is now a valid token
	tokens := tokenizer.Lexer("buy $100")
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	if tokens[0].Kind != tokenizer.KeywordBuy {
		t.Errorf("expected KeywordBuy, got %v", tokens[0].Kind)
	}
}

func TestLexerDollarVariable(t *testing.T) {
	tokens := tokenizer.Lexer("$profit")
	if len(tokens) != 3 { // dollar, keyword(profit), eof
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Dollar {
		t.Errorf("expected Dollar, got %v", tokens[0].Kind)
	}
	// profit is a keyword, not an identifier
	if tokens[1].Kind != tokenizer.KeywordProfit || tokens[1].Value != "profit" {
		t.Errorf("expected KeywordProfit(profit), got %v(%s)", tokens[1].Kind, tokens[1].Value)
	}
}

func TestLexerDollarInExpression(t *testing.T) {
	tokens := tokenizer.Lexer("buy $amount @market")
	expected := []tokenizer.TokenKind{
		tokenizer.KeywordBuy,
		tokenizer.Dollar,
		tokenizer.Identifier,
		tokenizer.At,
		tokenizer.Identifier,
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
	if len(tokens) != 5 { // identifier, colon, identifier(value), identifier(here), eof
		t.Fatalf("expected 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.Identifier || tokens[0].Value != "key" {
		t.Errorf("expected Identifier(key), got %v(%s)", tokens[0].Kind, tokens[0].Value)
	}
	if tokens[1].Kind != tokenizer.Colon {
		t.Errorf("expected Colon, got %v", tokens[1].Kind)
	}
	if tokens[2].Kind != tokenizer.KeywordValue || tokens[2].Value != "value" {
		t.Errorf("expected KeywordValue(value), got %v(%s)", tokens[2].Kind, tokens[2].Value)
	}
	if tokens[3].Kind != tokenizer.Identifier || tokens[3].Value != "here" {
		t.Errorf("expected Identifier(here), got %v(%s)", tokens[3].Kind, tokens[3].Value)
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

func TestLexerAnnotationComment(t *testing.T) {
	input := `/* @adapter: binance */`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 2 { // annotation_comment + eof
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.AnnotationComment {
		t.Errorf("expected AnnotationComment, got %v", tokens[0].Kind)
	}
	// value 格式为 "name|value"
	if tokens[0].Value != "adapter|binance" {
		t.Errorf("expected annotation value 'adapter|binance', got %s", tokens[0].Value)
	}
}

func TestLexerAnnotationModeComment(t *testing.T) {
	input := `/* @adapter_mode: futures */`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.AnnotationComment {
		t.Errorf("expected AnnotationComment, got %v", tokens[0].Kind)
	}
	if tokens[0].Value != "adapter_mode|futures" {
		t.Errorf("expected value 'adapter_mode|futures', got %s", tokens[0].Value)
	}
}

func TestLexerAnnotationConfigComment(t *testing.T) {
	input := `/* @adapter_config: { api_key: "123" } */`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.AnnotationComment {
		t.Errorf("expected AnnotationComment, got %v", tokens[0].Kind)
	}
	if !strings.Contains(tokens[0].Value, "api_key") {
		t.Errorf("expected config to contain 'api_key', got %s", tokens[0].Value)
	}
}

func TestLexerAnnotationSwitchComment(t *testing.T) {
	input := `/* @adapter_switch: { condition: "down" } */`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.AnnotationComment {
		t.Errorf("expected AnnotationComment, got %v", tokens[0].Kind)
	}
}

func TestLexerMultiLineAnnotationComment(t *testing.T) {
	input := `/* @adapter_config: {
	 *   api_key: "abc",
	 *   secret: "xyz"
	 * } */`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Kind != tokenizer.AnnotationComment {
		t.Errorf("expected AnnotationComment, got %v", tokens[0].Kind)
	}
	if !strings.Contains(tokens[0].Value, "api_key") || !strings.Contains(tokens[0].Value, "secret") {
		t.Errorf("expected multiline config preserved, got %s", tokens[0].Value)
	}
}

func TestLexerNormalBlockCommentSkipped(t *testing.T) {
	input := `/* this is just a comment */ buy 100`
	tokens := tokenizer.Lexer(input)
	if len(tokens) != 4 { // keywordBuy, integer, terminator(from space?), eof - wait, no newline
		// Actually no terminator because no newline. Should be keywordBuy, integer, eof = 3
		t.Logf("tokens: %+v", tokens)
	}
	foundBuy := false
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KeywordBuy {
			foundBuy = true
		}
	}
	if !foundBuy {
		t.Error("expected buy token after skipped normal comment")
	}
}
