package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	appLogger := logger.NewLogger(cfg.Server.Mode)
	appLogger.Info("Starting OpenMind server...")

	h := server.New(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithExitWaitTime(cfg.Server.IdleTimeout),
	)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		appLogger.Info("Shutting down server gracefully...")

		// TODO: Add other cleanup logic here (e.g., closing Database connections, Redis, etc.)

		appLogger.Info("Server resources released")
	})

	appLogger.Info("Server starting", zap.String("host", cfg.Server.Host), zap.Int("port", cfg.Server.Port))
	h.Spin()

	appLogger.Info("Server exited")
}
