// Command pgcraft starts the TUI.
//
// Usage:
//
//	pgcraft "postgres://user:pass@host:5432/dbname"
//	DATABASE_URL=... pgcraft
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lucasfrederico/pgcraft/internal/tui"
)

func main() {
	// Phase 1: pure TUI scaffold (sem DB connection ainda).
	// Phase 2 vai pegar connection string via argv ou DATABASE_URL.
	connStr := pickConnString()

	p := tea.NewProgram(
		tui.NewApp(connStr),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pgcraft: %v\n", err)
		os.Exit(1)
	}
}

func pickConnString() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return os.Getenv("DATABASE_URL")
}
