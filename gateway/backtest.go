package gateway

import (
	"fmt"
	"sync"
)

// BacktestGateway 回测网关，模拟交易执行并记录成交
type BacktestGateway struct {
	mu        sync.RWMutex
	connected bool
	config    map[string]any

	// 模拟持仓
	positions map[string]float64
	avgCosts  map[string]float64
	cash      float64

	// 模拟委托簿
	orders map[string]*backtestOrder
}

type backtestOrder struct {
	ID     string
	Symbol string
	Side   string
	Amount float64
	Price  float64
	Status OrderStatus
}

// NewBacktestGateway 创建回测网关
func NewBacktestGateway() *BacktestGateway {
	return &BacktestGateway{
		positions: make(map[string]float64),
		avgCosts:  make(map[string]float64),
		orders:    make(map[string]*backtestOrder),
		cash:      1_000_000, // 默认初始资金
	}
}

func (b *BacktestGateway) Connect(config map[string]any) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = config
	b.connected = true
	if v, ok := config["initial_cash"]; ok {
		if f, ok2 := v.(float64); ok2 {
			b.cash = f
		}
	}
	fmt.Println("[backtest] connected")
	return nil
}

func (b *BacktestGateway) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.connected = false
	fmt.Println("[backtest] disconnected")
	return nil
}

func (b *BacktestGateway) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

func (b *BacktestGateway) GetPosition(symbol string) (float64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.positions[symbol], nil
}

func (b *BacktestGateway) GetAverageCost(symbol string) (float64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.avgCosts[symbol], nil
}

func (b *BacktestGateway) GetAvailableBalance(symbol string) (float64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cash, nil
}

func (b *BacktestGateway) GetMarketPrice(symbol string) (float64, error) {
	// 回测引擎中由外部注入当前价格，此处返回 0 占位
	return 0, nil
}

func (b *BacktestGateway) BuyMarket(symbol string, amount float64) (string, error) {
	return b.fillOrder(symbol, "buy", amount, 0)
}

func (b *BacktestGateway) SellMarket(symbol string, amount float64) (string, error) {
	return b.fillOrder(symbol, "sell", amount, 0)
}

func (b *BacktestGateway) BuyLimit(symbol string, amount float64, price float64) (string, error) {
	return b.fillOrder(symbol, "buy", amount, price)
}

func (b *BacktestGateway) SellLimit(symbol string, amount float64, price float64) (string, error) {
	return b.fillOrder(symbol, "sell", amount, price)
}

func (b *BacktestGateway) SellShortMarket(symbol string, amount float64) (string, error) {
	return b.fillOrder(symbol, "sell_short", amount, 0)
}

func (b *BacktestGateway) SellShortLimit(symbol string, amount float64, price float64) (string, error) {
	return b.fillOrder(symbol, "sell_short", amount, price)
}

func (b *BacktestGateway) BuyCoverMarket(symbol string, amount float64) (string, error) {
	return b.fillOrder(symbol, "buy_cover", amount, 0)
}

func (b *BacktestGateway) BuyCoverLimit(symbol string, amount float64, price float64) (string, error) {
	return b.fillOrder(symbol, "buy_cover", amount, price)
}

func (b *BacktestGateway) CancelOrder(orderID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if o, ok := b.orders[orderID]; ok {
		o.Status = OrderCancelled
	}
	return nil
}

func (b *BacktestGateway) GetOrderStatus(orderID string) (OrderStatus, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if o, ok := b.orders[orderID]; ok {
		return o.Status, nil
	}
	return OrderRejected, fmt.Errorf("order not found: %s", orderID)
}

func (b *BacktestGateway) OnEngineStart() error {
	fmt.Println("[backtest] engine start")
	return nil
}

func (b *BacktestGateway) OnEngineStop() error {
	fmt.Println("[backtest] engine stop")
	return nil
}

func (b *BacktestGateway) OnError(symbol string, errMsg string) {
	fmt.Printf("[backtest] error on %s: %s\n", symbol, errMsg)
}

// fillOrder 模拟成交
func (b *BacktestGateway) fillOrder(symbol string, side string, amount, price float64) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := fmt.Sprintf("bt_%d", len(b.orders)+1)
	order := &backtestOrder{
		ID:     id,
		Symbol: symbol,
		Side:   side,
		Amount: amount,
		Price:  price,
		Status: OrderFilled,
	}
	b.orders[id] = order

	switch side {
	case "buy":
		b.positions[symbol] += amount
		// 简化：不更新均价
		fmt.Printf("[backtest] BUY %s %.2f @ %.2f\n", symbol, amount, price)
	case "sell":
		b.positions[symbol] -= amount
		fmt.Printf("[backtest] SELL %s %.2f @ %.2f\n", symbol, amount, price)
	case "sell_short":
		fmt.Printf("[backtest] SELL_SHORT %s %.2f @ %.2f\n", symbol, amount, price)
	case "buy_cover":
		fmt.Printf("[backtest] BUY_COVER %s %.2f @ %.2f\n", symbol, amount, price)
	}

	return id, nil
}
