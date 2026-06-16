package polymodel

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// HCL is the authored format (Decision D-0004). ParseHCL parses HCL source into the
// in-memory Model. An unknown top-level block or unsupported argument is rejected with
// a located diagnostic (AC: model-concepts, component-scope). The returned Model is
// not yet validated — call Model.Validate for that.
func ParseHCL(filename string, src []byte) (*Model, error) {
	f, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	var root hclRoot
	if d := gohcl.DecodeBody(f.Body, nil, &root); d.HasErrors() {
		return nil, d
	}
	return root.toModel(), nil
}

type hclRoot struct {
	Entities    []hclEntity     `hcl:"entity,block"`
	Collections []hclCollection `hcl:"collection,block"`
	Recordsets  []hclRecordset  `hcl:"recordset,block"`
	Components  []hclComponent  `hcl:"component,block"`
}

// hclAttr is the shared shape of a property, field, or column block. Each concept
// uses the subset of attributes that applies to it.
type hclAttr struct {
	Name       string   `hcl:"name,label"`
	Type       string   `hcl:"type,optional"`
	Component  string   `hcl:"component,optional"`
	Entity     string   `hcl:"entity,optional"`
	Required   bool     `hcl:"required,optional"`
	Unique     bool     `hcl:"unique,optional"`
	MinLen     *int     `hcl:"min_len,optional"`
	MaxLen     *int     `hcl:"max_len,optional"`
	Pattern    string   `hcl:"pattern,optional"`
	Enum       []string `hcl:"enum,optional"`
	Reuse      string   `hcl:"reuse,optional"`       // property
	Optional   bool     `hcl:"optional,optional"`    // field
	Bind       string   `hcl:"bind,optional"`        // field, column
	ForeignKey string   `hcl:"foreign_key,optional"` // field
	Source     string   `hcl:"source,optional"`      // column
	SoftRef    string   `hcl:"soft_ref,optional"`    // column
}

type hclEntity struct {
	Name       string    `hcl:"name,label"`
	Key        []string  `hcl:"key,optional"`
	Use        []string  `hcl:"use,optional"` // embedded components — Entities only
	Properties []hclAttr `hcl:"property,block"`
}

type hclCollection struct {
	Name   string    `hcl:"name,label"`
	Kind   string    `hcl:"kind"`
	Query  *string   `hcl:"query,optional"`
	Fields []hclAttr `hcl:"field,block"`
}

type hclRecordset struct {
	Name    string    `hcl:"name,label"`
	Key     []string  `hcl:"key,optional"`
	Query   *string   `hcl:"query,optional"`
	Columns []hclAttr `hcl:"column,block"`
}

type hclComponent struct {
	Name   string    `hcl:"name,label"`
	Fields []hclAttr `hcl:"field,block"`
}

func attrType(a hclAttr) TypeRef {
	return TypeRef{Primitive: a.Type, Component: a.Component, Entity: a.Entity}
}

func attrConstraints(a hclAttr) Constraints {
	return Constraints{
		Required: a.Required, Unique: a.Unique,
		MinLen: a.MinLen, MaxLen: a.MaxLen,
		Pattern: a.Pattern, Enum: a.Enum,
	}
}

func (r hclRoot) toModel() *Model {
	m := NewModel()
	for _, e := range r.Entities {
		ent := &Entity{Name: e.Name, Key: e.Key, Components: e.Use}
		for _, p := range e.Properties {
			ent.Properties = append(ent.Properties, Property{
				Name: p.Name, Type: attrType(p), Constraints: attrConstraints(p), Reuse: p.Reuse,
			})
		}
		m.Entities[e.Name] = ent
	}
	for _, c := range r.Collections {
		col := &Collection{Name: c.Name, Kind: CollectionKind(c.Kind)}
		if c.Query != nil {
			col.Query = &Query{Raw: *c.Query}
		}
		for _, f := range c.Fields {
			col.Fields = append(col.Fields, Field{
				Name: f.Name, Type: attrType(f), Optional: f.Optional,
				Bind: f.Bind, ForeignKey: f.ForeignKey, Constraints: attrConstraints(f),
			})
		}
		m.Collections[c.Name] = col
	}
	for _, rs := range r.Recordsets {
		set := &Recordset{Name: rs.Name, Keys: rs.Key}
		if rs.Query != nil {
			set.Query = &Query{Raw: *rs.Query}
		}
		for _, col := range rs.Columns {
			set.Columns = append(set.Columns, Column{
				Name: col.Name, Type: attrType(col), Bind: col.Bind,
				Source: col.Source, SoftRef: col.SoftRef,
			})
		}
		m.Recordsets[rs.Name] = set
	}
	for _, comp := range r.Components {
		cm := &Component{Name: comp.Name}
		for _, f := range comp.Fields {
			cm.Fields = append(cm.Fields, Property{Name: f.Name, Type: attrType(f), Constraints: attrConstraints(f)})
		}
		m.Components[comp.Name] = cm
	}
	return m
}
