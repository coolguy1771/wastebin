package main

import (
	"go.uber.org/zap"

	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/server"
)

func main() {
	log.Setup()

	// Create server instance with dependency injection
	srv, err := server.New()
	if err != nil {
		log.Fatal("Failed to create server", zap.Error(err))
	}

	// Start the server
	startErr := srv.Start()
	if startErr != nil {
		log.Fatal("Server failed to start", zap.Error(startErr))
	}
}
