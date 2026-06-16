package polymodel

import "fmt"

// Error is a located validation error: it names the object and (optionally) the
// attribute where the violation occurred (AC: validation-located).
type Error struct {
	Object string // e.g. "entity User", "collection Orders"
	Attr   string // field/property/column name, or ""
	Msg    string
}

func (e *Error) Error() string {
	if e.Attr != "" {
		return fmt.Sprintf("%s.%s: %s", e.Object, e.Attr, e.Msg)
	}
	return fmt.Sprintf("%s: %s", e.Object, e.Msg)
}
