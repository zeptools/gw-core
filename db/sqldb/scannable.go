package sqldb

import "github.com/zeptools/gw-core/orm"

type scanFieldsProvider interface {
	FieldsToScan() []any
}

type Scannable[T any] interface {
	~*T                // Type Constraint: Underlying Type(~) = *T
	scanFieldsProvider // must implement scanFieldsProvider
}

type ScannableIdentifiable[T any, ID comparable] interface {
	~*T
	scanFieldsProvider
	orm.Identifiable[ID]
}
