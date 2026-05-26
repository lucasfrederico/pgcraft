# pgcraft

> A lazygit-style TUI for PostgreSQL. Navigate schemas, run queries, view explain plans,
> peek at `pg_stat_*` — without ever leaving the terminal.

[![Go](https://github.com/lucasfrederico/pgcraft/actions/workflows/ci.yml/badge.svg)](https://github.com/lucasfrederico/pgcraft/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/lucasfrederico/pgcraft)](https://goreportcard.com/report/github.com/lucasfrederico/pgcraft)

## Why?

I work in Postgres every day across a few projects (game backend, microservices, side
projects). The existing tools are either:

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

## Screenshots

### Three-panel browser (schemas / tables / details)

```
┌─ Schemas ──────┬─ Tables: public ──────┬─ user ───────────────────────────┐
│ ▸ public       │   audit_log           │ Columns                          │
│   pg_catalog   │   auth_group          │ ──────                           │
│   information… │   bid                 │ id            bigint PK          │
│                │   django_admin_log    │ email         varchar(254) UNIQUE│
│                │   django_migrations   │ tenant_id     bigint FK→tenant   │
│                │   tenant              │ password      varchar(128)       │
│                │ ▸ user                │ is_active     bool               │
│                │   user_groups         │ date_joined   timestamptz        │
│                │                       │                                  │
│                │                       │ Indexes                          │
│                │                       │ ▸ user_pkey (id)                 │
│                │                       │   user_email_idx (email) UNIQUE  │
│                │                       │   user_tenant_idx (tenant_id)    │
└────────────────┴───────────────────────┴──────────────────────────────────┘
[j/k] nav  [h/l/Tab] cycle  [s] SQL  [e] explain  [T] stats  [A] activity  [/] filter
```

### SQL editor + paginated results (s, Ctrl+J)

```
┌─ SQL Editor ────────────────────────────────────────────────────────────┐
│ SELECT id, email, is_active                                             │
│ FROM "user"                                                             │
│ WHERE tenant_id = 1                                                     │
│ ORDER BY date_joined DESC                                               │
│ LIMIT 50;                                                               │
└─────────────────────────────────────────────────────────────────────────┘
 [Ctrl+J] execute · [Esc] cancel · query ok · 3 rows · 5ms

┌─ Results ────────────────────────────────────────────────────────────────┐
│ id │ email          │ is_active                                          │
│ ───────────────────────────────────────                                  │
│ 3  │ carl@flag.com  │ true                                               │
│ 2  │ bob@a.com      │ true                                               │
│ 1  │ alice@a.com    │ true                                               │
└──────────────────────────────────────────────────────────────────────────┘
```

### EXPLAIN view (e)

```
┌─ Results (EXPLAIN) ──────────────────────────────────────────────────────┐
│ QUERY PLAN                                                               │
│ ─────────                                                                │
│ Seq Scan on "user"  (cost=0.00..10.60 rows=60 width=1143)                │
│   Filter: (tenant_id = 1)                                                │
└──────────────────────────────────────────────────────────────────────────┘
 1 row · 8ms
```

### `pg_stat_user_tables` (T)

```
┌─ Results (TableStats) ────────────────────────────────────────────────────┐
│ table        │ live │ dead │ ins │ upd │ del │ last_vacuum               │
│ ─────────────────────────────────────────────────────────────────────     │
│ public.user  │ 3    │ 0    │ 3   │ 0   │ 0   │ 2026-05-25 12:01:33+00    │
└───────────────────────────────────────────────────────────────────────────┘
 1 row · 4ms
```

### `pg_stat_activity` (A)

```
┌─ Results (Activity) ─────────────────────────────────────────────────────┐
│ pid   │ user     │ app  │ state  │ wait     │ duration │ query           │
│ ────────────────────────────────────────────────────────────────────     │
│ 44367 │ dmt_user │ psql │ active │ PgSleep  │ 3s       │ SELECT pg_…     │
│ 44380 │ dmt_user │      │ idle   │ ClientR… │ NULL     │                 │
└──────────────────────────────────────────────────────────────────────────┘
 2 rows · 2ms
```

> Want a full animated demo? See [`vhs/demo.tape`](vhs/demo.tape) — install
> [vhs](https://github.com/charmbracelet/vhs) and run `vhs vhs/demo.tape` to regenerate.

## Roadmap

- [x] **Phase 1 — TUI scaffold:** layout, panels, vim-like navigation
- [x] **Phase 2 — Schema browser:** schemas → tables → columns/indexes
- [x] **Phase 3 — SQL editor:** write queries, execute, paginated results
- [x] **Phase 4 — EXPLAIN + search + Docker:** EXPLAIN view, live filter (`/`), multi-arch Docker
- [x] **Phase 5 — Postgres-deep features:** `pg_stat_user_tables`, `pg_stat_activity`
- [x] **Phase 6 — Launch:** CI, GHCR Docker push, screenshots, HN announcement
- [ ] Future: partitioning view, JSONB tree explorer, replication monitor, `EXPLAIN (ANALYZE, BUFFERS)`, query bookmarks

## Quick start

```bash
git clone https://github.com/lucasfrederico/pgcraft.git
cd pgcraft
go build -o pgcraft ./cmd/pgcraft

# Connect via URL
./pgcraft "postgres://user:pass@host:5432/dbname"

# Or via env
DATABASE_URL="postgres://localhost/mydb" ./pgcraft
```

## Keys

| Key | Action |
|-----|--------|
| `j` / `k` | navigate up/down |
| `h` / `l` / `Tab` | cycle panels |
| `Enter` | open selection |
| `/` | filter current panel (live) |
| `Esc` | close filter / cancel |
| `s` | open SQL editor |
| `Ctrl+J` | execute SQL |
| `e` | `EXPLAIN` selected table |
| `T` | `pg_stat_user_tables` for selected table |
| `A` | `pg_stat_activity` (live sessions) |
| `x` / `Esc` | close results panel |
| `r` | refresh |
| `q` | quit |

## Docker

```bash
# From source
docker build -t pgcraft .
docker run -it --network host pgcraft "postgres://localhost/mydb"

# From GHCR (after Phase 6 release)
docker run -it --network host ghcr.io/lucasfrederico/pgcraft:latest "postgres://..."
```

Multi-stage build → ~15MB final image (alpine + static binary).
ARM64 and amd64 supported via `docker buildx`.

## Install

```bash
# Go install (any platform)
go install github.com/lucasfrederico/pgcraft/cmd/pgcraft@latest

# Homebrew (planned)
brew install lucasfrederico/tap/pgcraft

# Docker
docker pull ghcr.io/lucasfrederico/pgcraft:latest
```

## Stack

- **Go** — single binary, cross-platform
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework from Charm
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — styling
- **[pgx v5](https://github.com/jackc/pgx)** — Postgres driver

## Inspired by

- [lazygit](https://github.com/jesseduffield/lazygit) — TUI for git (50k+ stars)
- [k9s](https://github.com/derailed/k9s) — TUI for Kubernetes (28k+ stars)
- [lazydocker](https://github.com/jesseduffield/lazydocker) — TUI for Docker (37k+ stars)

## Contributing

PRs welcome. Run tests with `go test ./...`, lint with `go vet ./...` and `gofmt -d .`
before opening a PR.

## License

MIT — see [LICENSE](LICENSE).
