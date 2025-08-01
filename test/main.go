package main

import (
	"context"
	"fmt"
	"oneclick-metrics-go/db"
	"oneclick-metrics-go/metrics"
)

func main() {
	if err := db.InitDb(); err != nil {
		fmt.Println("❌ 数据库初始化失败:", err)
		return
	}
	fmt.Println("✅ 数据库连接成功！")

	var db = db.DB
	ctx := context.Background()
	err := metrics.RegisterPreparedSQLs(ctx, db, true)
	if err != nil {
		return
	}
	//m := metrics.SetupMetrics()
	//
	//err = metrics.ExportPRMissingReport(ctx, m, db, StmtSlice["pr_missing_report_query"], "date_trunc('minute',current_timestamp AT TIME ZONE 'UTC')")
	//if err != nil {
	//	return
	//}
	//var pool = db.Pool
	//defer pool.Close()
	//row := pool.QueryRow(context.Background(), "select 1")
	//var result int
	//if err := row.Scan(&result); err != nil {
	//	fmt.Println("执行sql测试失败：%w", err)
	//	return
	//}

	//fmt.Println("测试 SQL 执行成功，结果为:", result)
}
