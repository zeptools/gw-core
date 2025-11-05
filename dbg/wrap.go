package dbg

type Packed[T any] struct {
	Data      T   `json:"data"`
	DebugData any `json:"debug_data,omitempty"`
}

func Pack[T any](data T) *Packed[T] {
	return &Packed[T]{
		Data: data,
	}
}
