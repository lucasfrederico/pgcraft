package tui

import (
	"github.com/lucasfrederico/pgcraft/internal/db"
)

// Mensagens Bubble Tea customizadas. Toda I/O de DB é async — disparada
// como tea.Cmd, resultado volta como uma dessas messages.

type connectedMsg struct {
	client *db.Client
}

type connectFailedMsg struct {
	err error
}

type schemasLoadedMsg struct {
	schemas []string
	err     error
}

type tablesLoadedMsg struct {
	schema string
	tables []string
	err    error
}

type tableDetailsLoadedMsg struct {
	schema  string
	table   string
	columns []db.Column
	indexes []db.Index
	err     error
}

// Phase 3: result de query ad-hoc do SQL editor.
type queryExecutedMsg struct {
	result *db.QueryResult
	err    error
}
