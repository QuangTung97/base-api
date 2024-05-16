package null

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	x := New("example")

	assert.Equal(t, Null[string]{
		Valid: true,
		Data:  "example",
	}, x)

	assert.Equal(t, false, x.IsNull())
}

func TestIsNullType(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		obj := &Null[string]{}

		val := reflect.ValueOf(obj).Elem()
		validVal, innerVal, ok := IsNullType(val)
		assert.Equal(t, true, ok)

		validVal.SetBool(true)
		innerVal.SetString("new string")

		assert.Equal(t, Null[string]{
			Valid: true,
			Data:  "new string",
		}, *obj)
	})

	t.Run("not a struct", func(t *testing.T) {
		val := reflect.ValueOf("")
		_, _, ok := IsNullType(val)
		assert.Equal(t, false, ok)
	})

	t.Run("num fields not = 2", func(t *testing.T) {
		type invalidStruct struct {
			Name string
		}

		val := reflect.ValueOf(invalidStruct{})
		_, _, ok := IsNullType(val)
		assert.Equal(t, false, ok)
	})

	t.Run("first field is not boolean", func(t *testing.T) {
		type invalidStruct struct {
			Valid int
			Name  string
		}

		val := reflect.ValueOf(invalidStruct{})
		_, _, ok := IsNullType(val)
		assert.Equal(t, false, ok)
	})

	t.Run("first field is not Valid", func(t *testing.T) {
		type invalidStruct struct {
			Name  bool
			Other string
		}

		val := reflect.ValueOf(invalidStruct{})
		_, _, ok := IsNullType(val)
		assert.Equal(t, false, ok)
	})
}
