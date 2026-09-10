// Package api configure le routeur HTTP principal et regroupe l'ensemble des gestionnaires
// de requêtes (ingestion de webhooks, API REST administrative et flux WebSockets).
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

// NewRouter initialise et configure le routeur Chi principal de ZeroDrop.
//
// Middlewares configurés :
//   - RequestID : Attribue un identifiant de traçabilité à chaque requête HTTP.
//   - RealIP    : Extrait l'adresse IP cliente réelle depuis les en-têtes de proxy inverses.
//   - Logger    : Journalise les accès de manière standardisée.
//   - Recoverer : Intercepte les paniques inattendues et retourne une erreur 500 propre.
//   - Timeout   : Définit un délai limite de traitement de 60 secondes par requête.
//   - CORS      : Autorise les requêtes cross-origin pour le Dashboard et le développement local.
func NewRouter(db *database.DB, h *hub.Hub, re *replay.Engine) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-ZeroDrop-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	ingestHandler := NewIngestHandler(db, h)
	endpointsHandler := NewEndpointsHandler(db, h)
	requestsHandler := NewRequestsHandler(db, re)
	wsHandler := NewWSHandler(h)

	// Passerelle d'ingestion publique universelle
	r.HandleFunc("/in/{slug}", ingestHandler.HandleIngest)
	r.HandleFunc("/in/{slug}/*", ingestHandler.HandleIngest)

	// Flux WebSockets temps réel
	r.Get("/ws/events", wsHandler.ServeEvents)
	r.Get("/ws/tunnel", wsHandler.ServeTunnel)

	// API REST administrative
	r.Route("/api", func(r chi.Router) {
		// Endpoints
		r.Get("/endpoints", endpointsHandler.List)
		r.Post("/endpoints", endpointsHandler.Create)
		r.Get("/endpoints/{id}", endpointsHandler.GetByID)
		r.Delete("/endpoints/{id}", endpointsHandler.Delete)

		// Requêtes & Rejeux
		r.Get("/requests", requestsHandler.List)
		r.Delete("/requests", requestsHandler.Clear)
		r.Get("/requests/{id}", requestsHandler.GetByID)
		r.Post("/requests/{id}/replay", requestsHandler.Replay)
		r.Get("/requests/{id}/replays", requestsHandler.GetReplays)

		// Vérification de santé (Healthcheck)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status": "healthy", "version": "0.1.0"}`))
		})
	})

	// Servir l'interface web React SPA embarquée pour toutes les autres routes
	r.Mount("/", ui.Handler())

	return r
}
