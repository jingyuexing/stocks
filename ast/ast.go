package ast

import (
	"fmt"
	"strconv"
)

// NodeType 表示 AST 节点的通用类型，
// 仅保留语法结构类型，不再为每个业务关键字单独定义类型。
type NodeType uint

type LiteralType uint

const (
	Float LiteralType = iota
	Integer
	Text
	Date
	Time
	Enum
	List
)

const (
	Program NodeType = iota + 1

	// 语句类型（Stmt）
	StrategyStmt // grid, long, short, both, portfolio, arbitrage
	BlockStmt    // { ... }
	FlowStmt     // level, trigger, until, atomic
	ActionStmt   // buy, sell, sell_short, buy_cover
	ConfigStmt   // 所有 key value... ; 形式的配置
	ImportStmt
	ExportStmt
	UseStmt
	TemplateStmt
	MacroStmt
	IfStmt
	AssertStmt
	AnnotationStmt
	DefineStmt // identifier: value
	SelectStmt

	// 表达式类型（Expr）
	LiteralExpr
	BinaryExpr
	UnaryExpr
	RangeExpr
	VariableExpr
	ReferenceExpr
	MacroExpandExpr
	CallExpr
	IdentifierExpr
	ListExpr

	Unknown
)

func (t NodeType) String() string {
	switch t {
	case Program:
		return "Program"
	case StrategyStmt:
		return "StrategyStmt"
	case BlockStmt:
		return "BlockStmt"
	case FlowStmt:
		return "FlowStmt"
	case ActionStmt:
		return "ActionStmt"
	case ConfigStmt:
		return "ConfigStmt"
	case ImportStmt:
		return "ImportStmt"
	case ExportStmt:
		return "ExportStmt"
	case UseStmt:
		return "UseStmt"
	case TemplateStmt:
		return "TemplateStmt"
	case MacroStmt:
		return "MacroStmt"
	case IfStmt:
		return "IfStmt"
	case AssertStmt:
		return "AssertStmt"
	case AnnotationStmt:
		return "AnnotationStmt"
	case DefineStmt:
		return "DefineStmt"
	case SelectStmt:
		return "SelectStmt"
	case LiteralExpr:
		return "LiteralExpr"
	case BinaryExpr:
		return "BinaryExpr"
	case UnaryExpr:
		return "UnaryExpr"
	case RangeExpr:
		return "RangeExpr"
	case VariableExpr:
		return "VariableExpr"
	case ReferenceExpr:
		return "ReferenceExpr"
	case MacroExpandExpr:
		return "MacroExpandExpr"
	case CallExpr:
		return "CallExpr"
	case IdentifierExpr:
		return "IdentifierExpr"
	case ListExpr:
		return "ListExpr"
	default:
		return "Unknown"
	}
}

// Position 源码位置信息
type Position struct {
	Line   int
	Column int
}

// Node 是所有 AST 节点的基类
type Node struct {
	Type NodeType
	Pos  Position
}

// Expr 表达式接口
type Expr interface {
	exprNode()
}

// Stmt 语句接口
type Stmt interface {
	stmtNode()
}

// BaseExpr 表达式基类
type BaseExpr struct {
	Node
}

func (BaseExpr) exprNode() {}

// BaseStmt 语句基类
type BaseStmt struct {
	Node
}

func (BaseStmt) stmtNode() {}

// ---------- 表达式节点 ----------

// Literal 字面量：数字、字符串、标识符、时间单位、百分比等
type Literal struct {
	BaseExpr
	LiteralType
	Value string // 原始值
	Unit  string // 单位：%, ns, us, ms, s, min, h, d, w 等
}

func (l Literal) String() string {
	if l.Unit == "%" {
		v, _ := strconv.ParseFloat(l.Value, 64)
		return fmt.Sprintf("%f", v/100)
	}
	return l.Value
}

func (l Literal) AsFloat() float64 {
	v, _ := strconv.ParseFloat(l.Value, 64)
	return v
}

func (l Literal) AsInt() int {
	v, _ := strconv.Atoi(l.Value)
	return v
}

func (l Literal) AsBool() bool {
	return l.Value == "true"
}

// RangeExprNode 范围表达式：begin…end / begin...end / …end / begin…
type RangeExprNode struct {
	BaseExpr
	Begin *Literal
	End   *Literal
}

// BinaryExprNode 二元表达式：left op right
type BinaryExprNode struct {
	BaseExpr
	Op    string // +, -, *, /, ==, !=, <, >, <=, >=, &&, ||
	Left  Expr
	Right Expr
}

// UnaryExprNode 一元表达式：op expr
type UnaryExprNode struct {
	BaseExpr
	Op   string // -, !
	Expr Expr
}

// VariableExprNode 变量表达式：$name
type VariableExprNode struct {
	BaseExpr
	Name string
}

// ReferenceExprNode 引用表达式：@name
type ReferenceExprNode struct {
	BaseExpr
	Name string
}

// MacroExpandExprNode 宏展开表达式：${name}
type MacroExpandExprNode struct {
	BaseExpr
	Name string
}

// CallExprNode 调用表达式：callee(arg1, arg2, ...)
type CallExprNode struct {
	BaseExpr
	Callee string
	Args   []Expr
}

// IdentifierExprNode 裸标识符表达式
type IdentifierExprNode struct {
	BaseExpr
	Name string
}

// ListLiteralNode 列表字面量：用于聚合连续的 duration 字面量等
type ListLiteralNode struct {
	BaseExpr
	Items []Expr
}

// ---------- 语句节点 ----------

// StrategyStmtNode 策略语句：grid/long/short/both/portfolio/arbitrage TARGET { BODY }
type StrategyStmtNode struct {
	BaseStmt
	Kind   string // "grid", "long", "short", "both", "portfolio", "arbitrage"
	Target string // 标的代码或标识符
	Body   []Stmt
}

// BlockStmtNode 块语句：{ STATEMENTS... }
type BlockStmtNode struct {
	BaseStmt
	Statements []Stmt
}

// FlowStmtNode 流程控制语句：
//
//	level CONDITION { BODY }
//	[once|twice|daily|weekly|monthly|yearly|hourly] trigger CONDITION { BODY }
//	until CONDITION { BODY }
//	atomic { BODY } [rollback|best_effort];
//	override LABEL+ RANGE { BODY }
type FlowStmtNode struct {
	BaseStmt
	Kind      string // "level", "trigger", "until", "atomic", "override"
	Condition Expr   // level 为 RangeExpr/LiteralExpr，trigger/until 为 BinaryExpr/VariableExpr
	Body      []Stmt
	Mode      string   // rollback, best_effort（仅 atomic）；profit, loss, price（仅 level）
	Labels    []string // 仅 override：标识符列表
	Frequency string   // 仅 trigger：once, twice, daily, weekly, monthly, yearly, hourly
}

// ActionStmtNode 动作语句：ACTION ARGS... [@ market|limit EXPR] [from oldest|newest|remaining] ;
type ActionStmtNode struct {
	BaseStmt
	Action    string // "buy", "sell", "sell_short", "buy_cover"
	Args      []Expr
	Price     Expr   // 可选
	PriceType string // "market", "limit", ""
	Modifier  string // "from_oldest", "from_newest", "remaining", ""
}

// ConfigStmtNode 通用配置语句：KEY VALUE... ;
type ConfigStmtNode struct {
	BaseStmt
	Key    string
	Params []Expr
	Range  *RangeExprNode
}

// ImportStmtNode 导入语句：import "PATH" ;
type ImportStmtNode struct {
	BaseStmt
	Path string
}

// ExportStmtNode 导出语句：export NAME ;
type ExportStmtNode struct {
	BaseStmt
	Name string
}

// UseStmtNode 使用语句：use NAME(ARGS...) as ALIAS ;
type UseStmtNode struct {
	BaseStmt
	Name  string
	Args  []Expr
	Alias string
}

// ParamDef 模板参数定义
type ParamDef struct {
	Name    string
	Type    string
	Default Expr
}

// TemplateStmtNode 模板定义：template NAME(PARAMS...) extends PARENT { BODY }
type TemplateStmtNode struct {
	BaseStmt
	Name    string
	Params  []ParamDef
	Extends string
	Body    []Stmt
}

// MacroStmtNode 宏定义：macro NAME { BODY }
type MacroStmtNode struct {
	BaseStmt
	Name string
	Body []Stmt
}

// IfStmtNode 条件语句：if CONDITION { BODY }
type IfStmtNode struct {
	BaseStmt
	Condition Expr
	Body      []Stmt
}

// AssertStmtNode 断言语句：assert CONDITION ;
type AssertStmtNode struct {
	BaseStmt
	Condition Expr
}

// AnnotationStmtNode 注解语句：/* @KEY: VALUE */
type AnnotationStmtNode struct {
	BaseStmt
	Key   string
	Value string
}

// DefineStmtNode 定义语句：NAME: VALUE
type DefineStmtNode struct {
	BaseStmt
	Name  string
	Value Expr
}

// SelectStmtNode 选择语句：select TARGET
type SelectStmtNode struct {
	BaseStmt
	Target string
}

// ProgramNode 程序根节点
type ProgramNode struct {
	Node
	Statements []Stmt
}

// RootNode 兼容旧名称的别名
type RootNode = ProgramNode
