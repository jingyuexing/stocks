package transformer_test

import (
	"testing"

	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
	"github.com/jingyuexing/stocks/transformer"
)

func TestTransfomer(t *testing.T) {
	result := transformer.Transformer(
		parser.Parser(tokenizer.Lexer("keep 1W")),
	)
}
