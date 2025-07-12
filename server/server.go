package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/observability"
	"github.com/coolguy1771/wastebin/routes"
	"github.com/coolguy1771/wastebin/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server represents the application server with its dependencies
type Server struct {
	config        *config.Config
	db            *gorm.DB
	logger        *log.Logger
	httpServer    *http.Server
	observability *observability.Provider
}

// New creates a new server instance with dependency injection
func New() (*Server, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize logger based on config
	logger := log.Default()

	// Initialize observability first
	observabilityConfig := observability.Config{
		Tracing: observability.TracingConfig{
			Enabled:     cfg.TracingEnabled,
			ServiceName: cfg.ServiceName,
			Version:     cfg.ServiceVersion,
			Environment: cfg.Environment,
			Endpoint:    cfg.OTLPTraceEndpoint,
			Headers:     make(map[string]string),
		},
		Metrics: observability.MetricsConfig{
			Enabled:  cfg.MetricsEnabled,
			Endpoint: cfg.OTLPMetricsEndpoint,
			Headers:  make(map[string]string),
			Interval: time.Duration(cfg.MetricsInterval) * time.Second,
		},
	}

	obs, err := observability.New(observabilityConfig, logger.ZapLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	// Initialize and connect to storage with observability
	if err := storage.ConnectWithRetry(3, obs); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Run migrations
	if err := storage.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	server := &Server{
		config:        cfg,
		db:            storage.DBConn,
		logger:        logger,
		observability: obs,
	}

	return server, nil
}

// Start starts the HTTP server and handles graceful shutdown
func (s *Server) Start() error {
	// Initialize Chi router with routes and observability middleware
	router := routes.AddRoutes(s.observability)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    ":" + s.config.WebappPort,
		Handler: router,
		// Add timeouts for security
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Setup graceful shutdown
	return s.startWithGracefulShutdown()
}

// startWithGracefulShutdown starts the server and handles graceful shutdown
func (s *Server) startWithGracefulShutdown() error {
	// Setup channel for OS signals
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		s.logger.Info("Received signal to shutdown server", zap.String("signal", sig.String()))

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown HTTP server
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		}

		// Close database connections
		if err := storage.Close(); err != nil {
			s.logger.Error("Error closing database connections", zap.Error(err))
		}

		// Shutdown observability
		if err := s.observability.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down observability", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	// Start the server
	s.logger.Info("Starting HTTP server",
		zap.String("port", s.config.WebappPort),
		zap.String("env", func() string {
			if s.config.Dev {
				return "development"
			}
			return "production"
		}()))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	// Wait for the server to shutdown gracefully
	<-idleConnsClosed
	s.logger.Info("Server shutdown completed")
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return storage.Close()
}

// HealthCheck performs a comprehensive health check
func (s *Server) HealthCheck(ctx context.Context) error {
	// Check database health
	if err := storage.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Add more health checks as needed (Redis, external services, etc.)

	return nil
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *config.Config {
	return s.config
}

// GetDB returns the database connection
func (s *Server) GetDB() *gorm.DB {
	return s.db
}

// GetLogger returns the logger instance
func (s *Server) GetLogger() *log.Logger {
	return s.logger
}
