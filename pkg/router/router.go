package router

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"html/template"
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

func HTMLGet[Req any](
	r *Router, pattern PathPattern,
	handler func(ctx *Context, req Req) (template.HTML, error),
) {
	r.mux.Get(string(pattern), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		var req Req

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

func assignParams(req any, getter func(key string) string) error {
	val := reflect.ValueOf(req)
	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		fieldType := typ.Field(i)

		jsonName := fieldType.Tag.Get("json")
		findIndex := strings.Index(jsonName, ",")
		if findIndex >= 0 {
			jsonName = jsonName[:findIndex]
		}

		fieldVal := getter(jsonName)
		if len(fieldVal) == 0 {
			continue
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(fieldVal)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(fieldVal, 10, 64)
			if err != nil {
				panic(err) // TODO
			}
			f.SetInt(intVal)

		case reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			intVal, err := strconv.ParseUint(fieldVal, 10, 64)
			if err != nil {
				panic(err) // TODO
			}
			f.SetUint(intVal)

		default:
			panic("TODO")
		}
	}

	return nil
}
