package null

import (
	"bytes"
	"encoding/json"
	"reflect"
)

type Null[T any] struct {
	Valid bool
	Data  T
}

func New[T any](val T) Null[T] {
	return Null[T]{
		Valid: true,
		Data:  val,
	}
}

func (n Null[T]) IsNull() bool {
	return !n.Valid
}

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Data)
}

func (n *Null[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*n = Null[T]{}
		return nil
	}
	err := json.Unmarshal(b, &n.Data)
	if err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// IsNullType should NOT be used directly
func IsNullType(obj reflect.Value) (validVal reflect.Value, dataVal reflect.Value, ok bool) {
	objType := obj.Type()
	if objType.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.Value{}, false
	}

	if objType.NumField() != 2 {
		return reflect.Value{}, reflect.Value{}, false
	}

	if objType.Field(0).Type.Kind() != reflect.Bool {
		return reflect.Value{}, reflect.Value{}, false
	}

	if objType.Field(0).Name != "Valid" {
		return reflect.Value{}, reflect.Value{}, false
	}

	return obj.Field(0), obj.Field(1), true
}
