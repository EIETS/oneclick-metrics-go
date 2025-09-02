package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"oneclick-metrics-go/db"
	"oneclick-metrics-go/global"
	"oneclick-metrics-go/initialize"
	"oneclick-metrics-go/metrics"
	"sync"
)

func main() {
	// 初始化配置
	initialize.InitConfig()

	// 初始化日志设置
	initialize.InitLogger()

	// 启动 Prometheus 指标服务
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		zap.S().Infof("Starting metrics server on %s:%s", global.ServerConfig.Host, global.ServerConfig.Port)
		zap.S().Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", global.ServerConfig.Host, global.ServerConfig.Port), nil))
	}()

	// 注册指标
	m := metrics.SetupMetrics()
	metrics.RegisterMetrics(m)

	if err := db.InitDb(); err != nil {
		zap.S().Errorf("数据库初始化失败:", err)
		return
	}
	zap.S().Info("数据库连接成功！")

	var dB = db.DB
	ctx := context.Background()
	for {
		wg := sync.WaitGroup{}
		metrics.CollectMetrics(ctx, m, dB, &wg)
	}

}
