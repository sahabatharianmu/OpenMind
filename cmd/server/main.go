package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/internal/core/database"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	"github.com/sahabatharianmu/OpenMind/internal/core/router"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/handler"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/repository"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/service"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	appLogger := logger.NewLogger(cfg.Server.Mode)
	appLogger.Info(
		"Starting OpenMind server",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
	)

	database.InitDB(cfg)
	db := database.GetDB()

	if err := database.RunMigrations(db, appLogger); err != nil {
		appLogger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db, appLogger)
	jwtService := security.NewJWTService(cfg)
	passwordService := crypto.NewPasswordService(cfg)
	authService := service.NewAuthService(userRepo, jwtService, passwordService, appLogger)
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	h := server.New(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithExitWaitTime(cfg.Server.ExitTimeout),
	)

	router.RegisterRoutes(h, authHandler, authMiddleware)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		appLogger.Info("Shutting down server gracefully...")

		// TODO: Add other cleanup logic here (e.g., closing Database connections, Redis, etc.)

		appLogger.Info("Server resources released")
	})

	appLogger.Info("Server starting", zap.String("host", cfg.Server.Host), zap.Int("port", cfg.Server.Port))
	h.Spin()

	appLogger.Info("Server exited")
}
