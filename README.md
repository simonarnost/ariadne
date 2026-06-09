# Ariadne

A Zanzibar-style ReBAC (Relationship-Based Access Control) authorization service in Go.

A `Check` walks the relationship graph to find a path from a subject to a permission — Ariadne's thread through the labyrinth.

## Status

Early scaffold — tuple store and Read/Write in place (M1).

## The model

Authorization is a **graph reachability** problem. Every fact is a *relation tuple* — one edge in the graph. The two core operations walk that graph:

- **Check** — can subject S do relation R on object O? (walk for a path → allow/deny)
- **Expand** — who has relation R on object O? (the userset subtree)

## Relation tuples

A tuple is `object#relation@subject`:

```
document:readme#viewer@user:alice          alice is a viewer of readme
document:readme#editor@group:eng#member    members of group:eng are editors of readme
```

The **subject** comes in two flavors, and the difference is what makes this a graph rather than a flat list:

- **Concrete user** — `user:alice`. A specific individual.
- **Userset** — `group:eng#member`. A *pointer* to "everyone with `member` on `group:eng`", whose membership lives in other tuples (`group:eng#member@user:bob`, …). A Check follows that pointer — that's the walk.

## Storage contract

Tuples live in a `TupleStore`, queried by object (a relation's subjects) or by subject (an object's relations). `Write` applies deletes then inserts, atomically and idempotently — re-inserting an existing tuple or deleting an absent one is a no-op.

See [CLAUDE.md](./CLAUDE.md) for the schema rewrite rules (union, computed_userset, tuple_to_userset), design decisions, and the milestone roadmap.
