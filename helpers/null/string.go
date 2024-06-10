package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

type String sql.NullString

func NewString(data string) String {
	return String{String: data, Valid: true}
}

// MarshalJSON implements the json.Marshaler interface.
func (s String) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *String) UnmarshalJSON(bytes []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		s.String = x
	case map[string]interface{}:
		err = json.Unmarshal(bytes, &s.String)
	case nil:
		s.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.String", reflect.TypeOf(v).Name())
	}
	s.Valid = err == nil
	return err
}

// IsZero returns true for null strings (omitempty support)
func (s String) IsZero() bool {
	return !s.Valid
}

// Scan implements the Scanner interface.
func (s *String) Scan(value interface{}) error {
	v := sql.NullString(*s)
	err := v.Scan(value)
	if err != nil {
		return err
	}
	*s = String(v)
	return nil
}

// Scan implements the Scanner interface.
func (s String) Value() (driver.Value, error) {
	return sql.NullString(s).Value()
}
