package tui

import (
	"fmt"
	"strings"
)

// EXPLAIN viz: parseamos saída texto de EXPLAIN ANALYZE e renderizamos
// com cost highlights. Phase 4 keep simples — texto colorizado.
// Phase 5 pode adicionar tree view interativo.

// formatExplainOutput pega rows do EXPLAIN (cada row é 1 col com "QUERY PLAN")
// e retorna string com indentação + colorização básica.
//
// Postgres EXPLAIN ANALYZE retorna várias linhas com indent crescente
// indicando children do plan. Já vem com ASCII art (->) bonito.
// O que adicionamos: highlight de "cost", "actual time", "rows", e cor
// pra Seq Scan vs Index Scan (heurística pra alertas perf).
func formatExplainOutput(rows [][]string) string {
	var b strings.Builder
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		line := row[0]
		b.WriteString(highlightExplainLine(line))
		b.WriteString("\n")
	}
	return b.String()
}

// highlightExplainLine aplica cor a partes da linha do plan.
// Heurística simples — palavras-chave problemáticas em vermelho/amarelo,
// times reais em ciano.
func highlightExplainLine(line string) string {
	out := line

	// alertas perf (vermelho)
	for _, alert := range []string{"Seq Scan", "Hash Join", "Sort"} {
		out = highlightSubstring(out, alert, "#ff8888")
	}

	// boas práticas (verde)
	for _, good := range []string{"Index Scan", "Index Only Scan", "Bitmap"} {
		out = highlightSubstring(out, good, "#88ff88")
	}

	// timing (ciano)
	for _, key := range []string{"actual time=", "actual rows=", "loops="} {
		out = highlightSubstring(out, key, "#88ddff")
	}

	return out
}

// highlightSubstring é wrapper simples — em Phase 4 keep ANSI direto
// pra não criar dependência circular com styles.go.
func highlightSubstring(s, sub, color string) string {
	if !strings.Contains(s, sub) {
		return s
	}
	colored := fmt.Sprintf("\x1b[38;2;%s;%s;%sm%s\x1b[0m", hex2(color[1:3]), hex2(color[3:5]), hex2(color[5:7]), sub)
	return strings.ReplaceAll(s, sub, colored)
}

func hex2(h string) string {
	v := 0
	for i := 0; i < len(h); i++ {
		c := h[i]
		v *= 16
		switch {
		case c >= '0' && c <= '9':
			v += int(c - '0')
		case c >= 'a' && c <= 'f':
			v += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v += int(c-'A') + 10
		}
	}
	return fmt.Sprintf("%d", v)
}
