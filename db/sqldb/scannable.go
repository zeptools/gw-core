package sqldb

import "github.com/zeptools/gw-core/orm"

type targetFieldsProvider interface {
	TargetFields() []any
}

type Scannable[T any] interface {
	~*T                  // Type Constraint: Underlying Type(~) = *T
	targetFieldsProvider // must implement targetFieldsProvider
}

type ScannableIdentifiable[T any, ID comparable] interface {
	~*T
	targetFieldsProvider
	orm.Identifiable[ID]
}
