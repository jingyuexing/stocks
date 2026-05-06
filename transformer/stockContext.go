package transformer

import (
	"errors"
	"sync"
	"time"

	"github.com/jingyuexing/go-utils"
)

// StockContext 存储策略执行过程中的状态与事件
type StockContext struct {
	mu sync.RWMutex

	Code string // 股票代码

	// 任务起始时间
	start time.Time // 起始时间
	// 起始价格
	beginPrice float64 // 起始价格
	// 持续时长
	KeepDuration time.Duration
	// 钱包资金
	amount float64 // 资金
	// 当前价格
	currentPrice float64 //当前价格
	// 事件发射器
	Emiter *utils.EventEmit

	stop     func() bool
	keep     func(duration int64) bool
	duration utils.DateTime
}

// NewStockContext 创建新的执行上下文
func NewStockContext() *StockContext {
	return &StockContext{
		start:  time.Now(),
		Emiter: utils.NewEventEmit(),
	}
}

func (s *StockContext) GetBeginTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.start
}

func (s *StockContext) SetCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if code != "" {
		s.Code = code
	}
}

func (s *StockContext) GetBeginPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.beginPrice
}

func (s *StockContext) SetBeginPrice(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.beginPrice = val
}

func (s *StockContext) GetAmount() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.amount
}

func (s *StockContext) SetAmount(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.amount = val
}

func (s *StockContext) GetCurrentPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentPrice
}

func (s *StockContext) SetCurrentPrice(val float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentPrice = val
	return s.currentPrice
}

func (s *StockContext) StopCallback(callback func() bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if callback != nil {
		s.stop = callback
	}
}

func (s *StockContext) KeepCallback(callback func(duration int64) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if callback != nil {
		s.keep = callback
	}
}

func (s *StockContext) SetDuration(dt utils.DateTime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.duration = dt
}

func (s *StockContext) GetDuration() utils.DateTime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.duration
}

func (s *StockContext) UseStop() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.stop == nil {
		return false, errors.New(`the callback "stop" is nil`)
	}
	return s.stop(), nil
}

func (s *StockContext) UseKeep() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.keep == nil {
		defaultKeep := func(duration int64) bool {
			// 注意：不在闭包内调用加锁方法，避免死锁
			return s.start.After(time.Now())
		}
		s.keep = defaultKeep
	}
	return s.keep(0), nil
}

func (s *StockContext) Sell() {
	s.Emiter.Emit("sell", s)
}

func (s *StockContext) Buy() {
	s.Emiter.Emit("buy", s)
}
