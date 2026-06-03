package engine

// ActionType 交易动作类型
type ActionType string

const (
	ActionBuy       ActionType = "buy"
	ActionSell      ActionType = "sell"
	ActionSellShort ActionType = "sell_short"
	ActionBuyCover  ActionType = "buy_cover"
)

// Action 标准化的交易动作
type Action struct {
	// Type 动作类型
	Type ActionType
	// Symbol 标的代码
	Symbol string
	// Amount 数量（具体含义由 IsPercent 决定）
	Amount float64
	// IsPercent 为 true 时 Amount 表示百分比（如 50%）
	IsPercent bool
	// Price 限价，0 表示市价
	Price float64
	// IsMarket 是否为市价单
	IsMarket bool
	// Source 触发该动作的 AST 节点类型（用于日志/回溯）
	Source string
}
