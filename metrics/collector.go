package metrics

import (
	"context"
	"database/sql"
	dbase "oneclick-metrics-go/db"
	"sync"
	"time"
)

func CollectMetrics(ctx context.Context, m *Metrics, db *sql.DB, wg *sync.WaitGroup) {
	ticker := time.NewTicker(3 * time.Second)

	// PRMissingReport
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
			ExportPRMissingReport(ctx, m, db)
		}

	}()

	// PRNum
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
			ExportPRNum(ctx, m, db)
		}
	}()

	// OpenPRReport
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
			ExportOpenPRReport(ctx, m, db)
		}
	}()

	// ClosedPRReport
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
			ExportClosedPRReport(ctx, m, db)
		}
	}()

	// ResultReport
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
			ExportResultReport(ctx, m, db)
		}
	}()

	wg.Wait()
	//log.Println("所有采集指标完成")
}
