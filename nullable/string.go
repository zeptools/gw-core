package nullable

import (
	"database/sql"
	"encoding/json/v2"
)

// String in `nullable` package
// implements: sql.Scanner by embedding sql.NullString
// implements: json.Marshaler and json.Unmarshaler
type String struct {
	sql.NullString
}

func (n *String) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return []byte("null"), nil
}

func (n *String) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.String = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	n.String = str
	n.Valid = true
	return nil
}

func (n *String) ForceValue() string {
	if !n.Valid {
		return ""
	}
	return n.String
}

func (n *String) IsNil() bool {
	return !n.Valid
}
