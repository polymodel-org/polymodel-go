package polymodel

import "testing"

// Sub-feature core-model/recordset, AC recordset-shape.
// These sharpen recordset-specific edge cases beyond the foundations tests.

func TestRecordsetColumnBindUnresolved(t *testing.T) {
	m := base()
	m.Recordsets["order_summary"].Columns[0].Bind = "Order.ghost"
	errs := m.Validate()
	if !hasErr(errs, "recordset order_summary", "orderId") {
		t.Errorf("expected located error for column binding to unknown property, got %v", errs)
	}
}

func TestRecordsetKeyUndeclaredColumn(t *testing.T) {
	m := base()
	m.Recordsets["order_summary"].Keys = []string{"nope"}
	errs := m.Validate()
	if !hasErr(errs, "recordset order_summary", "") {
		t.Errorf("expected located error for key naming an undeclared column, got %v", errs)
	}
}

func TestRecordsetDocumentColumn(t *testing.T) {
	m := base()
	// A recordset column may carry a document/JSON value.
	cols := m.Recordsets["order_summary"].Columns
	if cols[2].Type.Primitive != "document" {
		t.Fatal("expected a document-typed column in the base model")
	}
	if errs := m.Validate(); len(errs) != 0 {
		t.Fatalf("document-typed column should validate, got %v", errs)
	}
}
