package router

import (
	"github.com/stretchr/testify/assert"
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
	APIGet(r, userPath, func(ctx *Context, req userGetRequest) (userGetResponse, error) {
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
