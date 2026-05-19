package main

import (
	"net/http"
)

// routes method returns an http.Handler that routes incoming requests.
func (app *application) routes() http.Handler {
	// Initialize a new serve mux.
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	return mux
}
