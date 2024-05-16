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
	checkStructTypeWithParams(obj, params)

	return Path[T]{
		pathParams: params,
	}
}

func checkMustBeStruct(objType reflect.Type) {
	if objType.Kind() != reflect.Struct {
		panic("must be a struct type")
	}
}

func checkStructTypeWithParams(obj any, pathParams []string) {
	objType := reflect.TypeOf(obj)
	checkMustBeStruct(objType)

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

	for _, param := range pathParams {
		_, ok := fieldNames[param]
		if !ok {
			msg := fmt.Sprintf(
				"missing path param '%s' in struct '%s'",
				param, objType.Name(),
			)
			panic(msg)
		}
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

func CheckIsSubStruct(sub any, parent any) {
	subType := reflect.TypeOf(sub)
	parentType := reflect.TypeOf(parent)

	checkMustBeStruct(parentType)
	checkMustBeStruct(subType)

	type fieldInfo struct {
		Type reflect.Type
		Tag  string
	}

	subMap := map[string]fieldInfo{}
	for i := 0; i < subType.NumField(); i++ {
		f := subType.Field(i)
		subMap[f.Name] = fieldInfo{
			Type: f.Type,
			Tag:  string(f.Tag),
		}
	}

	for i := 0; i < parentType.NumField(); i++ {
		f := parentType.Field(i)
		info, ok := subMap[f.Name]
		if !ok {
			panic(fmt.Sprintf(
				"missing field '%s' in struct '%s'",
				f.Name, subType.Name(),
			))
		}
		if f.Type != info.Type {
			panic(fmt.Sprintf(
				"mismatch type of field '%s' in struct '%s'",
				f.Name, subType.Name(),
			))
		}
		if string(f.Tag) != info.Tag {
			panic(fmt.Sprintf(
				"mismatch struct tag of field '%s' in struct '%s'",
				f.Name, subType.Name(),
			))
		}
	}
}
