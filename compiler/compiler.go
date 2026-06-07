package compiler

import (
	"fmt"
	"strings"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/parser"
	"github.com/jingyuexing/stocks/tokenizer"
)

// CompileError 编译错误
type CompileError struct {
	Phase   string // tokenizer / parser / transformer / validator
	Message string
}

func (e CompileError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Phase, e.Message)
}

// Compile 将 DSL 源码编译为 StrategyBundle
// 完整的编译链路：Tokenizer -> Parser -> Builder -> BundleAssembler
func Compile(source string) (*StrategyBundle, error) {
	bundle := NewBundle()
	bundle.Source = source

	// 1. Tokenizer
	tokens := tokenizer.Lexer(source)

	// 2. Parser
	root := parser.Parse(tokens)
	bundle.AST = *root

	// 3. Builder：AST -> StockContext
	ctx := BuildContext(root)
	if ctx == nil {
		return nil, CompileError{Phase: "builder", Message: "failed to create StockContext from AST"}
	}
	bundle.Context = ctx

	// 4. 从 AST 提取 AdapterConfig（优先于 Context 中的零散字段）
	bundle.Adapter = extractAdapterConfig(root)

	// 5. 基础校验（骨架，后续扩展）
	if err := Validate(bundle); err != nil {
		return nil, err
	}

	// 6. 元数据
	bundle.Metadata["compiled_at"] = time.Now().Format(time.RFC3339)
	bundle.Metadata["version"] = "2.1"

	return bundle, nil
}

// CompileWithID 带指定 ID 的编译入口
func CompileWithID(id string, source string) (*StrategyBundle, error) {
	bundle, err := Compile(source)
	if err != nil {
		return nil, err
	}
	bundle.ID = id
	return bundle, nil
}

// extractAdapterConfig 从 AST 的 AnnotationStmt 中提取适配器配置
func extractAdapterConfig(root *ast.ProgramNode) AdapterConfig {
	var cfg AdapterConfig
	for _, stmt := range root.Statements {
		ann, ok := stmt.(*ast.AnnotationStmtNode)
		if !ok {
			continue
		}
		switch ann.Key {
		case "adapter":
			parseAdapterSpec(ann.Value, &cfg)
		case "adapter_mode":
			cfg.Mode = strings.TrimSpace(ann.Value)
		case "adapter_config":
			cfg.Params = ann.Value
		case "adapter_switch":
			cfg.Switch = ann.Value
		}
	}
	return cfg
}

// parseAdapterSpec 解析适配器声明值，支持多种形式：
//   - "binance"
//   - "primary=binance, fallback=okx"
//   - "paper=interactive_brokers, live=interactive_brokers"
func parseAdapterSpec(value string, cfg *AdapterConfig) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	// 如果包含等号，视为 key=value 对列表
	if strings.Contains(value, "=") {
		pairs := strings.Split(value, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}
			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			switch k {
			case "primary", "paper", "live":
				cfg.Primary = v
			case "fallback":
				cfg.Fallback = v
			}
		}
	} else {
		// 简单的适配器名称
		cfg.Primary = value
	}
}
