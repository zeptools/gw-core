package dbg

type Wrapped[T any] struct {
	Data      T   `json:"data"`
	DebugData any `json:"debug_data,omitempty"`
}

func Wrap[T any](data T) *Wrapped[T] {
	return &Wrapped[T]{
		Data: data,
	}
}
