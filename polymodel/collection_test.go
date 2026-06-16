package polymodel

import "testing"

// Sub-feature core-model/collection, ACs collection-kind, collection-fields, computed-query.
// These sharpen collection-specific edge cases beyond the foundations tests.

func TestCollectionFieldBindUnresolved(t *testing.T) {
	m := base()
	// A field binding to an Entity property that does not exist.
	m.Collections["orders"].Fields[0].Bind = "Order.ghost"
	errs := m.Validate()
	if !hasErr(errs, "collection orders", "id") {
		t.Errorf("expected located error for binding to unknown entity property, got %v", errs)
	}
}

func TestCollectionEditableMayOmitQuery(t *testing.T) {
	m := base()
	// Editable collections in base() carry no query and must validate.
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("editable collection without a query should validate, got %v", errs)
	}
}

func TestCollectionOptionalField(t *testing.T) {
	m := base()
	// The "note" field is Optional — a record may omit it (schemaless-capable).
	if !m.Collections["orders"].Fields[1].Optional {
		t.Fatal("expected the base model to declare an optional field")
	}
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("optional field should validate, got %v", errs)
	}
}
