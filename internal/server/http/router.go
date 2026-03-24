package http

import (
	"net/http"

	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/arvaliullin/goph-keeper/internal/server/http/handlers"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
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
) http.Handler {
	mux := http.NewServeMux()

	authHandler := handlers.NewAuthHandler(authService)
	secretHandler := handlers.NewSecretHandler(secretService)
	binaryHandler := handlers.NewBinaryHandler(binaryService, maxBinarySize)

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	auth := middleware.AuthMiddleware(secretKey)

	mux.Handle("POST /api/v1/secrets", auth(http.HandlerFunc(secretHandler.Create)))
	mux.Handle("GET /api/v1/secrets", auth(http.HandlerFunc(secretHandler.List)))
	mux.Handle("GET /api/v1/secrets/{id}", auth(http.HandlerFunc(secretHandler.Get)))
	mux.Handle("PUT /api/v1/secrets/{id}", auth(http.HandlerFunc(secretHandler.Update)))
	mux.Handle("DELETE /api/v1/secrets/{id}", auth(http.HandlerFunc(secretHandler.Delete)))

	mux.Handle("POST /api/v1/binary/{id}", auth(http.HandlerFunc(binaryHandler.Upload)))
	mux.Handle("GET /api/v1/binary/{id}", auth(http.HandlerFunc(binaryHandler.Download)))
	mux.Handle("DELETE /api/v1/binary/{id}", auth(http.HandlerFunc(binaryHandler.Delete)))

	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return middleware.WithRecovery(middleware.WithLogging(mux))
}
