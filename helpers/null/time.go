package null

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Time struct {
	Valid bool
	Time  time.Time
}

func NewTime(data time.Time) Time {
	return Time{Time: data, Valid: true}
}

const timeFormat = "2006-01-02 15:04:05.999999"

// Scan implements the Scanner interface.
// The value type must be time.Time or string / []byte (formatted time-string),
// otherwise Scan fails.
func (nt *Time) Scan(value interface{}) (err error) {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return
	}

	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return
	case []byte:
		nt.Time, err = parseDateTime(string(v), time.UTC)
		nt.Valid = (err == nil)
		return
	case string:
		nt.Time, err = parseDateTime(v, time.UTC)
		nt.Valid = (err == nil)
		return
	}

	nt.Valid = false
	return fmt.Errorf("can't convert %T to time.Time", value)
}

// Value implements the driver Valuer interface.
func (nt Time) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func parseDateTime(str string, loc *time.Location) (t time.Time, err error) {
	base := "0000-00-00 00:00:00.0000000"
	switch len(str) {
	case 10, 19, 21, 22, 23, 24, 25, 26: // up to "YYYY-MM-DD HH:MM:SS.MMMMMM"
		if str == base[:len(str)] {
			return
		}
		t, err = time.Parse(timeFormat[:len(str)], str)
	default:
		err = fmt.Errorf("invalid time string: %s", str)
		return
	}

	// Adjust location
	if err == nil && loc != time.UTC {
		y, mo, d := t.Date()
		h, mi, s := t.Clock()
		t, err = time.Date(y, mo, d, h, mi, s, t.Nanosecond(), loc), nil
	}

	return
}

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%d", t.Time.Unix())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(bytes []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		val, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return err
		}
		t.Valid = true
		t.Time = time.Unix(val, 0)
		return nil
	case int64, int:
		t.Valid = true
		t.Time = time.Unix(x.(int64), 0)
		return nil
	case float64:
		t.Valid = true
		t.Time = time.Unix(int64(x), 0)
		return nil
	case nil:
		t.Valid = false
		return nil
	default:
		return fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Time", reflect.TypeOf(v).Name())
	}
}

// IsZero returns true for null strings (omitempty support)
func (t Time) IsZero() bool {
	return t.Time.IsZero()
}
