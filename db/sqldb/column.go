package sqldb

import (
	"fmt"
)

// Column is a validated SQL identifier (e.g. "user.email").
// It cannot be created directly — only via NewColumn().
type Column struct {
	name string // unexported → cannot bypass validation
}

// Name returns the identifier string.
func (c Column) Name() string { return c.name }

func NewColumn(name string) (Column, error) {
	if !IdentifierRegexp.MatchString(name) {
		return Column{}, fmt.Errorf("invalid SQL identifier: %q", name)
	}
	return Column{name: name}, nil
}

// NewColumnOrPanic validates the name and returns a safe Column value.
// WARNING: This function panics if the given name is not a valid SQL identifier.
// You can use recover() to handle invalid input at runtime.
func NewColumnOrPanic(name string) Column {
	if !IdentifierRegexp.MatchString(name) {
		panic(fmt.Errorf("invalid SQL identifier: %q", name))
	}
	return Column{name: name}
}
