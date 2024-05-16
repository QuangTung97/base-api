package urls

import (
	"fmt"
	"reflect"
	"strings"
)

type Path[T any] struct {
	pathParams []string
}

type Empty struct{}

func New[T any](path string) Path[T] {
	params := findPathParams(path)

	var obj T
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Struct {
		panic("must be a struct type")
	}

	fieldNames := map[string]struct{}{}
	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		tagName := getTagName(f.Tag.Get("json"))
		if len(tagName) == 0 {
			msg := fmt.Sprintf(
				"missing json struct tag for type '%s'",
				objType.Name(),
			)
			panic(msg)
		}
		fieldNames[tagName] = struct{}{}
	}

	for _, param := range params {
		_, ok := fieldNames[param]
		if !ok {
			msg := fmt.Sprintf(
				"missing path param '%s' in struct '%s'",
				param, objType.Name(),
			)
			panic(msg)
		}
	}

	return Path[T]{
		pathParams: params,
	}
}

func getTagName(tag string) string {
	index := strings.Index(tag, ",")
	if index < 0 {
		return tag
	}
	return tag[:index]
}

func findPathParams(path string) []string {
	var result []string
	for {
		index := strings.Index(path, "{")
		if index < 0 {
			return result
		}
		end := strings.Index(path, "}")
		if end < 0 {
			panic("missing closing bracket")
		}

		colonIndex := strings.Index(path[:end], ":")
		if colonIndex >= 0 {
			result = append(result, path[index+1:colonIndex])
		} else {
			result = append(result, path[index+1:end])
		}

		path = path[end+1:]
	}
}
