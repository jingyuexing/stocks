package compiler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/tokenizer"
	"github.com/jingyuexing/stocks/transformer"
)

// BuildContext 根据 AST 构建运行时执行上下文
func BuildContext(root *ast.ProgramNode) *transformer.StockContext {
	ctx := transformer.NewStockContext()
	processStmts(ctx, root.Statements)
	return ctx
}

// duration 将 AST 表达式解析为 time.Duration，支持 ListLiteralNode 和带 unit 的 Literal
func duration(expr ast.Expr) time.Duration {
	var d time.Duration
	if expr == nil {
		return d
	}
	switch v := expr.(type) {
	case *ast.ListLiteralNode:
		for _, item := range v.Items {
			d += accumulateDuration(item)
		}
	default:
		d = accumulateDuration(expr)
	}
	return d
}

// accumulateDuration 累加单个 AST 表达式中的 duration
func accumulateDuration(expr ast.Expr) time.Duration {
	if lit, ok := expr.(*ast.Literal); ok && lit.Unit != "" {
		return toLiteralValue(lit).AsDuration()
	}
	return 0
}

// durationFromParams 从参数列表中累加所有 duration（支持混合参数）
func durationFromParams(params []ast.Expr) time.Duration {
	var d time.Duration
	for _, p := range params {
		d += duration(p)
	}
	return d
}

// parseFloatExpr 将 AST 表达式解析为 float64（仅支持 Literal）
func parseFloatExpr(expr ast.Expr) float64 {
	if expr == nil {
		return 0
	}
	if lit, ok := expr.(*ast.Literal); ok {
		if lit.Value == "" {
			return 0
		}
		v, _ := strconv.ParseFloat(lit.Value, 64)
		return v
	}
	return 0
}

// parseFloatLiteral 将 AST 字面量解析为 float64
func parseFloatLiteral(lit *ast.Literal) float64 {
	if lit == nil || lit.Value == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(lit.Value, 64)
	return v
}

// stringFromExpr 从表达式中提取字符串值
func stringFromExpr(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	if lit, ok := expr.(*ast.Literal); ok {
		return lit.Value
	}
	if id, ok := expr.(*ast.IdentifierExprNode); ok {
		return id.Name
	}
	return ""
}

// toLiteralValue 将 AST Literal 转换为 transformer 的 LiteralValue，保留完整类型信息
func toLiteralValue(lit *ast.Literal) transformer.LiteralValue {
	if lit == nil {
		return transformer.LiteralValue{}
	}
	typ := transformer.Text
	switch lit.LiteralType {
	case ast.Float:
		typ = transformer.Float
	case ast.Integer:
		typ = transformer.Integer
	case ast.Text:
		typ = transformer.Text
	case ast.Time:
		typ = transformer.Time
	case ast.Date:
		typ = transformer.Date
	case ast.Enum:
		typ = transformer.Enum
	case ast.List:
		typ = transformer.List
	}

	unit := lit.Unit
	// 根据 unit 进一步细化类型
	if unit == "%" {
		typ = transformer.Percentage
	} else if tokenizer.IsTimeUnit(unit) && (typ == transformer.Integer || typ == transformer.Float || typ == transformer.Text) {
		typ = transformer.Duration
	}

	return transformer.LiteralValue{
		Type:  typ,
		Value: lit.Value,
		Unit:  unit,
	}
}

// toRangeValue 将 AST RangeExprNode 转换为 transformer 的 RangeValue
func toRangeValue(r *ast.RangeExprNode) *transformer.RangeValue {
	if r == nil {
		return nil
	}
	rv := &transformer.RangeValue{}
	if r.Begin != nil {
		lit := toLiteralValue(r.Begin)
		rv.Begin = &lit
	}
	if r.End != nil {
		lit := toLiteralValue(r.End)
		rv.End = &lit
	}
	return rv
}

// processStmts 递归处理语句列表，将 AST 节点转换为执行上下文状态
func processStmts(ctx *transformer.StockContext, stmts []ast.Stmt) {
	for _, stmt := range stmts {
		processStmt(ctx, stmt)
	}
}

func processStmt(ctx *transformer.StockContext, stmt ast.Stmt) {
	switch s := stmt.(type) {
	// 策略声明
	case *ast.StrategyStmtNode:
		if s.Target != "" {
			ctx.SetCode(s.Target)
		}
		ctx.SetDirection(s.Kind)
		processStmts(ctx, s.Body)

	// Flow 语句
	case *ast.FlowStmtNode:
		processFlowStmt(ctx, s)

	// 交易动作
	case *ast.ActionStmtNode:
		processActionStmt(ctx, s)

	// 通用配置
	case *ast.ConfigStmtNode:
		processConfigStmt(ctx, s)

	// 选择标的
	case *ast.SelectStmtNode:
		if s.Target != "" {
			ctx.SetCode(s.Target)
		}

	// 变量定义
	case *ast.DefineStmtNode:
		if lit, ok := s.Value.(*ast.Literal); ok {
			ctx.Vars.SetUserVar(s.Name, toLiteralValue(lit))
		}

	// 导入/导出（运行时忽略）
	case *ast.ImportStmtNode, *ast.ExportStmtNode:
		// 运行时无需处理

	// use / template / macro（运行时忽略或扩展）
	case *ast.UseStmtNode, *ast.TemplateStmtNode, *ast.MacroStmtNode:
		// 运行时忽略

	// 条件编译
	case *ast.IfStmtNode:
		// TODO: 运行时求值决定是否激活子策略
		processStmts(ctx, s.Body)

	// 断言
	case *ast.AssertStmtNode:
		// TODO: 启动时求值，失败则报错

	// 注解
	case *ast.AnnotationStmtNode:
		switch s.Key {
		case "adapter":
			ctx.SetAdapter(s.Value)
		case "adapter_mode":
			ctx.SetAdapterMode(s.Value)
		case "adapter_config":
			ctx.SetAdapterConfig(s.Value)
		case "adapter_switch":
			ctx.SetAdapterSwitch(s.Value)
		}
	}
}

func processFlowStmt(ctx *transformer.StockContext, s *ast.FlowStmtNode) {
	switch s.Kind {
	case "level":
		ctx.Emiter.On("grid_check", func(args ...any) {
			// TODO: 根据 Range/Condition 和当前价格判断是否触发
		})
	case "trigger":
		ctx.Emiter.On("trigger", func(args ...any) {
			// TODO: 条件求值 + 频次控制（Frequency）
		})
	case "until":
		ctx.Emiter.On("until", func(args ...any) {
			// TODO: 条件终止逻辑
		})
	case "atomic":
		ctx.Emiter.On("atomic", func(args ...any) {
			// TODO: 原子执行 + rollback/best_effort 模式
		})
	case "override":
		ctx.Emiter.On("override", func(args ...any) {
			// TODO: 覆盖规则逻辑
		})
	}
	processStmts(ctx, s.Body)
}

func processActionStmt(ctx *transformer.StockContext, s *ast.ActionStmtNode) {
	switch s.Action {
	case "buy":
		ctx.Buy()
	case "sell":
		ctx.Sell()
	case "sell_short":
		ctx.SellShort()
	case "buy_cover":
		ctx.BuyCover()
	}
}

func processConfigStmt(ctx *transformer.StockContext, s *ast.ConfigStmtNode) {
	switch s.Key {
	case "keep":
		ctx.SetDuration(durationFromParams(s.Params))
	case "hold_max":
		ctx.SetHoldMax(durationFromParams(s.Params))
	case "cooldown":
		ctx.SetCooldown(durationFromParams(s.Params))
		for _, p := range s.Params {
			if id, ok := p.(*ast.IdentifierExprNode); ok && id.Name == "per_level" {
				ctx.SetCooldownPerLevel(true)
			}
		}
	case "session":
		for _, p := range s.Params {
			ctx.Time.Session = append(ctx.Time.Session, stringFromExpr(p))
		}
	case "active":
		for _, p := range s.Params {
			ctx.Time.ActiveRange = append(ctx.Time.ActiveRange, stringFromExpr(p))
		}
	case "pause":
		for _, p := range s.Params {
			ctx.Time.PauseRange = append(ctx.Time.PauseRange, stringFromExpr(p))
		}
	case "position":
		if len(s.Params) > 0 {
			ctx.Position.Type = stringFromExpr(s.Params[0])
		}
		if len(s.Params) > 1 {
			ctx.Position.MaxPosition = parseFloatExpr(s.Params[1])
		}
	case "sizing":
		if len(s.Params) > 0 {
			ctx.Position.SizingMethod = stringFromExpr(s.Params[0])
		}
	case "base_position":
		if len(s.Params) > 0 {
			ctx.Position.BasePosition = parseFloatExpr(s.Params[0])
		}
	case "pyramid_step":
		if len(s.Params) > 0 {
			ctx.Position.PyramidStep = parseFloatExpr(s.Params[0])
		}
	case "max_pyramid_layers":
		if len(s.Params) > 0 {
			ctx.Position.MaxPyramidLayers = int(parseFloatExpr(s.Params[0]))
		}
	case "fixed_fraction":
		if len(s.Params) > 0 {
			ctx.Position.FixedFraction = parseFloatExpr(s.Params[0])
		}
	case "volatility_target":
		if len(s.Params) > 0 {
			ctx.Position.VolatilityTarget = parseFloatExpr(s.Params[0])
		}
	case "atr_period":
		if len(s.Params) > 0 {
			ctx.Position.AtrPeriod = int(parseFloatExpr(s.Params[0]))
		}
	case "risk_per_trade":
		if len(s.Params) > 0 {
			ctx.Position.RiskPerTrade = parseFloatExpr(s.Params[0])
		}
	case "risk_per_grid":
		if len(s.Params) > 0 {
			ctx.Position.RiskPerGrid = parseFloatExpr(s.Params[0])
		}
	case "max_drawdown":
		if len(s.Params) > 0 {
			ctx.Position.MaxDrawdown = parseFloatExpr(s.Params[0])
		}
	case "position_decay":
		for _, p := range s.Params {
			switch v := p.(type) {
			case *ast.Literal:
				if v.Unit == "" || !tokenizer.IsTimeUnit(v.Unit) {
					ctx.Position.PositionDecay = parseFloatExpr(v)
				} else {
					ctx.Position.PositionDecayPer = duration(v)
				}
			case *ast.ListLiteralNode:
				ctx.Position.PositionDecayPer = duration(v)
			}
		}
	case "gross_exposure":
		if len(s.Params) > 0 {
			ctx.Position.GrossExposure = parseFloatExpr(s.Params[0])
		}
	case "net_exposure":
		if len(s.Params) > 0 {
			ctx.Position.NetExposure = parseFloatExpr(s.Params[0])
		}
	case "beta_neutral":
		if len(s.Params) > 0 {
			ctx.Position.BetaNeutral = stringFromExpr(s.Params[0]) == "true"
		}
	case "rebalance":
		if len(s.Params) > 0 {
			ctx.Position.Rebalance = stringFromExpr(s.Params[0])
		}
		for _, p := range s.Params {
			if lit, ok := p.(*ast.Literal); ok && len(lit.Value) == 5 && lit.Value[2] == ':' {
				ctx.Position.RebalanceAt = lit.Value
			}
		}
	case "leverage":
		if len(s.Params) > 0 {
			ctx.Leverage.Leverage = parseFloatExpr(s.Params[0])
		}
	case "margin":
		if len(s.Params) > 0 {
			ctx.Leverage.MarginMode = stringFromExpr(s.Params[0])
		}
	case "hedge":
		if len(s.Params) > 0 {
			ctx.Leverage.Hedge = stringFromExpr(s.Params[0])
		}
	case "funding_priority":
		if len(s.Params) > 0 {
			ctx.Leverage.FundingPriority = stringFromExpr(s.Params[0])
		}
	case "max_short":
		if len(s.Params) > 0 {
			ctx.Risk.MaxShort = parseFloatExpr(s.Params[0])
		}
	case "borrow_rate_limit":
		if len(s.Params) > 0 {
			ctx.Risk.BorrowRateLimit = parseFloatExpr(s.Params[0])
		}
	case "max_position":
		if len(s.Params) > 0 {
			ctx.Risk.MaxPosition = parseFloatExpr(s.Params[0])
			ctx.Position.MaxPosition = parseFloatExpr(s.Params[0])
		}
	case "stop_loss":
		if s.Range != nil {
			ctx.Risk.StopLossRange = toRangeValue(s.Range)
		} else if len(s.Params) > 0 {
			if lit, ok := s.Params[0].(*ast.Literal); ok {
				litVal := toLiteralValue(lit)
				ctx.Risk.StopLossRange = &transformer.RangeValue{
					Begin: &litVal,
				}
			}
		}
	case "slippage_tolerance":
		if len(s.Params) > 0 {
			ctx.Risk.SlippageTolerance = parseFloatExpr(s.Params[0])
		}
	case "circuit_breaker":
		for _, p := range s.Params {
			switch v := p.(type) {
			case *ast.Literal:
				if v.Unit == "" || !tokenizer.IsTimeUnit(v.Unit) {
					ctx.Risk.CircuitBreaker = parseFloatExpr(v)
				} else {
					ctx.Risk.CircuitBreakerIn = duration(v)
				}
			case *ast.ListLiteralNode:
				ctx.Risk.CircuitBreakerIn = duration(v)
			}
		}
		ctx.Emiter.On("circuit_breaker", func(args ...any) {})
	case "partial_fill":
		if len(s.Params) > 0 {
			ctx.Execution.PartialFill = stringFromExpr(s.Params[0])
		}
	case "compound_profit":
		if len(s.Params) > 0 {
			ctx.Execution.CompoundProfit = stringFromExpr(s.Params[0]) == "true"
		}
	case "skip_if_gapped":
		ctx.Execution.SkipIfGapped = true
	case "fallback":
		if len(s.Params) > 0 {
			ctx.Execution.Fallback = stringFromExpr(s.Params[0])
		}
	default:
		fmt.Printf("unknown config: %s\n", s.Key)
	}
}
