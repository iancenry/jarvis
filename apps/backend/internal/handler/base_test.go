package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reusablePayload struct {
	Name string `json:"name"`
}

func (p *reusablePayload) Validate() error {
	return nil
}

func TestHandleUsesFreshPayloadPerRequest(t *testing.T) {
	e := echo.New()
	handler := Handle(
		func(c echo.Context, payload *reusablePayload) (map[string]string, error) {
			return map[string]string{"name": payload.Name}, nil
		},
		http.StatusOK,
		&reusablePayload{},
	)

	firstResponse := performJSONRequest(t, e, handler, http.MethodPost, map[string]string{
		"name": "first",
	})
	assert.Equal(t, http.StatusOK, firstResponse.Code)
	assert.JSONEq(t, `{"name":"first"}`, firstResponse.Body.String())

	secondResponse := performJSONRequest(t, e, handler, http.MethodPost, map[string]string{})
	assert.Equal(t, http.StatusOK, secondResponse.Code)
	assert.JSONEq(t, `{"name":""}`, secondResponse.Body.String())
}

func TestHandleNoContentReturnsConfiguredStatus(t *testing.T) {
	e := echo.New()
	handler := HandleNoContent(
		func(c echo.Context, payload *reusablePayload) error {
			return nil
		},
		http.StatusNoContent,
		&reusablePayload{},
	)

	response := performJSONRequest(t, e, handler, http.MethodDelete, map[string]string{})
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())
}

func performJSONRequest(t *testing.T, e *echo.Echo, handler echo.HandlerFunc, method string, body any) *httptest.ResponseRecorder {
	t.Helper()

	requestBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(method, "/", bytes.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = handler(ctx)
	require.NoError(t, err)

	return rec
}
