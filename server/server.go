package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/observability"
	"github.com/coolguy1771/wastebin/routes"
	"github.com/coolguy1771/wastebin/storage"
)

const (
	// Server timeout constants.
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	readHeaderTimeout = 5 * time.Second
	maxHeaderBytes    = 1 << 20 // 1MB

	// Shutdown timeout constants.
	shutdownTimeout = 30 * time.Second

	// Storage retry constants.
	storageRetryAttempts = 3
)

// Server represents the application server with its dependencies.
type Server struct {
	config        *config.Config
	db            *gorm.DB
	logger        *log.Logger
	httpServer    *http.Server
	observability *observability.Provider
}

// New creates a new server instance with dependency injection.
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

	obs, err := observability.New(&observabilityConfig, logger.ZapLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	// Initialize and connect to storage with observability
	err = storage.ConnectWithRetry(storageRetryAttempts, obs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Run migrations
	err = storage.Migrate()
	if err != nil {
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

// Start starts the HTTP server and handles graceful shutdown.
func (s *Server) Start() error {
	// Initialize Chi router with routes and observability middleware
	router := routes.AddRoutes(s.observability)

	// Create HTTP server with enhanced security configuration
	s.httpServer = &http.Server{
		Addr:              ":" + s.config.WebappPort,
		Handler:           router,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_AES_128_GCM_SHA256,
			},
			ClientAuth:             tls.NoClientCert,
			ClientCAs:              nil,
			InsecureSkipVerify:     false,
			Certificates:           nil,
			GetCertificate:         nil,
			GetClientCertificate:   nil,
			GetConfigForClient:     nil,
			VerifyPeerCertificate:  nil,
			KeyLogWriter:           nil,
			SessionTicketsDisabled: false,
			SessionTicketKey:       [32]byte{},
			ClientSessionCache:     nil,
			Time:                   nil,
			Rand:                   nil,
			NextProtos:             nil,
		},
	}

	// Setup graceful shutdown
	return s.startWithGracefulShutdown()
}

// startWithGracefulShutdown starts the server and handles graceful shutdown.
func (s *Server) startWithGracefulShutdown() error {
	// Setup channel for OS signals
	idleConnsClosed := make(chan struct{})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		s.logger.Info("Received signal to shutdown server", zap.String("signal", sig.String()))

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Shutdown HTTP server
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		}

		// Close database connections
		err = storage.Close()
		if err != nil {
			s.logger.Error("Error closing database connections", zap.Error(err))
		}

		// Shutdown observability
		err = s.observability.Shutdown(ctx)
		if err != nil {
			s.logger.Error("Error shutting down observability", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	// Start the server with or without TLS
	protocol := "HTTP"
	if s.config.TLSEnabled {
		protocol = "HTTPS"
	}

	s.logger.Info("Starting server",
		zap.String("protocol", protocol),
		zap.String("port", s.config.WebappPort),
		zap.Bool("tls_enabled", s.config.TLSEnabled),
		zap.String("env", func() string {
			if s.config.Dev {
				return "development"
			}

			return "production"
		}()))

	var err error
	if s.config.TLSEnabled {
		err = s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	} else {
		err = s.httpServer.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s server failed: %w", protocol, err)
	}

	// Wait for the server to shutdown gracefully
	<-idleConnsClosed
	s.logger.Info("Server shutdown completed")

	return nil
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	err = storage.Close()
	if err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}

	return nil
}

// HealthCheck performs a comprehensive health check.
func (s *Server) HealthCheck(ctx context.Context) error {
	// Check database health
	err := storage.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Add more health checks as needed (Redis, external services, etc.)

	return nil
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *config.Config {
	return s.config
}

// GetDB returns the database connection.
func (s *Server) GetDB() *gorm.DB {
	return s.db
}

// GetLogger returns the logger instance.
func (s *Server) GetLogger() *log.Logger {
	return s.logger
}
