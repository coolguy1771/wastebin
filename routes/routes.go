package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/handlers"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/observability"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func initChiRouter(obs *observability.Provider) *chi.Mux {
	r := chi.NewRouter()

	// Apply core middlewares
	r.Use(middleware.RequestID)

	// Add observability middleware first
	if obs != nil {
		r.Use(obs.HTTPMiddleware())
	}

	r.Use(handlers.ZapLogger(log.Default())) // Log the start and end of each request with the elapsed processing time
	r.Use(middleware.Recoverer)              // Recover from panics without crashing server
	r.Use(middleware.Heartbeat("/healthz"))

	// Add security headers
	r.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	r.Use(middleware.SetHeader("X-Frame-Options", "DENY"))
	r.Use(middleware.SetHeader("X-XSS-Protection", "1; mode=block"))

	// Add rate limiting
	r.Use(httprate.LimitByIP(100, 1*time.Minute)) // 100 requests per minute per IP

	// Add API versioning middleware
	r.Use(APIVersionMiddleware)

	return r
}

// AddRoutes sets up all routes and middleware for the Chi router.
func AddRoutes(obs *observability.Provider) *chi.Mux {
	r := initChiRouter(obs)

	// Apply CORS middleware globally
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Set up API routes
	setupAPIRoutes(r, obs)

	// Serve static files and SPA
	setupStaticRoutes(r)

	return r
}

// setupAPIRoutes configures the API routes for the application.
func setupAPIRoutes(r *chi.Mux, obs *observability.Provider) {
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
		r.Get("/health", obs.HealthCheckMiddleware())
	} else {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]string{"status": "healthy", "service": "wastebin"})
		})
	}

	// Database health check endpoint
	r.Get("/health/db", handlers.DatabaseHealthCheck)
}

// setupStaticRoutes sets up routes to serve static files and SPA.
func setupStaticRoutes(r *chi.Mux) {
	// Determine static file path based on environment
	staticPath := "./web/dist/"
	if !config.Conf.Dev {
		staticPath = "/web/"
	}

	// Serve static files (assets, etc.)
	fileServer(r, "/assets/", http.Dir(staticPath+"assets/"))

	// Handle SPA routes by serving index.html
	r.Get("/", serveSPA)
	r.Get("/about", serveSPA)
	r.Get("/paste/new", serveSPA)
	r.Get("/paste/{uuid}", serveSPA)
	r.Get("/paste/{uuid}/raw", handlers.GetRawPaste) // This should serve raw text, not SPA
}

// serveSPA serves the Single Page Application (SPA) index file.
func serveSPA(w http.ResponseWriter, r *http.Request) {
	indexFilePath := "./web/dist/index.html"
	if !config.Conf.Dev {
		indexFilePath = "/web/index.html"
	}
	http.ServeFile(w, r, indexFilePath)
}

// jsonResponse sends a JSON response with the given data.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fileServer sets up a file server for static files.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	fs := http.StripPrefix(path, http.FileServer(root))
	r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

// getAllowedOrigins returns CORS allowed origins based on configuration
func getAllowedOrigins() []string {
	if config.Conf.Dev {
		// Allow localhost for development
		return []string{
			"http://localhost:*",
			"http://127.0.0.1:*",
			"https://localhost:*",
			"https://127.0.0.1:*",
		}
	}
	// In production, should be configured to specific domains
	// For now, allowing all origins for backward compatibility
	// TODO: Configure specific allowed origins via environment variable
	return []string{"*"}
}

// APIVersionMiddleware handles API versioning
func APIVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default to v1 if no version specified
		version := "v1"

		// Check for version in header
		if headerVersion := r.Header.Get("X-API-Version"); headerVersion != "" {
			version = headerVersion
		}

		// Check for version in Accept header (content negotiation)
		if accept := r.Header.Get("Accept"); accept != "" {
			if strings.Contains(accept, "application/vnd.wastebin.v2+json") {
				version = "v2"
			} else if strings.Contains(accept, "application/vnd.wastebin.v1+json") {
				version = "v1"
			}
		}

		// Add version to response headers
		w.Header().Set("X-API-Version", version)

		// Add version to context for handlers to use
		ctx := context.WithValue(r.Context(), "api-version", version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
