package router

import (
	"github.com/stretchr/testify/assert"
	"learn-gin/pkg/null"
	"testing"
)

type reqBody struct {
	Name    string `json:"name"`
	Age     int    `json:"age,omitempty"`
	Counter uint32 `json:"counter"`
}

type invalidBody struct {
	Name any `json:"name"`
}

func TestAssignParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name":    "user01",
			"age":     "1234",
			"counter": "89",
		}

		err := assignParams(&req, []string{"name", "age", "counter"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, reqBody{
			Name:    "user01",
			Age:     1234,
			Counter: 89,
		}, req)
	})

	t.Run("partial set", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name":    "user01",
			"age":     "1234",
			"counter": "89",
		}

		err := assignParams(&req, []string{"name", "counter"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, reqBody{
			Name:    "user01",
			Counter: 89,
		}, req)
	})

	t.Run("partial values", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name": "user01",
		}

		err := assignParams(&req, []string{"name", "age", "counter"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, reqBody{
			Name: "user01",
		}, req)
	})

	t.Run("not a number", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name": "user01",
			"age":  "A9",
		}

		err := assignParams(&req, []string{"name", "age", "counter"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, "router: can not parse value 'A9' into field 'Age'", err.Error())
		assert.Equal(t, reqBody{
			Name: "user01",
		}, req)
	})

	t.Run("not a unsigned number", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name":    "user01",
			"counter": "A9",
		}

		err := assignParams(&req, []string{"name", "age", "counter"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, "router: can not parse value 'A9' into field 'Counter'", err.Error())
		assert.Equal(t, reqBody{
			Name: "user01",
		}, req)
	})

	t.Run("invalid type", func(t *testing.T) {
		var req invalidBody
		m := map[string]string{
			"name": "user01",
		}

		err := assignParams(&req, []string{"name"}, func(key string) string {
			return m[key]
		})
		assert.Equal(t, "unrecognized field type 'interface' of field 'Name'", err.Error())
		assert.Equal(t, invalidBody{}, req)
	})
}

type nullReqBody struct {
	Name null.Null[string] `json:"name"`
}

func TestAssignParams_With_Null(t *testing.T) {
	var req nullReqBody
	m := map[string]string{
		"name": "user01",
	}
	err := assignParams(&req, []string{"name"}, func(key string) string {
		return m[key]
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, nullReqBody{
		Name: null.New("user01"),
	}, req)
}
