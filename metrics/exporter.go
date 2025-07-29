package metrics

func ExportPRMissingReport(m *Metrics) {
	defer wg.Done()
	m.OneClickPRMissingReport.WithLabelValues(
		"pr_state",
		"project",
	).Set(float64(1))
}

func ExportPRNum(m *Metrics) {
	defer wg.Done()
	m.OneClickPRNum.WithLabelValues(
		"pr_state",
		"project",
	).Set(float64(1))
}

func ExportOpenPRReport(m *Metrics) {
	defer wg.Done()
	m.OneClickOpenPRReport.WithLabelValues(
		"stash_repo",
		"pr_no",
		"valid_appearance",
		"present_reports",
		"project",
		"code_coverage",
	).Set(float64(1))
}

func ExportClosedPRReport(m *Metrics) {
	defer wg.Done()
	m.OneClickClosedPRReport.WithLabelValues(
		"stash_repo",
		"pr_no",
		"valid_appearance",
		"present_reports",
		"project",
		"code_coverage",
	).Set(float64(1))
}

func ExportResultReport(m *Metrics) {
	defer wg.Done()
	m.OneClickResultReport.WithLabelValues(
		"report",
		"status",
		"project",
	).Set(float64(1))
}

func ExportCheckSummary(m *Metrics) {
	defer wg.Done()
	m.OneClickCheckSummary.WithLabelValues(
		"stash_repo",
		"valid_appearance",
		"enabled_checks",
		"project",
	).Set(float64(1))

}
