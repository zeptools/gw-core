package constraints

import (
	"fmt"
	"strconv"
)

type UID interface {
	~int64 | ~string
}

func ParseUID[U UID](s string) (U, error) {
	var zero U
	switch any(zero).(type) {
	case string:
		v := ParseUIDToString(s)
		return any(v).(U), nil
	case int64:
		v, err := ParseUIDToInt64(s)
		return any(v).(U), err
	default:
		var zero U
		return zero, fmt.Errorf("unsupported type %T", zero)
	}
}

func ParseUIDToString(s string) string        { return s }
func ParseUIDToInt64(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }
