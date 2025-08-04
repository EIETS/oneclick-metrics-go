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

func ExportPRMissingReport(ctx context.Context, m *Metrics, db *sql.DB) {
	//defer wg.Done()
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

func ExportPRNum(ctx context.Context, m *Metrics, db *sql.DB) {
	//defer wg.Done()
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

func ExportOpenPRReport(ctx context.Context, m *Metrics, db *sql.DB) {
	m.OneClickOpenPRReport.Reset()
	pgsql := "EXECUTE open_pr_report_query"
	rows, err := db.QueryContext(ctx, pgsql)
	if err != nil {
		log.Printf("ExportOpenPRReport QueryContext 过程中发生错误: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stashRepo, prNo, project string
		var presentReportsNullString, coverageRaw sql.NullString
		if err = rows.Scan(&stashRepo, &prNo, &presentReportsNullString, &project, &coverageRaw); err != nil {
			log.Printf("ExportOpenPRReport scan 过程中发生错误: %v", err)
			continue
		}

		validAppearance := 0 // 统计经过检验的报告的数量
		presentReports := ""
		if presentReportsNullString.Valid {
			presentReports = presentReportsNullString.String
		}

		if presentReports != "" {
			validAppearance = strings.Count(presentReports, ",") + 1
		}

		var codeCoverage string // 如果原始数据为null，codeCoverage直接返回为空
		if !coverageRaw.Valid {
			codeCoverage = ""
		} else {
			codeCoverage = ExtractCodeCoverage(coverageRaw.String)
		}

		m.OneClickOpenPRReport.WithLabelValues(
			stashRepo,
			prNo,
			strconv.Itoa(validAppearance),
			presentReports,
			project,
			codeCoverage).Set(float64(1))
	}

	if err := rows.Err(); err != nil {
		log.Printf("ExportOpenPRReport rows.Err 过程中发生错误: %v", err)
		return
	}

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
