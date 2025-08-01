package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"oneclick-metrics-go/db"
	"oneclick-metrics-go/metrics"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 启动 Prometheus 指标服务
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting metrics server on :8000")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	// 注册指标
	m := metrics.SetupMetrics()
	metrics.RegisterMetrics(m)

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
	// 定时采集指标
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		log.Println("Collecting metrics...")
		//metrics.CollectMetrics(m)
		err = metrics.ExportPRMissingReport(ctx, m, db)
		if err != nil {
			fmt.Println(err)
		}
	}
}
