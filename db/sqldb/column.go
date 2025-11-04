package sqldb

import (
	"fmt"
	"regexp"
)

var regexIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// Column is a validated SQL identifier (e.g. "user.email").
// It cannot be created directly — only via NewColumn().
type Column struct {
	name string // unexported → cannot bypass validation
}

// NewColumn validates the name and returns a safe Column value.
// WARNING: This function panics if the given name is not a valid SQL identifier.
// You can use recover() to handle invalid input at runtime.
func NewColumn(name string) Column {
	if !regexIdentifier.MatchString(name) {
		panic(fmt.Errorf("invalid SQL identifier: %q", name))
	}
	return Column{name: name}
}

// Name returns the identifier string.
func (c Column) Name() string { return c.name }
