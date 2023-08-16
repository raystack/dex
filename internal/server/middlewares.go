package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/xid"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

const headerRequestID = "X-Request-Id"

type curRouteFn func(r *http.Request) string

type middleware func(http.Handler) http.Handler

type wrappedWriter struct {
	http.ResponseWriter

	Status int
}

func (wr *wrappedWriter) WriteHeader(statusCode int) {
	wr.Status = statusCode
	wr.ResponseWriter.WriteHeader(statusCode)
}

func newRelicAPM(nrApp *newrelic.Application, curRoute curRouteFn) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := curRoute(r)
			if route == "" {
				route = r.URL.Path
			}

			txn := nrApp.StartTransaction(r.Method + " " + route)
			defer txn.End()

			w = txn.SetWebResponse(w)
			txn.SetWebRequestHTTP(r)
			r = newrelic.RequestWithTransactionContext(r, txn)

			next.ServeHTTP(w, r)
		})
	}
}

func withOpenCensus(curRoute curRouteFn) middleware {
	return func(next http.Handler) http.Handler {
		oc := &ochttp.Handler{
			Handler: next,
			FormatSpanName: func(r *http.Request) string {
				route := curRoute(r)
				if route == "" {
					route = r.URL.Path
				}
				return fmt.Sprintf("%s %s", r.Method, route)
			},
			IsPublicEndpoint: false,
		}

		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			pathTpl := curRoute(req)
			if pathTpl == "" {
				pathTpl = req.URL.Path
			}

			ctx, _ := tag.New(req.Context(),
				tag.Insert(ochttp.KeyServerRoute, pathTpl),
				tag.Insert(ochttp.Method, req.Method),
			)

			oc.ServeHTTP(wr, req.WithContext(ctx))
		})
	}
}

func requestID() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			rid := strings.TrimSpace(req.Header.Get(headerRequestID))
			if rid == "" {
				rid = xid.New().String()

				headers := req.Header.Clone()
				headers.Set(headerRequestID, rid)
				req.Header = headers
			}

			wr.Header().Set(headerRequestID, rid)
			next.ServeHTTP(wr, req)
		})
	}
}

func requestLogger(lg *zap.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			t := time.Now()
			span := trace.FromContext(req.Context())

			clientID, _, _ := req.BasicAuth()

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

			if req.URL.Path == "/ping" {
				return
			}

			fields := []zap.Field{
				zap.String("request_path", req.URL.Path),
				zap.String("request_method", req.Method),
				zap.String("request_id", req.Header.Get(headerRequestID)),
				zap.String("client_id", clientID),
				zap.String("trace_id", span.SpanContext().TraceID.String()),
				zap.String("response_time", time.Since(t).String()),
				zap.Int("status", wrapped.Status),
			}

			switch req.Method {
			case http.MethodGet:
				break
			default:
				buf, err := io.ReadAll(req.Body)
				if err != nil {
					lg.Debug("error reading request body: %v", zap.String("error", err.Error()))
				} else if len(buf) > 0 {
					dst := &bytes.Buffer{}
					err := json.Compact(dst, buf)
					if err != nil {
						lg.Debug("error json compacting request body: %v", zap.String("error", err.Error()))
					} else {
						fields = append(fields, zap.String("request_body", dst.String()))
					}
				}

				reader := io.NopCloser(bytes.NewBuffer(buf))
				req.Body = reader
			}

			if !is2xx(wrapped.Status) {
				lg.Warn("request handled with non-2xx response", fields...)
			} else {
				lg.Info("request handled", fields...)
			}
		})
	}
}

func currentRouteGetter(router chi.Router) func(r *http.Request) string {
	return func(r *http.Request) string {
		rCtx := chi.NewRouteContext()
		if !router.Match(rCtx, r.Method, r.URL.Path) {
			return ""
		}
		return rCtx.RoutePattern()
	}
}

func is2xx(status int) bool {
	const max2xxCode = 299
	return status >= http.StatusOK && status < max2xxCode
}
