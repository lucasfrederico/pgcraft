# pgcraft

> A lazygit-style TUI for PostgreSQL. Navigate schemas, run queries, view explain plans —
> without ever leaving the terminal.

🚧 **Work in progress.** Building in the open. README is a roadmap right now.

## Why?

I work in Postgres every day across a few projects (game backend at LoverCraft, microservices
at FlagShip, side projects). The existing tools are either:

- **psql** — terrific but no navigation, every operation is a `\d users` command from memory
- **pgcli** — REPL with autocomplete; better but still REPL-shaped
- **pgAdmin / DBeaver / TablePlus / DataGrip** — heavy GUIs, slow startup, some paid

`lazygit` proved how nice it is to navigate git visually inside a terminal. `k9s` did the
same for Kubernetes. `lazydocker` for Docker. **No one did it for Postgres.** `pgcraft` is
that tool.

## Planned UX (mockup)

```
┌─ Schemas ──────┬─ Tables: public ──────┬─ users ──────────────────────────┐
│ ▸ public       │ ▸ users               │ Columns                          │
│   audit        │   posts               │ ──────                           │
│   reporting    │   comments            │ id            bigserial PK       │
│                │   tenants             │ email         varchar(255) UNIQUE│
│                │   tags                │ tenant_id     bigint FK→tenants  │
│ Views          │                       │ created_at    timestamptz        │
│   active_users │ Views                 │                                  │
│                │   active_users        │ Indexes                          │
│ Functions      │                       │ ▸ users_pkey (id)                │
│   audit_log    │                       │   users_email_idx (email)        │
│                │                       │   users_tenant_idx (tenant_id)   │
│                │                       │                                  │
│                │                       ├─ Sample data (5 rows) ───────────┤
│                │                       │ id   email          name         │
│                │                       │ 1    alice@a.com    Alice        │
│                │                       │ 2    bob@a.com      Bob          │
└────────────────┴───────────────────────┴──────────────────────────────────┘
[j/k] nav   [enter] open   [s] SQL editor   [e] explain   [/] search   [?] help
```

## Roadmap

- [ ] **Phase 1 — TUI scaffold:** layout, panels, vim-like navigation
- [ ] **Phase 2 — Schema browser:** schemas → tables → columns/indexes/sample data
- [ ] **Phase 3 — SQL editor:** write queries, execute, paginated results
- [ ] **Phase 4 — Explain plan view:** EXPLAIN ANALYZE visualized
- [ ] **Phase 5 — Polish:** search/filter, multi-connection, theme, Docker image
- [ ] **Phase 6 — Launch:** README with demo gif, ship to HN

## Install (planned)

```bash
# Homebrew
brew install lucasfrederico/tap/pgcraft

# Go install
go install github.com/lucasfrederico/pgcraft@latest

# Docker
docker run -it ghcr.io/lucasfrederico/pgcraft "postgres://localhost/mydb"
```

## Usage (planned)

```bash
# Connect via URL
pgcraft "postgres://user:pass@host:5432/dbname"

# Or via env
DATABASE_URL=... pgcraft

# Multiple connections (cycle with Tab)
pgcraft "postgres://local/dev" "postgres://staging/db"
```

## Stack

- **Go** — single binary, cross-platform
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework from Charm
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — styling
- **[pgx](https://github.com/jackc/pgx)** — Postgres driver

## Inspired by

- [lazygit](https://github.com/jesseduffield/lazygit) — TUI for git (50k+ stars)
- [k9s](https://github.com/derailed/k9s) — TUI for Kubernetes (28k+ stars)
- [lazydocker](https://github.com/jesseduffield/lazydocker) — TUI for Docker (37k+ stars)

## License

MIT
