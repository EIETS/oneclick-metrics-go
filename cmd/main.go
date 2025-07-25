package main

import (
	"log"
	"net/http"
	"time"

	"omclick-metrics-go/metrics"

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
	metrics.RegisterMetrics()

	// 定时采集指标
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		log.Println("Collecting metrics...")
		metrics.CollectDummyMetric()
	}
}
