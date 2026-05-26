package tui

// panelID enumera os painéis da UI. Em Phase 1 são 3:
// Schemas → Tables → Columns/Indexes/Sample (tabs no 3º painel virão Phase 3).
type panelID int

const (
	panelSchemas panelID = iota
	panelTables
	panelDetails
)

// String é útil pra footer ("Active: tables") e debug.
func (p panelID) String() string {
	switch p {
	case panelSchemas:
		return "schemas"
	case panelTables:
		return "tables"
	case panelDetails:
		return "details"
	}
	return "?"
}

// next/prev rotacionam o foco entre painéis (Tab / Shift+Tab).
func (p panelID) next() panelID {
	if p == panelDetails {
		return panelSchemas
	}
	return p + 1
}

func (p panelID) prev() panelID {
	if p == panelSchemas {
		return panelDetails
	}
	return p - 1
}

// panel é um helper simples: lista de items, item selecionado.
// Estado é mantido por painel (cada painel tem seu cursor).
type panel struct {
	title    string
	items    []string
	cursor   int
	viewport int // primeira linha visível (pra scroll quando lista é maior que altura)
}

func (p *panel) moveDown(maxVisible int) {
	if p.cursor < len(p.items)-1 {
		p.cursor++
	}
	// scroll se cursor saiu da viewport
	if p.cursor >= p.viewport+maxVisible {
		p.viewport = p.cursor - maxVisible + 1
	}
}

func (p *panel) moveUp() {
	if p.cursor > 0 {
		p.cursor--
	}
	if p.cursor < p.viewport {
		p.viewport = p.cursor
	}
}

func (p *panel) currentItem() string {
	if p.cursor >= 0 && p.cursor < len(p.items) {
		return p.items[p.cursor]
	}
	return ""
}
