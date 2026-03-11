package handler

import (
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gotogether/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(authService *service.AuthService, meetingService *service.MeetingService, linkService *service.LinkService, corsOrigin, botLinkSecret string, pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Register pgx pool metrics collector (nil-safe for tests).
	if pool != nil {
		prometheus.MustRegister(NewPgxPoolCollector(pool))
	}

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(OTelMiddleware)
	r.Use(PrometheusMiddleware)

	// Sentry error reporting middleware (no-op if Sentry is not initialized).
	if hub := sentry.CurrentHub(); hub != nil && hub.Client() != nil {
		sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
		r.Use(sentryHandler.Handle)
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Prometheus metrics endpoint — unauthenticated, outside /api.
	r.Handle("/metrics", promhttp.Handler())

	authHandler := NewAuthHandler(authService)
	meetingHandler := NewMeetingHandler(meetingService)
	linkHandler := NewLinkHandler(linkService)

	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddleware(authService))
				r.Get("/me", authHandler.Me)
				r.Post("/link/telegram", linkHandler.LinkTelegram)
			})
		})

		// Bot-to-backend link endpoints (authenticated by shared secret)
		r.Route("/link", func(r chi.Router) {
			r.Use(BotSecretMiddleware(botLinkSecret))
			r.Post("/bot/auth", linkHandler.AuthByTelegramID)
			r.Post("/bot/initiate", linkHandler.InitiateFromBot)
			r.Post("/bot/confirm", linkHandler.ConfirmFromBot)
		})

		// Protected user routes
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(authService))
			r.Get("/users/search", authHandler.SearchUsers)
		})

		// Protected meeting routes
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(authService))

			r.Route("/meetings", func(r chi.Router) {
				r.Post("/", meetingHandler.Create)
				r.Get("/", meetingHandler.List)
				r.Get("/all", meetingHandler.ListAll)
				r.Get("/tags/all", meetingHandler.GetAllTags)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", meetingHandler.Get)
					r.Put("/", meetingHandler.Update)
					r.Delete("/", meetingHandler.Delete)
					r.Post("/confirm", meetingHandler.Confirm)
					r.Post("/participants", meetingHandler.AddParticipants)
					r.Put("/participants/rsvp", meetingHandler.UpdateRSVP)
					r.Post("/votes", meetingHandler.Vote)
					r.Get("/votes", meetingHandler.GetVotes)
					r.Put("/tags", meetingHandler.SetTags)
				})
			})
		})
	})

	return r
}
