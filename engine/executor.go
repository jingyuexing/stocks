package engine

import (
	"fmt"

	"github.com/jingyuexing/stocks/adapter"
	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/transformer"
)

// Executor 策略动作执行器，负责订阅运行时事件并驱动 AST 动作执行
type Executor struct {
	ctx        *transformer.StockContext
	astRoot    ast.RootNode
	ruleEng    *RuleEngine
	riskGuard  *RiskGuard
	adapterMgr *adapter.Manager
}

// NewExecutor 创建执行器
func NewExecutor(ctx *transformer.StockContext, astRoot ast.RootNode) *Executor {
	return &Executor{
		ctx:       ctx,
		astRoot:   astRoot,
		ruleEng:   NewRuleEngine(ctx),
		riskGuard: NewRiskGuard(ctx),
	}
}

// SetAdapterManager 注入适配器管理器（由 Engine 调用）
func (ex *Executor) SetAdapterManager(mgr *adapter.Manager) {
	ex.adapterMgr = mgr
}

// Setup 注册所有事件监听器，应在 Engine Start 后调用
func (ex *Executor) Setup() {
	if ex.ctx == nil || ex.ctx.Emiter == nil {
		return
	}

	// 网格检查：遍历 AST 中的 LevelBlock，用 RuleEngine 判断范围命中
	ex.ctx.Emiter.On("grid_check", ex.onGridCheck)

	// 止损/止盈触发：遍历 Loss / Profit 节点
	ex.ctx.Emiter.On("stop_triggered", ex.onStopTriggered)

	// 持仓超时：执行 keep 相关的默认平仓逻辑
	ex.ctx.Emiter.On("keep_expired", ex.onKeepExpired)

	// 交易动作事件（用于日志/风控 hook，实际下单在 executeAction 中）
	ex.ctx.Emiter.On("buy", ex.onActionEvent)
	ex.ctx.Emiter.On("sell", ex.onActionEvent)
	ex.ctx.Emiter.On("sell_short", ex.onActionEvent)
	ex.ctx.Emiter.On("buy_cover", ex.onActionEvent)
}

// onGridCheck 处理网格定时检查事件
func (ex *Executor) onGridCheck(args ...any) {
	profit := ex.ruleEng.calcProfitPercent()
	for _, expr := range ex.astRoot.Expression {
		ex.evalGridExpression(expr, profit)
	}
}

// evalGridExpression 递归求值 Grid / LevelBlock / Loss / Profit
func (ex *Executor) evalGridExpression(expr ast.ExpressionNode, profit float64) {
	switch expr.Type {
	case ast.GridExpression:
		// LevelBlock：带 Range 的 GridExpression 子节点
		if expr.Range != nil {
			hit, err := ex.ruleEng.EvalCondition(expr, profit)
			if err != nil {
				fmt.Printf("[executor] grid level eval error: %v\n", err)
				return
			}
			if hit {
				ex.ExecuteBody(expr.Body)
			}
		} else {
			// 无 Range 的顶层 grid，递归处理 body
			ex.ExecuteBody(expr.Body)
		}
	case ast.LongExpression, ast.ShortExpression, ast.BothExpression, ast.PortfolioExpression:
		ex.ExecuteBody(expr.Body)
	case ast.LossExpression:
		// loss 通常由 stop_triggered 处理，但也可在 grid_check 中复合判断
		hit, err := ex.ruleEng.EvalCondition(expr, profit)
		if err == nil && hit {
			ex.ExecuteBody(expr.Body)
		}
	case ast.ProfitExpression:
		hit, err := ex.ruleEng.EvalCondition(expr, profit)
		if err == nil && hit {
			ex.ExecuteBody(expr.Body)
		}
	}
}

// onStopTriggered 处理止损/止盈事件
func (ex *Executor) onStopTriggered(args ...any) {
	profit := ex.ruleEng.calcProfitPercent()
	for _, expr := range ex.astRoot.Expression {
		ex.evalStopExpression(expr, profit)
	}
}

// evalStopExpression 递归查找 Loss / Profit / Stop 节点
func (ex *Executor) evalStopExpression(expr ast.ExpressionNode, profit float64) {
	switch expr.Type {
	case ast.GridExpression, ast.LongExpression, ast.ShortExpression, ast.BothExpression, ast.PortfolioExpression:
		for _, child := range expr.Body {
			ex.evalStopExpression(child, profit)
		}
	case ast.LossExpression:
		hit, err := ex.ruleEng.EvalCondition(expr, profit)
		if err != nil {
			fmt.Printf("[executor] loss eval error: %v\n", err)
			return
		}
		if hit {
			fmt.Printf("[executor] loss condition hit at %.2f%%\n", profit)
			ex.ExecuteBody(expr.Body)
		}
	case ast.ProfitExpression:
		hit, err := ex.ruleEng.EvalCondition(expr, profit)
		if err != nil {
			fmt.Printf("[executor] profit eval error: %v\n", err)
			return
		}
		if hit {
			fmt.Printf("[executor] profit condition hit at %.2f%%\n", profit)
			ex.ExecuteBody(expr.Body)
		}
	case ast.StopExpression:
		// stop 表达式本身由 Scheduler 触发 stop_triggered
		// 若 stop 出现在 body 中，直接执行子动作
		ex.ExecuteBody(expr.Body)
	}
}

// onKeepExpired 处理持仓超时事件
func (ex *Executor) onKeepExpired(args ...any) {
	// 默认行为：触发 stop 或平仓（当前仅日志，后续对接 RiskGuard / Adapter）
	fmt.Println("[executor] keep expired, consider closing position")
}

// onActionEvent 统一交易动作事件回调
func (ex *Executor) onActionEvent(args ...any) {
	// 目前用于日志与风控 hook 预留
	if len(args) > 0 {
		if ctx, ok := args[0].(*transformer.StockContext); ok {
			_ = ctx
			// TODO: 第四阶段对接 RiskGuard 与 AdapterManager
		}
	}
}

// ExecuteBody 顺序执行 AST Body 中的表达式
func (ex *Executor) ExecuteBody(body []ast.ExpressionNode) {
	for _, expr := range body {
		ex.executeExpression(expr)
	}
}

// executeExpression 执行单个表达式节点
func (ex *Executor) executeExpression(expr ast.ExpressionNode) {
	switch expr.Type {
	case ast.BuyExpression:
		act := ex.BuildAction(ActionBuy, expr)
		ex.dispatchAction(act)
	case ast.SellExpression:
		act := ex.BuildAction(ActionSell, expr)
		ex.dispatchAction(act)
	case ast.SellShortExpression:
		act := ex.BuildAction(ActionSellShort, expr)
		ex.dispatchAction(act)
	case ast.BuyCoverExpression:
		act := ex.BuildAction(ActionBuyCover, expr)
		ex.dispatchAction(act)
	case ast.GridExpression, ast.LossExpression, ast.ProfitExpression:
		// 嵌套块级表达式，递归执行
		ex.ExecuteBody(expr.Body)
	case ast.ConditionalExpression:
		// 条件编译/运行时条件判断
		if len(expr.Body) > 0 {
			cond := expr.Body[0]
			ok, err := ex.ruleEng.EvalCondition(cond, 0)
			if err == nil && ok {
				ex.ExecuteBody(expr.Body[1:])
			}
		}
	case ast.StopExpression:
		ex.ExecuteBody(expr.Body)
	default:
		// 全局配置、注释等无需执行
	}
}

// BuildAction 从 AST 动作表达式构造标准化 Action
func (ex *Executor) BuildAction(typ ActionType, expr ast.ExpressionNode) Action {
	act := Action{
		Type:   typ,
		Symbol: ex.ctx.Code,
		Source: expr.Type.String(),
	}
	if len(expr.Params) > 0 {
		p := expr.Params[0]
		act.Amount = p.AsFloat()
		if p.Node.Type == ast.PercentLiteral || p.Unit == "%" {
			act.IsPercent = true
		}
	}
	return act
}

// dispatchAction 分发动作：风控检查 -> 网关下单
func (ex *Executor) dispatchAction(act Action) {
	fmt.Printf("[executor] dispatch action: %s %s %.2f", act.Type, act.Symbol, act.Amount)
	if act.IsPercent {
		fmt.Print("%")
	}
	fmt.Println()

	// 1. 风控检查
	if ex.riskGuard != nil {
		if err := ex.riskGuard.Check(act); err != nil {
			fmt.Printf("[executor] risk guard blocked: %v\n", err)
			return
		}
	}

	// 2. 通过适配器发送交易指令
	if ex.adapterMgr != nil && ex.adapterMgr.IsLoaded() {
		ex.sendToGateway(act)
		return
	}

	// 3. 无网关时回退到 StockContext 原始事件（保持与现有测试兼容）
	switch act.Type {
	case ActionBuy:
		ex.ctx.Buy()
	case ActionSell:
		ex.ctx.Sell()
	case ActionSellShort:
		ex.ctx.SellShort()
	case ActionBuyCover:
		ex.ctx.BuyCover()
	}
}

// sendToGateway 将 Action 发送到 TradingGateway
func (ex *Executor) sendToGateway(act Action) {
	gw := ex.adapterMgr.Primary()
	if gw == nil {
		fmt.Println("[executor] gateway not available")
		return
	}

	var orderID string
	var err error

	switch act.Type {
	case ActionBuy:
		if act.IsMarket || act.Price == 0 {
			orderID, err = gw.BuyMarket(act.Symbol, act.Amount)
		} else {
			orderID, err = gw.BuyLimit(act.Symbol, act.Amount, act.Price)
		}
	case ActionSell:
		if act.IsMarket || act.Price == 0 {
			orderID, err = gw.SellMarket(act.Symbol, act.Amount)
		} else {
			orderID, err = gw.SellLimit(act.Symbol, act.Amount, act.Price)
		}
	case ActionSellShort:
		if act.IsMarket || act.Price == 0 {
			orderID, err = gw.SellShortMarket(act.Symbol, act.Amount)
		} else {
			orderID, err = gw.SellShortLimit(act.Symbol, act.Amount, act.Price)
		}
	case ActionBuyCover:
		if act.IsMarket || act.Price == 0 {
			orderID, err = gw.BuyCoverMarket(act.Symbol, act.Amount)
		} else {
			orderID, err = gw.BuyCoverLimit(act.Symbol, act.Amount, act.Price)
		}
	}

	if err != nil {
		fmt.Printf("[executor] gateway order error: %v\n", err)
		gw.OnError(act.Symbol, err.Error())
	} else {
		fmt.Printf("[executor] gateway order placed: %s\n", orderID)
	}
}
