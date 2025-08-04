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
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
			ExportPRMissingReport(ctx, m, db)
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
			ExportPRNum(ctx, m, db)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
			ExportOpenPRReport(ctx, m, db)
		}
	}()

	wg.Wait()
	//log.Println("所有采集指标完成")
}
