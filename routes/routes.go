package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

var db *gorm.DB

// AddRoutes adds all the routes to the router
func AddRoutes(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Mount("/api", api())
	r.Mount("/", web())
}

func api() chi.Router {
	r := chi.NewRouter()
	r.Mount("/v1", v1())

	return r
}

func v1() chi.Router {
	r := chi.NewRouter()
	r.Get("/paste/{uuid}", handlers.GetPaste(db))
	r.Post("/paste", handlers.CreatePaste(db))

	return r
}

func web() chi.Router {
	r := chi.NewRouter()

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "/web"))
	if config.Conf.Dev {
		filesDir = http.Dir(filepath.Join(workDir, "/web/build"))
	}

	// Serve static files
	r.Handle("/*", http.FileServer(filesDir))
	return r
}
