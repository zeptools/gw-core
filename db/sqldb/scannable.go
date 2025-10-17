package sqldb

type targetFieldsProvider interface {
	TargetFields() []any
}

type Scannable[T any] interface {
	~*T                  // any pointer to T, including aliases
	targetFieldsProvider // must implement targetFieldsProvider
}
