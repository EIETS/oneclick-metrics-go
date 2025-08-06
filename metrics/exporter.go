package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

var allProjects map[string]struct{} = make(map[string]struct{})

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
	StashRepo      string
	PrNo           string
	PresentReports sql.NullString
	Project        string
	ReportData     sql.NullString
}

type ClosedPrReport = OpenPrReport

type ResultReport struct {
	Report  string
	Status  sql.NullInt64
	Project string
	Count   int
}

type CheckSummary struct {
	StashRepo     string
	PresrntReport sql.NullString
	Project       string
}

// GetAllProjects 从数据库中获取所有项目的名称
func GetAllProjects(ctx context.Context, db *sql.Conn, knownProjects map[string]struct{}) (map[string]struct{}, error) {
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

// 将从sql中获取的原始数据解析为代码覆盖率
func ExtractCodeCoverage(rawData string) string {
	if rawData == "" {
		return ""
	}
	/*
		获取第一个 code coverage 值
		    元素类型：'[[{dict1}, {dict2}], [{dict3}, ...]]'
	*/

	var parsedData [][]map[string]interface{}
	if err := json.Unmarshal([]byte(rawData), &parsedData); err != nil {
		return ""
	}

	for _, subArray := range parsedData {
		for _, item := range subArray {
			if title, ok := item["title"].(string); ok && title == "Code Coverage" {
				if value, ok := item["value"]; ok {
					switch v := value.(type) {
					case float64:
						return fmt.Sprintf("%.2f", v)
					case int:
						return strconv.Itoa(v)
					}
				}
				break
			}
		}
	}

	return ""
}

func ExportPRMissingReport(ctx context.Context, m *Metrics, db *sql.Conn) {
	m.OneClickPRMissingReport.Reset()

	pgsql := "EXECUTE pr_missing_report_query(date_trunc('minute',current_timestamp AT TIME ZONE 'UTC'));"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		log.Printf("ExportPRMissingReport QueryContext 过程中发生错误: %v", err)
		return
	}
	var rawResults []PrMissingReport          // 获取本次sql查询的结果
	knowProjects := make(map[string]struct{}) // 保存本次查询中涉及的project
	for rows.Next() {
		var r PrMissingReport
		if err = rows.Scan(&r.Count, &r.State, &r.Project); err != nil {
			log.Printf("ExportPRMissingReport scan 过程中发生错误: %v", err)
			continue
		}
		rawResults = append(rawResults, r)
		knowProjects[r.Project] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		log.Printf("ExportPRMissingReport Err 过程中发生错误: %v", err)
		return
	}

	allProjects, err = GetAllProjects(ctx, db, knowProjects)
	if err != nil {
		log.Printf("ExportPRMissingReport GetAllProjects 过程中发生错误: %v", err)
		return
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

	return
}

func ExportPRNum(ctx context.Context, m *Metrics, db *sql.Conn) {
	m.OneClickPRNum.Reset()
	pgsql := "EXECUTE pr_counts_query(date_trunc('minute',current_timestamp AT TIME ZONE 'UTC'))"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		log.Printf("ExportPRNum QueryContext 过程中发生错误: %v", err)
		return
	}
	defer rows.Close()

	var rawResults []PrNum
	knowProjects := make(map[string]struct{})

	for rows.Next() {
		var r PrNum
		if err = rows.Scan(&r.State, &r.Project, &r.Count); err != nil {
			log.Printf("ExportPRNum scan 过程中发生错误: %v", err)
			continue
		}
		rawResults = append(rawResults, r)
		knowProjects[r.Project] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		log.Printf("ExportPRNum Err 过程中发生错误: %v", err)
		return
	}

	// 获取所有项目
	allProjects, err = GetAllProjects(ctx, db, knowProjects)
	if err != nil {
		log.Printf("ExportPRNum GetAllProjects 过程中发生错误: %v", err)
		return
	}

	// 初始化计数器
	cntArr := make(map[string][3]int)
	for project := range allProjects {
		cntArr[project] = [3]int{0, 0, 0}
	}

	// 填充数据
	for _, r := range rawResults {
		if state, _ := strconv.Atoi(r.State); state >= 0 && state <= 2 {
			arr := cntArr[r.Project]
			count, _ := strconv.Atoi(r.Count)
			arr[state] = count
			cntArr[r.Project] = arr
		}
	}

	for project, counts := range cntArr {
		m.OneClickPRNum.WithLabelValues("open", project).Set(float64(counts[0]))
		m.OneClickPRNum.WithLabelValues("merged", project).Set(float64(counts[1]))
		m.OneClickPRNum.WithLabelValues("declined", project).Set(float64(counts[2]))
	}

}

func ExportOpenPRReport(ctx context.Context, m *Metrics, db *sql.Conn) {
	m.OneClickOpenPRReport.Reset()
	pgsql := "EXECUTE open_pr_report_query"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		log.Printf("ExportOpenPRReport QueryContext 过程中发生错误: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r OpenPrReport
		if err = rows.Scan(&r.StashRepo, &r.PrNo, &r.PresentReports, &r.Project, &r.ReportData); err != nil {
			log.Printf("ExportOpenPRReport scan 过程中发生错误: %v", err)
			continue
		}

		validAppearance := 0 // 统计经过检验的报告的数量
		presentReports := ""
		if r.PresentReports.Valid {
			presentReports = r.PresentReports.String
		}

		if presentReports != "" {
			validAppearance = strings.Count(presentReports, ",") + 1
		}

		var codeCoverage string // 如果原始数据为null，codeCoverage直接返回为空
		if !r.ReportData.Valid {
			codeCoverage = ""
		} else {
			codeCoverage = ExtractCodeCoverage(r.ReportData.String)
		}

		m.OneClickOpenPRReport.WithLabelValues(
			r.StashRepo,
			r.PrNo,
			strconv.Itoa(validAppearance),
			presentReports,
			r.Project,
			codeCoverage).Set(float64(1))
	}

	if err := rows.Err(); err != nil {
		log.Printf("ExportOpenPRReport rows.Err 过程中发生错误: %v", err)
		return
	}

}

func ExportClosedPRReport(ctx context.Context, m *Metrics, db *sql.Conn) {
	m.OneClickOpenPRReport.Reset()
	pgsql := "EXECUTE closed_pr_report_query(date_trunc('minute',current_timestamp AT TIME ZONE 'UTC'))"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		log.Printf("ExportClosedPRReport QueryContext 过程中发生错误: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r ClosedPrReport
		if err = rows.Scan(&r.StashRepo, &r.PrNo, &r.PresentReports, &r.Project, &r.ReportData); err != nil {
			log.Printf("ExportClosedPRReport scan 过程中发生错误: %v", err)
			continue
		}

		validAppearance := 0 // 统计经过检验的报告的数量
		presentReports := ""
		if r.PresentReports.Valid {
			presentReports = r.PresentReports.String
		}

		if presentReports != "" {
			validAppearance = strings.Count(presentReports, ",") + 1
		}

		var codeCoverage string // 如果原始数据为null，codeCoverage直接返回为空
		if !r.ReportData.Valid {
			codeCoverage = ""
		} else {
			codeCoverage = ExtractCodeCoverage(r.ReportData.String)
		}

		m.OneClickClosedPRReport.WithLabelValues(
			r.StashRepo,
			r.PrNo,
			strconv.Itoa(validAppearance),
			presentReports,
			r.Project,
			codeCoverage).Set(float64(1))
	}

	if err := rows.Err(); err != nil {
		log.Printf("ExportClosedPRReport rows.Err 过程中发生错误: %v", err)
		return
	}
}

func ExportResultReport(ctx context.Context, m *Metrics, db *sql.Conn) {
	allReportKeys := []string{
		"ci", "codecoverage", "sast", "smoke", "snyk", "sabug", "sasmell", "savul",
	}
	// 用于收集所有项目
	knownProjects := make(map[string]struct{})
	reportResults := make(map[string][]ResultReport)

	for _, reportKey := range allReportKeys {
		pgsql := fmt.Sprintf("EXECUTE result_report_query('%s')", reportKey)
		rows, err := db.QueryContext(ctx, pgsql)
		if err != nil {
			log.Printf("ExportResultReport 查询 %s 时发生错误: %v", reportKey, err)
			continue
		}
		defer rows.Close()
		var results []ResultReport
		for rows.Next() {
			var r ResultReport
			if err = rows.Scan(&r.Report, &r.Status, &r.Project, &r.Count); err != nil {
				log.Printf("ExportResultReport 扫描 %s 时发生错误: %v", reportKey, err)
				continue
			}
			results = append(results, r)
			knownProjects[r.Project] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			log.Printf("ExportResultReport rows.Err %s 时发生错误: %v", reportKey, err)
			continue
		}
		reportResults[reportKey] = results
	}

	//获取所有项目
	var err error
	allProjects, err = GetAllProjects(ctx, db, knownProjects)
	if err != nil {
		log.Printf("ExportResultReport 获取所有项目时发生错误: %v", err)
		return
	}

	m.OneClickResultReport.Reset()

	for reportKey, results := range reportResults {
		//初始化计数器
		cntarr := make(map[string][3]int)
		for project := range allProjects {
			cntarr[project] = [3]int{0, 0, 0}
		}

		for _, r := range results {
			status := 2
			if r.Status.Valid {
				status = int(r.Status.Int64)
			}
			if status >= 0 && status <= 2 {
				arr := cntarr[r.Project]
				arr[status] = r.Count
				cntarr[r.Project] = arr
			}
		}

		for project := range allProjects {
			counts := cntarr[project]
			total := counts[0] + counts[1]

			m.OneClickResultReport.WithLabelValues(reportKey, "failure", project).Set(float64(counts[0]))
			m.OneClickResultReport.WithLabelValues(reportKey, "success", project).Set(float64(counts[1]))
			m.OneClickResultReport.WithLabelValues(reportKey, "total", project).Set(float64(total))
			m.OneClickResultReport.WithLabelValues(reportKey, "notAvailable", project).Set(float64(counts[2]))
		}
	}
}

func ExportCheckSummary(ctx context.Context, m *Metrics, db *sql.Conn) {

	// 获取所有项目
	allProjects, err := GetAllProjects(ctx, db, map[string]struct{}{})
	if err != nil {
		log.Printf("ExportCheckSummary 获取项目失败: %v", err)
		return
	}

	// 构造项目列表字符串
	var projectList []string
	for project := range allProjects {
		projectList = append(projectList, project)
	}
	projectListStr := strings.Join(projectList, ",")

	// 执行查询
	query := fmt.Sprintf("EXECUTE check_summary_query('{%s}')", projectListStr)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("ExportCheckSummary 查询失败: %v", err)
		return
	}
	defer rows.Close()

	m.OneClickCheckSummary.Reset()

	for rows.Next() {
		var r CheckSummary
		err = rows.Scan(&r.StashRepo, &r.PresrntReport, &r.Project)
		if err != nil {
			log.Printf("ExportCheckSummary 扫描失败: %v", err)
			continue
		}

		checkSummaryString := ""
		if r.PresrntReport.Valid {
			checkSummaryString = r.PresrntReport.String
		}

		validAppearance := 0
		if checkSummaryString != "" {
			validAppearance = strings.Count(checkSummaryString, ",") + 1
		}

		m.OneClickCheckSummary.WithLabelValues(
			r.StashRepo,
			strconv.Itoa(validAppearance),
			checkSummaryString,
			r.Project,
		).Set(float64(1))
	}

	if err = rows.Err(); err != nil {
		log.Printf("ExportCheckSummary rows.Err过程中发生错误: %v", err)
		return
	}

}
