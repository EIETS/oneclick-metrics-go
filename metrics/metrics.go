package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
)

var wg sync.WaitGroup

type Metrics struct {
	OneClickPRMissingReport *prometheus.GaugeVec
	OneClickPRNum           *prometheus.GaugeVec
	OneClickOpenPRReport    *prometheus.GaugeVec
	OneClickClosedPRReport  *prometheus.GaugeVec
	OneClickResultReport    *prometheus.GaugeVec
	OneClickCheckSummary    *prometheus.GaugeVec
}

func SetupMetrics() *Metrics {
	return &Metrics{
		OneClickPRMissingReport: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_pr_missing_report",
				Help: "The number of pull requests in one project with missing reports",
			},
			[]string{"pr_state", "project"},
		),
		OneClickPRNum: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_pr_num",
				Help: "The number of pull requests in one project by state",
			},
			[]string{"pr_state", "project"},
		),
		OneClickOpenPRReport: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_open_pr_report",
				Help: "Report detail for open state pull requests by pull request id",
			},
			[]string{"stash_repo", "pr_no", "valid_appearance", "present_reports", "project", "code_coverage"},
		),
		OneClickClosedPRReport: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_closed_pr_report",
				Help: "Report detail for recent one minute closed pull requests by pull request id",
			},
			[]string{"stash_repo", "pr_no", "valid_appearance", "present_reports", "project", "code_coverage"},
		),
		OneClickResultReport: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_result_report",
				Help: "The number of reports for open state pull requests in one project by report result category",
			},
			[]string{"report", "status", "project"},
		),
		OneClickCheckSummary: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "oneclick_check_summary",
				Help: "Report detail for enabled code insight checks by stash_repo",
			},
			[]string{"stash_repo", "valid_appearance", "enabled_checks", "project"},
		),
	}
}

func RegisterMetrics(m *Metrics) {
	prometheus.MustRegister(
		m.OneClickPRMissingReport,
		m.OneClickPRNum,
		m.OneClickOpenPRReport,
		m.OneClickClosedPRReport,
		m.OneClickResultReport,
		m.OneClickCheckSummary,
	)
	zap.S().Info("Registered Prometheus metrics")
}
