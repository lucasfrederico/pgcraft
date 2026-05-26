# pgcraft

> A lazygit-style TUI for PostgreSQL. Navigate schemas, run queries, view explain plans —
> without ever leaving the terminal.

🚧 **Work in progress.** Building in the open. Phase 3 done (SQL editor + results).

## Why?

I work in Postgres every day across a few projects (game backend at LoverCraft, microservices
at FlagShip, side projects). The existing tools are either:

- **psql** — terrific but no navigation, every operation is a `\d users` command from memory
- **pgcli** — REPL with autocomplete; better but still REPL-shaped
- **pgAdmin / DBeaver / TablePlus / DataGrip** — heavy GUIs, slow startup, some paid
- **lazysql / Sqlit** — multi-DB but shallow on each one; can't surface Postgres-specific stuff

`lazygit` proved how nice it is to navigate git visually inside a terminal. `k9s` did the
same for Kubernetes. `lazydocker` for Docker. **No one did it for Postgres specifically.**
`pgcraft` is that tool.

## Postgres-only by design

This is **not** a generic SQL TUI. Going multi-DB means either shallow features or
abstracting away exactly the parts that make Postgres worth using —
`EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)`, `pg_stat_*`, partitioning, logical replication,
JSONB ops, vector indexes. The roadmap leans into those.

If you need multi-DB CRUD, `lazysql` is the right tool. `pgcraft` is for people who
actually live in Postgres.

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

- [x] **Phase 1 — TUI scaffold:** layout, panels, vim-like navigation
- [x] **Phase 2 — Schema browser:** schemas → tables → columns/indexes
- [x] **Phase 3 — SQL editor:** write queries, execute, paginated results
- [ ] **Phase 4 — EXPLAIN view + search:** EXPLAIN ANALYZE visualized, table filter, Docker image
- [ ] **Phase 5 — Postgres-deep features:** `pg_stat_*` dashboard, partitioning view, JSONB helper, replication monitor
- [ ] **Phase 6 — Launch:** README demo gif, HN announcement

## Usage today

```bash
git clone https://github.com/lucasfrederico/pgcraft.git
cd pgcraft
go build -o pgcraft ./cmd/pgcraft

# Connect via URL
./pgcraft "postgres://user:pass@host:5432/dbname"

# Or via env
DATABASE_URL="postgres://localhost/mydb" ./pgcraft
```

Keys:
- `j/k` navigate · `h/l` / `Tab` cycle panels · `Enter` open
- `s` open SQL editor · `Ctrl+J` execute · `Esc` cancel
- `esc/x` close results · `r` refresh · `q` quit

## Install (planned for Phase 6 launch)

```bash
brew install lucasfrederico/tap/pgcraft
go install github.com/lucasfrederico/pgcraft@latest
docker run -it ghcr.io/lucasfrederico/pgcraft "postgres://..."
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
