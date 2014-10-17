package json

import (
	"encoding/json"
	"strconv"
)

// Used as a JSON type
//
// The Set flag indicates whether an unmarshaling actually happened on the type
type RequiredInt64 struct {
	Set   bool
	Int64 int64
}

func (r RequiredInt64) MarshalJSON() ([]byte, error) {
	lit := strconv.FormatInt(r.Int64, 10)
	return json.Marshal(lit)
}

func (r *RequiredInt64) UnmarshalJSON(raw []byte) error {
	var lit string
	var err error
	if err = json.Unmarshal(raw, &lit); err != nil {
		return err
	}
	r.Int64, err = strconv.ParseInt(lit, 10, 64)
	if err != nil {
		return err
	}
	r.Set = true
	return nil
}

type RequiredInt8 struct {
	Set  bool
	Int8 int8
}

func (r RequiredInt8) MarshalJSON() ([]byte, error) {
	lit := strconv.Itoa(int(r.Int8))
	return json.Marshal(lit)
}

func (r *RequiredInt8) UnmarshalJSON(raw []byte) error {
	var lit string
	if err := json.Unmarshal(raw, &lit); err != nil {
		return err
	}
	intVal, err := strconv.ParseInt(lit, 10, 8)
	if err != nil {
		return err
	}
	r.Int8 = int8(intVal)
	r.Set = true
	return nil
}
