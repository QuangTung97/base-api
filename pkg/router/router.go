package router

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"learn-gin/pkg/null"
	"learn-gin/pkg/urls"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type PathPattern string

type HTMLHandler[T any] func(ctx *Context, req T) (template.HTML, error)

type Router struct {
	mux *chi.Mux
}

func NewRouter() *Router {
	mux := chi.NewRouter()
	return &Router{
		mux: mux,
	}
}

func (r *Router) Mux() *chi.Mux {
	return r.mux
}

type ErrorBody struct {
	Error string `json:"error"`
}

func inList(list []string, str string) bool {
	for _, e := range list {
		if e == str {
			return true
		}
	}
	return false
}

func HTMLGet[T any, Req any](
	r *Router, pattern urls.Path[T],
	handler func(ctx *Context, req Req) (template.HTML, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	r.mux.Get(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		var req Req

		pathParams := pattern.GetPathParams()
		err := assignParams(&req, pattern.GetAllParams(), func(key string) string {
			if inList(pathParams, key) {
				return chi.URLParam(ctx.Request, key)
			}
			return ctx.Request.URL.Query().Get(key)
		})
		if err != nil {
			r.writeResponse(ctx, "", err)
			return
		}

		respBody, err := handler(ctx, req)
		r.writeResponse(ctx, respBody, err)
	})
}

func (r *Router) writeResponse(ctx *Context, respBody template.HTML, err error) {
	writer := ctx.Writer

	if err != nil {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(writer).Encode(ErrorBody{
			Error: err.Error(),
		})
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = writer.Write([]byte(respBody))
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
