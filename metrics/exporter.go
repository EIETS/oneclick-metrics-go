package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

var dummyGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "oneclick_dummy_metric",
		Help: "A dummy metric for testing",
	},
	[]string{"label"},
)

func RegisterMetrics() {
	prometheus.MustRegister(dummyGauge)
	log.Println("Registered Prometheus metrics")
}

func CollectDummyMetric() {
	dummyGauge.WithLabelValues("test").Set(float64(1))
}
