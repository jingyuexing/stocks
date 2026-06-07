package scheduler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jingyuexing/stocks/ast"
	"github.com/jingyuexing/stocks/transformer"
)

func TestNewScheduler(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	if s == nil {
		t.Fatal("NewScheduler returned nil")
	}
	if s.ctx != ctx {
		t.Error("scheduler context mismatch")
	}
	if s.cron == nil {
		t.Error("scheduler cron is nil")
	}
}

func TestScheduler_AddJob_ListJobs_RemoveJob(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	job := Job{
		ID:   "test_job",
		Spec: "*/1 * * * * *", // valid 6-field cron expression with seconds
		Type: JobTypePriceCheck,
		Callback: func(ctx *transformer.StockContext, expr any) {
			// no-op
		},
	}

	id, err := s.AddJob(job)
	if err != nil {
		t.Fatalf("AddJob failed: %v", err)
	}
	if id != "test_job" {
		t.Errorf("expected job id %q, got %q", "test_job", id)
	}

	// brief sleep to let cron internal state settle
	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == "test_job" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job %q in ListJobs, got %v", "test_job", jobs)
	}

	// remove existing job
	if err := s.RemoveJob("test_job"); err != nil {
		t.Fatalf("RemoveJob failed: %v", err)
	}

	jobs = s.ListJobs()
	for _, j := range jobs {
		if j == "test_job" {
			t.Error("expected job to be removed")
		}
	}

	// removing non-existent job should error
	if err := s.RemoveJob("not_exist"); err == nil {
		t.Error("expected error when removing non-existent job")
	}
}

func TestScheduler_AddJob_AutoID(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	job := Job{
		Spec: "*/1 * * * * *",
		Type: JobTypePriceCheck,
		Callback: func(ctx *transformer.StockContext, expr any) {
			// no-op
		},
	}

	id, err := s.AddJob(job)
	if err != nil {
		t.Fatalf("AddJob failed: %v", err)
	}
	if id == "" {
		t.Error("expected auto-generated id, got empty string")
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected auto-generated job %q in ListJobs, got %v", id, jobs)
	}
}

func TestScheduler_AddJob_InvalidSpec(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)

	job := Job{
		ID:   "bad_job",
		Spec: "invalid cron spec",
		Type: JobTypePriceCheck,
	}

	_, err := s.AddJob(job)
	if err == nil {
		t.Error("expected error for invalid cron spec")
	}
}

func TestScheduler_NewPriceCheckJob(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	id, err := s.NewPriceCheckJob("*/1 * * * * *", func(ctx *transformer.StockContext) {
		// no-op
	})
	if err != nil {
		t.Fatalf("NewPriceCheckJob failed: %v", err)
	}
	if id != "price_check" {
		t.Errorf("expected job id %q, got %q", "price_check", id)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == "price_check" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job %q in ListJobs, got %v", "price_check", jobs)
	}
}

func TestScheduler_ScheduleFromAST(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.ConfigStmtNode{Key: "keep"},
			&ast.ConfigStmtNode{Key: "stop"},
			&ast.StrategyStmtNode{Kind: "grid"},
		},
	}

	err := s.ScheduleFromAST(root, "0 */1 * * * *")
	if err != nil {
		t.Fatalf("ScheduleFromAST failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	expected := map[string]bool{
		"keep_check": false,
		"stop_check": false,
		"grid_check": false,
	}

	for _, j := range jobs {
		if _, ok := expected[j]; ok {
			expected[j] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected job %q to be scheduled", name)
		}
	}
}

func TestScheduler_ScheduleFromAST_DefaultInterval(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.ConfigStmtNode{Key: "stop"},
		},
	}

	// empty priceCheckInterval should fall back to default
	err := s.ScheduleFromAST(root, "")
	if err != nil {
		t.Fatalf("ScheduleFromAST failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == "stop_check" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job %q to be scheduled with default interval", "stop_check")
	}
}

func TestScheduler_ScheduleFromAST_Empty(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{},
	}

	err := s.ScheduleFromAST(root, "*/1 * * * * *")
	if err != nil {
		t.Fatalf("ScheduleFromAST with empty AST should not error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("expected no jobs for empty AST, got %v", jobs)
	}
}

func TestScheduler_ScheduleFromAST_OnlyBuy(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.ActionStmtNode{Action: "buy"},
		},
	}

	err := s.ScheduleFromAST(root, "*/1 * * * * *")
	if err != nil {
		t.Fatalf("ScheduleFromAST with only buy action should not error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("expected no jobs for buy action AST, got %v", jobs)
	}
}

func TestScheduler_ScheduleFromAST_MultipleKeep(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.ConfigStmtNode{Key: "keep"},
			&ast.ConfigStmtNode{Key: "keep"},
		},
	}

	err := s.ScheduleFromAST(root, "*/1 * * * * *")
	if err != nil {
		t.Fatalf("ScheduleFromAST with multiple keep config should not error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == "keep_check" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job %q to be scheduled, got %v", "keep_check", jobs)
	}
}

func TestScheduler_ScheduleFromAST_MultipleStop(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	root := &ast.ProgramNode{
		Statements: []ast.Stmt{
			&ast.ConfigStmtNode{Key: "stop"},
			&ast.ConfigStmtNode{Key: "stop"},
		},
	}

	err := s.ScheduleFromAST(root, "*/1 * * * * *")
	if err != nil {
		t.Fatalf("ScheduleFromAST with multiple stop config should not error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	jobs := s.ListJobs()
	found := false
	for _, j := range jobs {
		if j == "stop_check" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job %q to be scheduled, got %v", "stop_check", jobs)
	}
}

func TestScheduler_NewPriceCheckJob_Callback(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	var called bool
	id, err := s.NewPriceCheckJob("*/1 * * * * *", func(ctx *transformer.StockContext) {
		called = true
	})
	if err != nil {
		t.Fatalf("NewPriceCheckJob failed: %v", err)
	}
	if id != "price_check" {
		t.Errorf("expected job id %q, got %q", "price_check", id)
	}

	time.Sleep(1500 * time.Millisecond)

	if !called {
		t.Error("expected callback to be invoked")
	}
}

func TestScheduler_AddJob_NilContext(t *testing.T) {
	s := NewScheduler(nil)
	s.Start()
	defer s.Stop()

	var called bool
	job := Job{
		ID:   "nil_ctx_job",
		Spec: "*/1 * * * * *",
		Type: JobTypePriceCheck,
		Callback: func(ctx *transformer.StockContext, expr any) {
			called = true
		},
	}

	_, err := s.AddJob(job)
	if err != nil {
		t.Fatalf("AddJob should not fail when scheduler ctx is nil: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	if called {
		t.Error("callback should not be invoked when scheduler ctx is nil")
	}
}

func TestScheduler_ConcurrentAccess(t *testing.T) {
	ctx := transformer.NewStockContext()
	s := NewScheduler(ctx)
	s.Start()
	defer s.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			s.AddJob(Job{
				ID:   fmt.Sprintf("job_%d", i),
				Spec: "*/1 * * * * *",
				Type: JobTypePriceCheck,
			})
		}(i)
		go func(i int) {
			defer wg.Done()
			s.ListJobs()
		}(i)
		go func(i int) {
			defer wg.Done()
			s.RemoveJob(fmt.Sprintf("job_%d", i))
		}(i)
	}
	wg.Wait()
}
