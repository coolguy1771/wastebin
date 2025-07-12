package main

import (
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/server"
	"go.uber.org/zap"
)

func main() {
	// Create server instance with dependency injection
	srv, err := server.New()
	if err != nil {
		log.Fatal("Failed to create server", zap.Error(err))
	}

	// Start the server
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
	}
}
