package types

type Adapter interface {
	Sell(code string, volumn float64) error
	Buy(code string, volumn float64) error
}
