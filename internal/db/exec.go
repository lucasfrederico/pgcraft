package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// QueryResult representa output de uma query ad-hoc — colunas, rows
// formatadas como string, timing, e flag se foi truncated.
type QueryResult struct {
	Cols      []string
	Rows      [][]string
	RowCount  int
	Truncated bool
	TimeMs    int64
}

// MaxRows é o cap defensivo. Phase 3 não tem paginação backend ainda;
// 1000 rows cobre maioria dos cases sem estourar memory na TUI.
const MaxRows = 1000

// ExecQuery roda SQL arbitrária. Retorna QueryResult formatado.
// Trim de whitespace ajuda quando user copia SQL multi-linha com tabs.
func (c *Client) ExecQuery(ctx context.Context, sql string) (*QueryResult, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, fmt.Errorf("empty query")
	}

	start := time.Now()

	rows, err := c.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	// header names from FieldDescriptions
	fds := rows.FieldDescriptions()
	cols := make([]string, len(fds))
	for i, fd := range fds {
		cols[i] = string(fd.Name)
	}

	result := &QueryResult{Cols: cols}

	for rows.Next() {
		if result.RowCount >= MaxRows {
			result.Truncated = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("scan row %d: %w", result.RowCount, err)
		}
		strRow := make([]string, len(values))
		for i, v := range values {
			strRow[i] = formatValue(v)
		}
		result.Rows = append(result.Rows, strRow)
		result.RowCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iter: %w", err)
	}

	result.TimeMs = time.Since(start).Milliseconds()
	return result, nil
}

// formatValue convert pgx values pra string display-friendly.
// Phase 3 keep simples — Phase 4 pode adicionar JSON pretty, bytea hex, etc.
func formatValue(v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case string:
		// truncate strings longas no display, evitar uma row gigante quebrar layout
		if len(x) > 80 {
			return x[:77] + "..."
		}
		return x
	case time.Time:
		return x.Format(time.RFC3339)
	case time.Duration:
		return x.Round(time.Second).String()
	case pgtype.Interval:
		// pgx Interval { Microseconds, Days, Months, Valid }
		if !x.Valid {
			return "NULL"
		}
		d := time.Duration(x.Microseconds) * time.Microsecond
		if x.Days != 0 || x.Months != 0 {
			return fmt.Sprintf("%dmo %dd %s", x.Months, x.Days, d.Round(time.Second))
		}
		return d.Round(time.Second).String()
	case []byte:
		if len(x) > 16 {
			return fmt.Sprintf("\\x%x...", x[:16])
		}
		return fmt.Sprintf("\\x%x", x)
	default:
		s := fmt.Sprintf("%v", x)
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	}
}

// IsReadOnly retorna true se a query parece SELECT/SHOW/EXPLAIN.
// Heurística simples baseada no prefixo após whitespace stripping.
// Phase 3 ainda permite write queries (sem confirmation), Phase 4 pode
// adicionar prompt "are you sure?" pra INSERT/UPDATE/DELETE/DROP.
func IsReadOnly(sql string) bool {
	sql = strings.TrimSpace(strings.ToLower(sql))
	prefixes := []string{"select ", "show ", "explain ", "with "}
	for _, p := range prefixes {
		if strings.HasPrefix(sql, p) {
			return true
		}
	}
	return false
}

// EnsureMaxConn assegura que ExecQuery usa só 1 conn — não interfere
// com queries de schema browser rolando em paralelo. (Por enquanto
// pool maxConns=2 cobre isso naturalmente.)
var _ = pgx.RowToMap // garantir que pgx import é usado caso refatore depois
