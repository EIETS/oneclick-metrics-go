package metrics

import (
	"context"
	"database/sql"
	"log"
	"strconv"
)

var allProjects map[string]struct{} = make(map[string]struct{})

type FirstCall struct {
	flag bool `default:"true"`
}

type PrMissingReport struct {
	Count   string
	State   string
	Project string
}

type PrNum struct {
	State   string
	Project string
	Count   string
}

type OpenPrReport struct {
	StashRepo     string
	PrNo          string
	PresentReport string
	Project       string
	ReportData    string
}

type ClosedPrReport = OpenPrReport

type ResultReport struct {
	Report  string
	Status  string
	Project string
	Count   string
}

type CheckSummary struct {
	StashRepo     string
	PresrntReport string
	Project       string
}

// GetAllProjects 从数据库中获取所有项目的名称
func GetAllProjects(ctx context.Context, db *sql.DB, knownProjects map[string]struct{}) (map[string]struct{}, error) {
	// 如果缓存为空，则从数据库查询
	if len(allProjects) == 0 {
		query := `
            SELECT DISTINCT project.project_key
            FROM sta_pull_request pr
            INNER JOIN repository repo ON pr.from_repository_id = repo.id
            INNER JOIN project ON repo.project_id = project.id
            WHERE project.project_key NOT LIKE '~%';
        `
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("获取项目名称过程中有误: %v", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var projectKey string
			if err = rows.Scan(&projectKey); err != nil {
				log.Printf("浏览项目key时出错: %v", err)
				continue
			}
			allProjects[projectKey] = struct{}{}
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	// 合并 knownProjects
	for k := range knownProjects {
		allProjects[k] = struct{}{}
	}

	return allProjects, nil
}

func ExportPRMissingReport(ctx context.Context, m *Metrics, db *sql.DB) error {
	//defer wg.Done()

	pgsql := "EXECUTE pr_missing_report_query(date_trunc('minute',current_timestamp AT TIME ZONE 'UTC'));"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		return err
	}
	var rawResults []PrMissingReport          // 获取本次sql查询的结果
	knowProjects := make(map[string]struct{}) // 保存本次查询中涉及的project
	for rows.Next() {
		var r PrMissingReport
		if err = rows.Scan(&r.Count, &r.State, &r.Project); err != nil {
			return err
		}
		rawResults = append(rawResults, r)
		knowProjects[r.Project] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return err
	}

	allProjects, err := GetAllProjects(ctx, db, knowProjects)
	if err != nil {
		return err
	}

	// 初始化计数器
	cntArr := make(map[string][3]int)
	for project := range allProjects {
		cntArr[project] = [3]int{0, 0, 0}
	}

	// 填充数据, cntArr 中的数据是每一个project中每一个状态的丢失报告的pr的数量
	for _, r := range rawResults {
		if state, _ := strconv.Atoi(r.State); state >= 0 && state <= 2 {
			arr := cntArr[r.Project]
			count, _ := strconv.Atoi(r.Count)
			arr[state] = count
			cntArr[r.Project] = arr
		}
	}

	for project, counts := range cntArr {
		m.OneClickPRMissingReport.WithLabelValues("open", project).Set(float64(counts[0]))
		m.OneClickPRMissingReport.WithLabelValues("merged", project).Set(float64(counts[1]))
		m.OneClickPRMissingReport.WithLabelValues("declined", project).Set(float64(counts[2]))
	}

	return nil
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
