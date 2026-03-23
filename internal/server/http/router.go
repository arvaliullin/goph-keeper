package http

import (
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/arvaliullin/goph-keeper/internal/server/http/handlers"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title GophKeeper API
// @version 1.0
// @description REST API для менеджера паролей GophKeeper.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
//
// NewRouter создает HTTP-маршрутизатор сервера.
func NewRouter(
	authService ports.AuthService,
	secretService ports.SecretService,
	binaryService ports.BinaryService,
	secretKey string,
	maxBinarySize int64,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	authHandler := handlers.NewAuthHandler(authService)
	secretHandler := handlers.NewSecretHandler(secretService)
	binaryHandler := handlers.NewBinaryHandler(binaryService, maxBinarySize)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(secretKey))

			r.Route("/secrets", func(r chi.Router) {
				r.Post("/", secretHandler.Create)
				r.Get("/", secretHandler.List)
				r.Get("/{id}", secretHandler.Get)
				r.Put("/{id}", secretHandler.Update)
				r.Delete("/{id}", secretHandler.Delete)
			})

			r.Route("/binary", func(r chi.Router) {
				r.Post("/{id}", binaryHandler.Upload)
				r.Get("/{id}", binaryHandler.Download)
				r.Delete("/{id}", binaryHandler.Delete)
			})
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}
