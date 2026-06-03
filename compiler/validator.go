package compiler

import (
	"strings"

	"github.com/jingyuexing/stocks/ast"
)

// Validator 编译期校验器接口
type Validator interface {
	Validate(bundle *StrategyBundle) error
}

// Validate 对 StrategyBundle 执行基础校验
// 当前为骨架实现，后续可扩展：类型检查、循环依赖检测、断言求值等
func Validate(bundle *StrategyBundle) error {
	if bundle == nil {
		return CompileError{Phase: "validator", Message: "bundle is nil"}
	}
	if bundle.Context == nil {
		return CompileError{Phase: "validator", Message: "context is nil"}
	}

	// 1. 校验断言（assert）表达式
	for _, expr := range bundle.AST.Expression {
		if expr.Type == ast.AssertExpression {
			if err := validateAssert(bundle, expr); err != nil {
				return err
			}
		}
	}

	// 2. 校验策略声明完整性
	foundStrategy := false
	for _, expr := range bundle.AST.Expression {
		switch expr.Type {
		case ast.GridExpression, ast.LongExpression, ast.ShortExpression,
			ast.BothExpression, ast.PortfolioExpression:
			foundStrategy = true
		}
	}
	// 如果没有显式策略声明但包含交易动作，视为合法（简化脚本）
	_ = foundStrategy

	// 3. 校验 Adapter 配置有效性（仅当存在配置时）
	if !bundle.Adapter.IsEmpty() {
		if bundle.Adapter.Primary != "" {
			// 内置适配器白名单（可扩展）
			knownAdapters := []string{
				"binance", "okx", "bybit", "interactive_brokers",
				"alpaca", "xtquant", "tdx", "backtest", "paper",
			}
			found := false
			for _, name := range knownAdapters {
				if strings.EqualFold(bundle.Adapter.Primary, name) {
					found = true
					break
				}
			}
			if !found {
				// 非内置适配器仅警告，不阻断（支持自定义扩展）
				// 可通过严格模式开关控制
			}
		}
	}

	return nil
}

// validateAssert 对 assert 表达式进行编译期求值（常量表达式）
// 目前为骨架：若 assert 条件包含运行时变量，则跳过（留到运行期检查）
func validateAssert(bundle *StrategyBundle, expr ast.ExpressionNode) error {
	// TODO: 实现常量表达式求值
	// 如果 assert 条件中不含运行时变量，可在编译期直接判断
	_ = bundle
	_ = expr
	return nil
}
