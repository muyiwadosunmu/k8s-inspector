package main

import (
	"context"
	"net/http"

	"muyiwadosunmu/k8s-inspector/internal/pkg/web"
)

func (app *application) healthcheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	data := envelope{
		"status":      "available",
		"environment": "development",
		"version":     "1.0.0",
	}
	return web.Respond(ctx, w, data, http.StatusOK)
}
