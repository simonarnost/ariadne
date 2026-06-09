# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Ariadne

A Zanzibar-style **ReBAC** (Relationship-Based Access Control) authorization service, written in Go.

## Status & commands

Bare scaffold: `go.mod` (module `github.com/simonarnost/ariadne`, Go 1.23) only — no Go source, proto, or tooling yet. The directory layout below is the *intended* target, not what exists. Standard Go workflow applies as packages land:

```
go build ./...
go test ./...
go test ./internal/engine/ -run TestCheck      # single test by name
go test ./internal/engine/ -run TestCheck/union # single table-driven subtest
go vet ./...
```

No proto-gen, lint, Docker, or DB-migration tooling is wired yet. When adding it (e.g. `protoc`/`buf` for `proto/`, a Postgres migration tool for `internal/storage/`), record the exact invocation here.

> *Ariadne gave Theseus the thread to find his way through the labyrinth.*
> A `Check` walks the relationship graph to find a path from a subject to a permission — the thread through the maze.

## What it does

Authorization as a **graph reachability** problem. Everything is a relation tuple:

```
object#relation@subject
document:readme#viewer@user:alice
document:readme#editor@group:eng#member
```

Two core operations:
- **Check** — can subject S do relation R on object O? (graph walk → verdict)
- **Expand** — who has relation R on object O? (userset subtree)

## Design decisions (locked)

- **Transport: gRPC.** Authz sits on the hot path of every request; low-latency backend-to-backend RPC + native server-streaming for the Watch API. Engine stays **transport-agnostic** (`Check(ctx, req)` Go methods); gRPC is a thin adapter. A grpc-gateway REST proxy can be added later if needed.
- **Store: Postgres** to start. Single tuples table, indexed by object and by subject.
- **Model: aim to be conceptually compatible with OpenFGA** so we can borrow their test cases — reference *after* attempting, not before.

## Architecture

```
gRPC API  (Check / Expand / Write / Read / Watch)   <- thin adapter
   │
Schema / IR  (parse DSL -> rewrite rules)
   │
Check engine (recursive graph walk, concurrent fan-out, cycle guard)
   │
Tuple store  (Postgres; indexes on object & subject)
```

### Schema rewrite rules that matter
- **union / computed_userset** — `editor` implies `viewer`.
- **tuple_to_userset** — "viewers of a folder are viewers of its documents" (hierarchies — the powerful one).

```
definition document {
  relation parent: folder
  relation editor: user
  relation viewer: user | group#member
  permission view = viewer + editor + parent->view   // tuple_to_userset
}
```

## Roadmap (milestones)

- **M1** — Tuple store + Write/Read API. Postgres table `(object_type, object_id, relation, subject_type, subject_id, subject_relation)`.
- **M2** — Direct Check (single-hop: exact tuple present?).
- **M3** — Schema DSL + rewrites: union + computed_userset; recursive Check.
- **M4** — tuple_to_userset (hierarchies) + cycle detection + concurrent fan-out (`errgroup`, bounded goroutines). **← MVP / showcase target (M1–M4).**
- **M5** — Expand (userset tree).
- **M6** — depth (pick any): zookies/consistency (the "new enemy" problem), Leopard-style transitive-closure index, check caching/memoization, Watch streaming API.

## Layout (intended)

```
cmd/ariadne/        main — wire config, store, engine, gRPC server
internal/engine/    transport-agnostic core: Check, Expand
internal/storage/   TupleStore interface + Postgres impl
internal/schema/    DSL parser -> rewrite-rule IR
proto/              .proto defs (Check/Expand/Write/Read/Watch) + generated code
```

## Conventions

- Engine exposes plain Go methods; no transport types leak into `internal/engine`.
- `context.Context` first arg everywhere; respect cancellation in graph walks.
- Errors wrapped with `%w`; sentinel errors for not-found / cycle.
- Table-driven tests; the check engine is the thing to test hardest.

## Learning resources
- **Zanzibar paper** (Pang et al., 2019) — read after attempting M1–M5.
- **OpenFGA** (CNCF) and **SpiceDB** — Go reference implementations; consult after a first attempt, don't copy.
