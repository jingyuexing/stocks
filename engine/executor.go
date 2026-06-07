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
	astRoot    *ast.ProgramNode
	ruleEng    *RuleEngine
	riskGuard  *RiskGuard
	adapterMgr *adapter.Manager
}

// NewExecutor 创建执行器
func NewExecutor(ctx *transformer.StockContext, astRoot *ast.ProgramNode) *Executor {
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

	// 网格检查：遍历 AST 中的 level FlowStmt，用 RuleEngine 判断范围命中
	ex.ctx.Emiter.On("grid_check", ex.onGridCheck)

	// 止损/止盈触发：遍历 level / trigger FlowStmt 节点
	ex.ctx.Emiter.On("stop_triggered", ex.onStopTriggered)

	// 持仓超时：执行 keep 相关的默认平仓逻辑
	ex.ctx.Emiter.On("keep_expired", ex.onKeepExpired)

	// 交易动作事件（用于日志/风控 hook，实际下单在 dispatchAction 中）
	ex.ctx.Emiter.On("buy", ex.onActionEvent)
	ex.ctx.Emiter.On("sell", ex.onActionEvent)
	ex.ctx.Emiter.On("sell_short", ex.onActionEvent)
	ex.ctx.Emiter.On("buy_cover", ex.onActionEvent)
}

// onGridCheck 处理网格定时检查事件
func (ex *Executor) onGridCheck(args ...any) {
	profit := ex.calcProfitPercent()
	if ex.astRoot == nil {
		return
	}
	// 使用 goroutine 避免在 EventEmit 回调中嵌套触发事件导致的死锁
	go func() {
		for _, stmt := range ex.astRoot.Statements {
			ex.evalStmtForGrid(stmt, profit)
		}
	}()
}

// evalStmtForGrid 递归查找 Strategy / Level 节点并评估
func (ex *Executor) evalStmtForGrid(stmt ast.Stmt, profit float64) {
	switch s := stmt.(type) {
	case *ast.StrategyStmtNode:
		for _, child := range s.Body {
			ex.evalStmtForGrid(child, profit)
		}
	case *ast.FlowStmtNode:
		if s.Kind == "level" {
			ex.evalLevelFlow(s, profit)
		}
	}
}

// evalLevelFlow 评估 level 条件并执行 body
func (ex *Executor) evalLevelFlow(s *ast.FlowStmtNode, profit float64) {
	if s.Condition == nil {
		return
	}
	var hit bool
	switch cond := s.Condition.(type) {
	case *ast.RangeExprNode:
		hit = ex.ruleEng.EvalRange(cond, profit)
	case *ast.BinaryExprNode:
		ok, err := ex.ruleEng.EvalCondition(cond)
		if err != nil {
			fmt.Printf("[executor] level condition eval error: %v\n", err)
			return
		}
		hit = ok
	case *ast.Literal:
		// 单点阈值：负数表示 loss，正数表示 profit
		threshold := cond.AsFloat()
		if threshold < 0 {
			hit = profit <= threshold
		} else {
			hit = profit >= threshold
		}
	}
	if hit {
		ex.ExecuteBody(s.Body)
	}
}

// onStopTriggered 处理止损/止盈事件
func (ex *Executor) onStopTriggered(args ...any) {
	profit := ex.calcProfitPercent()
	if ex.astRoot == nil {
		return
	}
	// 使用 goroutine 避免在 EventEmit 回调中嵌套触发事件导致的死锁
	go func() {
		for _, stmt := range ex.astRoot.Statements {
			ex.evalStmtForStop(stmt, profit)
		}
	}()
}

// evalStmtForStop 递归查找 Level / Trigger 节点
func (ex *Executor) evalStmtForStop(stmt ast.Stmt, profit float64) {
	switch s := stmt.(type) {
	case *ast.StrategyStmtNode:
		for _, child := range s.Body {
			ex.evalStmtForStop(child, profit)
		}
	case *ast.FlowStmtNode:
		if s.Kind == "level" || s.Kind == "trigger" {
			ex.evalLevelFlow(s, profit)
		}
	}
}

// onKeepExpired 处理持仓超时事件
func (ex *Executor) onKeepExpired(args ...any) {
	fmt.Println("[executor] keep expired, consider closing position")
}

// onActionEvent 统一交易动作事件回调
func (ex *Executor) onActionEvent(args ...any) {
	if len(args) == 0 {
		return
	}
	switch v := args[0].(type) {
	case *transformer.StockContext:
		_ = v
	case *transformer.Event:
		_ = v
	}
}

// ExecuteBody 顺序执行 AST Body 中的语句
func (ex *Executor) ExecuteBody(body []ast.Stmt) {
	for _, stmt := range body {
		ex.executeStmt(stmt)
	}
}

// executeStmt 执行单个语句节点
func (ex *Executor) executeStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ActionStmtNode:
		ex.executeAction(s)
	case *ast.FlowStmtNode:
		// 嵌套 flow，按当前 profit 评估
		ex.evalLevelFlow(s, ex.calcProfitPercent())
	case *ast.StrategyStmtNode:
		ex.ExecuteBody(s.Body)
	}
}

// executeAction 执行单个动作
func (ex *Executor) executeAction(s *ast.ActionStmtNode) {
	switch s.Action {
	case "buy":
		act := ex.BuildAction(ActionBuy, s)
		ex.dispatchAction(act)
	case "sell":
		act := ex.BuildAction(ActionSell, s)
		ex.dispatchAction(act)
	case "sell_short":
		act := ex.BuildAction(ActionSellShort, s)
		ex.dispatchAction(act)
	case "buy_cover":
		act := ex.BuildAction(ActionBuyCover, s)
		ex.dispatchAction(act)
	}
}

// BuildAction 从 AST 动作语句构造标准化 Action
func (ex *Executor) BuildAction(typ ActionType, s *ast.ActionStmtNode) Action {
	act := Action{
		Type:   typ,
		Symbol: ex.ctx.GetCode(),
		Source: s.Action,
	}
	if len(s.Args) > 0 {
		if lit, ok := s.Args[0].(*ast.Literal); ok {
			act.Amount = lit.AsFloat()
			if lit.Unit == "%" {
				act.IsPercent = true
			}
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

// calcProfitPercent 计算当前盈亏百分比
func (ex *Executor) calcProfitPercent() float64 {
	begin := ex.ctx.GetBeginPrice()
	current := ex.ctx.GetCurrentPrice()
	if begin <= 0 || current <= 0 {
		return 0
	}
	return (current - begin) / begin * 100
}
