package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/transformer"

	"github.com/robfig/cron/v3"
)

// JobType 表示调度任务类型
type JobType string

const (
	JobTypePriceCheck JobType = "price_check" // 定时检查股价
	JobTypeKeepCheck  JobType = "keep_check"  // 定时检查持仓时间
	JobTypeStopCheck  JobType = "stop_check"  // 定时检查止损/止盈
	JobTypeGridCheck  JobType = "grid_check"  // 定时检查网格交易条件
)

// Job 表示一个调度任务
type Job struct {
	ID       string
	Type     JobType
	Spec     string // Cron 表达式
	Expr     ast.ExpressionNode
	Context  *transformer.StockContext
	Callback func(ctx *transformer.StockContext, expr ast.ExpressionNode)
}

// Scheduler 基于 cron 的定时任务调度器
type Scheduler struct {
	cron *cron.Cron
	jobs map[string]cron.EntryID
	mu   sync.RWMutex
	ctx  *transformer.StockContext
}

// NewScheduler 创建新的调度器
func NewScheduler(ctx *transformer.StockContext) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
		jobs: make(map[string]cron.EntryID),
		ctx:  ctx,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// AddJob 添加定时任务
// spec 支持标准 cron 表达式，如:
//
//	"0 */5 * * * *" 每5分钟
//	"0 0 9 * * 1-5" 工作日9点
func (s *Scheduler) AddJob(job Job) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job.ID == "" {
		job.ID = fmt.Sprintf("job_%d", time.Now().UnixNano())
	}

	entryID, err := s.cron.AddFunc(job.Spec, func() {
		if job.Callback != nil && s.ctx != nil {
			job.Callback(s.ctx, job.Expr)
		}
	})
	if err != nil {
		return "", fmt.Errorf("failed to add cron job: %w", err)
	}

	s.jobs[job.ID] = entryID
	return job.ID, nil
}

// RemoveJob 移除定时任务
func (s *Scheduler) RemoveJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %s not found", jobID)
	}

	s.cron.Remove(entryID)
	delete(s.jobs, jobID)
	return nil
}

// ListJobs 列出所有任务
func (s *Scheduler) ListJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.jobs))
	for id := range s.jobs {
		ids = append(ids, id)
	}
	return ids
}

// ScheduleFromAST 根据 AST 表达式自动生成定时任务
func (s *Scheduler) ScheduleFromAST(root ast.RootNode, priceCheckInterval string) error {
	if priceCheckInterval == "" {
		priceCheckInterval = "0 */5 * * * *" // 默认每5分钟
	}

	for _, expr := range root.Expression {
		switch expr.Type {
		case ast.KeepExpression:
			// 为 keep 表达式添加持仓时间检查
			_, err := s.AddJob(Job{
				ID:   "keep_check",
				Type: JobTypeKeepCheck,
				Spec: "0 * * * * *", // 每分钟检查一次持仓时间
				Expr: expr,
				Callback: func(ctx *transformer.StockContext, e ast.ExpressionNode) {
					ok, err := ctx.UseKeep()
					if err != nil {
						fmt.Printf("[scheduler] keep check error: %v\n", err)
						return
					}
					if !ok {
						fmt.Println("[scheduler] keep duration expired, triggering stop")
						ctx.Emiter.Emit("keep_expired", ctx)
					}
				},
			})
			if err != nil {
				return fmt.Errorf("failed to schedule keep check: %w", err)
			}

		case ast.StopExpression:
			// 为止损/止盈添加股价检查
			_, err := s.AddJob(Job{
				ID:   "stop_check",
				Type: JobTypeStopCheck,
				Spec: priceCheckInterval,
				Expr: expr,
				Callback: func(ctx *transformer.StockContext, e ast.ExpressionNode) {
					ok, err := ctx.UseStop()
					if err != nil {
						fmt.Printf("[scheduler] stop check error: %v\n", err)
						return
					}
					if ok {
						fmt.Println("[scheduler] stop condition triggered")
						ctx.Emiter.Emit("stop_triggered", ctx)
					}
				},
			})
			if err != nil {
				return fmt.Errorf("failed to schedule stop check: %w", err)
			}

		case ast.GridExpression:
			// 为网格交易添加检查
			_, err := s.AddJob(Job{
				ID:   "grid_check",
				Type: JobTypeGridCheck,
				Spec: priceCheckInterval,
				Expr: expr,
				Callback: func(ctx *transformer.StockContext, e ast.ExpressionNode) {
					fmt.Println("[scheduler] grid check running")
					ctx.Emiter.Emit("grid_check", ctx, e)
				},
			})
			if err != nil {
				return fmt.Errorf("failed to schedule grid check: %w", err)
			}
		}
	}

	return nil
}

// NewPriceCheckJob 创建股价检查任务（通用）
func (s *Scheduler) NewPriceCheckJob(spec string, callback func(ctx *transformer.StockContext)) (string, error) {
	return s.AddJob(Job{
		ID:   "price_check",
		Type: JobTypePriceCheck,
		Spec: spec,
		Callback: func(ctx *transformer.StockContext, _ ast.ExpressionNode) {
			if callback != nil {
				callback(ctx)
			}
		},
	})
}
