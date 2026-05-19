package web

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// App is the entrypoint for the web framework.
type App struct {
	mux          *http.ServeMux
	mw           []Middleware
	errorHandler func(context.Context, http.ResponseWriter, error)
	tracer       trace.Tracer
}

// NewApp creates a new App with global middleware and an error handler.
func NewApp(mux *http.ServeMux, errorHandler func(context.Context, http.ResponseWriter, error), mw ...Middleware) *App {
	return &App{
		mux:          mux,
		mw:           mw,
		errorHandler: errorHandler,
		tracer:       otel.GetTracerProvider().Tracer("web-app"),
	}
}

// Handle sets a handler for a given method and pattern.
func (a *App) Handle(method string, pattern string, handler Handler, mw ...Middleware) {
	// 1. Wrap with route-specific middleware.
	handler = wrapMiddleware(mw, handler)

	// 2. Wrap with app-level global middleware.
	handler = wrapMiddleware(a.mw, handler)

	// Final handler function.
	h := func(w http.ResponseWriter, r *http.Request) {
		// Start a span for every request (Observability)
		ctx, span := a.tracer.Start(r.Context(), r.Method+" "+pattern)
		defer span.End()

		// Set the context values for this request.
		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Now:     time.Now().UTC(),
		}
		ctx = context.WithValue(ctx, key, &v)

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		)

		if r.Method != method && method != "" {
			a.errorHandler(ctx, w, NewRequestError(http.StatusMethodNotAllowed, "Method not allowed"))
			return
		}

		if err := handler(ctx, w, r); err != nil {
			a.errorHandler(ctx, w, err)
			return
		}
	}

	a.mux.HandleFunc(method+" "+pattern, h)
}

func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}
	return handler
}
