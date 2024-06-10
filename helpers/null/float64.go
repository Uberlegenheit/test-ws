package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type Float64 sql.NullFloat64

func NewFloat64(data float64) Float64 {
	return Float64{Float64: data, Valid: true}
}

// MarshalJSON implements the json.Marshaler interface.
func (f Float64) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	if math.IsInf(f.Float64, 0) || math.IsNaN(f.Float64) {
		return nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float64),
			Str:   strconv.FormatFloat(f.Float64, 'g', -1, 64),
		}
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *Float64) UnmarshalJSON(bytes []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		f.Float64 = float64(x)
	case string:
		str := string(x)
		if len(str) == 0 {
			f.Valid = false
			return nil
		}
		f.Float64, err = strconv.ParseFloat(str, 64)
	case map[string]interface{}:
		err = json.Unmarshal(bytes, &f.Float64)
	case nil:
		f.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Float", reflect.TypeOf(v).Name())
	}
	f.Valid = err == nil
	return err
}

// IsZero returns true for null strings (omitempty support)
func (f Float64) IsZero() bool {
	return !f.Valid
}

// Scan implements the Scanner interface.
func (f *Float64) Scan(value interface{}) error {
	v := sql.NullFloat64(*f)
	err := v.Scan(value)
	if err != nil {
		return err
	}
	*f = Float64(v)
	return nil
}

// Scan implements the Scanner interface.
func (f Float64) Value() (driver.Value, error) {
	return sql.NullFloat64(f).Value()
}
