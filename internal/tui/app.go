// Package tui contém o modelo Bubble Tea raiz e os painéis.
//
// Phase 1: scaffold com 3 painéis (schemas/tables/details) populados com
// mock data. Vim-style nav (j/k/h/l), Tab cicla painéis, q sai.
// Phase 2 vai trocar mocks por dados reais do Postgres.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App é o modelo raiz Bubble Tea.
type App struct {
	connStr string

	// dimensões da janela (atualizadas via tea.WindowSizeMsg)
	width  int
	height int

	// painéis e foco
	schemas panel
	tables  panel
	details panel
	active  panelID

	// status bar message (transitory, ex: "connected to dev DB")
	statusMsg string
}

// NewApp constrói o modelo raiz. connStr ainda não é usada em Phase 1
// (mocks). Phase 2 vai abrir pool aqui.
func NewApp(connStr string) *App {
	a := &App{
		connStr: connStr,
		active:  panelSchemas,
		schemas: panel{
			title: "Schemas",
			items: []string{"public", "audit", "reporting", "auth"},
		},
		tables: panel{
			title: "Tables (public)",
			items: []string{"users", "posts", "comments", "tenants", "tags", "audit_log"},
		},
		details: panel{
			title: "Details (users)",
			items: []string{
				"Columns",
				"  id            bigserial PK",
				"  email         varchar(255) UNIQUE",
				"  tenant_id     bigint FK → tenants",
				"  created_at    timestamptz",
				"",
				"Indexes",
				"  users_pkey (id)",
				"  users_email_idx (email)",
				"  users_tenant_idx (tenant_id)",
			},
		},
	}
	if connStr == "" {
		a.statusMsg = "no connection (Phase 1 mock mode)"
	} else {
		a.statusMsg = "connection string loaded (real DB in Phase 2)"
	}
	return a
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)
	}
	return a, nil
}

// handleKey é a tabela de teclado. Vim-style nav + Tab/Shift+Tab pra
// cycle entre painéis. Mantida pequena por enquanto; Phase 3 vai
// adicionar SQL editor mode com keymap diferente.
func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "tab":
		a.active = a.active.next()
	case "shift+tab":
		a.active = a.active.prev()
	case "j", "down":
		a.activePanel().moveDown(a.panelHeight())
	case "k", "up":
		a.activePanel().moveUp()
	case "l", "right", "enter":
		// Phase 1: enter no schema "abre" tables; enter na table "abre" details.
		// Tudo mock por enquanto. Phase 2 dispara query.
		if a.active == panelSchemas {
			a.active = panelTables
		} else if a.active == panelTables {
			a.active = panelDetails
		}
	case "h", "left":
		a.active = a.active.prev()
	case "?":
		a.statusMsg = "[j/k] nav  [h/l/tab] cycle panels  [enter] open  [q] quit"
	}
	return a, nil
}

func (a *App) activePanel() *panel {
	switch a.active {
	case panelSchemas:
		return &a.schemas
	case panelTables:
		return &a.tables
	case panelDetails:
		return &a.details
	}
	return &a.schemas
}

// View renderiza o frame. Três painéis lado a lado + status bar.
// Layout: 20% / 30% / 50% da largura. Altura = window - 2 (status bar).
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "loading..."
	}

	bodyHeight := a.height - 2 // 1 linha status + 1 hint
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	wSchemas := a.width * 20 / 100
	wTables := a.width * 30 / 100
	wDetails := a.width - wSchemas - wTables - 6 // 6 = bordas/padding

	schemasView := renderPanel(&a.schemas, a.active == panelSchemas, wSchemas, bodyHeight)
	tablesView := renderPanel(&a.tables, a.active == panelTables, wTables, bodyHeight)
	detailsView := renderPanel(&a.details, a.active == panelDetails, wDetails, bodyHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, schemasView, tablesView, detailsView)

	hint := keyHintStyle.Render("[j/k]") + mutedStyle.Render(" nav   ") +
		keyHintStyle.Render("[tab]") + mutedStyle.Render(" cycle panels   ") +
		keyHintStyle.Render("[enter]") + mutedStyle.Render(" open   ") +
		keyHintStyle.Render("[?]") + mutedStyle.Render(" help   ") +
		keyHintStyle.Render("[q]") + mutedStyle.Render(" quit")

	statusLine := statusBarStyle.Render(fmt.Sprintf(
		"pgcraft · active: %s · %s",
		a.active.String(),
		a.statusMsg,
	))

	return body + "\n" + hint + "\n" + statusLine
}

func (a *App) panelHeight() int {
	h := a.height - 4
	if h < 3 {
		return 3
	}
	return h
}

// renderPanel desenha um painel com borda, título, e itens com cursor.
func renderPanel(p *panel, active bool, width, height int) string {
	style := panelStyle
	if active {
		style = activePanelStyle
	}
	style = style.Width(width).Height(height)

	var lines []string
	lines = append(lines, panelTitleStyle.Render(p.title))
	lines = append(lines, mutedStyle.Render(strings.Repeat("─", width-4)))

	innerHeight := height - 3 // header + sep + bottom border
	if innerHeight < 1 {
		innerHeight = 1
	}

	visibleEnd := p.viewport + innerHeight
	if visibleEnd > len(p.items) {
		visibleEnd = len(p.items)
	}

	for i := p.viewport; i < visibleEnd; i++ {
		line := p.items[i]
		// truncate se passar da largura
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		if i == p.cursor && active {
			lines = append(lines, selectedItemStyle.Render("▸ "+line))
		} else if i == p.cursor {
			lines = append(lines, itemStyle.Render("▸ "+line))
		} else {
			lines = append(lines, itemStyle.Render("  "+line))
		}
	}

	// pad to height
	for len(lines) < height-1 {
		lines = append(lines, "")
	}

	return style.Render(strings.Join(lines, "\n"))
}
