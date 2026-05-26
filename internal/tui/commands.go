package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lucasfrederico/pgcraft/internal/db"
)

// Commands Bubble Tea — chamam o package db e empacotam resultado
// numa message. Mantidos sem state pra serem reentrant.

// connectCmd dispara conexão. Timeout 8s. Resultado volta como
// connectedMsg ou connectFailedMsg.
func connectCmd(connStr string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		c, err := db.Connect(ctx, connStr)
		if err != nil {
			return connectFailedMsg{err: err}
		}
		return connectedMsg{client: c}
	}
}

func loadSchemasCmd(c *db.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		schemas, err := c.ListSchemas(ctx)
		return schemasLoadedMsg{schemas: schemas, err: err}
	}
}

func loadTablesCmd(c *db.Client, schema string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tables, err := c.ListTables(ctx, schema)
		return tablesLoadedMsg{schema: schema, tables: tables, err: err}
	}
}

func loadTableDetailsCmd(c *db.Client, schema, table string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cols, errC := c.ListColumns(ctx, schema, table)
		if errC != nil {
			return tableDetailsLoadedMsg{schema: schema, table: table, err: errC}
		}
		idx, errI := c.ListIndexes(ctx, schema, table)
		if errI != nil {
			return tableDetailsLoadedMsg{schema: schema, table: table, columns: cols, err: errI}
		}
		return tableDetailsLoadedMsg{schema: schema, table: table, columns: cols, indexes: idx}
	}
}

// execQueryCmd roda SQL ad-hoc do editor. 30s timeout — queries lentas
// ainda matam o request mas dão chance pra ANALYZE/explorações.
func execQueryCmd(c *db.Client, sql string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err := c.ExecQuery(ctx, sql)
		return queryExecutedMsg{result: result, err: err}
	}
}

// tableStatsCmd e activityCmd reusam queryExecutedMsg — UI trata igual
// a uma query do editor (mesmo render panel).
func tableStatsCmd(c *db.Client, schema, table string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := c.TableStats(ctx, schema, table)
		return queryExecutedMsg{result: result, err: err}
	}
}

func activityCmd(c *db.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := c.Activity(ctx)
		return queryExecutedMsg{result: result, err: err}
	}
}
