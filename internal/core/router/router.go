package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	"github.com/sahabatharianmu/OpenMind/internal/modules/auth/handler"
)

func RegisterRoutes(h *server.Hertz, authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	api := h.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
	protected := v1.Group("/")
	protected.Use(authMiddleware.Middleware())
	{
		// TODO: Add protected routes here
	}
}
