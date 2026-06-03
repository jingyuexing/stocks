package types

type Adapter interface {
	Sell(code string, volumn float64) error
	Buy(code string, volumn float64) error
	SellShort(code string, volumn float64) error // v2.1
	BuyCover(code string, volumn float64) error  // v2.1
}
