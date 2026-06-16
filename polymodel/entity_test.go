package polymodel

import "testing"

// Sub-feature core-model/entity, AC entity-definition.
// These sharpen the entity-specific edge cases beyond the foundations tests:
// association and component-embed references must resolve.

func TestEntityAssociationUnresolved(t *testing.T) {
	m := base()
	// Order.customer associates to an Entity that does not exist.
	m.Entities["Order"].Properties[1].Type = TypeRef{Entity: "Ghost"}
	errs := m.Validate()
	if !hasErr(errs, "entity Order", "customer") {
		t.Errorf("expected located error for association to unknown entity, got %v", errs)
	}
}

func TestEntityComponentEmbedUnresolved(t *testing.T) {
	m := base()
	m.Entities["Order"].Components = []string{"Ghost"}
	errs := m.Validate()
	if !hasErr(errs, "entity Order", "") {
		t.Errorf("expected located error for embedding an unknown component, got %v", errs)
	}
}

func TestEntityValidAssociationAndReuse(t *testing.T) {
	m := base()
	// base() already wires Order.customer -> Customer (association) and
	// Order.billingEmail reuse of Customer.email; both must resolve cleanly.
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("association + reuse should resolve, got %v", errs)
	}
}
