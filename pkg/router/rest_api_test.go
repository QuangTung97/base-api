package router

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

type userGetResponse struct {
	UserID   userID `json:"user_id"`
	Username string `json:"username"`
}

func TestAPIGet_Success(t *testing.T) {
	r := NewRouter()

	var inputReq userGetRequest
	APIGet(r, userPath, func(ctx Context, req userGetRequest) (userGetResponse, error) {
		inputReq = req
		return userGetResponse{
			UserID:   234,
			Username: "username01",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userGetRequest{
		UserID: 123,
	}, inputReq)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, `{"user_id":234,"username":"username01"}`+"\n", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}

func TestAPIGet_Not_A_Sub_Struct(t *testing.T) {
	r := NewRouter()

	type invalidRequest struct {
		Name string `json:"name"`
	}

	assert.PanicsWithValue(t, "missing field 'UserID' in struct 'invalidRequest'", func() {
		APIGet(r, userPath, func(ctx Context, req invalidRequest) (userGetResponse, error) {
			return userGetResponse{}, nil
		})
	})
}

func TestAPIGet_With_Error(t *testing.T) {
	r := NewRouter()

	var inputReq userGetRequest
	APIGet(r, userPath, func(ctx Context, req userGetRequest) (userGetResponse, error) {
		inputReq = req
		return userGetResponse{}, errors.New("some handler error")
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

func TestAPIGet_With_Invalid_Query_Param(t *testing.T) {
	r := NewRouter()

	APIGet(r, userPath, func(ctx Context, req userGetRequest) (template.HTML, error) {
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

func TestAPIPost_Success(t *testing.T) {
	r := NewRouter()

	var inputReq userPostRequest
	APIPost(r, userPath, func(ctx Context, req userPostRequest) (userGetResponse, error) {
		inputReq = req
		return userGetResponse{
			UserID:   2345,
			Username: "some-user",
		}, nil
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
	assert.Equal(t, `{"user_id":2345,"username":"some-user"}`+"\n", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}

func TestAPIPost_Parse_JSON_Request_Error(t *testing.T) {
	r := NewRouter()

	APIPost(r, userPath, func(ctx Context, req userPostRequest) (userGetResponse, error) {
		return userGetResponse{}, nil
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

func TestAPIGet_Success_With_Middlewares(t *testing.T) {
	r := NewRouter()

	var steps []string

	r = r.WithMiddlewares(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (resp any, err error) {
				steps = append(steps, "step01")
				return handler(ctx, req)
			}
		},
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (resp any, err error) {
				steps = append(steps, "step02")
				return handler(ctx, req)
			}
		},
	)

	var inputReq userGetRequest
	APIGet(r, userPath, func(ctx Context, req userGetRequest) (userGetResponse, error) {
		inputReq = req
		return userGetResponse{
			UserID:   234,
			Username: "username01",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userGetRequest{
		UserID: 123,
	}, inputReq)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, `{"user_id":234,"username":"username01"}`+"\n", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())

	assert.Equal(t, []string{"step01", "step02"}, steps)
}

func TestAPIGet_Success_With_Middlewares_Return_Error(t *testing.T) {
	r := NewRouter()

	r = r.WithMiddlewares(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (resp any, err error) {
				ctx.SetStatusCode(http.StatusForbidden)
				return nil, errors.New("some middleware error")
			}
		},
	)

	var inputReq userGetRequest
	APIGet(r, userPath, func(ctx Context, req userGetRequest) (userGetResponse, error) {
		inputReq = req
		return userGetResponse{
			UserID:   234,
			Username: "username01",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	writer := httptest.NewRecorder()
	r.Mux().ServeHTTP(writer, req)

	assert.Equal(t, userGetRequest{}, inputReq)

	assert.Equal(t, http.StatusForbidden, writer.Code)
	assert.Equal(t, `{"error":"some middleware error"}`+"\n", writer.Body.String())
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
	}, writer.Header())
}
