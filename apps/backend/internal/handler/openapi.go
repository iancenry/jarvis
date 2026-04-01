package handler

import (
	"fmt"
	"net/http"

	"github.com/iancenry/jarvis/internal/assets"
	"github.com/iancenry/jarvis/internal/server"

	"github.com/labstack/echo/v4"
)

type OpenAPIHandler struct {
	Handler
}

func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

func (h *OpenAPIHandler) ServeOpenAPIUI(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-cache")
	templateBytes, err := assets.Files.ReadFile(assets.OpenAPIUIPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded OpenAPI UI template: %w", err)
	}

	templateString := string(templateBytes)

	err = c.HTML(http.StatusOK, templateString)
	if err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}

	return nil
}

func (h *OpenAPIHandler) ServeOpenAPISpec(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-cache")

	specBytes, err := assets.Files.ReadFile(assets.OpenAPISpecPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded OpenAPI spec: %w", err)
	}

	return c.Blob(http.StatusOK, echo.MIMEApplicationJSONCharsetUTF8, specBytes)
}
