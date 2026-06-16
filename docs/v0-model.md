# polymodel-go v0 Model Design

Date: 2026-06-16
Status: Approved design — basis for the v0 implementation.

This document defines the core model `polymodel-go` implements for v0: the in-memory
types for a PolyModel definition, how they relate, and the decisions behind them.

## Purpose

PolyModel defines canonical data models once and projects them into many storage,
transport, and language targets. A recurring problem (visible in DataTug) is
**tangling the semantic model with the physical schema**. The v0 model keeps three
concerns explicitly separate:

- **Entity** — what a thing *is* (semantic).
- **Collection** — a named, queryable data source that holds records (table or view).
- **Recordset** — the shape of a *result* (a query or stored-procedure output).

## Attribute types: property, field, column

Each layer has its own attribute type, named for what that layer actually is. They
are deliberately distinct, not one type reused:

| Type | Belongs to | Character |
|---|---|---|
| **`Property`** | Entity | The canonical, *semantic* definition of an attribute. Owns identity participation, authoritative constraints, and true relationships. Everything else binds back to it. |
| **`Field`** | Collection | A *loose, schemaless-capable* attribute. A record may carry more or fewer fields; a field may be optional. Suits document / Git / Firestore collections, not just rigid tables. |
| **`Column`** | Recordset | A *strict tabular* attribute. Every row of a recordset has the same columns. A column's type may itself be a document/JSON value. |

### Why three, not one

The cleavage is **semantic vs tabular vs loose-container**:

- `column` (RDBMS connotation) implies a fixed schema every record conforms to —
  wrong for a **Collection**, which can be schemaless or hold heterogeneous
  documents. `field` is the fluent, less-restrictive term.
- A **Recordset** is tabular *by definition* (a result set); `column` is precise, and
  using a different word from `field` keeps "loose container" and "tabular result"
  visibly distinct.
- An **Entity** attribute is the only one that is canonical/authoritative — it carries
  identity and true constraints; the others *bind* to it. `property` names that
  semantic role.

## Structural types

### Entity

A logical, identity-bearing business object; the semantic anchor.

- `Properties []Property`
- `Components []ComponentRef` — composition (reusable field groups)
- `Key []string` — identity

A `Property` may type to a primitive, a component, or an **association to another
entity** (a relationship), and may reuse another property's definition (`Ref`).

### Collection

A **named, addressable, FROM-able data source** that holds records. To a consumer,
`SELECT FROM` a view is indistinguishable from a table, so both are Collections with a
**uniform `field` model**.

- `Kind CollectionKind` — `Editable` (tables / FS collections) | `Computed` (views).
  Partly derivable (a collection with a query is computed) but kept explicit for
  clarity.
- `Fields []Field` — uniform across both kinds.
- `Query *dtql.Query` — present when `Kind == Computed` (the view definition).
- `Storage *StorageMapping` — optional per-target storage hints.

### Recordset

The **shape of a result** — a query result or **stored-procedure / function output**.
It is *not* a FROM-able named source; it describes output `column`s.

- `Columns []Column`
- `Keys *Keys` — optional logical key(s)
- `Query *dtql.Query` — what produces the rows (when known)

### Collection (Computed/View) vs Recordset — the line

Both involve a query, so the distinction matters:

- A **Computed Collection (view)** is a *persistent, named data source* you query like
  a table → uniform **fields**.
- A **Recordset** is the *shape of a result* (e.g. a stored procedure's output) → it
  has **columns** and is not itself queried `FROM`.

## Binding and references

- **Binding:** a `Field` or `Column` may bind to an entity `Property`
  (e.g. `@pm:Order.currency`) to inherit its semantics. Both can also exist with no
  entity behind them (join keys, technical fields, computed/aggregated columns).
- **Relationships / FKs — three flavors, intentionally different:**
  - Entity `Property` → **association** to another entity (semantic relationship).
  - Collection `Field` → **enforced** foreign key to another collection.
  - Recordset `Column` → **soft / navigational** reference (for lookups), not enforced.

## Type sketch (illustrative)

```go
// ---- attributes ----
type Property struct {
    Name        string
    Type        TypeRef
    Title       string
    Constraints Constraints   // required, unique, length, pattern, enum
    Ref         *PropertyRef  // reuse/inherit another property's definition
}

type Field struct {           // collection attribute — loose, schemaless-capable
    Name       string
    Type       TypeRef
    Property   *PropertyRef    // optional bind to an entity property
    Optional   bool            // may be absent on a record
    ForeignKey *Reference      // enforced
}

type Column struct {          // recordset attribute — strict tabular
    Name     string
    Type     TypeRef           // may be a document/JSON type
    Property *PropertyRef       // optional semantic bind
    Source   *dtql.Expr         // produced-by expression
    Ref      *Reference         // soft navigational reference
}

// ---- structures ----
type Entity struct {
    Name       string
    Properties []Property
    Components []ComponentRef
    Key        []string
}

type CollectionKind int
const ( Editable CollectionKind = iota; Computed )

type Collection struct {
    Name    string
    Kind    CollectionKind
    Fields  []Field
    Query   *dtql.Query         // when Kind == Computed
    Storage *StorageMapping
}

type Recordset struct {
    Name    string
    Columns []Column
    Keys    *Keys
    Query   *dtql.Query
}

// ---- supporting ----
type Component struct { Name string; Fields []Property } // reusable group
type TypeRef   struct {
    Primitive string             // string,int,decimal,bool,uuid,timestamp,date,...
    Component *ComponentRef       // embed a reusable group
    Entity    *EntityRef          // association to another entity
}
type Constraints struct { Required, Unique bool; MinLen, MaxLen *int; Pattern string; Enum []string }
```

(Names/shapes are indicative; the implementation plan refines them.)

## Serialization

- **HCL-first.** The authored format is HCL, matching the spec. v0 parses HCL into the
  in-memory model via `hashicorp/hcl/v2`.
- **Go consumers use the in-memory model directly** — the inGitDB schema generator and
  the SpecScore `@pm:` resolver are Go and consume `polymodel-go` types, not a file
  format. So "all consumers are YAML/JSON" is not a v0 constraint.
- **JSON/YAML emission** is deferred to when a non-Go consumer (e.g. a TS tool) needs
  it.

## Queries

View definitions (computed collections) and recordset sources use **DTQL** — the
dialect-agnostic query AST — rather than a PolyModel-native query language. This adds
a dependency on `dtql-go`. Watch-item: DTQL is approved-but-unfinished, so v0 uses
whatever is stable in `dtql-go`.

## Dependencies

- `github.com/hashicorp/hcl/v2` — parse the authored HCL.
- `dtql-go` — query AST for views / recordsets.

## v0 scope

In:

- The types above (Entity/Property, Collection/Field, Recordset/Column, Component,
  TypeRef, Constraints).
- HCL parsing into the model.
- Reference resolution within a model (property binds, components, associations,
  collection FKs).
- Structural + constraint **validation** with located errors.

Out (later):

- JSON/YAML serialization.
- Target generators (inGitDB, SQL, etc.) — those live in their own tools.
- Migrations.
- A catalog / remote module resolution.
- Deriving computed-collection fields from the query (v0 declares fields explicitly).

## Open items (deferred, not blocking v0)

- Exact constraint vocabulary and primitive type list (align with the spec's
  type-system work).
- Whether components can be embedded into collections (v0: entities only).
- Catalog reference syntax (`@pm:` and module/version addressing) — needed by the
  SpecScore integration, designed separately.
