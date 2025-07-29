package main

import (
	"log"
	"net/http"
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

	// 定时采集指标
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		log.Println("Collecting metrics...")
		metrics.CollectMetrics(m)
	}
}
