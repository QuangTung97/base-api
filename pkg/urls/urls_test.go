package urls

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindPathParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		params := findPathParams("/api/users/{user_id}/list")
		assert.Equal(t, []string{"user_id"}, params)
	})

	t.Run("multi", func(t *testing.T) {
		params := findPathParams("/api/users/{user_id}/get/{ds_id}")
		assert.Equal(t, []string{"user_id", "ds_id"}, params)
	})

	t.Run("with regex", func(t *testing.T) {
		params := findPathParams("/api/users/{user_id:[a-z]+}/get/{ds_id}")
		assert.Equal(t, []string{"user_id", "ds_id"}, params)
	})

	t.Run("with regex multi", func(t *testing.T) {
		params := findPathParams("/api/users/{user_id:[a-z]+}/get/{ds_id:[0-9]+}")
		assert.Equal(t, []string{"user_id", "ds_id"}, params)
	})

	t.Run("missing closing bracket", func(t *testing.T) {
		assert.PanicsWithValue(t, "missing closing bracket", func() {
			findPathParams("/api/users/{user_id")
		})
	})
}

type userPath struct {
	UserID int `json:"user_id"`
}

type userPath2 struct {
	UserID int
}

type userPath3 struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name,omitempty"`
	Age    int    `json:"age"`
}

func TestNew(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		p := New[userPath]("/api/users/{user_id}/list")
		assert.Equal(t, []string{"user_id"}, p.pathParams)
	})

	t.Run("missing path param", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"missing path param 'name' in struct 'userPath'",
			func() {
				New[userPath]("/api/users/{name}/list")
			},
		)
	})

	t.Run("not struct", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"must be a struct type",
			func() {
				New[string]("/api/users/{name}/list")
			},
		)
	})

	t.Run("missing json tag", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"missing json struct tag for type 'userPath2'",
			func() {
				New[userPath2]("/api/users/{name}/list")
			},
		)
	})

	t.Run("multi", func(t *testing.T) {
		p := New[userPath3]("/api/users/{user_id}/get/{name}")
		assert.Equal(t, []string{"user_id", "name"}, p.pathParams)
	})
}

type structA struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type structB struct {
	Value int `json:"value"`
}

type structC struct {
	Name int `json:"name"`
	Age  int `json:"age"`
}

type structD struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age"`
}

type structE struct {
	Name string  `json:"name"`
	Age  int     `json:"age"`
	Val  float64 `json:"val"`
}

func TestCheckIsSubStruct(t *testing.T) {
	t.Run("missing field", func(t *testing.T) {
		assert.PanicsWithValue(t, "missing field 'Name' in struct 'structB'", func() {
			CheckIsSubStruct(structB{}, structA{})
		})
	})

	t.Run("invalid type", func(t *testing.T) {
		assert.PanicsWithValue(t, "mismatch type of field 'Name' in struct 'structC'", func() {
			CheckIsSubStruct(structC{}, structA{})
		})
	})

	t.Run("invalid tag", func(t *testing.T) {
		assert.PanicsWithValue(t, "mismatch struct tag of field 'Name' in struct 'structD'", func() {
			CheckIsSubStruct(structD{}, structA{})
		})
	})

	t.Run("same", func(t *testing.T) {
		CheckIsSubStruct(structA{}, structA{})
	})

	t.Run("sub", func(t *testing.T) {
		CheckIsSubStruct(structE{}, structA{})
	})

	t.Run("parent not a struct", func(t *testing.T) {
		assert.PanicsWithValue(t, "must be a struct type", func() {
			CheckIsSubStruct(structA{}, "")
		})
	})

	t.Run("sub not a struct", func(t *testing.T) {
		assert.PanicsWithValue(t, "must be a struct type", func() {
			CheckIsSubStruct("", structA{})
		})
	})
}
