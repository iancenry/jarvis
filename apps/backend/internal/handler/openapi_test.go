package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIHandlerServeOpenAPIUIUsesEmbeddedAsset(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewOpenAPIHandler(nil)

	err := handler.ServeOpenAPIUI(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Body.String(), `data-url="/static/openapi.json"`)
}

func TestOpenAPIHandlerServeOpenAPISpecUsesEmbeddedAsset(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/static/openapi.json", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewOpenAPIHandler(nil)

	err := handler.ServeOpenAPISpec(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Header().Get("Content-Type"), echo.MIMEApplicationJSON)
	assert.Contains(t, rec.Body.String(), `"openapi": "3.0.2"`)
}
