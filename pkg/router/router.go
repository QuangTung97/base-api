package router

import (
	"github.com/go-chi/chi/v5"
	"slices"
)

type Router struct {
	mux         *chi.Mux
	middlewares []MiddlewareFunc
	wrapFunc    func(handler GenericHandler) GenericHandler
}

func NewRouter() *Router {
	mux := chi.NewRouter()
	return &Router{
		mux: mux,
	}
}

type GenericHandler = func(ctx Context, req any) (resp any, err error)

type MiddlewareFunc = func(handler GenericHandler) GenericHandler

func (r *Router) WithMiddlewares(
	middlewares ...MiddlewareFunc,
) *Router {
	newR := *r
	newR.middlewares = slices.Clone(r.middlewares)
	newR.middlewares = append(newR.middlewares, middlewares...)
	return &newR
}

func (r *Router) Mux() *chi.Mux {
	return r.mux
}

func (r *Router) wrapHandler(handler GenericHandler) GenericHandler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	return handler
}

type ErrorBody struct {
	Error string `json:"error"`
}
