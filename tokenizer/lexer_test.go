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
	if len(tokens) != 4 {
		t.Error("not pass tokenizer")
	}

	tokens2 := tokenizer.Lexer(`buy 20%`)
	if len(tokens2) != 3 {
		t.Error("lexer not pas")
	}

	tokens3 := tokenizer.Lexer("stop 30%...50%")
	if len(tokens3) != 6 {
		t.Error("lexer is not pass")
	}
	tokens4 := tokenizer.Lexer("keep 1H 1m 1s")
	if len(tokens4) != 7 {
		t.Error("lexer is not pass")
	}

}
