# Launch checklist & HN draft

Internal doc — not part of the product. Tracks Phase 6 launch.

## Pre-launch checklist

- [x] CI passing (build + test on Linux + macOS)
- [x] Docker multi-arch build (amd64 + arm64) pushed to GHCR on every push to main
- [x] Release workflow on tags (cross-compile binaries → GitHub release)
- [x] README has screenshots and keymap table
- [x] Roadmap up to date (Phase 1-6 marked done)
- [x] LICENSE present (MIT)
- [ ] Cut `v0.1.0` tag → triggers release workflow + first Docker tag
- [ ] vhs demo gif recorded and embedded in README (optional, soft launch first)
- [ ] Show HN post

## v0.1.0 release steps

```bash
# Verify everything builds clean
go vet ./...
gofmt -d .
go test ./...
go build -o pgcraft ./cmd/pgcraft

# Tag and push
git tag -a v0.1.0 -m "v0.1.0 — initial public release"
git push origin v0.1.0

# This triggers:
#   - .github/workflows/release.yml → cross-platform binaries on GitHub Releases
#   - .github/workflows/ci.yml      → docker build with version + latest tags on GHCR
```

## Show HN draft

**Title (under 80 chars):**

> Show HN: pgcraft – a lazygit-style TUI for Postgres

**Body:**

```
Hey HN,

I work in Postgres every day across a few backend projects and the existing tools
felt either too thin (psql, pgcli) or too heavy (pgAdmin, DBeaver, DataGrip).
lazygit / k9s / lazydocker proved how nice navigating a complex system can be from
the terminal — nobody had done that for Postgres specifically.

pgcraft is opinionated about that scope: Postgres-only by design. Going multi-DB
would mean abstracting away exactly the things that make Postgres worth using —
EXPLAIN plans, pg_stat_*, partitioning, JSONB ops, vector indexes. The roadmap
leans into those.

What's in v0.1.0:
- Three-panel browser (schemas → tables → columns/indexes)
- Vim-style keys (j/k/h/l, Tab cycles panels)
- SQL editor + paginated results (s to open, Ctrl+J to execute)
- EXPLAIN view (e)
- pg_stat_user_tables snapshot (T)
- pg_stat_activity live sessions (A)
- Live filter on any panel (/)
- ~15MB Docker image (multi-arch), single ~14MB Go binary

What's next:
- Partitioning view
- JSONB tree explorer
- Replication monitor
- EXPLAIN (ANALYZE, BUFFERS)
- Query bookmarks

It's a side project — feedback, issues, and PRs welcome.

Repo:  https://github.com/lucasfrederico/pgcraft
Docs:  README has screenshots + full keymap
Built with: Go + Bubble Tea + pgx v5

Happy to answer questions on design decisions (especially "why not multi-DB").
```

**First comment (self-reply) to seed discussion:**

> Quick FAQ:
>
> **Why not multi-DB like lazysql?** I covered this in the README but tl;dr: going
> multi-DB means I'd be abstracting away exactly the Postgres-specific stuff that
> motivated the tool. lazysql is great if you need CRUD across MySQL/PG/SQLite.
> pgcraft is for people who live in Postgres.
>
> **Why a TUI in 2026?** Same reason lazygit/k9s/lazydocker exist. Terminal-native
> tools are fast, scriptable, work over SSH, and don't break your context.
>
> **Read-only?** No, it'll run any SQL you type. There's no "are you sure?" yet for
> DDL/DML — that's on the roadmap.

## Marketing follow-ups

- [ ] Post in `r/PostgreSQL` (use "Show HN" body, adapt title)
- [ ] Tweet/Bsky from Charm community handle once vhs gif is ready
- [ ] Discord: post in #postgres on the Bubble Tea / Charm server
- [ ] charmbracelet/showcase PR (if accepting)
