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


## 注释

```dsl
// 这是行注释
/* 这是
   块注释 */
```

## 策略声明

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `grid` | `grid <symbol> { ... }` | 单方向网格策略（默认做多） |
| `long` | `long <symbol> { ... }` | 显式做多策略 |
| `short` | `short <symbol> { ... }` | 做空策略 |
| `both` | `both <symbol> { long { ... } short { ... } }` | 双向同时运行 |
| `portfolio` | `portfolio <name> { ... }` | 多标的组合策略 |

## 模板与复用

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `template` | `template <name>(<params>) [extends <name>] { ... }` | 策略模板定义 |
| `use` | `use <name>(<args>) [as <alias>];` | 模板实例化 |
| `import` | `import "<path>";` | 文件导入 |
| `export` | `export <name>;` | 文件导出 |
| `macro` | `macro <name> { ... }` | 代码宏定义 |
| `if` | `if <expr> { ... }` | 条件编译 |
| `assert` | `assert <condition>;` | 断言校验 |

## 时间控制

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `keep` | `keep <time>;` | 最小持仓时间 |
| `hold_max` | `hold_max <time>;` | 最大持仓时间 |
| `cooldown` | `cooldown <time> [per level];` | 层级冷却 |
| `session` | `session <time_point>…<time_point> [<TZ>];` | 交易时段 |
| `active` | `active <date>…<date>;` | 策略有效日期 |
| `pause` | `pause <date>…<date>;` | 策略暂停日期 |

时间表达式：`<number> <unit>`，支持组合如 `2h 30min`。
时间单位：`ms`, `s`, `min`, `h`, `H`, `d`, `W`, `M`, `Y`。

## 交易动作

| 关键字 | 方向 | 语法 | 说明 |
|--------|------|------|------|
| `buy` | 做多 | `buy <amount> [shares/capital] @ [market/limit <price%>];` | 开多/加仓 |
| `sell` | 做多 | `sell <amount> [shares] [from oldest/newest];` | 平多/减仓 |
| `sell_short` | 做空 | `sell_short <amount> [shares/capital] @ [market/limit <price%>];` | 开空/加仓 |
| `buy_cover` | 做空 | `buy_cover <amount> [shares] [from newest/oldest];` | 平空/减仓 |

## 头寸管理

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `position` | `position [fixed/dynamic] <amount> [shares/USD/capital];` | 头寸类型 |
| `sizing` | `sizing <method>;` | 头寸规模算法 |
| `base_position` | `base_position <n>% capital;` | 金字塔基础头寸 |
| `pyramid_step` | `pyramid_step <n>x;` | 金字塔递增倍数 |
| `max_pyramid_layers` | `max_pyramid_layers <n>;` | 金字塔最大层数 |
| `fixed_fraction` | `fixed_fraction <n>%;` | 固定分数 |
| `volatility_target` | `volatility_target <n>%;` | 目标波动率 |
| `atr_period` | `atr_period <n>;` | ATR计算周期 |
| `risk_per_trade` | `risk_per_trade <n>% capital;` | 单笔交易风险 |
| `risk_per_grid` | `risk_per_grid <n>% capital;` | 单层网格风险 |
| `max_drawdown` | `max_drawdown <n>%;` | 最大回撤限制 |
| `position_decay` | `position_decay <n>% per <time>;` | 头寸衰减 |
| `gross_exposure` | `gross_exposure <n>% capital;` | 总敞口上限 |
| `net_exposure` | `net_exposure <n>% capital;` | 净敞口上限 |
| `beta_neutral` | `beta_neutral [true/false];` | Beta中性 |
| `rebalance` | `rebalance [daily/weekly/monthly] [at <time_point>];` | 动态再平衡 |

`sizing` 支持的算法：`equal`, `pyramid`, `pyramid_inverse`, `martingale`, `anti_martingale`, `kelly_half`, `fixed_fractional`, `fixed_ratio`, `volatility_target`。

## 杠杆与保证金

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `leverage` | `leverage <n>x;` | 杠杆倍数 |
| `margin` | `margin [isolated/cross];` | 保证金模式 |
| `hedge` | `hedge <ratio>;` | 多空对冲比例 |
| `funding_priority` | `funding_priority [long/short];` | 资金优先方向 |
| `max_short` | `max_short <n> shares;` | 融券额度上限 |
| `borrow_rate_limit` | `borrow_rate_limit <n>%;` | 融券费率上限 |

## 风控

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `max_position` | `max_position <amount> [shares/capital];` | 最大持仓 |
| `stop_loss` | `stop_loss <profit_range> [on total/total_position];` | 止损线 |
| `slippage_tolerance` | `slippage_tolerance <price%>;` | 滑点容忍 |
| `partial_fill` | `partial_fill [accept/reject];` | 部分成交处理 |
| `circuit_breaker` | `circuit_breaker <profit%> in <time>;` | 熔断机制 |

## 执行偏好

| 关键字 | 语法 | 说明 |
|--------|------|------|
| `compound_profit` | `compound_profit [true/false];` | 利润再投入 |
| `skip_if_gapped` | `skip_if_gapped;` | 跳空跳过 |
| `fallback` | `fallback [skip/reduce];` | 资金不足回退策略 |

## 范围表达式

```dsl
-10%…-5%      // 跌幅10%到5%之间
-20%…         // 跌幅≥20%（开放上限）
…-30%         // 跌幅≤30%（开放下限）
5%…10%        // 涨幅5%到10%之间
```

范围可使用 `cross` 修饰，表示仅穿越边界时触发：
```dsl
5%…10% cross priority 1 { sell 50% }
```

也可使用 `override` 定义时段/日期覆盖规则：
```dsl
override 09:30…10:00 -5%…-2% { buy 20% }
```

## 变量定义与引用

### 定义（define）

```dsl
amount: hello
risk: 10%
```

标识符后跟冒号，再跟值。值可以是标识符、数字或字符串。

### 引用（reference）

```dsl
@ork
```

`@` 符号用于引用已定义的变量。

### 内置变量（variable）

在 grid / long / short / both 策略作用域中，支持以 `$` 开头的内置运行时变量：

```dsl
grid AAPL {
    // 在交易动作中使用变量
    buy $amount
    sell $profit

    // 在条件表达式中使用变量
    if $price > 150 {
        sell 50%
    }

    // 变量也可作为独立表达式
    $high
    $low
}
```

**常用内置变量**：

| 变量 | 说明 |
|------|------|
| `$profit` | 当前浮动盈亏（百分比或金额） |
| `$price` | 当前市价 |
| `$high` | 周期最高价 |
| `$low` | 周期最低价 |
| `$open` | 周期开盘价 |
| `$close` | 周期收盘价 |
| `$volume` | 成交量 |
| `$amount` | 可用资金/头寸金额 |
| `$position` | 当前持仓数量 |
| `$cost` | 持仓成本价 |
| `$grid_level` | 当前所在网格层级 |
| `$time` | 当前时间戳 |

> 内置变量由运行时（Transformer / Scheduler）注入具体值，Parser 阶段仅将其解析为 `VariableExpression` 节点。

## 网格交易（grid / loss / profit）

```dsl
grid AAPL {
    session 09:30…16:00 EST
    position dynamic 1000 capital
    sizing equal

    loss 10% {
        sell 100%
    }
    profit 10% {
        sell 50%
    }
}
```

`grid` 内部支持：
- `loss` / `profit` 分支（可带范围）
- 全局配置（`keep`, `session`, `position`, `sizing`, `leverage` 等）
- 风控配置（`max_position`, `stop_loss`, `circuit_breaker` 等）
- 范围级别块（如 `-10%…-5% { buy 20% }`）
- `override` 覆盖规则
