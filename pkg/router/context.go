package router

import (
	"context"
	"net/http"
)

type requestState struct {
	code int
}

type Context struct {
	request *http.Request
	writer  http.ResponseWriter
	state   *requestState
}

func NewContext(writer http.ResponseWriter, req *http.Request) Context {
	return Context{
		request: req,
		writer:  writer,
		state:   &requestState{},
	}
}

func (c Context) Context() context.Context {
	return c.request.Context()
}

func (c Context) WithContext(ctx context.Context) Context {
	newCtx := c
	newCtx.request = c.request.WithContext(ctx)
	return newCtx
}

func (c Context) SetStatusCode(code int) {
	c.state.code = code
}
