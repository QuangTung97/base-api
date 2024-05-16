package router

import (
	"encoding/json"
	"learn-gin/pkg/urls"
	"net/http"
)

func APIGet[T any, Req any, Resp any](
	r *Router, pattern urls.Path[T],
	handler func(ctx Context, req Req) (Resp, error),
) {
	apiDoAction(r, r.mux.Get, false, pattern, handler)
}

func APIPost[T any, Req any, Resp any](
	r *Router, pattern urls.Path[T],
	handler func(ctx Context, req Req) (Resp, error),
) {
	apiDoAction(r, r.mux.Post, true, pattern, handler)
}

func apiDoAction[T any, Req any, Resp any](
	r *Router,
	registerFunc func(pattern string, handler http.HandlerFunc),
	decodeBody bool,
	pattern urls.Path[T],
	handler func(ctx Context, req Req) (Resp, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	genericHandler := func(ctx Context, req any) (resp any, err error) {
		return handler(ctx, req.(Req))
	}
	genericHandler = r.wrapHandler(genericHandler) // TODO Testing

	registerFunc(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		var req Req

		if decodeBody {
			if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
				r.writeAPIResp(ctx, nil, err, http.StatusBadRequest)
				return
			}
		}

		if err := doAssignParams(ctx, &req, pattern); err != nil {
			r.writeAPIResp(ctx, nil, err, http.StatusBadRequest)
			return
		}

		respBody, err := genericHandler(ctx, req)
		r.writeAPIResp(ctx, respBody, err, http.StatusInternalServerError)
	})
}

func (r *Router) writeAPIResp(ctx Context, respBody any, err error, status int) {
	writer := ctx.writer
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		if ctx.state.code != 0 {
			status = ctx.state.code
		}
		writer.WriteHeader(status)
		_ = json.NewEncoder(writer).Encode(ErrorBody{
			Error: err.Error(),
		})
		return
	}

	_ = json.NewEncoder(writer).Encode(respBody)
}
