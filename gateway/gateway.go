package gateway

import "fmt"

// TradingGateway 交易网关接口，对接真实交易所或回测引擎
type TradingGateway interface {
	// 连接与生命周期
	Connect(config map[string]any) error
	Disconnect() error
	IsConnected() bool

	// 行情查询
	GetPosition(symbol string) (float64, error)
	GetAverageCost(symbol string) (float64, error)
	GetAvailableBalance(symbol string) (float64, error)
	GetMarketPrice(symbol string) (float64, error)

	// 下单（市价/限价）
	BuyMarket(symbol string, amount float64) (string, error)
	SellMarket(symbol string, amount float64) (string, error)
	BuyLimit(symbol string, amount float64, price float64) (string, error)
	SellLimit(symbol string, amount float64, price float64) (string, error)
	SellShortMarket(symbol string, amount float64) (string, error)
	SellShortLimit(symbol string, amount float64, price float64) (string, error)
	BuyCoverMarket(symbol string, amount float64) (string, error)
	BuyCoverLimit(symbol string, amount float64, price float64) (string, error)

	// 委托管理
	CancelOrder(orderID string) error
	GetOrderStatus(orderID string) (OrderStatus, error)

	// 事件回调（引擎注入）
	OnEngineStart() error
	OnEngineStop() error
	OnError(symbol string, errMsg string)
}

// OrderStatus 委托状态
type OrderStatus string

const (
	OrderPending     OrderStatus = "PENDING"
	OrderPartialFill OrderStatus = "PARTIAL_FILL"
	OrderFilled      OrderStatus = "FILLED"
	OrderCancelled   OrderStatus = "CANCELLED"
	OrderRejected    OrderStatus = "REJECTED"
)

// Factory 网关工厂函数签名
type Factory func() TradingGateway

// registry 内置网关注册表
var registry = map[string]Factory{
	"backtest": func() TradingGateway { return NewBacktestGateway() },
}

// Register 注册自定义网关
func Register(name string, f Factory) {
	registry[name] = f
}

// Create 根据名称创建网关实例
func Create(name string) (TradingGateway, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown gateway: %s", name)
	}
	return f(), nil
}

// KnownGateways 返回已注册网关名称列表
func KnownGateways() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}
