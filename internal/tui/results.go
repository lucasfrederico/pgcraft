package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/lucasfrederico/pgcraft/internal/db"
)

// resultsView mantém estado do scroll vertical na tabela de results.
// Lipgloss não tem tabela paginada built-in adequada — fizemos render
// manual com viewport scroll por linha.
type resultsView struct {
	result *db.QueryResult
	cursor int // first visible row index
}

func (r *resultsView) Set(qr *db.QueryResult) {
	r.result = qr
	r.cursor = 0
}

func (r *resultsView) Clear() {
	r.result = nil
	r.cursor = 0
}

func (r *resultsView) HasResult() bool { return r.result != nil }

func (r *resultsView) ScrollDown(maxVisible int) {
	if r.result == nil {
		return
	}
	if r.cursor < r.result.RowCount-maxVisible {
		r.cursor++
	}
}

func (r *resultsView) ScrollUp() {
	if r.cursor > 0 {
		r.cursor--
	}
}

// View renderiza a tabela com header + rows visíveis + status line.
// Layout: largura por coluna calculada uma vez (max de header + sample
// das primeiras 50 rows pra não correr toda a tabela).
func (r *resultsView) View(width, height int) string {
	if r.result == nil {
		return ""
	}

	colWidths := computeColWidths(r.result, width)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderHot).
		Padding(0, 1).
		Width(width).
		Height(height)

	var lines []string

	// header
	headerCells := make([]string, len(r.result.Cols))
	for i, c := range r.result.Cols {
		headerCells[i] = padOrTruncate(c, colWidths[i])
	}
	header := panelTitleStyle.Render(strings.Join(headerCells, " │ "))
	lines = append(lines, header)
	lines = append(lines, mutedStyle.Render(strings.Repeat("─", width-4)))

	// visible rows
	innerHeight := height - 5 // borders + header + sep + status
	if innerHeight < 1 {
		innerHeight = 1
	}
	end := r.cursor + innerHeight
	if end > r.result.RowCount {
		end = r.result.RowCount
	}
	for i := r.cursor; i < end; i++ {
		row := r.result.Rows[i]
		cells := make([]string, len(row))
		for j, v := range row {
			width := colWidths[j]
			if j >= len(colWidths) {
				width = 20
			}
			cells[j] = padOrTruncate(v, width)
		}
		lines = append(lines, strings.Join(cells, " │ "))
	}

	// pad to height
	for len(lines) < height-3 {
		lines = append(lines, "")
	}

	// status line: "1234 rows · 47ms · truncated at 1000"
	statusParts := []string{
		fmt.Sprintf("%d rows", r.result.RowCount),
		fmt.Sprintf("%dms", r.result.TimeMs),
	}
	if r.result.Truncated {
		statusParts = append(statusParts, fmt.Sprintf("truncated at %d", db.MaxRows))
	}
	if r.result.RowCount > innerHeight {
		statusParts = append(statusParts, fmt.Sprintf("rows %d-%d/%d", r.cursor+1, end, r.result.RowCount))
	}
	statusLine := mutedStyle.Render(strings.Join(statusParts, " · "))
	lines = append(lines, statusLine)

	return style.Render(strings.Join(lines, "\n"))
}

// computeColWidths usa max(header, sample first 50 rows) por coluna,
// depois distribui largura proporcional pro fit no width disponível.
func computeColWidths(r *db.QueryResult, totalWidth int) []int {
	n := len(r.Cols)
	if n == 0 {
		return nil
	}
	widths := make([]int, n)
	for i, c := range r.Cols {
		widths[i] = len(c)
	}
	sampleEnd := 50
	if sampleEnd > r.RowCount {
		sampleEnd = r.RowCount
	}
	for i := 0; i < sampleEnd; i++ {
		row := r.Rows[i]
		for j, v := range row {
			if j < n && len(v) > widths[j] {
				widths[j] = len(v)
			}
		}
	}
	// reserva espaço pra separadores (3 chars por sep entre N colunas)
	avail := totalWidth - 4 - (n-1)*3
	total := 0
	for _, w := range widths {
		total += w
	}
	if total <= avail {
		return widths
	}
	// shrink proporcionalmente
	for i := range widths {
		widths[i] = widths[i] * avail / total
		if widths[i] < 4 {
			widths[i] = 4
		}
	}
	return widths
}

func padOrTruncate(s string, width int) string {
	if width < 1 {
		return ""
	}
	if len(s) > width {
		if width <= 3 {
			return s[:width]
		}
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}
