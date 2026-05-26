package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// editor é um wrapper sobre bubbles/textarea com prompt + dimensões.
// Open=true significa que está em foco — UI captura todos keys pra ele.
type editor struct {
	ta   textarea.Model
	open bool
}

func newEditor() editor {
	ta := textarea.New()
	ta.Placeholder = "Type SQL here. Ctrl+J to execute. Esc to cancel."
	ta.CharLimit = 0 // no limit
	ta.ShowLineNumbers = true
	ta.Prompt = "│ "
	ta.SetWidth(80)
	ta.SetHeight(8)
	return editor{ta: ta}
}

func (e *editor) Focus() {
	e.open = true
	e.ta.Focus()
}

func (e *editor) Blur() {
	e.open = false
	e.ta.Blur()
}

func (e *editor) Value() string { return e.ta.Value() }
func (e *editor) Reset()        { e.ta.Reset() }

func (e *editor) SetSize(w, h int) {
	if w < 20 {
		w = 20
	}
	if h < 4 {
		h = 4
	}
	e.ta.SetWidth(w)
	e.ta.SetHeight(h)
}

// View renderiza com borda + título.
func (e *editor) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderHot).
		Padding(0, 1)

	title := panelTitleStyle.Render("SQL Editor") +
		mutedStyle.Render("   [Ctrl+J] execute   [Esc] cancel")

	return style.Render(title + "\n" + e.ta.View())
}
