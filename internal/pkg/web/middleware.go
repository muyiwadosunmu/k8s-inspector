package web

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

// Logger writes information about every request to the log.
func Logger(log *zap.SugaredLogger) Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v, err := GetValues(ctx)
			if err != nil {
				return NewShutdownError("web value missing from context")
			}

			log.Infow("request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
			)

			err = handler(ctx, w, r)

			log.Infow("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
				"statuscode", v.StatusCode,
				"since", time.Since(v.Now),
			)

			return err
		}
	}
}

// Errors handles errors coming out of the call chain.
func Errors(log *zap.SugaredLogger) Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("error", "traceid", GetTraceID(ctx), "message", err)

				if re, ok := IsRequestError(err); ok {
					er := ErrorResponse{
						Error:  re.Message,
						Fields: re.Fields,
					}
					if err := Respond(ctx, w, er, re.Status); err != nil {
						return err
					}
					return nil
				}

				// If it's not a trusted error, respond with a generic 500
				if err := Respond(ctx, w, ErrorResponse{Error: "Internal Server Error"}, http.StatusInternalServerError); err != nil {
					return err
				}
			}
			return nil
		}
	}
}

// Panics recovers from panics and converts them to errors.
func Panics() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					trace := make([]byte, 1024)
					runtime.Stack(trace, false)
					err = fmt.Errorf("PANIC RECOVERED: %v\n%s", rec, trace)
				}
			}()

			return handler(ctx, w, r)
		}
	}
}
