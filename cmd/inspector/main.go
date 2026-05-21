package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"muyiwadosunmu/k8s-inspector/internal/k8s"
	"muyiwadosunmu/k8s-inspector/internal/pkg/web"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

/*
curl http://localhost:3000/healthz
curl http://localhost:3000/summary
curl "http://localhost:3000/pods?namespace=default"
*/

type config struct {
	port int
	env  string
}

type application struct {
	Config    config
	logger    *zap.Logger
	k8sClient interface{}
	wg        sync.WaitGroup
}

// build metadata injected at build time via -ldflags
var buildTime string
var gitCommit string

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	// Create the Kubernetes client
	k8sClient, err := k8s.NewClient("")
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", zap.Error(err))
	}

	app := &application{
		Config: config{
			port: 3000,
			env:  "development",
		},
		logger:    logger,
		k8sClient: k8sClient,
	}

	err = app.serve()
	if err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}

func (app *application) serve() error {
	// Declare a HTTP server
	srv := &http.Server{
		Addr:         ":3000",
		Handler:      otelhttp.NewHandler(app.routes(), "inspector-server"),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Create a shutdownError channel
	shutdownError := make(chan error)

	go func() {
		// Create a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)
		// Listen for SIGINT and SIGTERM signals
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// Read the signal from the quit channel
		s := <-quit
		// Log shutdown signal
		app.logger.Info("shutting down server", zap.String("signal", s.String()))
		// Create a context with a 10-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		err := srv.Shutdown(ctx)
		shutdownError <- err
	}()

	app.logger.Info("starting server", zap.String("addr", srv.Addr))

	// Start the server in a goroutine to prevent blocking
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdownError <- err
		}
	}()

	// Wait for shutdown or error
	err := <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("server stopped")
	return nil
}

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

	webApp.Handle(http.MethodGet, "/healthz", app.healthzHandler)
	webApp.Handle(http.MethodGet, "/summary", app.summaryHandler)
	webApp.Handle(http.MethodGet, "/pods", app.podsHandler)

	return mux
}

func (app *application) healthzHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	data := map[string]string{
		"status":    "ok",
		"buildTime": buildTime,
		"gitCommit": gitCommit,
	}
	return web.Respond(ctx, w, data, http.StatusOK)
}

// webErrorHandler is the centralized error handler for the web framework
func (app *application) webErrorHandler(ctx context.Context, w http.ResponseWriter, err error) {
	var status int
	var message any
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

	errResponse := map[string]interface{}{
		"error": message,
	}
	if len(fields) > 0 {
		errResponse["fields"] = fields
	}

	if writeErr := web.Respond(ctx, w, errResponse, status); writeErr != nil {
		app.logger.Error("Failed to write error response", zap.Error(writeErr))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
