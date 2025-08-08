package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"oneclick-metrics-go/db"
	"oneclick-metrics-go/metrics"
	"sync"
)

func main() {
	// 启动 Prometheus 指标服务
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting metrics server on :8000")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	// 指标注册
	m := metrics.SetupMetrics()
	metrics.RegisterMetrics(m)

	if err := db.InitDb(); err != nil {
		fmt.Println("❌ 数据库初始化失败:", err)
		return
	}
	fmt.Println("✅ 数据库连接成功！")

	var dB = db.DB
	ctx := context.Background()
	for {
		wg := sync.WaitGroup{}
		metrics.CollectMetrics(ctx, m, dB, &wg)
	}

}
