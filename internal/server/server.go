package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
)

func StartHttpServer(logger hclog.Logger, defaultRouter http.Handler) {
	port := 8080

	// create a new server
	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", port),                              // configure the bind address
		Handler:  defaultRouter,                                         // set the default handler
		ErrorLog: logger.StandardLogger(&hclog.StandardLoggerOptions{}), // set the logger for the server
	}

	// start the server
	go func() {
		logger.Info(fmt.Sprintf("Starting server on port: %d", port))

		err := server.ListenAndServe()
		if err != nil {
			logger.Error("Error starting server", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	// Block until a signal is received.
	sig := <-c
	logger.Info("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
