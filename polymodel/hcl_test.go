package polymodel

import "testing"

const sampleHCL = `
component "Auditable" {
  field "createdAt" {
    type = "datetime"
  }
}

entity "Customer" {
  key = ["id"]
  property "id" {
    type     = "uuid"
    required = true
  }
  property "email" {
    type     = "string"
    required = true
    max_len  = 320
  }
}

entity "Order" {
  key = ["id"]
  use = ["Auditable"]
  property "id" {
    type = "uuid"
  }
  property "customer" {
    entity = "Customer"
  }
}

collection "orders" {
  kind = "editable"
  field "id" {
    type = "uuid"
    bind = "Order.id"
  }
  field "customerId" {
    type        = "uuid"
    foreign_key = "customers"
  }
}

collection "customers" {
  kind = "editable"
  field "id" {
    type = "uuid"
  }
}

collection "active_orders" {
  kind  = "computed"
  query = "from orders where status = 'active'"
  field "id" {
    type = "uuid"
  }
}

recordset "order_total" {
  key = ["orderId"]
  column "orderId" {
    type = "uuid"
    bind = "Order.id"
  }
  column "total" {
    type   = "decimal"
    source = "SUM(items.amount)"
  }
  column "payload" {
    type = "document"
  }
}
`

// AC: authoring + model-concepts — HCL parses into the in-memory model, usable directly.
func TestParseHCL(t *testing.T) {
	m, err := ParseHCL("sample.pm.hcl", []byte(sampleHCL))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("parsed model should validate, got %v", errs)
	}
	if _, ok := m.Entities["Order"]; !ok {
		t.Error("expected Entity Order")
	}
	if m.Collections["active_orders"].Kind != Computed || m.Collections["active_orders"].Query == nil {
		t.Error("computed collection should carry a query")
	}
	if m.Collections["active_orders"].Query.Raw != "from orders where status = 'active'" {
		t.Error("query must be retained verbatim (opaque seam)")
	}
	if len(m.Recordsets["order_total"].Columns) != 3 {
		t.Error("recordset columns not parsed")
	}
}

// AC: model-concepts — an unknown top-level concept is rejected with a located error.
func TestParseUnknownConcept(t *testing.T) {
	_, err := ParseHCL("bad.pm.hcl", []byte(`widget "x" {}`))
	if err == nil {
		t.Fatal("expected error for unknown top-level block")
	}
}

// AC: component-scope — components embed only on Entities; an embed elsewhere is rejected.
func TestParseComponentScopeRejected(t *testing.T) {
	src := `
collection "orders" {
  kind = "editable"
  use  = ["Auditable"]
}`
	if _, err := ParseHCL("bad.pm.hcl", []byte(src)); err == nil {
		t.Fatal("expected error: a collection cannot embed a component (use is entity-only)")
	}
}
