package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/go-chi/chi/v5"

	"github.com/coolguy1771/wastebin/routes"
	"go.uber.org/zap"
)

func main() {
	config.Load()

	err := storage.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", zap.Error(err))
	}

	defer storage.Close()

	err = storage.Migrate()
	if err != nil {
		log.Fatal("Error migrating the database", zap.Error(err))
	}

	r := chi.NewRouter()

	routes.AddRoutes(r)

	log.Info("Starting the server", zap.String("port", config.Conf.WebappPort))

	// The HTTP Server
	server := &http.Server{Addr: ":" + config.Conf.WebappPort, Handler: r}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Use a separate goroutine to listen for signals and shutdown the server gracefully
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			log.Info("Received signal to shutdown server", zap.String("signal", shutdownCtx.Err().Error()))
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal("error", zap.Error(err))
		}
		serverStopCtx()
	}()

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("error", zap.Error(err))
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
