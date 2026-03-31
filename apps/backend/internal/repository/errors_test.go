package repository_test

import (
	"net/http"
	"testing"

	"github.com/iancenry/jarvis/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertRepositoryHTTPError(t *testing.T, err error, status int, code string, message string) {
	t.Helper()

	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, status, httpErr.Status)
	assert.Equal(t, code, httpErr.Code)
	assert.Equal(t, message, httpErr.Message)
}

func assertRepositoryNotFoundError(t *testing.T, err error, code string, message string) {
	t.Helper()
	assertRepositoryHTTPError(t, err, http.StatusNotFound, code, message)
}

func assertRepositoryConflictError(t *testing.T, err error, code string, message string) {
	t.Helper()
	assertRepositoryHTTPError(t, err, http.StatusConflict, code, message)
}
