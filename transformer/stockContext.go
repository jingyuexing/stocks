package transformer

import (
	"errors"
	"time"

	"github.com/jingyuexing/go-utils"
)

type StockContext struct {
	// 任务起始时间
	start time.Time // 起始时间
	// 起始价格
	beginPrice float64 // 起始价格
	// 钱包资金
	amount       float64 // 资金
	currentPrice float64 //当前价格
	Emiter       *utils.EventEmit

	stop     func() bool
	keep     func() bool
	duration utils.DateTime
}

func (s StockContext) GetBeginTime() time.Time {
	return s.start
}

func (s StockContext) GetBeginPrice() float64 {
	return s.beginPrice
}
func (s StockContext) GetAmount() float64 {
	return s.amount
}
func (s StockContext) SetCurrentPrice(val float64) float64 {
	s.currentPrice = val
	return s.currentPrice
}

func (s StockContext) StopCallback(callback func() bool) {
	if callback != nil {
		s.stop = callback
	}
}

func (s StockContext) KeepCallback(callback func() bool) {
	if callback != nil {
		s.keep = callback
	}
}

func (s StockContext) SetDuration(time utils.DateTime) {
	s.duration = time
}

func (s StockContext) UseStop() (bool, error) {
	if s.stop == nil {
		return false, errors.New(`the callback "stop" is nil`)
	}
	return s.stop(), nil
}
func (s StockContext) UseKeep() (bool, error) {
	defaultKeep := func() bool {
		return s.GetBeginTime().After(time.Now())
	}
	if s.keep == nil {
		s.keep = defaultKeep
	}
	return s.keep(), nil
}

func (s StockContext) Sell() {
	s.Emiter.Emit("sell", s)
}

func (s StockContext) Buy() {
	s.Emiter.Emit("buy", s)
}
