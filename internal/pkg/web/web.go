package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Handler is the signature used by all API handlers.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Middleware defines the signature for a middleware function.
type Middleware func(Handler) Handler

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// GetValues retrieves the Values from the context.
func GetValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, errors.New("web values not found in context")
	}
	return v, nil
}

type ctxKey int

const key ctxKey = 1

// Respond sends a JSON response to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	ctx, span := otel.GetTracerProvider().Tracer("web").Start(ctx, "web.respond")
	defer span.End()

	// Set the status code for the logger to synthetically access.
	if v, err := GetValues(ctx); err == nil {
		v.StatusCode = statusCode
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Use MarshalIndent for prettier formatting, matched to your helpers.go
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(js); err != nil {
		return err
	}

	span.SetAttributes(attribute.Int("http.status_code", statusCode))
	return nil
}

// Decode reads JSON from the request into a value, including strict validation.
func Decode(r *http.Request, w http.ResponseWriter, dst any) error {
	_, span := otel.GetTracerProvider().Tracer("web").Start(r.Context(), "web.decode")
	defer span.End()

	// Match your helpers.go 512KB limit
	const maxBytes = 1_048_576 >> 1
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		default:
			return err
		}
	}
	return nil
}

// GetTraceID retrieves the trace ID from the context for observability.
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// Param retrieves a path parameter from the request using http.ServeMux patterns.
func Param(r *http.Request, name string) string {
	return r.PathValue(name)
}
