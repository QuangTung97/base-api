package router

import (
	"context"
	"net/http"
)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
}

func NewContext(writer http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Request: req,
		Writer:  writer,
	}
}

func (c *Context) Context() context.Context {
	return c.Request.Context()
}
