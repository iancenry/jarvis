package repository

import (
	"errors"
	"strings"

	"github.com/iancenry/jarvis/internal/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func newDomainNotFoundError(resource string) error {
	code := domainErrorCode(resource, "NOT_FOUND")
	return errs.NewNotFoundError(strings.ToLower(resource)+" not found", false, &code)
}

func newDomainConflictError(resource, message string) error {
	code := domainErrorCode(resource, "ALREADY_EXISTS")
	return errs.NewConflictError(message, false, &code)
}

func domainErrorCode(resource, suffix string) string {
	resource = strings.ToUpper(strings.ReplaceAll(resource, " ", "_"))
	return resource + "_" + suffix
}

func isNoRowsError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
