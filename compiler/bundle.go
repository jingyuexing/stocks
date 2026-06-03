package compiler

import (
	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/transformer"
)

// AdapterConfig 从文档注释中提取的适配器配置
type AdapterConfig struct {
	// Primary 主适配器名称，如 binance / interactive_brokers / backtest
	Primary string
	// Fallback 备用适配器名称
	Fallback string
	// Mode 交易模式：spot / futures / margin / stock / option
	Mode string
	// Params 适配器初始化参数（JSON 或 key-value 文本）
	Params string
	// Switch 主备切换配置
	Switch string
}

// IsEmpty 判断适配器配置是否为空
func (ac AdapterConfig) IsEmpty() bool {
	return ac.Primary == "" && ac.Mode == ""
}

// StrategyBundle 编译期产物，包含 AST 快照、运行时上下文与 Adapter 配置
type StrategyBundle struct {
	// ID 策略唯一标识（可由调用方指定或自动生成）
	ID string

	// Source 原始 DSL 源码文本，用于日志、回放与调试
	Source string

	// AST 原始抽象语法树快照，Executor 运行时可回溯原始规则
	AST ast.RootNode

	// Context 运行时执行上下文，由 Transformer 生成
	Context *transformer.StockContext

	// Adapter 适配器配置（从 Annotation 注释提取）
	Adapter AdapterConfig

	// Metadata 编译期附加元数据，如编译时间、版本、校验和等
	Metadata map[string]string
}

// NewBundle 创建空的 StrategyBundle
func NewBundle() *StrategyBundle {
	return &StrategyBundle{
		Metadata: make(map[string]string),
	}
}
