package tui

import "github.com/charmbracelet/lipgloss"

// Palette — minimal, dark-mode-first.
// Tom-de-cinza pra UI chrome, accent ciano pra foco/seleção.
var (
	colorAccent     = lipgloss.Color("#5fafff") // ciano claro
	colorAccentDim  = lipgloss.Color("#3070a0")
	colorMuted      = lipgloss.Color("#888888")
	colorBorder     = lipgloss.Color("#3a3a3a")
	colorBorderHot  = lipgloss.Color("#5fafff")
	colorBg         = lipgloss.Color("#1a1a1a")
	colorFg         = lipgloss.Color("#e6e6e6")
	colorSelectedBg = lipgloss.Color("#2a3a4a")
)

// Panel styles. activePanelStyle desenha borda destacada quando o painel
// está com foco.
var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePanelStyle = panelStyle.
				BorderForeground(colorBorderHot)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	itemStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Background(colorSelectedBg).
				Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)
)
