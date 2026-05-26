package db

import (
	"context"
	"fmt"
)

// Postgres-deep queries — pg_stat_user_tables, pg_stat_activity etc.
// Cada função retorna QueryResult pra reusar render do results panel.

// TableStats retorna stats da table (insert/update/delete count, vacuum,
// dead tuples). Mostra um snapshot que ajuda a entender saúde da table.
func (c *Client) TableStats(ctx context.Context, schema, table string) (*QueryResult, error) {
	sql := `
		SELECT
			schemaname || '.' || relname AS table,
			n_live_tup AS live_rows,
			n_dead_tup AS dead_rows,
			n_tup_ins AS inserts,
			n_tup_upd AS updates,
			n_tup_del AS deletes,
			n_tup_hot_upd AS hot_updates,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables
		WHERE schemaname = $1 AND relname = $2
	`
	return c.execQueryWithArgs(ctx, sql, schema, table)
}

// Activity retorna sessões/queries ativas. Filtros: exclui background workers
// e a própria conexão pra não poluir.
func (c *Client) Activity(ctx context.Context) (*QueryResult, error) {
	sql := `
		SELECT
			pid,
			usename AS user,
			application_name AS app,
			client_addr::text AS client,
			state,
			wait_event_type AS wait_type,
			wait_event,
			now() - query_start AS duration,
			LEFT(query, 80) AS query
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'
		  AND pid <> pg_backend_pid()
		ORDER BY query_start DESC NULLS LAST
		LIMIT 100
	`
	return c.execQueryWithArgs(ctx, sql)
}

// execQueryWithArgs é versão privada de ExecQuery que aceita args.
// ExecQuery (em exec.go) é só pra SQL livre do editor, sem args.
func (c *Client) execQueryWithArgs(ctx context.Context, sql string, args ...any) (*QueryResult, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fds := rows.FieldDescriptions()
	cols := make([]string, len(fds))
	for i, fd := range fds {
		cols[i] = string(fd.Name)
	}
	result := &QueryResult{Cols: cols}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		strRow := make([]string, len(values))
		for i, v := range values {
			strRow[i] = formatValue(v)
		}
		result.Rows = append(result.Rows, strRow)
		result.RowCount++
	}
	return result, rows.Err()
}
