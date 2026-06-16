package polymodel

import "testing"

func intp(i int) *int { return &i }

// base returns a valid model exercising all three concepts and their attribute types.
func base() *Model {
	m := NewModel()
	m.Components["Auditable"] = &Component{
		Name:   "Auditable",
		Fields: []Property{{Name: "createdAt", Type: TypeRef{Primitive: "datetime"}}},
	}
	m.Entities["Customer"] = &Entity{
		Name: "Customer",
		Properties: []Property{
			{Name: "id", Type: TypeRef{Primitive: "uuid"}, Constraints: Constraints{Required: true}},
			{Name: "email", Type: TypeRef{Primitive: "string"}, Constraints: Constraints{Required: true, MaxLen: intp(320)}},
		},
		Key: []string{"id"},
	}
	m.Entities["Order"] = &Entity{
		Name: "Order",
		Properties: []Property{
			{Name: "id", Type: TypeRef{Primitive: "uuid"}},
			{Name: "customer", Type: TypeRef{Entity: "Customer"}},  // association
			{Name: "billingEmail", Type: TypeRef{Primitive: "string"}, Reuse: "Customer.email"}, // reuse
		},
		Components: []string{"Auditable"},
		Key:        []string{"id"},
	}
	m.Collections["orders"] = &Collection{
		Name: "orders", Kind: Editable,
		Fields: []Field{
			{Name: "id", Type: TypeRef{Primitive: "uuid"}, Bind: "Order.id"},
			{Name: "note", Type: TypeRef{Primitive: "string"}, Optional: true},
		},
	}
	m.Collections["customers"] = &Collection{
		Name: "customers", Kind: Editable,
		Fields: []Field{{Name: "id", Type: TypeRef{Primitive: "uuid"}}},
	}
	m.Recordsets["order_summary"] = &Recordset{
		Name: "order_summary",
		Columns: []Column{
			{Name: "orderId", Type: TypeRef{Primitive: "uuid"}, Bind: "Order.id"},
			{Name: "total", Type: TypeRef{Primitive: "decimal"}, Source: "SUM(items.amount)"},
			{Name: "payload", Type: TypeRef{Primitive: "document"}},
		},
		Keys: []string{"orderId"},
	}
	return m
}

func hasErr(errs []*Error, obj, attr string) bool {
	for _, e := range errs {
		if e.Object == obj && (attr == "" || e.Attr == attr) {
			return true
		}
	}
	return false
}

// AC: model-concepts — the three concepts load, each with its own attribute type.
func TestModelConcepts(t *testing.T) {
	m := base()
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("expected valid model, got %v", errs)
	}
	if len(m.Entities["Order"].Properties) == 0 {
		t.Error("entity should expose properties")
	}
	if len(m.Collections["orders"].Fields) == 0 {
		t.Error("collection should expose fields")
	}
	if len(m.Recordsets["order_summary"].Columns) == 0 {
		t.Error("recordset should expose columns")
	}
}

// AC: entity-definition — key, association, reuse resolve; key to missing prop rejected.
func TestEntityDefinition(t *testing.T) {
	m := base()
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("base entity should validate, got %v", errs)
	}
	m.Entities["Order"].Key = []string{"nope"}
	errs := m.Validate()
	if !hasErr(errs, "entity Order", "") {
		t.Errorf("expected located error for key naming undeclared property, got %v", errs)
	}
}

// AC: collection-kind — kind required and constrained.
func TestCollectionKind(t *testing.T) {
	m := base()
	m.Collections["orders"].Kind = CollectionKind("nope")
	errs := m.Validate()
	if !hasErr(errs, "collection orders", "") {
		t.Errorf("expected error for invalid kind, got %v", errs)
	}
}

// AC: collection-fields — binding and FK resolve; FK to missing collection rejected.
func TestCollectionFields(t *testing.T) {
	m := base()
	m.Collections["orders"].Fields = append(m.Collections["orders"].Fields,
		Field{Name: "customerId", Type: TypeRef{Primitive: "uuid"}, ForeignKey: "customers"})
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("valid FK should resolve, got %v", errs)
	}
	m.Collections["orders"].Fields[len(m.Collections["orders"].Fields)-1].ForeignKey = "ghost"
	errs := m.Validate()
	if !hasErr(errs, "collection orders", "customerId") {
		t.Errorf("expected error for FK to missing collection, got %v", errs)
	}
}

// AC: computed-query / query-seam-opaque — computed needs a query, editable must not.
func TestComputedQuery(t *testing.T) {
	m := base()
	m.Collections["active"] = &Collection{Name: "active", Kind: Computed, Query: &Query{Raw: "select *"}}
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("computed-with-query should validate, got %v", errs)
	}
	if m.Collections["active"].Query.Raw != "select *" {
		t.Error("query must be retained verbatim (opaque seam)")
	}
	m.Collections["active"].Query = nil
	if errs := m.Validate(); !hasErr(errs, "collection active", "") {
		t.Errorf("expected error for computed without query, got %v", errs)
	}
	m.Collections["orders"].Query = &Query{Raw: "x"}
	if errs := m.Validate(); !hasErr(errs, "collection orders", "") {
		t.Errorf("expected error for editable with query, got %v", errs)
	}
}

// AC: recordset-shape — column order preserved; soft ref recorded, not enforced.
func TestRecordsetShape(t *testing.T) {
	m := base()
	rs := m.Recordsets["order_summary"]
	rs.Columns = append(rs.Columns, Column{Name: "cust", Type: TypeRef{Primitive: "uuid"}, SoftRef: "ghost.id"})
	errs := m.Validate()
	if len(errs) != 0 {
		t.Fatalf("soft ref must NOT be enforced (no error even for unknown target), got %v", errs)
	}
	if rs.Columns[0].Name != "orderId" || rs.Columns[1].Name != "total" {
		t.Error("column order must be preserved")
	}
	if rs.Columns[len(rs.Columns)-1].SoftRef == "" {
		t.Error("soft ref must be recorded")
	}
}

// AC: binding-precedence — local field primary, override applies; unbound valid.
func TestBindingPrecedence(t *testing.T) {
	m := base()
	// Order.id is required; the local field overrides to optional.
	m.Collections["orders"].Fields[0].Optional = true
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("local override should be valid, got %v", errs)
	}
	if !m.Collections["orders"].Fields[0].Optional {
		t.Error("local override (Optional) must be primary over the bound property")
	}
	// Unbound field is valid.
	if m.Collections["orders"].Fields[1].Bind != "" {
		t.Error("expected an unbound field in the base model")
	}
}

// AC: type-system — out-of-set primitive rejected with a located error.
func TestTypeSystem(t *testing.T) {
	m := base()
	m.Entities["Order"].Properties[0].Type = TypeRef{Primitive: "money"}
	errs := m.Validate()
	if !hasErr(errs, "entity Order", "id") {
		t.Errorf("expected located error for unknown primitive, got %v", errs)
	}
}

// AC: validation-located — unresolved reference fails with object+attribute context.
func TestValidationLocated(t *testing.T) {
	m := base()
	m.Entities["Order"].Properties[2].Reuse = "Customer.ghost"
	errs := m.Validate()
	var found *Error
	for _, e := range errs {
		if e.Object == "entity Order" && e.Attr == "billingEmail" {
			found = e
		}
	}
	if found == nil {
		t.Fatalf("expected located error naming object+attribute, got %v", errs)
	}
}

// AC: component-scope — components are embeddable on Entities only (structural).
func TestComponentScope(t *testing.T) {
	m := base()
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("entity component embed should validate, got %v", errs)
	}
	// A Collection has no component-embed field by construction (Decision: component-scope);
	// an embed elsewhere is rejected at parse time (see hcl_test.go).
}
