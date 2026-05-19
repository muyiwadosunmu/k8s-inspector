package main

import (
	"log"
	"net/http"
	"os"
)

type application struct {
	logger *log.Logger
}

func main() {

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Declare an instance of the application struct, containing the pointer to our logger.
	app := &application{
		logger: logger,
	}

	srv := &http.Server{
		Addr:     ":3000",
		Handler:  app.routes(),
		ErrorLog: logger,
	}

	// Start the HTTP server.
	logger.Printf("starting server on %s", srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
