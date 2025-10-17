package nullable

import (
	"database/sql"
	"encoding/json/v2"
)

// Int in `nullable` package
// implements: sql.Scanner by embedding sql.NullInt64
// implements: json.Marshaler and json.Unmarshaler
type Int struct {
	sql.NullInt64
}

func (n *Int) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int64)
	}
	return []byte("null"), nil
}

func (n *Int) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Int64 = 0
		return nil
	}
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	n.Int64 = i
	n.Valid = true
	return nil
}

func (n *Int) ForceValue() int64 {
	if !n.Valid {
		return 0
	}
	return n.Int64
}

func (n *Int) IsNil() bool {
	return !n.Valid
}
