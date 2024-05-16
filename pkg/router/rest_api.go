package router

import (
	"encoding/json"
	"learn-gin/pkg/urls"
	"net/http"
)

func APIGet[T any, Req any, Resp any](
	r *Router, pattern urls.Path[T],
	handler func(ctx *Context, req Req) (Resp, error),
) {
	var testPathVal T
	var testReqVal Req
	urls.CheckIsSubStruct(testReqVal, testPathVal)

	r.mux.Get(pattern.GetPattern(), func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		var req Req

		if err := doAssignParams(ctx, &req, pattern); err != nil {
			r.writeAPIResp(ctx, nil, err, http.StatusBadRequest)
			return
		}

		respBody, err := handler(ctx, req)
		r.writeAPIResp(ctx, respBody, err, http.StatusInternalServerError)
	})
}

func (r *Router) writeAPIResp(ctx *Context, respBody any, err error, status int) {
	writer := ctx.Writer
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		writer.WriteHeader(status)
		_ = json.NewEncoder(writer).Encode(ErrorBody{
			Error: err.Error(),
		})
		return
	}

	_ = json.NewEncoder(writer).Encode(respBody)
}
