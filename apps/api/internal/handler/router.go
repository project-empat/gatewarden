package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/gatewarden/api/internal/config"
	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/service"
)

func NewRouter(svc *service.Service, log *zap.SugaredLogger, cfg *config.Config, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logging(log))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authHandler := NewAuthHandler(svc.Auth)
	agentHandler := NewAgentHandler(svc.Agent)
	nodeHandler := NewNodeHandler(svc.Node)
	incidentHandler := NewIncidentHandler(db)
	settingsHandler := NewSettingsHandler()
	sseHandler := NewSSEHandler(svc.Event)

	r.Route("/api", func(r chi.Router) {
		// Public agent endpoints (API key auth)
		r.Post("/agent/register", agentHandler.Register)
		r.With(middleware.AgentAuth(svc.Agent)).Post("/agent/report", agentHandler.Report)
		r.With(middleware.AgentAuth(svc.Agent)).Post("/agent/heartbeat", agentHandler.Heartbeat)

		// User auth
		r.Post("/auth/login", authHandler.Login)

		// Protected user endpoints (JWT auth)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(svc.Auth))

			r.Get("/dashboard/stats", nodeHandler.DashboardStats)

			r.Get("/nodes", nodeHandler.List)
			r.Get("/nodes/{id}", nodeHandler.Get)

			r.Get("/incidents", incidentHandler.List)
			r.Put("/incidents/{id}/resolve", incidentHandler.Resolve)

			r.Get("/settings", settingsHandler.Get)

			r.Get("/events", sseHandler.Stream)
			r.Get("/events/history", sseHandler.History)
		})
	})

	return r
}
