# Stocks DSL 项目概述

## 简介

本项目是一个用 **Go** 语言编写的**股票交易策略 DSL（领域特定语言）**编译器与执行框架。用户可以通过简洁的声明式语法描述股票交易策略（如买入、卖出、持仓、止损、网格交易等），系统将其编译为 AST（抽象语法树）并转换为可执行的上下文，最终通过定时任务调度器自动监控与执行。

## 核心目标

- **声明式策略编写**：通过类自然语言的 DSL 描述交易逻辑，无需编写复杂代码。
- **自动解析与执行**：从文本到 AST 再到执行上下文的全链路支持。
- **定时任务驱动**：内置 Cron 调度器，支持定时检查股价、持仓时间、止损止盈等条件。
- **可扩展架构**：模块化设计（tokenizer → parser → transformer → scheduler），便于后续扩展新语法与功能。

## 项目结构

```
stocks/
├── ast/               # 抽象语法树（AST）节点定义
├── tokenizer/         # 词法分析器（Lexer）
├── parser/            # 语法解析器（Parser）
├── transformer/       # AST 转换器与执行上下文（StockContext）
├── scheduler/         # Cron 定时任务调度器
├── types/             # 公共接口定义（如 Adapter）
├── utils/             # 工具函数（金额转换、时间计算等）
├── docs/              # 项目文档（本目录）
├── stocks.g4          # ANTLR 语法文件（预留）
├── readme.md          # 项目简介与 DSL 示例
├── go.mod / go.sum    # Go 模块依赖
└── index.ts           # TypeScript 入口（预留）
```

## 技术栈

| 层级 | 技术/包 |
|------|---------|
| 语言 | Go 1.22+ |
| 事件驱动 | `github.com/jingyuexing/go-utils` |
| 定时任务 | `github.com/robfig/cron/v3` |
| 测试 | Go 标准库 testing |

## 快速示例

```dsl
select AUL          // 选择股票代码
value ork 1000      // 配置资金参数
buy @ork            // 买入（引用变量）
keep 10W            // 持仓 10 周
sell AUL +30%       // 盈利 30% 时卖出

grid {              // 网格交易策略
    loss 10% {
        sell 100%
    }
    profit 10% {
        sell 50%
    }
}
```
