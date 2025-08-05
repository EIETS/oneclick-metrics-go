package metrics

import (
	"context"
	"database/sql"
	dbase "oneclick-metrics-go/db"
	"sync"
	"time"
)

func CollectMetrics(ctx0 context.Context, m *Metrics, db *sql.DB, wg *sync.WaitGroup) {
	ticker := time.NewTicker(5 * time.Second)

	ctxs := make([]context.Context, 6)
	//cancels := make([]context.CancelFunc, 6)

	// 创建 6 个子 context，每个带 5 秒超时
	for i := 0; i < 6; i++ {
		ctxs[i] = context.Background()
	}

	// PRMissingReport
	wg.Add(1)
	go func() {
		ctx := ctxs[0]
		defer wg.Done()
		//defer cancels[0]()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
			ExportPRMissingReport(ctx, m, db)
		}

	}()

	// PRNum
	wg.Add(1)
	go func() {
		ctx := ctxs[1]
		defer wg.Done()
		//defer cancels[1]()
		_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
			ExportPRNum(ctx, m, db)
		}
	}()

	// OpenPRReport
	wg.Add(1)
	go func() {
		ctx := ctxs[2]

		defer wg.Done()
		//defer cancels[2]()
		_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
			ExportOpenPRReport(ctx, m, db)
		}
	}()

	// ClosedPRReport
	wg.Add(1)
	go func() {
		ctx := ctxs[3]
		defer wg.Done()
		//defer cancels[3]()
		_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
			ExportClosedPRReport(ctx, m, db)
		}
	}()

	// ResultReport
	wg.Add(1)
	go func() {
		ctx := ctxs[4]
		defer wg.Done()
		//defer cancels[4]()
		_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
		for range ticker.C {
			_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
			ExportResultReport(ctx, m, db)
		}
	}()

	wg.Wait()
	//log.Println("所有采集指标完成")
}
