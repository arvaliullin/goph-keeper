// Package main запускает HTTP-сервер GophKeeper.
package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/config"
	"github.com/arvaliullin/goph-keeper/internal/core/services"
	"github.com/arvaliullin/goph-keeper/internal/crypto"
	"github.com/arvaliullin/goph-keeper/internal/repository/minio"
	"github.com/arvaliullin/goph-keeper/internal/repository/postgres"
	serverhttp "github.com/arvaliullin/goph-keeper/internal/server/http"
	"github.com/jackc/pgx/v5/pgxpool"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"

	_ "github.com/arvaliullin/goph-keeper/docs"       // swagger docs
	_ "github.com/arvaliullin/goph-keeper/migrations" // goose Go migrations
	_ "github.com/jackc/pgx/v5/stdlib"                // postgres driver for goose
)

func main() {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db for migrations")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close db for migrations")
		}
	}()
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal().Err(err).Msg("failed to set goose dialect")
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create db pool")
	}
	defer pool.Close()

	minioClient, err := miniogo.New(cfg.MinioEndpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioSecure,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create minio client")
	}

	userRepo := postgres.NewUserRepository(pool)
	secretRepo := postgres.NewSecretRepository(pool)
	binaryRepo, err := minio.NewBinaryRepository(ctx, minioClient, "gophkeeper-binary")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize binary repository")
	}

	cryptoService, err := crypto.New(cfg.EncryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize crypto service")
	}

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	secretService := services.NewSecretService(secretRepo, binaryRepo, cryptoService)
	binaryService := services.NewBinaryService(binaryRepo, secretRepo, cryptoService)

	router := serverhttp.NewRouter(authService, secretService, binaryService, cfg.JWTSecret, cfg.MaxBinarySize)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		log.Info().Str("address", cfg.Address).Msg("starting server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("server stopped")
}
