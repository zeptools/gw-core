package sqldb

import "strings"

// OrderBy defines a validated ORDER BY clause.
type OrderBy struct {
	Column Column
	Desc   bool
}

// String returns the safe ORDER BY clause fragment (without the "ORDER BY" prefix).
func (o OrderBy) String() string {
	if o.Desc {
		return o.Column.Name() + " DESC"
	}
	return o.Column.Name() + " ASC"
}

// OrderByClause joins multiple OrderBy items into a valid ORDER BY SQL fragment.
func OrderByClause(orders []OrderBy) string {
	if len(orders) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(16 * len(orders)) // rough prealloc: " column DESC, "
	b.WriteString(" ORDER BY ")
	for i, o := range orders {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(o.Column.Name())
		if o.Desc {
			b.WriteString(" DESC")
		} else {
			b.WriteString(" ASC")
		}
	}
	return b.String()
}
