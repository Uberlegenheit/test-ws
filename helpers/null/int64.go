package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type Int64 sql.NullInt64

func NewInt64(data int64) Int64 {
	return Int64{Int64: data, Valid: true}
}

// MarshalJSON implements the json.Marshaler interface.
func (i Int64) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *Int64) UnmarshalJSON(bytes []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		// Unmarshal again, directly to int64, to avoid intermediate float64
		err = json.Unmarshal(bytes, &i.Int64)
	case string:
		str := string(x)
		if len(str) == 0 {
			i.Valid = false
			return nil
		}
		i.Int64, err = strconv.ParseInt(str, 10, 64)
	case map[string]interface{}:
		err = json.Unmarshal(bytes, &i.Int64)
	case nil:
		i.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Int", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil
	return err
}

// IsZero returns true for null strings (omitempty support)
func (i Int64) IsZero() bool {
	return !i.Valid
}

// Scan implements the Scanner interface.
func (i *Int64) Scan(value interface{}) error {
	v := sql.NullInt64(*i)
	err := v.Scan(value)
	if err != nil {
		return err
	}
	*i = Int64(v)
	return nil
}

// Scan implements the Scanner interface.
func (i Int64) Value() (driver.Value, error) {
	return sql.NullInt64(i).Value()
}
