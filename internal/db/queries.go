package db

import (
	"context"
	"fmt"
)

// Queries em information_schema / pg_catalog.
// Filtros default escondem schemas internos do Postgres (pg_*, information_schema).
// Phase 4 pode adicionar flag pra mostrar schemas internos.

// ListSchemas retorna schemas user-facing ordenados alfabeticamente,
// com 'public' primeiro (convenção).
func (c *Client) ListSchemas(ctx context.Context) ([]string, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%'
		  AND schema_name <> 'information_schema'
		ORDER BY
			CASE WHEN schema_name = 'public' THEN 0 ELSE 1 END,
			schema_name
	`)
	if err != nil {
		return nil, fmt.Errorf("list schemas: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ListTables retorna tables + views em ordem alfabética.
// Em Phase 3 podemos quebrar em (tables, views) separados.
func (c *Client) ListTables(ctx context.Context, schema string) ([]string, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("list tables in %s: %w", schema, err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// Column é uma linha de information_schema.columns + flag de PK
// (separar via pg_constraint seria mais preciso; isso basta pra Phase 2).
type Column struct {
	Name     string
	DataType string
	Nullable bool
	IsPK     bool
}

// ListColumns retorna colunas com tipo. PK é detectado via pg_constraint.
func (c *Client) ListColumns(ctx context.Context, schema, table string) ([]Column, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS nullable,
			EXISTS (
				SELECT 1
				FROM pg_constraint con
				JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = ANY(con.conkey)
				WHERE con.contype = 'p'
				  AND con.conrelid = (
					SELECT c2.oid FROM pg_class c2
					JOIN pg_namespace n ON n.oid = c2.relnamespace
					WHERE n.nspname = $1 AND c2.relname = $2
				  )
				  AND a.attname = c.column_name
			) AS is_pk
		FROM information_schema.columns c
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("list columns of %s.%s: %w", schema, table, err)
	}
	defer rows.Close()

	var out []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Name, &col.DataType, &col.Nullable, &col.IsPK); err != nil {
			return nil, err
		}
		out = append(out, col)
	}
	return out, rows.Err()
}

// Index representa pg_indexes row simplificado.
type Index struct {
	Name       string
	Definition string
	IsPrimary  bool
}

// ListIndexes retorna indexes da table.
func (c *Client) ListIndexes(ctx context.Context, schema, table string) ([]Index, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT
			i.indexname,
			i.indexdef,
			COALESCE(c.contype = 'p', false) AS is_primary
		FROM pg_indexes i
		LEFT JOIN pg_constraint c
			ON c.conname = i.indexname
		   AND c.contype = 'p'
		WHERE i.schemaname = $1 AND i.tablename = $2
		ORDER BY is_primary DESC, i.indexname
	`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("list indexes of %s.%s: %w", schema, table, err)
	}
	defer rows.Close()

	var out []Index
	for rows.Next() {
		var idx Index
		if err := rows.Scan(&idx.Name, &idx.Definition, &idx.IsPrimary); err != nil {
			return nil, err
		}
		out = append(out, idx)
	}
	return out, rows.Err()
}
