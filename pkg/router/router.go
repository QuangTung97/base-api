package router

import (
	"github.com/go-chi/chi/v5"
	"html/template"
)

type PathPattern string

type HTMLHandler[T any] func(ctx *Context, req T) (template.HTML, error)

type Router struct {
	mux *chi.Mux
}

func NewRouter() *Router {
	mux := chi.NewRouter()
	return &Router{
		mux: mux,
	}
}

func (r *Router) Mux() *chi.Mux {
	return r.mux
}

type ErrorBody struct {
	Error string `json:"error"`
}
