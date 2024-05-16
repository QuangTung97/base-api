package router

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type reqBody struct {
	Name    string `json:"name"`
	Age     int    `json:"age,omitempty"`
	Counter uint32 `json:"counter"`
}

func TestAssignParams(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var req reqBody
		m := map[string]string{
			"name":    "user01",
			"age":     "1234",
			"counter": "89",
		}

		err := assignParams(&req, func(key string) string {
			return m[key]
		})
		assert.Equal(t, nil, err)
		assert.Equal(t, reqBody{
			Name:    "user01",
			Age:     1234,
			Counter: 89,
		}, req)
	})
}
