package polymodel

// CollectionKind distinguishes a stored (Editable) collection from a query-defined
// (Computed) one. There is no separate Table or View type — the kind is an attribute
// (Decision D-0002).
type CollectionKind string

const (
	// Editable is a stored table / file-system collection.
	Editable CollectionKind = "editable"
	// Computed is a view: its content is defined by a query.
	Computed CollectionKind = "computed"
)

// Primitives is the v0 primitive type set (Decision: type-system; fuller
// reconciliation deferred).
var Primitives = map[string]bool{
	"string": true, "int": true, "float": true, "bool": true,
	"decimal": true, "uuid": true, "date": true, "time": true,
	"datetime": true, "document": true, "json": true, "any": true,
}

// TypeRef is the type of a property, field, or column: exactly one of a v0
// primitive, a Component embed, or (properties only) an Entity association.
type TypeRef struct {
	Primitive string // a v0 primitive, or ""
	Component string // a Component name, or ""
	Entity    string // an Entity name (association); properties only, or ""
}

// Constraints is the v0 constraint vocabulary. Anything outside this set is not
// expressible (Decision: type-system).
type Constraints struct {
	Required bool
	Unique   bool
	MinLen   *int
	MaxLen   *int
	Pattern  string
	Enum     []string
}

// Property is an Entity attribute — the canonical, semantic definition (Decision
// D-0001). It may associate to another Entity (via Type.Entity) or reuse another
// property's definition (via Reuse).
type Property struct {
	Name        string
	Type        TypeRef
	Constraints Constraints
	Reuse       string // "Entity.property" reuse reference, or ""
}

// Entity is a logical, identity-bearing business object — the semantic anchor.
type Entity struct {
	Name       string
	Properties []Property
	Components []string // embedded Component names (Entities only — Decision: component-scope)
	Key        []string // property names forming the identity
}

// Query is the opaque placeholder seam (Decision D-0003). v0 never interprets it; a
// DTQL AST can replace it later without reshaping the model.
type Query struct {
	Raw string
}

// Field is a Collection attribute — loose and schemaless-capable (Decision D-0001).
// The local field is primary; Bind is an optional reference to an Entity property
// (Decision: binding-local-primary).
type Field struct {
	Name        string
	Type        TypeRef
	Optional    bool
	Bind        string // "Entity.property" binding, or ""
	ForeignKey  string // target Collection name (enforced), or ""
	Constraints Constraints
}

// Collection is a named, FROM-able data source. Computed collections carry a Query.
type Collection struct {
	Name   string
	Kind   CollectionKind
	Fields []Field
	Query  *Query // present iff Kind == Computed
}

// Column is a Recordset attribute — strict and tabular (Decision D-0001). SoftRef is
// a navigational hint, not an enforced constraint.
type Column struct {
	Name    string
	Type    TypeRef
	Bind    string // "Entity.property" semantic binding, or ""
	Source  string // source expression, or ""
	SoftRef string // soft navigational reference, or "" (recorded, not enforced)
}

// Recordset is the tabular shape of a result (a query or stored-procedure output).
type Recordset struct {
	Name    string
	Columns []Column
	Keys    []string
	Query   *Query
}

// Component is a reusable, named group of fields (Entities only in v0).
type Component struct {
	Name   string
	Fields []Property
}

// Model is a whole PolyModel definition: the three concepts plus Components.
type Model struct {
	Entities    map[string]*Entity
	Collections map[string]*Collection
	Recordsets  map[string]*Recordset
	Components  map[string]*Component
}

// NewModel returns an empty, ready-to-populate Model.
func NewModel() *Model {
	return &Model{
		Entities:    map[string]*Entity{},
		Collections: map[string]*Collection{},
		Recordsets:  map[string]*Recordset{},
		Components:  map[string]*Component{},
	}
}
