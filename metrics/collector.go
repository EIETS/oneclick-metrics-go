package metrics

import (
	"context"
	"database/sql"
	"fmt"
	dbase "oneclick-metrics-go/db"
	"sync"
	"time"
)

func CollectMetrics(ctx0 context.Context, m *Metrics, db0 *sql.DB, wg *sync.WaitGroup) {
	ticker := time.NewTicker(15 * time.Second)

	ctxs := make([]context.Context, 6)
	conns := make([]*sql.Conn, 6)
	//cancels := make([]context.CancelFunc, 6)

	// 创建 6 个子 context，每个带 5 秒超时
	for i := 0; i < 6; i++ {
		ctxs[i] = context.Background()
		conns[i], _ = db0.Conn(ctxs[i])
	}

	for range ticker.C {
		// PRMissingReport
		wg.Add(5)
		go func() {
			ctx := ctxs[0]
			db := conns[0]
			defer wg.Done()
			//defer db.Close()
			//defer cancels[0]()
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
			//_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
			start := time.Now()
			ExportPRMissingReport(ctx, m, db)
			fmt.Println("ExportPRMissingReport", time.Since(start))

		}()

		// PRNum
		go func() {
			ctx := ctxs[1]
			db := conns[1]
			defer wg.Done()
			//defer db.Close()
			//defer cancels[1]()
			_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
			//_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
			start := time.Now()
			ExportPRNum(ctx, m, db)
			fmt.Println("ExportPRNum", time.Since(start))
		}()

		// OpenPRReport
		go func() {
			ctx := ctxs[2]
			db := conns[2]
			defer wg.Done()
			//defer db.Close()
			//defer cancels[2]()
			_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
			//_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
			start := time.Now()
			ExportOpenPRReport(ctx, m, db)
			fmt.Println("ExportOpenPRReport", time.Since(start))
		}()

		// ClosedPRReport
		go func() {
			ctx := ctxs[3]
			db := conns[3]
			defer wg.Done()
			//defer db.Close()
			//defer cancels[3]()
			_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
			//_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
			start := time.Now()
			ExportClosedPRReport(ctx, m, db)
			fmt.Println("ExportClosedPRReport", time.Since(start))
		}()

		// ResultReport
		go func() {
			ctx := ctxs[4]
			db := conns[4]
			defer wg.Done()
			//defer db.Close()
			//defer cancels[4]()
			_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
			//_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
			start := time.Now()
			ExportResultReport(ctx, m, db)
			fmt.Println("ExportResultReport", time.Since(start))
		}()
	}

	//// PRMissingReport
	//wg.Add(5)
	//go func() {
	//	ctx := ctxs[0]
	//	db := conns[0]
	//	defer wg.Done()
	//	defer db.Close()
	//	//defer cancels[0]()
	//	_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
	//	for range ticker.C {
	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
	//		start := time.Now()
	//		ExportPRMissingReport(ctx, m, db)
	//		fmt.Println("ExportPRMissingReport", time.Since(start))
	//	}
	//
	//}()
	//
	//// PRNum
	//go func() {
	//	ctx := ctxs[1]
	//	db := conns[1]
	//	defer wg.Done()
	//	defer db.Close()
	//	//defer cancels[1]()
	//	_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
	//	for range ticker.C {
	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
	//		start := time.Now()
	//		ExportPRNum(ctx, m, db)
	//		fmt.Println("ExportPRNum", time.Since(start))
	//	}
	//}()
	//
	//// OpenPRReport
	//go func() {
	//	ctx := ctxs[2]
	//	db := conns[2]
	//	defer wg.Done()
	//	defer db.Close()
	//	//defer cancels[2]()
	//	_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
	//	for range ticker.C {
	//		//_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
	//		start := time.Now()
	//		ExportOpenPRReport(ctx, m, db)
	//		fmt.Println("ExportOpenPRReport", time.Since(start))
	//	}
	//}()
	//
	//// ClosedPRReport
	//go func() {
	//	ctx := ctxs[3]
	//	db := conns[3]
	//	defer wg.Done()
	//	defer db.Close()
	//	//defer cancels[3]()
	//	_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
	//	for range ticker.C {
	//		//_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
	//		start := time.Now()
	//		ExportClosedPRReport(ctx, m, db)
	//		fmt.Println("ExportClosedPRReport", time.Since(start))
	//	}
	//}()
	//
	//// ResultReport
	//go func() {
	//	ctx := ctxs[4]
	//	db := conns[4]
	//	defer wg.Done()
	//	defer db.Close()
	//	//defer cancels[4]()
	//	_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
	//	for range ticker.C {
	//		//_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
	//		start := time.Now()
	//		ExportResultReport(ctx, m, db)
	//		fmt.Println("ExportResultReport", time.Since(start))
	//	}
	//}()

	wg.Wait()
	//log.Println("所有采集指标完成")
}
