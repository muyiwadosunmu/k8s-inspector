package main

import (
	"net/http"

	"muyiwadosunmu/k8s-inspector/internal/pkg/web"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	sugar := app.logger.Sugar()

	webApp := web.NewApp(
		mux, 
		app.webErrorHandler,
		web.Logger(sugar),
		web.Errors(sugar),
		web.Panics(),
	)

	webApp.Handle(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	return mux
}
