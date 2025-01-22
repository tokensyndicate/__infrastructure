// internal/ports/http/server.go
package http

import (
	"access-manager/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewServer(manager *service.AccessManager) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	handlers := NewHandlers(manager)

	// Routes
	e.POST("/api/v1/instances", handlers.RegisterInstance)
	e.GET("/api/v1/instances/:id", handlers.GetInstance)
	e.POST("/api/v1/credentials", handlers.AssignCredentials)
	e.GET("/api/v1/instances/:id/credentials", handlers.GetInstanceCredentials)
	e.POST("/api/v1/instances/:id/metrics", handlers.UpdateInstanceMetrics)
	e.GET("/api/v1/instances/:id/bindings", handlers.GetInstanceBindings)

	return e
}
