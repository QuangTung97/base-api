package router

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"learn-gin/pkg/null"
	"learn-gin/pkg/urls"
	"reflect"
	"strconv"
	"strings"
)

func inList(list []string, str string) bool {
	for _, e := range list {
		if e == str {
			return true
		}
	}
	return false
}

func doAssignParams[T any](ctx *Context, req any, pattern urls.Path[T]) error {
	pathParams := pattern.GetPathParams()
	return assignParams(req, pattern.GetAllParams(), func(key string) string {
		if inList(pathParams, key) {
			return chi.URLParam(ctx.Request, key)
		}
		return ctx.Request.URL.Query().Get(key)
	})
}

func computeJsonName(tag string) string {
	findIndex := strings.Index(tag, ",")
	if findIndex >= 0 {
		return tag[:findIndex]
	}
	return tag
}

type ParseError struct {
	Field string
	Value string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("router: can not parse value '%s' into field '%s'", e.Value, e.Field)
}

func assignParams(req any, params []string, getter func(key string) string) error {
	val := reflect.ValueOf(req)
	val = val.Elem()
	typ := val.Type()

	paramSet := map[string]struct{}{}
	for _, p := range params {
		paramSet[p] = struct{}{}
	}

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		fieldType := typ.Field(i)

		jsonName := computeJsonName(fieldType.Tag.Get("json"))
		_, ok := paramSet[jsonName]
		if !ok {
			continue
		}

		fieldVal := getter(jsonName)
		if len(fieldVal) == 0 {
			continue
		}

		if err := setFieldData(f, fieldVal, fieldType); err != nil {
			return err
		}
	}

	return nil
}

func setFieldData(f reflect.Value, fieldVal string, fieldType reflect.StructField) error {
	switch f.Kind() {
	case reflect.String:
		f.SetString(fieldVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(fieldVal, 10, 64)
		if err != nil {
			return &ParseError{Value: fieldVal, Field: fieldType.Name}
		}
		f.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		intVal, err := strconv.ParseUint(fieldVal, 10, 64)
		if err != nil {
			return &ParseError{Value: fieldVal, Field: fieldType.Name}
		}
		f.SetUint(intVal)

	default:
		validVal, dataVal, ok := null.IsNullType(f)
		if ok {
			validVal.SetBool(true)
			return setFieldData(dataVal, fieldVal, fieldType)
		}
		return fmt.Errorf(
			"unrecognized field type '%s' of field '%s'",
			fieldType.Type.Kind(), fieldType.Name,
		)
	}
	return nil
}
