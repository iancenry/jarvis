package validation

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancenry/jarvis/internal/errs"
	"github.com/labstack/echo/v4"
)

type Validatable interface {
	Validate() error
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

func BindAndValidate(c echo.Context, payload Validatable) error {
	if err := c.Bind(payload); err != nil {
		return errs.NewBadRequestError(extractBindErrorMessage(err), false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); len(fieldErrors) > 0 {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}

	return nil
}

func validateStruct(v Validatable) (string, []errs.FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []errs.FieldError) {
	if err == nil {
		return "", nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return "Validation failed", buildValidatorFieldErrors(validationErrors)
	}

	var customValidationErrors CustomValidationErrors
	if errors.As(err, &customValidationErrors) {
		fieldErrors := make([]errs.FieldError, 0, len(customValidationErrors))
		for _, validationErr := range customValidationErrors {
			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: validationErr.Field,
				Error: validationErr.Message,
			})
		}

		return "Validation failed", fieldErrors
	}

	return "Validation failed", []errs.FieldError{
		{
			Field: "request",
			Error: err.Error(),
		},
	}
}

func buildValidatorFieldErrors(validationErrors validator.ValidationErrors) []errs.FieldError {
	fieldErrors := make([]errs.FieldError, 0, len(validationErrors))

	for _, err := range validationErrors {
		field := strings.ToLower(err.Field())
		var msg string

		switch err.Tag() {
		case "required":
			msg = "is required"
		case "min":
			if err.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must be at least %s characters", err.Param())
			} else {
				msg = fmt.Sprintf("must be at least %s", err.Param())
			}
		case "max":
			if err.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must not exceed %s characters", err.Param())
			} else {
				msg = fmt.Sprintf("must not exceed %s", err.Param())
			}
		case "oneof":
			msg = fmt.Sprintf("must be one of: %s", err.Param())
		case "email":
			msg = "must be a valid email address"
		case "e164":
			msg = "must be a valid phone number with country code"
		case "uuid":
			msg = "must be a valid UUID"
		case "uuidList":
			msg = "must be a comma-separated list of valid UUIDs"
		case "dive":
			msg = "some items are invalid"
		default:
			if err.Param() != "" {
				msg = fmt.Sprintf("%s: %s:%s", field, err.Tag(), err.Param())
			} else {
				msg = fmt.Sprintf("%s: %s", field, err.Tag())
			}
		}

		fieldErrors = append(fieldErrors, errs.FieldError{
			Field: strings.ToLower(err.Field()),
			Error: msg,
		})
	}

	return fieldErrors
}

func extractBindErrorMessage(err error) string {
	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		return bindHTTPErrorMessage(echoErr)
	}

	if err != nil && strings.TrimSpace(err.Error()) != "" {
		return err.Error()
	}

	return "invalid request payload"
}

func bindHTTPErrorMessage(err *echo.HTTPError) string {
	if err == nil {
		return "invalid request payload"
	}

	switch message := err.Message.(type) {
	case string:
		if strings.TrimSpace(message) != "" {
			return message
		}
	case error:
		if strings.TrimSpace(message.Error()) != "" {
			return message.Error()
		}
	case fmt.Stringer:
		if strings.TrimSpace(message.String()) != "" {
			return message.String()
		}
	}

	if err.Internal != nil && strings.TrimSpace(err.Internal.Error()) != "" {
		return err.Internal.Error()
	}

	if err.Code > 0 {
		return http.StatusText(err.Code)
	}

	return "invalid request payload"
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
