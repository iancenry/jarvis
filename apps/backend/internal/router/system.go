package router

import (
	"github.com/iancenry/jarvis/internal/handler"

	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	r.GET("/status", h.Health.CheckHealth)
	r.GET("/static/openapi.json", h.OpenAPI.ServeOpenAPISpec)
	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
}
