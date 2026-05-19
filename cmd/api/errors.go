package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"muyiwadosunmu/k8s-inspector/internal/pkg/web"
)

// webErrorHandler is the standard error handler for our web framework.
func (app *application) webErrorHandler(ctx context.Context, w http.ResponseWriter, err error) {
	var status int
	var message interface{}
	var fields map[string]string

	if webErr, ok := err.(*web.Error); ok {
		status = webErr.Status
		message = webErr.Message
		if len(webErr.Fields) > 0 {
			fields = webErr.Fields
		}
	} else {
		status = http.StatusInternalServerError
		message = "the server encountered a problem and could not process your request"
		app.logger.Error("Internal server error", zap.Error(err))
	}

	env := envelope{"error": message}
	if len(fields) > 0 {
		env["fields"] = fields
	}
	if writeErr := app.writeJSON(w, status, env, nil); writeErr != nil {
		app.logger.Error("Failed to write error response", zap.Error(writeErr))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
