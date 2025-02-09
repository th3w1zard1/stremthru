package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type reqCtxKey struct{}

type ReqCtx struct {
	StartTime time.Time
	RequestId string
	Error     error
	URL       string
	Log       *slog.Logger
}

func (ctx *ReqCtx) RedactURLPathValues(r *http.Request, names ...string) {
	for _, name := range names {
		if value := r.PathValue(name); value != "" {
			ctx.URL = strings.Replace(ctx.URL, value, "{"+name+"}", 1)
		}
	}
}

func SetReqCtx(r *http.Request, reqCtx *ReqCtx) *http.Request {
	ctx := context.WithValue(r.Context(), reqCtxKey{}, reqCtx)
	return r.WithContext(ctx)
}

func GetReqCtx(r *http.Request) *ReqCtx {
	return r.Context().Value(reqCtxKey{}).(*ReqCtx)
}
