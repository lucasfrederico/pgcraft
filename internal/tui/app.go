// Package tui contém o modelo Bubble Tea raiz e os painéis.
//
// Phase 2: connection real ao Postgres via pgx + queries em information_schema.
// I/O é async via tea.Cmd; UI mostra "loading..." enquanto query roda.
// Vim-style nav (j/k/h/l), Tab cicla painéis, q sai.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lucasfrederico/pgcraft/internal/db"
)

// App é o modelo raiz Bubble Tea.
type App struct {
	connStr string
	client  *db.Client // populado após connectedMsg

	// dimensões da janela
	width  int
	height int

	// painéis e foco
	schemas panel
	tables  panel
	details panel
	active  panelID

	// estado de loading por painel — não bloqueia UI, mostra "loading..."
	loadingTables  bool
	loadingDetails bool

	// status bar
	statusMsg string
	errMsg    string // último erro a mostrar (vermelho)
}

// NewApp constrói o modelo raiz. Em Phase 2 sem connStr cai em mock mode
// (útil pra screenshots e dev sem DB local). Com connStr real dispara
// connectCmd no Init.
func NewApp(connStr string) *App {
	a := &App{
		connStr: connStr,
		active:  panelSchemas,
		schemas: panel{title: "Schemas"},
		tables:  panel{title: "Tables"},
		details: panel{title: "Details"},
	}

	if connStr == "" {
		// modo mock — pra dev sem Postgres rodando
		a.statusMsg = "no DATABASE_URL (mock mode)"
		a.schemas.items = []string{"public", "audit", "reporting", "auth"}
		a.tables.items = []string{"users", "posts", "comments", "tenants"}
		a.details.items = []string{
			"Columns",
			"  id            bigserial PK",
			"  email         varchar(255) UNIQUE",
			"  tenant_id     bigint FK → tenants",
			"  created_at    timestamptz",
		}
	} else {
		a.statusMsg = "connecting..."
	}
	return a
}

func (a *App) Init() tea.Cmd {
	if a.connStr != "" {
		return connectCmd(a.connStr)
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case connectedMsg:
		a.client = msg.client
		a.statusMsg = "connected · " + a.client.HostInfo()
		return a, loadSchemasCmd(a.client)

	case connectFailedMsg:
		a.errMsg = "connect failed: " + msg.err.Error()
		a.statusMsg = "disconnected"
		return a, nil

	case schemasLoadedMsg:
		if msg.err != nil {
			a.errMsg = "load schemas: " + msg.err.Error()
			return a, nil
		}
		a.schemas.items = msg.schemas
		a.schemas.cursor = 0
		a.tables.items = nil
		a.details.items = nil
		// auto-carrega tables do primeiro schema (UX default)
		if len(msg.schemas) > 0 && a.client != nil {
			a.loadingTables = true
			a.tables.title = "Tables (" + msg.schemas[0] + ")"
			return a, loadTablesCmd(a.client, msg.schemas[0])
		}
		return a, nil

	case tablesLoadedMsg:
		a.loadingTables = false
		if msg.err != nil {
			a.errMsg = "load tables: " + msg.err.Error()
			return a, nil
		}
		a.tables.items = msg.tables
		a.tables.cursor = 0
		a.tables.title = "Tables (" + msg.schema + ")"
		a.details.items = nil
		return a, nil

	case tableDetailsLoadedMsg:
		a.loadingDetails = false
		if msg.err != nil {
			a.errMsg = "load details: " + msg.err.Error()
			return a, nil
		}
		a.details.items = formatDetails(msg.columns, msg.indexes)
		a.details.cursor = 0
		a.details.viewport = 0
		a.details.title = "Details (" + msg.table + ")"
		return a, nil
	}
	return a, nil
}

// handleKey é a tabela de teclado. Phase 2 conecta Enter à query real.
func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	a.errMsg = "" // qualquer keypress limpa erro stale
	switch msg.String() {
	case "q", "ctrl+c":
		if a.client != nil {
			a.client.Close()
		}
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
		return a, a.openSelected()
	case "h", "left":
		a.active = a.active.prev()
	case "?":
		a.statusMsg = "[j/k] nav · [h/l/tab] cycle · [enter] open · [q] quit"
	case "r":
		// refresh — reaproveita Phase 2 já que é só re-disparar load
		if a.client != nil {
			return a, loadSchemasCmd(a.client)
		}
	}
	return a, nil
}

// openSelected é o "Enter" semântico: depende de qual painel está ativo.
// schemas → carrega tables; tables → carrega details; details → no-op
// (Phase 3 vai abrir sample data ou SQL editor).
func (a *App) openSelected() tea.Cmd {
	if a.client == nil {
		// mock mode: só rotaciona pra próximo painel
		a.active = a.active.next()
		return nil
	}
	switch a.active {
	case panelSchemas:
		schema := a.schemas.currentItem()
		if schema == "" {
			return nil
		}
		a.loadingTables = true
		a.tables.title = "Tables (" + schema + ")"
		a.active = panelTables
		return loadTablesCmd(a.client, schema)
	case panelTables:
		schema := a.schemas.currentItem()
		table := a.tables.currentItem()
		if schema == "" || table == "" {
			return nil
		}
		a.loadingDetails = true
		a.details.title = "Details (" + table + ")"
		a.active = panelDetails
		return loadTableDetailsCmd(a.client, schema, table)
	}
	return nil
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

// formatDetails transforma columns + indexes em []string que o painel renderiza.
// Manter formatado pra alinhar nomes/tipos visualmente.
func formatDetails(cols []db.Column, idxs []db.Index) []string {
	var out []string

	if len(cols) > 0 {
		out = append(out, "Columns")
		// calcula largura máx do nome pra alinhar
		maxNameLen := 0
		for _, c := range cols {
			if len(c.Name) > maxNameLen {
				maxNameLen = len(c.Name)
			}
		}
		for _, c := range cols {
			flags := ""
			if c.IsPK {
				flags = " PK"
			} else if !c.Nullable {
				flags = " NOT NULL"
			}
			padded := c.Name + strings.Repeat(" ", maxNameLen-len(c.Name))
			out = append(out, fmt.Sprintf("  %s  %s%s", padded, c.DataType, flags))
		}
		out = append(out, "")
	}

	if len(idxs) > 0 {
		out = append(out, "Indexes")
		for _, i := range idxs {
			marker := "  "
			if i.IsPrimary {
				marker = "* "
			}
			out = append(out, marker+i.Name)
		}
	}

	if len(out) == 0 {
		out = append(out, "(no columns or indexes)")
	}
	return out
}

// View renderiza o frame. 3 painéis + status bar + hint.
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "loading..."
	}

	bodyHeight := a.height - 2
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	wSchemas := a.width * 20 / 100
	wTables := a.width * 30 / 100
	wDetails := a.width - wSchemas - wTables - 6

	// loading overlays — itens do painel ficam "loading..." enquanto query roda
	schemasItems := a.schemas
	tablesItems := a.tables
	detailsItems := a.details
	if a.loadingTables && len(a.tables.items) == 0 {
		tablesItems.items = []string{"loading..."}
	}
	if a.loadingDetails && len(a.details.items) == 0 {
		detailsItems.items = []string{"loading..."}
	}

	schemasView := renderPanel(&schemasItems, a.active == panelSchemas, wSchemas, bodyHeight)
	tablesView := renderPanel(&tablesItems, a.active == panelTables, wTables, bodyHeight)
	detailsView := renderPanel(&detailsItems, a.active == panelDetails, wDetails, bodyHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, schemasView, tablesView, detailsView)

	hint := keyHintStyle.Render("[j/k]") + mutedStyle.Render(" nav   ") +
		keyHintStyle.Render("[tab]") + mutedStyle.Render(" cycle   ") +
		keyHintStyle.Render("[enter]") + mutedStyle.Render(" open   ") +
		keyHintStyle.Render("[r]") + mutedStyle.Render(" refresh   ") +
		keyHintStyle.Render("[q]") + mutedStyle.Render(" quit")

	statusContent := fmt.Sprintf("pgcraft · %s", a.statusMsg)
	if a.errMsg != "" {
		statusContent = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")).Render("✘ " + a.errMsg)
	}
	statusLine := statusBarStyle.Render(statusContent)

	return body + "\n" + hint + "\n" + statusLine
}

func (a *App) panelHeight() int {
	h := a.height - 4
	if h < 3 {
		return 3
	}
	return h
}

// renderPanel desenha um painel com borda, título e itens.
func renderPanel(p *panel, active bool, width, height int) string {
	style := panelStyle
	if active {
		style = activePanelStyle
	}
	style = style.Width(width).Height(height)

	var lines []string
	lines = append(lines, panelTitleStyle.Render(p.title))
	lines = append(lines, mutedStyle.Render(strings.Repeat("─", width-4)))

	innerHeight := height - 3
	if innerHeight < 1 {
		innerHeight = 1
	}

	visibleEnd := p.viewport + innerHeight
	if visibleEnd > len(p.items) {
		visibleEnd = len(p.items)
	}

	for i := p.viewport; i < visibleEnd; i++ {
		line := p.items[i]
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

	for len(lines) < height-1 {
		lines = append(lines, "")
	}

	return style.Render(strings.Join(lines, "\n"))
}
