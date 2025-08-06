package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	dbase "oneclick-metrics-go/db"
	"sync"
	"time"
)

func CollectMetrics(ctx0 context.Context, m *Metrics, db0 *sql.DB, wg *sync.WaitGroup) {
	ticker := time.NewTicker(15 * time.Second)

	type MetricTask struct {
		Name     string
		QueryKey string
		ExportFn func(ctx context.Context, m *Metrics, db *sql.Conn)
	}

	tasks := []MetricTask{
		{"PRMissingReport", "pr_missing_report_query", ExportPRMissingReport},
		{"PRNum", "pr_counts_query", ExportPRNum},
		{"OpenPRReport", "open_pr_report_query", ExportOpenPRReport},
		{"ClosedPRReport", "closed_pr_report_query", ExportClosedPRReport},
		{"ResultReport", "result_report_query", ExportResultReport},
		{"CheckSummary", "check_summary_query", ExportCheckSummary},
	}

	// 初始化连接和 context
	conns := make([]*sql.Conn, len(tasks))

	for i, task := range tasks {
		conn, err := ensureConn(ctx0, db0, task.Name)
		if err != nil {
			log.Printf("初始化连接失败 [%s]: %v", tasks[i].Name, err)
			continue
		}
		conns[i] = conn
		defer conn.Close()

		// 初次 prepare
		if err := dbase.RegisterPreparedSQLsWithRetry(ctx0, task.QueryKey, conn, true); err != nil {
			log.Printf("初次 prepare 失败 [%s]: %v", task.Name, err)
		}
	}

	for range ticker.C {
		for i, task := range tasks {
			// 检查conn连接初始化是否完成
			if conns[i] == nil {
				log.Printf("[%s] 跳过执行：连接未初始化", task.Name)
				continue
			}
			wg.Add(1)
			go func(i int, task MetricTask) {
				defer wg.Done()
				start := time.Now()

				// 检查连接是否有效
				if err := conns[i].PingContext(ctx0); err != nil {
					log.Printf("[%s] 连接失效，尝试重连: %v", task.Name, err)
					newConn, err := ensureConn(ctx0, db0, task.Name)
					if err != nil {
						log.Printf("[%s] 重连失败: %v", task.Name, err)
						return
					}
					conns[i].Close()
					conns[i] = newConn
				}

				//_ = dbase.RegisterPreparedSQLs(ctx0, task.QueryKey, conns[i], false)
				task.ExportFn(ctx0, m, conns[i])
				log.Printf("%s 执行耗时: %v", task.Name, time.Since(start))
			}(i, task)
		}
		wg.Wait() // 等待所有任务完成后再进入下一轮 tick
	}
}

// 连接初始化和重连机制
func ensureConn(ctx context.Context, db *sql.DB, taskName string) (*sql.Conn, error) {
	var conn *sql.Conn
	var err error
	for retry := 0; retry < 3; retry++ {
		conn, err = db.Conn(ctx)
		if err == nil {
			return conn, nil
		}
		log.Printf("[%s] 获取连接失败（重试 %d）: %v", taskName, retry+1, err)
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("[%s] 获取连接失败，已重试 3 次: %w", taskName, err)
}

//func CollectMetrics(ctx0 context.Context, m *Metrics, db0 *sql.DB, wg *sync.WaitGroup) {
//
//	//ticker := time.NewTicker(15 * time.Second)
//	//
//	//ctxs := make([]context.Context, 6)
//	//conns := make([]*sql.Conn, 6)
//	////cancels := make([]context.CancelFunc, 6)
//	//
//	//// 创建 6 个子 context，每个带 5 秒超时
//	//for i := 0; i < 6; i++ {
//	//	ctxs[i] = context.Background()
//	//	conns[i], _ = db0.Conn(ctxs[i])
//	//}
//	//
//	//for range ticker.C {
//	//	// PRMissingReport
//	//	wg.Add(5)
//	//	go func() {
//	//		ctx := ctxs[0]
//	//		db := conns[0]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[0]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
//	//		start := time.Now()
//	//		ExportPRMissingReport(ctx, m, db)
//	//		fmt.Println("ExportPRMissingReport", time.Since(start))
//	//
//	//	}()
//	//
//	//	// PRNum
//	//	go func() {
//	//		ctx := ctxs[1]
//	//		db := conns[1]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[1]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
//	//		start := time.Now()
//	//		ExportPRNum(ctx, m, db)
//	//		fmt.Println("ExportPRNum", time.Since(start))
//	//	}()
//	//
//	//	// OpenPRReport
//	//	go func() {
//	//		ctx := ctxs[2]
//	//		db := conns[2]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[2]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
//	//		start := time.Now()
//	//		ExportOpenPRReport(ctx, m, db)
//	//		fmt.Println("ExportOpenPRReport", time.Since(start))
//	//	}()
//	//
//	//	// ClosedPRReport
//	//	go func() {
//	//		ctx := ctxs[3]
//	//		db := conns[3]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[3]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
//	//		start := time.Now()
//	//		ExportClosedPRReport(ctx, m, db)
//	//		fmt.Println("ExportClosedPRReport", time.Since(start))
//	//	}()
//	//
//	//	// ResultReport
//	//	go func() {
//	//		ctx := ctxs[4]
//	//		db := conns[4]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[4]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
//	//		start := time.Now()
//	//		ExportResultReport(ctx, m, db)
//	//		fmt.Println("ExportResultReport", time.Since(start))
//	//	}()
//	//
//	//	// CheckSummary
//	//	go func() {
//	//		ctx := ctxs[5]
//	//		db := conns[5]
//	//		defer wg.Done()
//	//		//defer db.Close()
//	//		//defer cancels[4]()
//	//		_ = dbase.RegisterPreparedSQLs(ctx, "check_summary_query", db, true)
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
//	//		start := time.Now()
//	//		ExportCheckSummary(ctx, m, db)
//	//		fmt.Println("ExportResultReport", time.Since(start))
//	//	}()
//	//}
//	//
//	//wg.Wait()
//
//	//// PRMissingReport
//	//wg.Add(5)
//	//go func() {
//	//	ctx := ctxs[0]
//	//	db := conns[0]
//	//	defer wg.Done()
//	//	defer db.Close()
//	//	//defer cancels[0]()
//	//	_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, true)
//	//	for range ticker.C {
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_missing_report_query", db, false)
//	//		start := time.Now()
//	//		ExportPRMissingReport(ctx, m, db)
//	//		fmt.Println("ExportPRMissingReport", time.Since(start))
//	//	}
//	//
//	//}()
//	//
//	//// PRNum
//	//go func() {
//	//	ctx := ctxs[1]
//	//	db := conns[1]
//	//	defer wg.Done()
//	//	defer db.Close()
//	//	//defer cancels[1]()
//	//	_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, true)
//	//	for range ticker.C {
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "pr_counts_query", db, false)
//	//		start := time.Now()
//	//		ExportPRNum(ctx, m, db)
//	//		fmt.Println("ExportPRNum", time.Since(start))
//	//	}
//	//}()
//	//
//	//// OpenPRReport
//	//go func() {
//	//	ctx := ctxs[2]
//	//	db := conns[2]
//	//	defer wg.Done()
//	//	defer db.Close()
//	//	//defer cancels[2]()
//	//	_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, true)
//	//	for range ticker.C {
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "open_pr_report_query", db, false)
//	//		start := time.Now()
//	//		ExportOpenPRReport(ctx, m, db)
//	//		fmt.Println("ExportOpenPRReport", time.Since(start))
//	//	}
//	//}()
//	//
//	//// ClosedPRReport
//	//go func() {
//	//	ctx := ctxs[3]
//	//	db := conns[3]
//	//	defer wg.Done()
//	//	defer db.Close()
//	//	//defer cancels[3]()
//	//	_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, true)
//	//	for range ticker.C {
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "closed_pr_report_query", db, false)
//	//		start := time.Now()
//	//		ExportClosedPRReport(ctx, m, db)
//	//		fmt.Println("ExportClosedPRReport", time.Since(start))
//	//	}
//	//}()
//	//
//	//// ResultReport
//	//go func() {
//	//	ctx := ctxs[4]
//	//	db := conns[4]
//	//	defer wg.Done()
//	//	defer db.Close()
//	//	//defer cancels[4]()
//	//	_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, true)
//	//	for range ticker.C {
//	//		//_ = dbase.RegisterPreparedSQLs(ctx, "result_report_query", db, false)
//	//		start := time.Now()
//	//		ExportResultReport(ctx, m, db)
//	//		fmt.Println("ExportResultReport", time.Since(start))
//	//	}
//	//}()
//
//	//log.Println("所有采集指标完成")
//}
