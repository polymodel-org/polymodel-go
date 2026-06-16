package polymodel

import (
	"fmt"
	"regexp"
	"strings"
)

// Validate checks structural and constraint rules and resolves intra-model
// references, returning located errors (AC: validation-located). An empty result
// means the model is valid.
func (m *Model) Validate() []*Error {
	var errs []*Error

	for _, e := range m.Entities {
		obj := "entity " + e.Name
		declared := map[string]bool{}
		for _, p := range e.Properties {
			declared[p.Name] = true
		}
		for _, p := range e.Properties {
			m.checkType(&errs, obj, p.Name, p.Type, true)
			m.checkConstraints(&errs, obj, p.Name, p.Constraints)
			if p.Reuse != "" && !m.resolveProperty(p.Reuse) {
				errs = append(errs, &Error{obj, p.Name,
					fmt.Sprintf("reuse target %q does not resolve to an entity property", p.Reuse)})
			}
		}
		for _, c := range e.Components {
			if _, ok := m.Components[c]; !ok {
				errs = append(errs, &Error{obj, "", fmt.Sprintf("unknown component %q", c)})
			}
		}
		for _, k := range e.Key {
			if !declared[k] {
				errs = append(errs, &Error{obj, "",
					fmt.Sprintf("key references undeclared property %q", k)})
			}
		}
	}

	for _, comp := range m.Components {
		obj := "component " + comp.Name
		for _, p := range comp.Fields {
			m.checkType(&errs, obj, p.Name, p.Type, false)
			m.checkConstraints(&errs, obj, p.Name, p.Constraints)
		}
	}

	for _, c := range m.Collections {
		obj := "collection " + c.Name
		switch c.Kind {
		case Editable:
			if c.Query != nil {
				errs = append(errs, &Error{obj, "", "editable collection must not carry a query"})
			}
		case Computed:
			if c.Query == nil {
				errs = append(errs, &Error{obj, "", "computed collection must carry a query"})
			}
		default:
			errs = append(errs, &Error{obj, "",
				fmt.Sprintf("invalid kind %q (must be %q or %q)", c.Kind, Editable, Computed)})
		}
		for _, f := range c.Fields {
			m.checkType(&errs, obj, f.Name, f.Type, false)
			m.checkConstraints(&errs, obj, f.Name, f.Constraints)
			if f.Bind != "" && !m.resolveProperty(f.Bind) {
				errs = append(errs, &Error{obj, f.Name,
					fmt.Sprintf("binding %q does not resolve to an entity property", f.Bind)})
			}
			if f.ForeignKey != "" {
				if _, ok := m.Collections[f.ForeignKey]; !ok {
					errs = append(errs, &Error{obj, f.Name,
						fmt.Sprintf("foreign key target collection %q does not exist", f.ForeignKey)})
				}
			}
		}
	}

	for _, r := range m.Recordsets {
		obj := "recordset " + r.Name
		declared := map[string]bool{}
		for _, col := range r.Columns {
			declared[col.Name] = true
		}
		for _, col := range r.Columns {
			m.checkType(&errs, obj, col.Name, col.Type, false)
			if col.Bind != "" && !m.resolveProperty(col.Bind) {
				errs = append(errs, &Error{obj, col.Name,
					fmt.Sprintf("binding %q does not resolve to an entity property", col.Bind)})
			}
			// SoftRef is a navigational hint — recorded, never enforced (AC: recordset-shape).
		}
		for _, k := range r.Keys {
			if !declared[k] {
				errs = append(errs, &Error{obj, "",
					fmt.Sprintf("key references undeclared column %q", k)})
			}
		}
	}

	return errs
}

// resolveProperty reports whether "Entity.property" names an existing entity property.
func (m *Model) resolveProperty(ref string) bool {
	en, prop, ok := strings.Cut(ref, ".")
	if !ok {
		return false
	}
	e, ok := m.Entities[en]
	if !ok {
		return false
	}
	for _, p := range e.Properties {
		if p.Name == prop {
			return true
		}
	}
	return false
}

func (m *Model) checkType(errs *[]*Error, obj, attr string, t TypeRef, allowEntity bool) {
	set := 0
	if t.Primitive != "" {
		set++
	}
	if t.Component != "" {
		set++
	}
	if t.Entity != "" {
		set++
	}
	if set != 1 {
		*errs = append(*errs, &Error{obj, attr, "type must be exactly one of primitive, component, or entity"})
		return
	}
	switch {
	case t.Primitive != "":
		if !Primitives[t.Primitive] {
			*errs = append(*errs, &Error{obj, attr, fmt.Sprintf("unknown primitive type %q", t.Primitive)})
		}
	case t.Component != "":
		if _, ok := m.Components[t.Component]; !ok {
			*errs = append(*errs, &Error{obj, attr, fmt.Sprintf("unknown component %q", t.Component)})
		}
	case t.Entity != "":
		if !allowEntity {
			*errs = append(*errs, &Error{obj, attr, "entity association is only allowed on entity properties"})
		} else if _, ok := m.Entities[t.Entity]; !ok {
			*errs = append(*errs, &Error{obj, attr, fmt.Sprintf("unknown entity %q", t.Entity)})
		}
	}
}

func (m *Model) checkConstraints(errs *[]*Error, obj, attr string, c Constraints) {
	if c.Pattern != "" {
		if _, err := regexp.Compile(c.Pattern); err != nil {
			*errs = append(*errs, &Error{obj, attr, "invalid pattern: " + err.Error()})
		}
	}
	if c.MinLen != nil && c.MaxLen != nil && *c.MinLen > *c.MaxLen {
		*errs = append(*errs, &Error{obj, attr, "min length exceeds max length"})
	}
}
