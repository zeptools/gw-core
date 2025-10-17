package nullable

import (
	"database/sql"
	"encoding/json/v2"
	"time"
)

// Time in `nullable` package
// implements: sql.Scanner by embedding sql.NullTime
// implements: json/v2.Marshaler and json/v2.Unmarshaler
type Time struct {
	sql.NullTime
}

func (n *Time) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time.Format(time.RFC3339))
	}
	return []byte("null"), nil
}

func (n *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Time = time.Time{}
		return nil
	}
	var str string // to string, then, to time.Time
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, str) // time.Time
	if err != nil {
		return err
	}
	n.Time = t
	n.Valid = true
	return nil
}

func (n *Time) ForceValue() time.Time {
	if !n.Valid {
		return time.Time{}
	}
	return n.Time
}

func (n *Time) IsNil() bool {
	return !n.Valid
}
