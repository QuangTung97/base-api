package router

import (
	"encoding/json"
	"html/template"
	"learn-gin/pkg/urls"
	"net/http"
)

func HTMLGet[T any, Req any](
	r *Router, pattern urls.Path[T],
	handler func(ctx *Context, req Req) (template.HTML, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	r.mux.Get(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		var req Req

		if err := doAssignParams(ctx, &req, pattern); err != nil {
			r.writeHTMLResp(ctx, "", err, http.StatusBadRequest)
			return
		}

		respBody, err := handler(ctx, req)
		r.writeHTMLResp(ctx, respBody, err, http.StatusInternalServerError)
	})
}

func HTMLPost[T any, Req any](
	r *Router, pattern urls.Path[T],
	handler func(ctx *Context, req Req) (template.HTML, error),
) {
	htmlChangeAction(r, r.mux.Post, pattern, handler)
}

func htmlChangeAction[T any, Req any](
	r *Router,
	actionFn func(pattern string, handler http.HandlerFunc),
	pattern urls.Path[T],
	handler func(ctx *Context, req Req) (template.HTML, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	actionFn(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)

		var req Req
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			r.writeHTMLResp(ctx, "", err, http.StatusBadRequest)
			return
		}

		if err := doAssignParams(ctx, &req, pattern); err != nil {
			r.writeHTMLResp(ctx, "", err, http.StatusBadRequest)
			return
		}

		respBody, err := handler(ctx, req)
		r.writeHTMLResp(ctx, respBody, err, http.StatusInternalServerError)
	})
}

func (r *Router) writeHTMLResp(ctx *Context, respBody template.HTML, err error, status int) {
	writer := ctx.Writer

	if err != nil {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(status)
		_ = json.NewEncoder(writer).Encode(ErrorBody{
			Error: err.Error(),
		})
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = writer.Write([]byte(respBody))
}
