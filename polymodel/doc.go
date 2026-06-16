// Package polymodel is the reference Go implementation of PolyModel: the
// canonical in-memory model and (eventually) serialization and validation for
// PolyModel definitions.
//
// PolyModel defines canonical data models once and projects them into many
// storage, transport, and language targets. See https://polymodel.org.
//
// Core concepts (v0):
//
//   - Property   — an Entity attribute: the canonical, semantic definition
//   - Field      — a Collection attribute: loose, schemaless-capable
//   - Column     — a Recordset attribute: strict and tabular
//   - Component  — a reusable group of fields (Entities only in v0)
//   - Entity     — a logical, identity-bearing business object (the semantic anchor)
//   - Collection — a named, FROM-able data source (Kind: Editable | Computed)
//   - Recordset  — the tabular shape of a query / stored-procedure result
//
// v0 provides the in-memory model (model.go), HCL parsing (hcl.go, Decision D-0004),
// and structural + reference + constraint validation with located errors
// (validate.go). The query body is held behind an opaque seam (Decision D-0003).
//
// This package is experimental and pre-v0; the API will change as the spec settles.
package polymodel
