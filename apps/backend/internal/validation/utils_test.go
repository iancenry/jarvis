package validation

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validatorPayload struct {
	Name string `json:"name" validate:"required"`
}

func (p *validatorPayload) Validate() error {
	return validator.New().Struct(p)
}

type customValidationPayload struct{}

func (p *customValidationPayload) Validate() error {
	return CustomValidationErrors{
		{Field: "title", Message: "is required"},
		{Field: "priority", Message: "must be one of: low medium high"},
	}
}

type genericValidationPayload struct{}

func (p *genericValidationPayload) Validate() error {
	return errors.New("validation blew up")
}

func TestBindAndValidateReturnsEchoBindMessageWithoutPanicking(t *testing.T) {
	ctx := newValidationContext(http.MethodPost, "/", `{"name":123}`, echo.MIMEApplicationJSON)
	payload := &validatorPayload{}

	err := BindAndValidate(ctx, payload)

	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Contains(t, httpErr.Message, "expected=string")
	assert.Empty(t, httpErr.Errors)
}

func TestBindAndValidateReturnsValidatorFieldErrors(t *testing.T) {
	ctx := newValidationContext(http.MethodPost, "/", `{}`, echo.MIMEApplicationJSON)
	payload := &validatorPayload{}

	err := BindAndValidate(ctx, payload)

	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	require.Len(t, httpErr.Errors, 1)
	assert.Equal(t, "name", httpErr.Errors[0].Field)
	assert.Equal(t, "is required", httpErr.Errors[0].Error)
}

func TestBindAndValidateReturnsCustomValidationErrorsWithoutPanicking(t *testing.T) {
	ctx := newValidationContext(http.MethodGet, "/", "", "")
	payload := &customValidationPayload{}

	err := BindAndValidate(ctx, payload)

	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	require.Len(t, httpErr.Errors, 2)
	assert.Equal(t, "title", httpErr.Errors[0].Field)
	assert.Equal(t, "is required", httpErr.Errors[0].Error)
	assert.Equal(t, "priority", httpErr.Errors[1].Field)
}

func TestBindAndValidateReturnsGenericValidationErrorsWithoutPanicking(t *testing.T) {
	ctx := newValidationContext(http.MethodGet, "/", "", "")
	payload := &genericValidationPayload{}

	err := BindAndValidate(ctx, payload)

	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	require.Len(t, httpErr.Errors, 1)
	assert.Equal(t, "request", httpErr.Errors[0].Field)
	assert.Equal(t, "validation blew up", httpErr.Errors[0].Error)
}

func newValidationContext(method string, target string, body string, contentType string) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}

	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}
