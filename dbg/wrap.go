package dbg

import "encoding/json/v2"

type Wrapped[T any] struct {
	Data      T
	DebugData any
}

func Wrap[T any](data T) *Wrapped[T] {
	return &Wrapped[T]{
		Data: data,
	}
}

func (w *Wrapped[T]) MarshalJSON() ([]byte, error) {
	if w.DebugData == nil {
		return json.Marshal(w.Data)
	}
	return json.Marshal(map[string]any{
		"data":       w.Data,
		"debug_data": w.DebugData,
	})
}
