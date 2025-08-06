package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

func ParseSqlDict(s string) (sqlNameAndParam, sqlText string) {
	sqlNameMap := map[string]string{
		"check_summary_query":     "check_summary_query(text[])",
		"closed_pr_report_query":  "closed_pr_report_query(timestamp)",
		"open_pr_report_query":    "open_pr_report_query",
		"result_report_query":     "result_report_query(text)",
		"pr_counts_query":         "pr_counts_query(timestamp)",
		"pr_missing_report_query": "pr_missing_report_query(timestamp)",
	}
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

	sqlNameAndParam = sqlNameMap[s]
	sqlText = regSQLs[sqlNameAndParam]
	return sqlNameAndParam, sqlText
}

// 带重连机制的RegisterPreparedSQLs
func RegisterPreparedSQLsWithRetry(ctx context.Context, queryKey string, conn *sql.Conn, isInit bool) error {
	const maxRetries = 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = RegisterPreparedSQLs(ctx, queryKey, conn, isInit)
		if err == nil {
			return nil
		}

		log.Printf("RegisterPreparedSQLs 失败 [%s]（第 %d 次尝试）: %v", queryKey, attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second) // 指数退避策略可选
	}

	return fmt.Errorf("RegisterPreparedSQLs 最终失败 [%s]: %w", queryKey, err)
}

// 注册pgsql语句
func RegisterPreparedSQLs(ctx context.Context, sqlName string, db *sql.Conn, firstCall bool) error {
	proto, sqlText := ParseSqlDict(sqlName)
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
