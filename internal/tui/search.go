package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// search é uma bar de filtro inline. Quando aberta, captura keys
// (similar ao editor mas single-line). Filtra o painel ativo.
type search struct {
	ti   textinput.Model
	open bool
}

func newSearch() search {
	ti := textinput.New()
	ti.Placeholder = "filter… (Esc to close)"
	ti.CharLimit = 64
	ti.Prompt = "/ "
	return search{ti: ti}
}

func (s *search) Focus() {
	s.open = true
	s.ti.Focus()
	s.ti.SetValue("")
}

func (s *search) Blur() {
	s.open = false
	s.ti.Blur()
	s.ti.SetValue("")
}

func (s *search) Query() string {
	return strings.TrimSpace(strings.ToLower(s.ti.Value()))
}

// filterItems aplica search à lista. Case-insensitive substring match.
// Retorna lista original se search vazia.
func filterItems(items []string, query string) []string {
	if query == "" {
		return items
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(it), query) {
			out = append(out, it)
		}
	}
	return out
}

// View renderiza bar de search no top do painel ativo (single-line).
func (s *search) View(width int) string {
	if !s.open {
		return ""
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderHot).
		Padding(0, 1).
		Width(width)
	return style.Render(s.ti.View())
}
