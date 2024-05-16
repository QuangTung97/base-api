package router

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"html/template"
	"learn-gin/pkg/null"
	"learn-gin/pkg/urls"
	"net/http"
	"net/http/httptest"
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

type userID int64

type userParams struct {
	UserID userID `json:"user_id"`
	Search string `json:"search"`
	Age    int    `json:"age"`
}

type userGetRequest struct {
	UserID userID `json:"user_id"`
	Search string `json:"search"`
	Age    int    `json:"age"`
}

var userPath = urls.New[userParams]("/api/users/{user_id}")

func TestHTMLGet_Success(t *testing.T) {
	r := NewRouter()

	var inputReq userGetRequest
	HTMLGet(r, userPath, func(ctx *Context, req userGetRequest) (template.HTML, error) {
		inputReq = req
		return "<div>Hello</div>", nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userGetRequest{
		UserID: 123,
	}, inputReq)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, "<div>Hello</div>", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"text/html; charset=utf-8"},
	}, writer.Header())
}

func TestHTMLGet_Not_A_Sub_Struct(t *testing.T) {
	r := NewRouter()

	type invalidRequest struct {
		Name string `json:"name"`
	}

	assert.PanicsWithValue(t, "missing field 'UserID' in struct 'invalidRequest'", func() {
		HTMLGet(r, userPath, func(ctx *Context, req invalidRequest) (template.HTML, error) {
			return "", nil
		})
	})
}

func TestHTMLGet_With_Error(t *testing.T) {
	r := NewRouter()

	var inputReq userGetRequest
	HTMLGet(r, userPath, func(ctx *Context, req userGetRequest) (template.HTML, error) {
		inputReq = req
		return "", errors.New("some handler error")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		userPath.Eval(userParams{UserID: 555, Search: "<div>hello</div>"}),
		nil,
	)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userGetRequest{
		UserID: 555,
		Search: "<div>hello</div>",
	}, inputReq)

	assert.Equal(t, http.StatusInternalServerError, writer.Code)
	assert.Equal(t, `{"error":"some handler error"}`+"\n", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}

func TestHTMLGet_With_Invalid_Query_Param(t *testing.T) {
	r := NewRouter()

	HTMLGet(r, userPath, func(ctx *Context, req userGetRequest) (template.HTML, error) {
		return "<div>Hello</div>", nil
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/users/123?age=AB",
		nil,
	)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Equal(t,
		`{"error":"router: can not parse value 'AB' into field 'Age'"}`+"\n",
		writer.Body.String(),
	)
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}

type userSetting struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type userPostRequest struct {
	UserID  userID      `json:"user_id"`
	Search  string      `json:"search"`
	Age     int         `json:"age"`
	Body    string      `json:"body"`
	Setting userSetting `json:"setting"`
}

func TestHTMLPost_Success(t *testing.T) {
	r := NewRouter()

	var inputReq userPostRequest
	HTMLPost(r, userPath, func(ctx *Context, req userPostRequest) (template.HTML, error) {
		inputReq = req
		return "<div>Hello</div>", nil
	})

	body := `
{
  "user_id": 33,
  "search": "search text",
  "body": "Some Body",
  "setting": {
    "path": "some path",
    "count": 8899
  }
}
`

	req := httptest.NewRequest(http.MethodPost, "/api/users/123", bytes.NewBufferString(body))
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userPostRequest{
		UserID: 123,
		Search: "search text",
		Body:   "Some Body",
		Setting: userSetting{
			Path:  "some path",
			Count: 8899,
		},
	}, inputReq)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, "<div>Hello</div>", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"text/html; charset=utf-8"},
	}, writer.Header())
}

func TestHTMLPost_Parse_JSON_Request_Error(t *testing.T) {
	r := NewRouter()

	HTMLPost(r, userPath, func(ctx *Context, req userPostRequest) (template.HTML, error) {
		return "<div>Hello</div>", nil
	})

	body := `
{
  "user_id": 33,
  "search": "search text",
  "body": "Some Body",
  "setting": {
    "path": "some path",
    "count": "mm"
  }
}
`

	req := httptest.NewRequest(http.MethodPost, "/api/users/123", bytes.NewBufferString(body))
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Equal(t,
		`{"error":"json: cannot unmarshal string into Go struct field userSetting.setting.count of type int"}`+"\n",
		writer.Body.String(),
	)
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}

func TestHTMLPost_Parse_Query_Param_Error(t *testing.T) {
	r := NewRouter()

	HTMLPost(r, userPath, func(ctx *Context, req userPostRequest) (template.HTML, error) {
		return "<div>Hello</div>", nil
	})

	body := `
{
  "user_id": 33,
  "body": "Some Body"
}
`

	req := httptest.NewRequest(http.MethodPost, "/api/users/123?age=AA", bytes.NewBufferString(body))
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, http.StatusBadRequest, writer.Code)
	assert.Equal(t,
		`{"error":"router: can not parse value 'AA' into field 'Age'"}`+"\n",
		writer.Body.String(),
	)
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}
