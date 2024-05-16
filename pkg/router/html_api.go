package router

import (
	"encoding/json"
	"html/template"
	"learn-gin/pkg/urls"
	"net/http"
)

func HTMLGet[T any, Req any](
	r *Router, pattern urls.Path[T],
	handler func(ctx Context, req Req) (template.HTML, error),
) {
	htmlDoAction(r, r.mux.Get, false, pattern, handler)
}

func HTMLPost[T any, Req any](
	r *Router, pattern urls.Path[T],
	handler func(ctx Context, req Req) (template.HTML, error),
) {
	htmlDoAction(r, r.mux.Post, true, pattern, handler)
}

func htmlDoAction[T any, Req any](
	r *Router,
	registerFunc func(pattern string, handler http.HandlerFunc),
	decodeBody bool,
	pattern urls.Path[T],
	handler func(ctx Context, req Req) (template.HTML, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	genericHandler := func(ctx Context, req any) (resp any, err error) {
		return handler(ctx, req.(Req))
	}
	genericHandler = r.wrapHandler(genericHandler)

	registerFunc(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)

		var req Req
		if decodeBody {
			if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
				r.writeHTMLResp(ctx, "", err, http.StatusBadRequest)
				return
			}
		}

		if err := doAssignParams(ctx, &req, pattern); err != nil {
			r.writeHTMLResp(ctx, "", err, http.StatusBadRequest)
			return
		}

		respBody, err := genericHandler(ctx, req)
		r.writeHTMLResp(ctx, respBody.(template.HTML), err, http.StatusInternalServerError)
	})
}

func (r *Router) writeHTMLResp(ctx Context, respBody template.HTML, err error, status int) {
	writer := ctx.writer

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
