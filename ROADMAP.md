# polymodel-go Roadmap

Pre-v0 scaffold. This document tracks the MVP scope and the decisions that must be
settled before the core model is implemented.

## MVP scope (v0)

1. **Core model types** for the five concepts:
   - `Field` — name, type, constraints (`required`, length bounds, `pattern`, `enum`,
     references).
   - `Component` — reusable named group of fields.
   - `Entity` — identity-bearing business object composed of fields/components.
   - `Recordset` — ordered columns + optional keys; the shape of a query or
     stored-procedure output. No storage, no identity.
   - `Collection` — persistent container of records: `Table` or `View` (a View's
     shape is a Recordset).
2. **Serialization** — load/save a PolyModel model from a canonical YAML/JSON
   representation (the authored HCL form, if kept, compiles to this).
3. **Validation** — structural validation with clear, located errors (the capability
   the CLI and CI gate on).
4. **Resolution** — resolve references between entities, components, collections, and
   recordsets (and, later, catalog references).

## Decisions to settle before implementing the model

These come from the PolyModel concept work and directly shape the Go types:

- **Terminology** — `field` vs `column` vs `property` across Entity / Collection /
  Recordset. Pick one canonical term (or define `column` = a field in a
  recordset/collection context).
- **Recordset vs Collection** — model `Recordset` as the base "shape" type that a
  `Collection` *has* and a `View` *is*, to avoid duplicating the column model.
- **Collection schema source** — allow defining a collection's shape by entity
  reference, component composition, and/or inline fields.
- **Keys & identity** — Entity has identity; Collection has keys; Recordset may have
  keys but no identity. Confirm and encode.
- **View / Recordset query language** — reference DTQL (a dialect-agnostic query AST)
  for view/recordset definitions, or define a PolyModel-native one.
- **Serialization** — confirm the canonical YAML/JSON shape; whether HCL remains the
  authored form with YAML/JSON as the compiled artifact.

## Non-goals (for now)

- Code generators for specific targets (those live in their own repos/tools, e.g.
  inGitDB schema generation lives in inGitDB).
- A migration engine.
