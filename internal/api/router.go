package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/replay"
	"github.com/nosleepman1/zerodrop/ui"
)

// NewRouter configure le routeur Chi principal avec tous les middlewares et routes.
func NewRouter(db *database.DB, h *hub.Hub, re *replay.Engine) http.Handler {
	r := chi.NewRouter()

	// Middlewares de base
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Configuration CORS pour autoriser le Dashboard en dev et en prod
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-ZeroDrop-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Initialisation des Handlers
	ingestHandler := NewIngestHandler(db, h)
	endpointsHandler := NewEndpointsHandler(db, h)
	requestsHandler := NewRequestsHandler(db, re)
	wsHandler := NewWSHandler(h)

	// Route d'ingestion publique universelle
	r.HandleFunc("/in/{slug}", ingestHandler.HandleIngest)
	r.HandleFunc("/in/{slug}/*", ingestHandler.HandleIngest)

	// Routes WebSockets temps réel
	r.Get("/ws/events", wsHandler.ServeEvents)
	r.Get("/ws/tunnel", wsHandler.ServeTunnel)

	// API REST
	r.Route("/api", func(r chi.Router) {
		// Endpoints CRUD
		r.Get("/endpoints", endpointsHandler.List)
		r.Post("/endpoints", endpointsHandler.Create)
		r.Get("/endpoints/{id}", endpointsHandler.GetByID)
		r.Delete("/endpoints/{id}", endpointsHandler.Delete)

		// Requests & Replay
		r.Get("/requests", requestsHandler.List)
		r.Delete("/requests", requestsHandler.Clear)
		r.Get("/requests/{id}", requestsHandler.GetByID)
		r.Post("/requests/{id}/replay", requestsHandler.Replay)
		r.Get("/requests/{id}/replays", requestsHandler.GetReplays)

		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status": "healthy", "version": "0.1.0"}`))
		})
	})

	// Montage du Dashboard Web React embarqué pour toutes les autres routes
	r.Mount("/", ui.Handler())

	return r
}
