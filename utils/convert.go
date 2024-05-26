package utils

import (
	"strconv"
	"strings"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
)

func Size(value float64, from, to string) float64 {
	units := map[string]float64{
		"":  1,
		"k": 1000,
		"w": 10000,
		"m": 1000000,
		"b": 1000000000,
	}
	from = strings.ToLower(from)
	to = strings.ToLower(to)
	
	fromFactor, fromValid := units[from]
	toFactor, toValid := units[to]
	
	if fromValid && toValid {
		return (value * fromFactor) / toFactor
	} else {
		return -1
	}
}

func ConvertAmount(val ast.Literal) float64 {
	value, _ := strconv.ParseFloat(val.Value, 64)
	if !tokenizer.IsAmountUnit(val.Unit) {
		return value
	}
	switch val.Unit {
	case "k", "K":
		return Size(value, "k", "")
	case "w", "W":
		return Size(value, "w", "")
	case "m", "M":
		return Size(value, "m", "")
	case "b", "B":
		return Size(value, "b", "")
	default:
		return value
	}
}