package parser_test

import (
	"testing"

	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
)

func TestParser(t *testing.T) {
	token := tokenizer.Lexer("buy 100")
	ast := parser.Parser(token)

	if len(ast.Expression) != 1 {
		t.Error("parser has wrong")
	}

	ast2 := parser.Parser(tokenizer.Lexer(`stop 20%`))
	if len(ast2.Expression[0].Params) != 1 {
		t.Error("has wrong")
	}
	ast3 := parser.Parser(tokenizer.Lexer(`keep 1M 1h 1m 1s`))
	if len(ast3.Expression[0].Params) != 4 {
		t.Error("parser has wrong")
	}

}
