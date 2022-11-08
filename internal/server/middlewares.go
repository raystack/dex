package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	gorillamux "github.com/gorilla/mux"
	"github.com/rs/xid"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

const (
	grpcGatewayPrefix = "/api"
	headerRequestID   = "X-Request-Id"
)

type wrappedWriter struct {
	http.ResponseWriter

	Status int
}

func (wr *wrappedWriter) WriteHeader(statusCode int) {
	wr.Status = statusCode
	wr.ResponseWriter.WriteHeader(statusCode)
}

func withOpenCensus() gorillamux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		oc := &ochttp.Handler{
			Handler:          next,
			FormatSpanName:   formatSpanName,
			IsPublicEndpoint: false,
		}
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			route := gorillamux.CurrentRoute(req)

			pathTpl := req.URL.Path
			if route != nil {
				pathTpl, _ = route.GetPathTemplate()
			}

			if strings.HasPrefix(pathTpl, grpcGatewayPrefix) {
				// FIX: figure out a way to extract path-pattern from gateway requests.
				pathTpl = "/api/"
			}

			ctx, _ := tag.New(req.Context(),
				tag.Insert(ochttp.KeyServerRoute, pathTpl),
				tag.Insert(ochttp.Method, req.Method),
			)

			oc.ServeHTTP(wr, req.WithContext(ctx))
		})
	}
}

func requestID() gorillamux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			rid := strings.TrimSpace(req.Header.Get(headerRequestID))
			if rid == "" {
				rid = xid.New().String()
			}

			headers := req.Header.Clone()
			headers.Set(headerRequestID, rid)

			wr.Header().Set(headerRequestID, rid)
			req.Header = headers
			next.ServeHTTP(wr, req)
		})
	}
}

func requestLogger(lg *zap.Logger) gorillamux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			t := time.Now()
			span := trace.FromContext(req.Context())

			clientID, _, _ := req.BasicAuth()
			fields := []zap.Field{
				zap.String("path", req.URL.Path),
				zap.String("method", req.Method),
				zap.String("request_id", req.Header.Get(headerRequestID)),
				zap.String("client_id", clientID),
				zap.String("trace_id", span.SpanContext().TraceID.String()),
			}

			wrapped := &wrappedWriter{ResponseWriter: wr, Status: http.StatusOK}

			var fr http.ResponseWriter
			flusher, ok := wr.(http.Flusher)
			if !ok {
				fr = wrapped
			} else {
				fr = struct {
					*wrappedWriter
					http.Flusher
				}{wrapped, flusher}
			}

			next.ServeHTTP(fr, req)
			fields = append(fields,
				zap.String("response_time", time.Since(t).String()),
				zap.Int("status", wrapped.Status),
			)

			if !is2xx(wrapped.Status) {
				lg.Warn("request handled with non-2xx response", fields...)
			} else {
				lg.Info("request handled", fields...)
			}
		})
	}
}

func formatSpanName(req *http.Request) string {
	route := gorillamux.CurrentRoute(req)

	pathTpl := req.URL.Path
	if route != nil {
		pathTpl, _ = route.GetPathTemplate()
	}

	return fmt.Sprintf("%s %s", req.Method, pathTpl)
}

func is2xx(status int) bool {
	const max2xxCode = 299
	return status >= http.StatusOK && status < max2xxCode
}
