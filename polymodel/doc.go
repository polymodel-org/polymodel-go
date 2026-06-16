// Package polymodel is the reference Go implementation of PolyModel: the
// canonical in-memory model and (eventually) serialization and validation for
// PolyModel definitions.
//
// PolyModel defines canonical data models once and projects them into many
// storage, transport, and language targets. See https://polymodel.org.
//
// Planned core concepts (see ROADMAP.md; NOT yet implemented):
//
//   - Field      — atomic typed value with constraints
//   - Component  — reusable group of fields (composition; no identity)
//   - Entity     — logical, identity-bearing business object (the semantic anchor)
//   - Recordset  — a row shape (query / stored-procedure output); no storage, no identity
//   - Collection — a persistent container of records: a Table or a View
//
// This package is experimental and pre-v0. The API will change as the spec and
// the core concept model settle.
package polymodel
