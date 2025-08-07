package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/handlers"
	"github.com/coolguy1771/wastebin/observability"
)

const (
	// Rate limiting constants.
	requestsPerMinute = 100

	// CORS constants.
	corsMaxAge = 300 // Maximum value not ignored by any of major browsers

	// API version constants.
	defaultAPIVersion = "v1"
	apiVersionV2      = "v2"
)

// initChiRouter initializes and configures a Chi router with core middleware for
// security, observability, logging, panic recovery, rate limiting, and API versioning.
// The router includes security headers, request size limiting, security audit logging,
// optional basic authentication, CSRF protection, and a heartbeat endpoint.
// If an observability provider is supplied, its HTTP middleware is applied for tracing and metrics.
// Returns the fully configured Chi router instance.
func initChiRouter(obs *observability.Provider) *chi.Mux {
	router := chi.NewRouter()

	// Apply core middlewares
	router.Use(middleware.RequestID)

	// Add observability middleware first
	if obs != nil {
		router.Use(obs.HTTPMiddleware())
	}

	// Security middleware stack
	router.Use(handlers.SecurityHeadersMiddleware)                              // Add comprehensive security headers
	router.Use(handlers.RequestSizeLimitMiddleware(config.Conf.MaxRequestSize)) // Global request size limits
	router.Use(handlers.SecurityAuditMiddleware)                                // Security audit logging
	router.Use(handlers.BasicAuthMiddleware)                                    // Optional basic authentication
	router.Use(handlers.CSRFProtectionMiddleware)                               // CSRF protection for web forms

	router.Use(middleware.Logger)    // Log the start and end of each request with the elapsed processing time
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.Heartbeat("/healthz"))

	// Add rate limiting
	router.Use(httprate.LimitByIP(requestsPerMinute, 1*time.Minute)) // 100 requests per minute per IP

	// Add API versioning middleware
	router.Use(APIVersionMiddleware)

	return router
}

// AddRoutes configures the Chi router with all middleware, CORS settings, API endpoints, and static file routes.
// Returns the fully initialized router ready to serve HTTP requests.
func AddRoutes(obs *observability.Provider) *chi.Mux {
	router := initChiRouter(obs)

	// Apply CORS middleware globally with secure configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:     getAllowedOrigins(),
		AllowedMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:     []string{"X-Request-ID", "X-API-Version"},
		AllowCredentials:   false,
		MaxAge:             corsMaxAge,
		OptionsPassthrough: false,
		Debug:              false,
	}))

	// Set up API routes
	setupAPIRoutes(router, obs)

	// Serve static files and SPA
	setupStaticRoutes(router)

	return router
}

// setupAPIRoutes configures the API routes for the application.
func setupAPIRoutes(router *chi.Mux, obs *observability.Provider) {
	router.Route("/api/v1", func(api chi.Router) {
		api.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			// Respond with a simple message for version check
			jsonResponse(w, map[string]string{"message": "🐣 v1"})
		})

		api.Get("/paste/{uuid}", handlers.GetPaste)       // Retrieve paste by UUID
		api.Post("/paste", handlers.CreatePaste)          // Create a new paste
		api.Delete("/paste/{uuid}", handlers.DeletePaste) // Delete a paste by UUID
	})

	// Raw paste endpoint is handled in setupStaticRoutes to avoid conflicts

	// Health check endpoint for monitoring
	if obs != nil {
		router.Get("/health", obs.HealthCheckMiddleware())
	} else {
		router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, map[string]string{"status": "healthy", "service": "wastebin"})
		})
	}

	// Database health check endpoint
	router.Get("/health/db", handlers.DatabaseHealthCheck)
}

// setupStaticRoutes sets up routes to serve static files and SPA.
func setupStaticRoutes(router *chi.Mux) {
	// Determine static file path based on environment
	staticPath := "./web/dist/"
	if !config.Conf.Dev {
		staticPath = "/web/"
	}

	// Serve static files (assets, etc.)
	fileServer(router, "/assets/", http.Dir(staticPath+"assets/"))

	// Handle SPA routes by serving index.html
	router.Get("/", serveSPA)
	router.Get("/about", serveSPA)
	router.Get("/paste/new", serveSPA)
	router.Get("/paste/{uuid}", serveSPA)
	router.Get("/paste/{uuid}/raw", handlers.GetRawPaste) // This should serve raw text, not SPA
}

// serveSPA serves the Single Page Application (SPA) index file.
func serveSPA(w http.ResponseWriter, req *http.Request) {
	indexFilePath := "./web/dist/index.html"
	if !config.Conf.Dev {
		indexFilePath = "/web/index.html"
	}

	http.ServeFile(w, req, indexFilePath)
}

// jsonResponse sends a JSON response with the given data.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fileServer sets up a file server for static files.
func fileServer(router chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		router.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}

	fs := http.StripPrefix(path, http.FileServer(root))
	router.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fs.ServeHTTP(w, req)
	}))
}

// getAllowedOrigins determines the list of allowed CORS origins based on configuration and environment.
// If explicit origins are configured, it returns those; in development mode, it returns common localhost origins;
// otherwise, it returns an empty list for maximum security.
func getAllowedOrigins() []string {
	// Check for explicitly configured origins
	if config.Conf.AllowedOrigins != "" {
		origins := strings.Split(config.Conf.AllowedOrigins, ",")
		// Trim whitespace from each origin
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}

		return origins
	}

	if config.Conf.Dev {
		// Allow localhost and common development ports
		return []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"https://localhost:3000",
			"https://localhost:5173",
		}
	}

	// In production, default to no origins if not explicitly configured
	// This is more secure than allowing all origins
	return []string{}
}

// APIVersionMiddleware handles API versioning.
func APIVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Default to v1 if no version specified
		version := defaultAPIVersion

		// Check for version in header
		if headerVersion := req.Header.Get("X-Api-Version"); headerVersion != "" {
			version = headerVersion
		}

		// Check for version in Accept header (content negotiation)
		if accept := req.Header.Get("Accept"); accept != "" {
			if strings.Contains(accept, "application/vnd.wastebin.v2+json") {
				version = apiVersionV2
			} else if strings.Contains(accept, "application/vnd.wastebin.v1+json") {
				version = defaultAPIVersion
			}
		}

		// Add version to response headers
		w.Header().Set("X-Api-Version", version)

		// Add version to context for handlers to use
		ctx := context.WithValue(req.Context(), contextKey("api-version"), version)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string
