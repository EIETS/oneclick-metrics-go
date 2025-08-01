package metrics

import (
	"context"
	"database/sql"
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

// 注册pgsql语句
func RegisterPreparedSQLs(ctx context.Context, db *sql.DB, firstCall bool) error {
	regSQLs := map[string]string{
		"check_summary_query(text[])": `
			SELECT
                repo.slug as stash_repo,
				string_agg(DISTINCT "REPORT_KEY",',') as present_report,
				project.project_key as project
	        FROM 
                "AO_2AD648_MERGE_CHECK" as merge_check
                RIGHT JOIN repository repo on merge_check."RESOURCE_ID" = repo.id AND merge_check."SCOPE_TYPE" = 'REPOSITORY'
                INNER JOIN project on repo.project_id = project.id
	        WHERE 
                project.project_key = ANY($1)
	        GROUP BY
                project.project_key,
				repo.slug	`,
		"closed_pr_report_query(timestamp)": `
			SELECT
                repo.slug AS stash_repo,
                pr.scoped_id AS prno,
                string_agg (DISTINCT insrep."REPORT_KEY", ',') FILTER (WHERE insrep."REPORT_KEY" in ( 'sast','sabug','savul','sasmell','smoke','ci','codecoverage','snyk' )) AS present_reports,
                project.project_key AS "project",
				'[' || string_agg(DISTINCT nullif(insrep."DATA",''),',') || ']' AS report_data
            FROM
                sta_pull_request AS pr
                INNER JOIN repository repo on pr.to_repository_id = repo.id
                INNER JOIN project on repo.project_id = project.id
                LEFT  JOIN "AO_2AD648_INSIGHT_REPORT" insrep on pr.from_hash = insrep."COMMIT_ID"
                and pr.to_repository_id = insrep."REPOSITORY_ID"
            WHERE
                project.project_key not like '~%' AND (pr.closed_timestamp >= (date_trunc('minute', $1) - INTERVAL '1 minute')
            AND pr.closed_timestamp < date_trunc('minute', $1))
            GROUP BY
                stash_repo,
                prno,
                project.project_key
            ORDER BY
                stash_repo,
                prno,
                project.project_key`,
		"open_pr_report_query": `
			SELECT
                repo.slug AS stash_repo,
                pr.scoped_id AS prno,
                string_agg (DISTINCT insrep."REPORT_KEY", ',') FILTER (WHERE insrep."REPORT_KEY" in ( 'sast','sabug','savul','sasmell','smoke','ci','codecoverage','snyk' )) AS present_reports,
                project.project_key AS "project",
                '[' || string_agg(DISTINCT nullif(insrep."DATA",''),',') || ']' AS report_data
            FROM
                sta_pull_request AS pr
                INNER JOIN repository repo on pr.to_repository_id = repo.id
                INNER JOIN project on repo.project_id = project.id
                LEFT  JOIN "AO_2AD648_INSIGHT_REPORT" insrep on pr.from_hash = insrep."COMMIT_ID"
                and pr.to_repository_id = insrep."REPOSITORY_ID"
            WHERE
                pr.pr_state = 0 AND project.project_key not like '~%'
            GROUP BY
                stash_repo,
                prno,
                project.project_key
            ORDER BY
                stash_repo,
                prno,
                project.project_key`,
		"result_report_query(text)": `
			WITH prbase AS(
                SELECT
				    pr.id AS pr_id,
                    CASE WHEN insrep."REPORT_KEY" = $1 THEN insrep."REPORT_KEY" ELSE null END AS report,
					string_agg(CASE WHEN insrep."REPORT_KEY" = $1 THEN insrep."REPORT_KEY" ELSE null END,',') OVER prw AS allrepts,
                    insrep."RESULT_ID" AS status,
                    project.project_key AS project_key
                FROM
                    sta_pull_request pr
                    INNER JOIN repository repo ON pr.to_repository_id = repo.id
                    INNER JOIN project ON repo.project_id = project.id
                    LEFT JOIN "AO_2AD648_INSIGHT_REPORT" insrep ON pr.from_hash = insrep."COMMIT_ID"
                    AND pr.to_repository_id = insrep."REPOSITORY_ID"
                WHERE
                    pr.pr_state = 0 AND project.project_key NOT LIKE '~%'
                    AND ( insrep."REPORT_KEY" = $1 OR repo.id = ANY ( SELECT "RESOURCE_ID" FROM "AO_2AD648_MERGE_CHECK" WHERE "RESOURCE_ID" = repo.id AND "REPORT_KEY" = $1 ) )
                GROUP BY
                    project.project_key,
                    pr.id,
                    report,
                    status
				WINDOW prw AS (PARTITION BY pr.id)
			)
            SELECT
                    report,
                    status,
                    project_key AS project,
                    COUNT(*) AS cnt
                FROM
				    prbase
                WHERE
                    report = $1
                GROUP BY
                    status,
					report,
                    project_key
			UNION
				SELECT
                    $1 AS report,
                    2 AS status,
                    project_key AS project,
                    COUNT(*) AS cnt
                FROM
				    prbase
				WHERE
				    allrepts IS NULL
                GROUP BY
                    project_key`,
		"pr_counts_query(timestamp)": `
			SELECT
                pr_state,
                project.project_key AS project,
                COUNT(*)
            FROM
                sta_pull_request AS pr
                INNER JOIN repository AS repo ON pr.to_repository_id = repo.id
                INNER JOIN project AS project ON repo."project_id" = project.id
            WHERE project.project_key not like '~%'
                AND ( pr.closed_timestamp is null OR (pr.closed_timestamp >= ($1 - interval '1 minute') and pr.closed_timestamp < $1 ) )
            GROUP BY
                pr_state,
                project.project_key`,
		"pr_missing_report_query(timestamp)": `
			SELECT SUM(cnt),pr_state,project_key
        FROM (
        -- part 1: count all pr(s) with partial reports
        SELECT COUNT(*) AS cnt,pr_state,project_key
        FROM (
            SELECT
                pr.id,
                pr.pr_state,
                repo.id AS repo_id,
                project.project_key
            FROM
                "AO_2AD648_INSIGHT_REPORT" AS insrep
                INNER JOIN sta_pull_request AS pr ON insrep."COMMIT_ID" = pr.from_hash
                INNER JOIN repository AS repo ON pr.to_repository_id = repo.id
                INNER JOIN project AS project ON repo.project_id = project.id
                INNER JOIN (
                    SELECT
                        merge_check."RESOURCE_ID" AS repo_id,
                        array_agg("REPORT_KEY") AS req_arr
                    FROM
                        "AO_2AD648_MERGE_CHECK" AS merge_check
                    WHERE
                        "SCOPE_TYPE" = 'REPOSITORY'
                        AND "REPORT_KEY" IN ('sast','sabug','savul','sasmell','smoke','ci','codecoverage','snyk')
                    GROUP BY
                        merge_check."RESOURCE_ID"
                ) AS repreq ON repo.id = repreq.repo_id
            WHERE
                project.project_key NOT LIKE '~%'
                AND ( pr.closed_timestamp is NULL OR (pr.closed_timestamp >= ($1 - interval '1 minute') and pr.closed_timestamp < $1 ) )
            GROUP BY
                pr.id,
                pr.pr_state,
                repo.id,
                project.project_key,
                repreq.req_arr
            HAVING NOT (repreq.req_arr <@ array_agg( DISTINCT
				CASE
					WHEN insrep."REPORT_KEY" in ('sast','sabug','savul','sasmell','smoke','ci','codecoverage','snyk') THEN insrep."REPORT_KEY"
					ELSE 'other'::VARCHAR
				END ))
        ) AS partquery
        GROUP BY
          project_key,
          pr_state
        UNION
        -- part 2: count pr(s) without any report
        SELECT count(pr.id) AS cnt, pr.pr_state, project.project_key
        FROM
          "AO_2AD648_INSIGHT_REPORT" AS insrep
          RIGHT JOIN sta_pull_request AS pr ON insrep."COMMIT_ID" = pr.from_hash
          INNER JOIN repository AS repo ON pr.to_repository_id = repo.id
          INNER JOIN project AS project ON repo.project_id = project.id
        WHERE
          insrep."REPORT_KEY" is null
          AND repo.id = ANY ( select "RESOURCE_ID"
                from "AO_2AD648_MERGE_CHECK"
                where "RESOURCE_ID" = repo.id
                AND "REPORT_KEY" in ('sast','sabug','savul','sasmell','smoke','ci','codecoverage','snyk') )
          AND ( pr.closed_timestamp is null OR (pr.closed_timestamp >= ($1 - interval '1 minute') and pr.closed_timestamp < $1 ) )
        GROUP BY project.project_key,pr.pr_state
        ) AS uniontable
        GROUP BY project_key,pr_state`,
	}

	for proto, sqlText := range regSQLs {
		name, param := parseSQLName(proto) // 把sql名称和参数拆分开来

		if !firstCall {
			// 如果不是首次执行sql，sql已经注册过prepare，这时需要先释放prepare过的sql，重新进行prepare
			_, err := db.ExecContext(ctx, "DEALLOCATE "+name)
			if err != nil {
				log.Printf("释放之前 prepare 过的 SQL 时出现错误 %s: %v", name, err)
			}
		}

		// 拼接完整的pgsql prepare语句
		prepareStmt := "PREPARE " + name + param + " AS " + sqlText
		_, err := db.ExecContext(ctx, prepareStmt)
		if err != nil {
			log.Printf("preparing SQL 时出错 %s: %v", name, err)
		}
		//fmt.Println(prepareStmt)
	}

	return nil
}

// 解析sql名称和参数
func parseSQLName(proto string) (name string, param string) {
	idx := strings.Index(proto, "(")
	if idx == -1 {
		return proto, ""
	}
	return proto[:idx], proto[idx:]
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
