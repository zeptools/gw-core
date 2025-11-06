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
func OrderByClause(orderBys []OrderBy) string {
	if len(orderBys) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(16 * len(orderBys)) // rough prealloc: " column DESC, "
	b.WriteString(" ORDER BY ")
	for i, o := range orderBys {
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
