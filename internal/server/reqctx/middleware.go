package reqctx

import (
	"net/http"
	"strings"
)

// shield header names.
// Refer https://github.com/goto/shield
const (
	headerUserEmail = "X-Auth-Email"
	headerShieldID  = "X-Shield-User-Id"
	headerRequestID = "X-Request-Id"
)

func WithRequestCtx() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := strings.TrimSpace(r.Header.Get(headerRequestID))

			ctx := withReqCtx(r.Context(), ReqCtx{
				UserID:    r.Header.Get(headerShieldID),
				UserEmail: r.Header.Get(headerUserEmail),
				RequestID: reqID,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
