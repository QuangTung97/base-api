package urls

import (
	"fmt"
	"learn-gin/pkg/null"
	"net/url"
	"reflect"
	"slices"
	"strings"
)

type Path[T any] struct {
	pattern    string
	pathParams []string
	allParams  []string
}

type Empty struct{}

func New[T any](pattern string) Path[T] {
	params := findPathParams(pattern)
	var obj T
	allParams := checkStructTypeWithParams(obj, params)

	return Path[T]{
		pattern:    pattern,
		pathParams: params,
		allParams:  allParams,
	}
}

func (p Path[T]) GetPattern() string {
	return p.pattern
}

func (p Path[T]) GetPathParams() []string {
	return slices.Clone(p.pathParams)
}

func (p Path[T]) GetAllParams() []string {
	return slices.Clone(p.allParams)
}

func getFieldValues(obj any) map[string]string {
	objVal := reflect.ValueOf(obj)
	objType := objVal.Type()

	result := map[string]string{}

	for i := 0; i < objVal.NumField(); i++ {
		f := objVal.Field(i)
		if f.IsZero() {
			continue
		}

		fieldType := objType.Field(i)
		jsonName := computeJSONName(fieldType.Tag.Get("json"))

		var val string

		validVal, dataVal, ok := null.IsNullType(f)
		if ok {
			if validVal.Bool() {
				val = fmt.Sprint(dataVal.Interface())
			}
		} else {
			val = fmt.Sprint(f.Interface())
		}

		result[jsonName] = val
	}

	return result
}

func (p Path[T]) Eval(param T) string {
	var buf strings.Builder

	paramVales := getFieldValues(param)

	pathParamSet := map[string]struct{}{}
	for _, pathParam := range p.pathParams {
		pathParamSet[pathParam] = struct{}{}
	}

	pattern := p.pattern
	for {
		index := strings.Index(pattern, "{")
		if index < 0 {
			buf.WriteString(pattern)
			break
		}

		end := strings.Index(pattern, "}")

		buf.WriteString(pattern[:index])

		key := pattern[index+1 : end]
		colonIndex := strings.Index(key, ":")
		if colonIndex >= 0 {
			key = key[:colonIndex]
		}

		buf.WriteString(paramVales[key])
		delete(paramVales, key)

		pattern = pattern[end+1:]
	}

	if len(paramVales) == 0 {
		return buf.String()
	}

	keys := make([]string, 0, len(paramVales))
	for k := range paramVales {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	queryParams := url.Values{}
	for _, k := range keys {
		queryParams.Set(k, paramVales[k])
	}

	buf.WriteString("?")
	buf.WriteString(queryParams.Encode())

	return buf.String()
}

func checkMustBeStruct(objType reflect.Type) {
	if objType.Kind() != reflect.Struct {
		panic("must be a struct type")
	}
}

func checkStructTypeWithParams(obj any, pathParams []string) []string {
	objType := reflect.TypeOf(obj)
	checkMustBeStruct(objType)

	fieldNames := map[string]struct{}{}
	var allParams []string

	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		tagName := computeJSONName(f.Tag.Get("json"))
		if len(tagName) == 0 {
			msg := fmt.Sprintf(
				"missing json struct tag for type '%s'",
				objType.Name(),
			)
			panic(msg)
		}
		fieldNames[tagName] = struct{}{}
		allParams = append(allParams, tagName)
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

	return allParams
}

func computeJSONName(tag string) string {
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

		key := path[index+1 : end]
		colonIndex := strings.Index(key, ":")
		if colonIndex >= 0 {
			key = key[:colonIndex]
		}
		result = append(result, key)

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
