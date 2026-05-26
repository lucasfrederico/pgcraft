// Package db wraps pgx pool + queries em information_schema/pg_catalog.
//
// Tudo aqui é puro Go — não conhece Bubble Tea. A UI envelopa estas
// chamadas em tea.Cmd async pra não travar o frame.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client é um wrapper fino sobre pgxpool. Phase 2 fica com queries
// síncronas simples; futuro pode adicionar prepared statements + cache.
type Client struct {
	pool *pgxpool.Pool
	url  string // mantido só pra exibir host/db no status bar
}

// Connect abre o pool e dá ping. Timeout curto pra não travar UI;
// erros voltam pra UI mostrar mensagem.
func Connect(ctx context.Context, connStr string) (*Client, error) {
	if connStr == "" {
		return nil, fmt.Errorf("empty connection string (set DATABASE_URL or pass as argv)")
	}

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse conn string: %w", err)
	}
	// 2 conexões basta pra TUI single-user
	cfg.MaxConns = 2
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Client{pool: pool, url: connStr}, nil
}

func (c *Client) Close() {
	if c != nil && c.pool != nil {
		c.pool.Close()
	}
}

// HostInfo retorna "user@host:port/db" sem senha — pra status bar.
// Truncado se ficar grande; é decorativo, não autoritativo.
func (c *Client) HostInfo() string {
	if c == nil || c.pool == nil {
		return ""
	}
	cfg := c.pool.Config().ConnConfig
	out := fmt.Sprintf("%s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.Database)
	if len(out) > 60 {
		out = out[:57] + "..."
	}
	return out
}
