// Command piplos runs the site backend API (admin + public endpoints).
package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/database"
	"github.com/piplos/piplos.media/internal/handlers"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
	"github.com/piplos/piplos.media/internal/server"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
	"github.com/piplos/piplos.media/internal/services/mailer"
	"github.com/piplos/piplos.media/internal/services/translate"
	"github.com/piplos/piplos.media/internal/utils"
)

// Version is the application version.
const Version = "1.0.0"

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Kitchen}).
		With().Timestamp().Logger()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid config")
	}

	// Migrations are applied by the one-shot `migrate` container in production
	// (see compose.yml) or manually via ./scripts/migrate in development.

	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect database")
	}

	repo := repository.New(db.Pool)
	encKey, err := utils.EncryptionKeyFromString(cfg.EncryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid encryption key")
	}
	repo.SetEncryptionKey(encKey)
	authService := authsvc.New(&cfg)

	if err := seedAdmin(ctx, &cfg, repo, authService, log); err != nil {
		log.Fatal().Err(err).Msg("seed admin user")
	}

	uploadDir, err := filepath.Abs(cfg.UploadDir)
	if err != nil {
		log.Fatal().Err(err).Msg("resolve upload dir")
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatal().Err(err).Str("dir", uploadDir).Msg("create upload dir")
	}

	publicAPIURL := cfg.PublicAPIURL
	if publicAPIURL == "" {
		publicAPIURL = "http://localhost:" + cfg.Port
	}

	app := fiber.New(fiber.Config{
		AppName:      "piplos-api v" + Version,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		ProxyHeader:  fiber.HeaderXForwardedFor,
		TrustProxy:   true,
	})

	app.Use(recover.New())
	app.Use(middleware.CORS(cfg.CORSOrigins))
	app.Use(middleware.ErrorHandler(log))
	app.Use("/uploads", static.New(uploadDir, static.Config{MaxAge: 86400}))

	authMw := middleware.NewAuth(authService, repo)
	mailService := mailer.NewService(repo, repo, cfg.AdminURL, log)
	h := &server.Handlers{
		Auth:     handlers.NewAuthHandler(authService, repo),
		Users:    handlers.NewUsersHandler(authService, repo),
		Content:  handlers.NewContentHandler(repo),
		Leads:    handlers.NewLeadsHandler(repo, mailService),
		Settings: handlers.NewSettingsHandler(repo, translate.New(repo)),
		Public:   handlers.NewPublicHandler(repo),
		Uploads:  handlers.NewUploadsHandler(uploadDir, publicAPIURL),
		Files:    handlers.NewFilesHandler(uploadDir, publicAPIURL),
		AIModels: handlers.NewAIModelsHandler(repo),
	}
	server.Register(app, h, authMw)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "piplos-api", "version": Version})
	})

	log.Info().Str("port", cfg.Port).Msg("server starting")
	go func() {
		if err := app.Listen("0.0.0.0:"+cfg.Port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
			log.Fatal().Err(err).Msg("start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}
	db.Close()
}

// seedAdmin creates the initial admin account when the users table is empty.
func seedAdmin(ctx context.Context, cfg *config.Config, repo *repository.Repository, auth *authsvc.Service, log zerolog.Logger) error {
	count, err := repo.CountUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	email := cfg.AdminEmail
	password := cfg.AdminPassword
	if email == "" || password == "" {
		log.Warn().Msg("no users exist and admin_email/admin_password are not set — skipping admin seed")
		return nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	if _, err := repo.CreateUser(ctx, email, hash, "Administrator", models.RoleAdmin); err != nil {
		return err
	}
	log.Info().Str("email", email).Msg("initial admin user created")
	return nil
}
