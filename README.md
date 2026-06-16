# polymodel-go

Reference Go implementation of [PolyModel](https://polymodel.org) — the canonical
in-memory model, serialization, and validation for PolyModel definitions.

> **Status: experimental scaffold (pre-v0).** No model code yet. The core concept
> model is still being settled; types and APIs will change. See
> [ROADMAP.md](ROADMAP.md).

## What is PolyModel?

PolyModel is an open specification for defining canonical data models once and
projecting them into many storage, transport, and language targets (PostgreSQL,
SQLite, Firestore, OpenAPI, Go, TypeScript, and more). Spec repo:
<https://github.com/polymodel-org/polymodel>.

This package is the Go library other tools build on — notably the
[`polymodel-cli`](https://github.com/polymodel-org/polymodel-cli) validator and
ecosystem generators (e.g. inGitDB schema generation).

## Planned layout

```
polymodel/      core model types: Field, Component, Entity, Recordset, Collection
                (+ serialization and validation)
```

## Install

```sh
go get github.com/polymodel-org/polymodel-go@latest
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
