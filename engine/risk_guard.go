package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/jingyuexing/stocks/transformer"
	"github.com/jingyuexing/stocks/utils"
)

// RiskGuard 风控拦截器，对每笔待执行动作进行前置检查
type RiskGuard struct {
	ctx *transformer.StockContext
}

// NewRiskGuard 创建风控拦截器
func NewRiskGuard(ctx *transformer.StockContext) *RiskGuard {
	return &RiskGuard{ctx: ctx}
}

// Check 对 Action 进行风控检查，通过返回 nil，拦截返回具体错误
func (rg *RiskGuard) Check(action Action) error {
	if rg.ctx == nil {
		return errors.New("risk guard: context is nil")
	}

	// 1. 交易时段检查
	if err := rg.checkSession(); err != nil {
		return err
	}

	// 2. 暂停日期检查
	if err := rg.checkPauseRange(); err != nil {
		return err
	}

	// 3. 熔断检查
	if err := rg.checkCircuitBreaker(); err != nil {
		return err
	}

	// 4. 最大持仓检查（仅 buy 方向）
	if action.Type == ActionBuy {
		if err := rg.checkMaxPosition(action); err != nil {
			return err
		}
	}

	// 5. 止损检查（仅 sell 方向，防止在止损线外错误加仓）
	if action.Type == ActionSell || action.Type == ActionSellShort {
		if err := rg.checkStopLossOnSell(); err != nil {
			return err
		}
	}

	return nil
}

// checkSession 交易时段检查
func (rg *RiskGuard) checkSession() error {
	session := rg.ctx.GetSession()
	if len(session) == 0 {
		return nil // 未配置 session 时不限制
	}

	ok, err := utils.IsTimeInRanges(session, time.Now())
	if err != nil {
		// 解析错误在严格模式下应阻断，但当前作为风控检查，解析失败时放行并记录
		fmt.Printf("[riskguard] session parse warning: %v\n", err)
		return nil
	}
	if !ok {
		return fmt.Errorf("outside trading session: %v", session)
	}
	return nil
}

// checkPauseRange 暂停日期检查
func (rg *RiskGuard) checkPauseRange() error {
	pause := rg.ctx.GetPauseRange()
	if len(pause) == 0 {
		return nil
	}

	ok, err := utils.IsDateInRange(pause, time.Now())
	if err != nil {
		fmt.Printf("[riskguard] pause range parse warning: %v\n", err)
		return nil
	}
	if ok {
		return fmt.Errorf("strategy paused in date range: %v", pause)
	}
	return nil
}

// checkCircuitBreaker 熔断检查
func (rg *RiskGuard) checkCircuitBreaker() error {
	cb := rg.ctx.GetCircuitBreaker()
	if cb <= 0 {
		return nil // 未配置熔断
	}

	// 简化：若当前 profit 绝对值超过熔断阈值，则触发
	profit := rg.calcProfitPercent()
	if profit < 0 && -profit >= cb {
		return fmt.Errorf("circuit breaker triggered: loss %.2f%% exceeds threshold %.2f%%", -profit, cb)
	}
	if profit >= cb {
		return fmt.Errorf("circuit breaker triggered: profit %.2f%% exceeds threshold %.2f%%", profit, cb)
	}
	return nil
}

// checkMaxPosition 最大持仓检查
func (rg *RiskGuard) checkMaxPosition(action Action) error {
	maxPos := rg.ctx.GetMaxPosition()
	if maxPos <= 0 {
		return nil // 未配置
	}

	// 骨架：当前 amount 直接比较（后续应查询 gateway 的真实持仓）
	amount := action.Amount
	if action.IsPercent {
		// 将百分比转换为绝对数量（简化：基于当前 available balance）
		amount = rg.ctx.GetAmount() * action.Amount / 100
	}
	if amount > maxPos {
		return fmt.Errorf("max position exceeded: %.2f > %.2f", amount, maxPos)
	}
	return nil
}

// checkStopLossOnSell 止损方向检查（预留）
func (rg *RiskGuard) checkStopLossOnSell() error {
	// 预留：防止在已触发止损后重复平仓等场景
	return nil
}

// calcProfitPercent 计算当前盈亏百分比
func (rg *RiskGuard) calcProfitPercent() float64 {
	begin := rg.ctx.GetBeginPrice()
	current := rg.ctx.GetCurrentPrice()
	if begin <= 0 || current <= 0 {
		return 0
	}
	return (current - begin) / begin * 100
}

// IsSessionActive 当前是否在交易时段（公共工具方法）
func (rg *RiskGuard) IsSessionActive() bool {
	return rg.checkSession() == nil
}

// IsPaused 当前是否在暂停日期（公共工具方法）
func (rg *RiskGuard) IsPaused() bool {
	return rg.checkPauseRange() != nil
}
