package transformer

import (
	"strconv"

	"github.com/jingyuexing/go-utils"
	"github.com/jingyuexing/stocks/ast"
)

// NewStockContextFromAST 根据 AST 创建新的执行上下文
func NewStockContextFromAST() *StockContext {
	ctx := NewStockContext()
	return ctx
}

func duration(params []ast.Literal) utils.DateTime {
	date := utils.NewDateTime()
	for _, param := range params {
		if param.Node.Type == ast.DurationLiteral {
			val, _ := strconv.Atoi(param.Value)
			date = date.Add(val, utils.AddUnits(param.Unit))
		}
	}
	return date
}

func parseFloatLiteral(lit ast.Literal) float64 {
	if lit.Value == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(lit.Value, 64)
	return v
}

// processExpressions 递归处理表达式列表，将 AST 节点转换为执行上下文状态
func processExpressions(ctx *StockContext, expressions []ast.ExpressionNode) {
	for _, exp := range expressions {
		switch exp.Type {
		// 基础表达式
		case ast.SelectExpression:
			ctx.SetCode(exp.Name)
		case ast.ValueStatement:
			if exp.Value.Value != "" {
				val, _ := strconv.ParseFloat(exp.Value.Value, 64)
				ctx.SetAmount(val)
			}
			if exp.Name != "" {
				ctx.AddVariable(exp.Name, exp.Value)
			}
		case ast.DefineStatement:
			ctx.AddVariable(exp.Name, exp.Value)
		case ast.ReferenceStatement:
			// 引用变量，运行时由调度器解析
		case ast.VariableExpression:
			// 内置变量（如 $profit），运行时由调度器注入

		// 交易动作
		case ast.BuyExpression:
			ctx.Buy()
		case ast.SellExpression:
			ctx.Sell()
		case ast.SellShortExpression:
			ctx.SellShort()
		case ast.BuyCoverExpression:
			ctx.BuyCover()

		// 持仓控制
		case ast.KeepExpression:
			if len(exp.Params) > 0 {
				ctx.SetDuration(duration(exp.Params))
			}
		case ast.StopExpression:
			ctx.Emiter.On("stop", func(args ...any) {
				currentValue, ok := args[0].(*StockContext)
				if !ok {
					return
				}
				isKeep, _ := currentValue.UseKeep()
				if isKeep {
					// [datetime] 当前 止盈/止损 于 {currentPrice} 总计 盈利 /亏损 {cha}, 总盈利率/ 亏损率 为 {}%
				}
			})

		// 策略声明
		case ast.GridExpression:
			if exp.Name != "" {
				ctx.SetCode(exp.Name)
			}
			ctx.Emiter.Emit("grid_check", ctx, exp)
			processExpressions(ctx, exp.Body)
		case ast.LongExpression:
			ctx.SetDirection("long")
			if exp.Name != "" {
				ctx.SetCode(exp.Name)
			}
			processExpressions(ctx, exp.Body)
		case ast.ShortExpression:
			ctx.SetDirection("short")
			if exp.Name != "" {
				ctx.SetCode(exp.Name)
			}
			processExpressions(ctx, exp.Body)
		case ast.BothExpression:
			ctx.SetDirection("both")
			if exp.Name != "" {
				ctx.SetCode(exp.Name)
			}
			processExpressions(ctx, exp.Body)
		case ast.PortfolioExpression:
			processExpressions(ctx, exp.Body)

		// 模板与复用
		case ast.TemplateExpression:
			// 注册模板定义（供后续 use 解析）
		case ast.UseExpression:
			// 实例化模板并合并到当前上下文
		case ast.ImportExpression:
			// 文件导入（运行时忽略或记录）
		case ast.ExportExpression:
			// 导出标记（运行时忽略）
		case ast.MacroExpression:
			// 宏定义（运行时忽略或展开）
		case ast.ConditionalExpression:
			// 条件编译：运行时求值决定是否激活子策略
		case ast.AssertExpression:
			// 断言校验：启动时检查条件，失败则 panic 或报错

		// 范围与优先级（通常作为子节点出现，运行时由调度器处理）
		case ast.CrossExpression, ast.PriorityExpression, ast.OverrideExpression,
			ast.RangeExpression, ast.LiteralExpression, ast.ArithExpression,
			ast.ComparisonExpression, ast.LogicalExpression, ast.MacroExpandExpression:
			// 这些表达式通常嵌套在其他节点中，由调度器或上级节点处理

		// 时间控制
		case ast.HoldMaxExpression:
			if len(exp.Params) > 0 {
				ctx.SetHoldMax(duration(exp.Params))
			}
		case ast.CooldownExpression:
			if len(exp.Params) > 0 {
				ctx.SetCooldown(duration(exp.Params))
			}
			if exp.Name == "per_level" {
				ctx.SetCooldownPerLevel(true)
			}
		case ast.SessionExpression:
			ctx.SetSession(exp.Params)
		case ast.ActiveExpression:
			ctx.SetActiveRange(exp.Params)
		case ast.PauseExpression:
			ctx.SetPauseRange(exp.Params)

		// 头寸管理
		case ast.PositionExpression:
			if len(exp.Params) > 0 {
				ctx.SetPositionType(exp.Params[0].Value)
			}
			if len(exp.Params) > 1 {
				ctx.SetMaxPosition(parseFloatLiteral(exp.Params[1]))
			}
		case ast.SizingExpression:
			if len(exp.Params) > 0 {
				ctx.SetSizingMethod(exp.Params[0].Value)
			}
		case ast.BasePositionExpression:
			if len(exp.Params) > 0 {
				ctx.SetBasePosition(parseFloatLiteral(exp.Params[0]))
			}
		case ast.PyramidStepExpression:
			if len(exp.Params) > 0 {
				ctx.SetPyramidStep(parseFloatLiteral(exp.Params[0]))
			}
		case ast.MaxPyramidLayersExpression:
			if len(exp.Params) > 0 {
				ctx.SetMaxPyramidLayers(int(parseFloatLiteral(exp.Params[0])))
			}
		case ast.FixedFractionExpression:
			if len(exp.Params) > 0 {
				ctx.SetFixedFraction(parseFloatLiteral(exp.Params[0]))
			}
		case ast.VolatilityTargetExpression:
			if len(exp.Params) > 0 {
				ctx.SetVolatilityTarget(parseFloatLiteral(exp.Params[0]))
			}
		case ast.AtrPeriodExpression:
			if len(exp.Params) > 0 {
				ctx.SetAtrPeriod(int(parseFloatLiteral(exp.Params[0])))
			}
		case ast.RiskPerTradeExpression:
			if len(exp.Params) > 0 {
				ctx.SetRiskPerTrade(parseFloatLiteral(exp.Params[0]))
			}
		case ast.RiskPerGridExpression:
			if len(exp.Params) > 0 {
				ctx.SetRiskPerGrid(parseFloatLiteral(exp.Params[0]))
			}
		case ast.MaxDrawdownExpression:
			if len(exp.Params) > 0 {
				ctx.SetMaxDrawdown(parseFloatLiteral(exp.Params[0]))
			}
		case ast.PositionDecayExpression:
			if len(exp.Params) > 0 {
				ctx.SetPositionDecay(parseFloatLiteral(exp.Params[0]))
			}
			// PositionDecayPer 需要解析后续 duration 参数，由 parser 放入 Params
			for i, p := range exp.Params {
				if p.Node.Type == ast.DurationLiteral && i > 0 {
					ctx.SetPositionDecayPer(duration([]ast.Literal{p}))
					break
				}
			}
		case ast.GrossExposureExpression:
			if len(exp.Params) > 0 {
				ctx.SetGrossExposure(parseFloatLiteral(exp.Params[0]))
			}
		case ast.NetExposureExpression:
			if len(exp.Params) > 0 {
				ctx.SetNetExposure(parseFloatLiteral(exp.Params[0]))
			}
		case ast.BetaNeutralExpression:
			if len(exp.Params) > 0 {
				ctx.SetBetaNeutral(exp.Params[0].Value == "true")
			}
		case ast.RebalanceExpression:
			if len(exp.Params) > 0 {
				ctx.SetRebalance(exp.Params[0].Value)
			}
			for _, p := range exp.Params {
				if p.Node.Type == ast.TimePointLiteral {
					ctx.SetRebalanceAt(p.Value)
				}
			}

		// 杠杆与保证金
		case ast.LeverageExpression:
			if len(exp.Params) > 0 {
				ctx.SetLeverage(parseFloatLiteral(exp.Params[0]))
			}
		case ast.MarginExpression:
			if len(exp.Params) > 0 {
				ctx.SetMarginMode(exp.Params[0].Value)
			}
		case ast.HedgeExpression:
			if len(exp.Params) > 0 {
				ctx.SetHedge(exp.Params[0].Value)
			}
		case ast.FundingPriorityExpression:
			if len(exp.Params) > 0 {
				ctx.SetFundingPriority(exp.Params[0].Value)
			}
		case ast.MaxShortExpression:
			if len(exp.Params) > 0 {
				ctx.SetMaxShort(parseFloatLiteral(exp.Params[0]))
			}
		case ast.BorrowRateLimitExpression:
			if len(exp.Params) > 0 {
				ctx.SetBorrowRateLimit(parseFloatLiteral(exp.Params[0]))
			}

		// 风控
		case ast.MaxPositionExpression:
			if len(exp.Params) > 0 {
				ctx.SetMaxPosition(parseFloatLiteral(exp.Params[0]))
			}
		case ast.StopLossExpression:
			ctx.SetStopLossRange(exp.Range)
		case ast.SlippageToleranceExpression:
			if len(exp.Params) > 0 {
				ctx.SetSlippageTolerance(parseFloatLiteral(exp.Params[0]))
			}
		case ast.CircuitBreakerExpression:
			if len(exp.Params) > 0 {
				ctx.SetCircuitBreaker(parseFloatLiteral(exp.Params[0]))
			}
			for i, p := range exp.Params {
				if p.Node.Type == ast.DurationLiteral && i > 0 {
					ctx.SetCircuitBreakerIn(duration([]ast.Literal{p}))
					break
				}
			}
			ctx.Emiter.On("circuit_breaker", func(args ...any) {
				// 熔断触发逻辑，可由调度器或外部系统调用
			})

		// 执行偏好
		case ast.CompoundProfitExpression:
			if len(exp.Params) > 0 {
				ctx.SetCompoundProfit(exp.Params[0].Value == "true")
			}
		case ast.SkipIfGappedExpression:
			ctx.SetSkipIfGapped(true)
		case ast.FallbackExpression:
			if len(exp.Params) > 0 {
				ctx.SetFallback(exp.Params[0].Value)
			}

		// Loss / Profit 分支
		case ast.LossExpression:
			ctx.Emiter.On("stop_triggered", func(args ...any) {
				processExpressions(ctx, exp.Body)
			})
		case ast.ProfitExpression:
			ctx.Emiter.On("stop_triggered", func(args ...any) {
				processExpressions(ctx, exp.Body)
			})

		// 组合配置
		case ast.PortfolioHeatExpression:
			if len(exp.Params) > 0 {
				ctx.SetPortfolioHeat(parseFloatLiteral(exp.Params[0]))
			}
		case ast.CorrelationLimitExpression:
			if len(exp.Params) > 0 {
				ctx.SetCorrelationLimit(parseFloatLiteral(exp.Params[0]))
			}

		// 文档注释 (通用适配器与扩展配置)
		case ast.AnnotationExpression:
			switch exp.Name {
			case "adapter":
				ctx.SetAdapter(exp.Value.Value)
			case "adapter_mode":
				ctx.SetAdapterMode(exp.Value.Value)
			case "adapter_config":
				ctx.SetAdapterConfig(exp.Value.Value)
			case "adapter_switch":
				ctx.SetAdapterSwitch(exp.Value.Value)
			}

		default:
			continue
		}
	}
}

// Transformer 将 AST 转换为可执行的 StockContext
func Transformer(root ast.RootNode) *StockContext {
	stocksContext := NewStockContextFromAST()
	processExpressions(stocksContext, root.Expression)
	return stocksContext
}
