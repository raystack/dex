package reqctx

import "context"

type reqCtxKeyType string

var reqCtxKey = reqCtxKeyType("request_context")

type ReqCtx struct {
	UserID    string
	UserEmail string
	RequestID string
}

func withReqCtx(ctx context.Context, reqCtx ReqCtx) context.Context {
	return context.WithValue(ctx, reqCtxKey, reqCtx)
}

// From returns the ReqCtx from the given go context. Returns zero-value
// if not available.
func From(ctx context.Context) ReqCtx {
	rCtx, _ := ctx.Value(reqCtxKey).(ReqCtx)
	return rCtx
}
