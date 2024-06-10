package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

type Bool sql.NullBool

func NewBool(data bool) Bool {
	return Bool{Bool: data, Valid: true}
}

// MarshalJSON implements the json.Marshaler interface.
func (b Bool) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	if !b.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Bool) UnmarshalJSON(bytes []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case bool:
		b.Bool = x
	case map[string]interface{}:
		err = json.Unmarshal(bytes, &b.Bool)
	case nil:
		b.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Bool", reflect.TypeOf(v).Name())
	}
	b.Valid = err == nil
	return err
}

// IsZero returns true for null strings (omitempty support)
func (b Bool) IsZero() bool {
	return !b.Valid
}

// Scan implements the Scanner interface.
func (b *Bool) Scan(value interface{}) error {
	v := sql.NullBool(*b)
	err := v.Scan(value)
	if err != nil {
		return err
	}
	*b = Bool(v)
	return nil
}

// Scan implements the Scanner interface.
func (b Bool) Value() (driver.Value, error) {
	return sql.NullBool(b).Value()
}
