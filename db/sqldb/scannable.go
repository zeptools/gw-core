package sqldb

type targetFieldsProvider interface {
	TargetFields() []any
}

type Scannable[T any] interface {
	~*T                  // any pointer to T, including aliases
	targetFieldsProvider // must implement targetFieldsProvider
}

type Identifiable[ID comparable] interface {
	GetID() ID
}

type ScannableIdentifiable[T any, ID comparable] interface {
	~*T
	targetFieldsProvider
	Identifiable[ID]
}
