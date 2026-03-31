package handler

import (
	"reflect"

	"github.com/iancenry/jarvis/internal/errs"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/iancenry/jarvis/internal/validation"
	"github.com/labstack/echo/v4"
)

// Handler provides base functionality for all handlers
type Handler struct {
	server *server.Server
}

// NewHandler creates a new base handler
func NewHandler(s *server.Server) Handler {
	return Handler{server: s}
}

// HandlerFunc represents a typed handler function that processes a request and returns a response
type HandlerFunc[Req validation.Validatable, Res any] func(c echo.Context, req Req) (Res, error)

// HandlerFuncNoContent represents a typed handler function that processes a request without returning content
type HandlerFuncNoContent[Req validation.Validatable] func(c echo.Context, req Req) error

type responseWriter[Res any] func(c echo.Context, result Res) error

// handleRequest centralizes request construction, binding, validation, and response writing.
func handleRequest[Req validation.Validatable, Res any](
	c echo.Context,
	prototype Req,
	handler HandlerFunc[Req, Res],
	writeResponse responseWriter[Res],
) error {
	req, err := newRequestPayload(prototype)
	if err != nil {
		return err
	}

	if err := validation.BindAndValidate(c, req); err != nil {
		return err
	}

	result, err := handler(c, req)
	if err != nil {
		return err
	}

	return writeResponse(c, result)
}

func newRequestPayload[Req validation.Validatable](prototype Req) (Req, error) {
	var zero Req
	prototypeType := reflect.TypeOf(prototype)
	if prototypeType == nil {
		return zero, errs.NewInternalServerError()
	}

	if prototypeType.Kind() == reflect.Ptr {
		request, ok := reflect.New(prototypeType.Elem()).Interface().(Req)
		if !ok {
			return zero, errs.NewInternalServerError()
		}
		return request, nil
	}

	request, ok := reflect.Zero(prototypeType).Interface().(Req)
	if !ok {
		return zero, errs.NewInternalServerError()
	}
	return request, nil
}

// Handle wraps a handler with request construction, validation, and JSON response writing.
func Handle[Req validation.Validatable, Res any](
	handler HandlerFunc[Req, Res],
	status int,
	prototype Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, prototype, handler, func(c echo.Context, result Res) error {
			return c.JSON(status, result)
		})
	}
}

func HandleFile[Req validation.Validatable](
	handler HandlerFunc[Req, []byte],
	status int,
	prototype Req,
	filename string,
	contentType string,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, prototype, handler, func(c echo.Context, result []byte) error {
			c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
			return c.Blob(status, contentType, result)
		})
	}
}

// HandleNoContent wraps a handler with request construction, validation, and no-content response writing.
func HandleNoContent[Req validation.Validatable](
	handler HandlerFuncNoContent[Req],
	status int,
	prototype Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, prototype, func(c echo.Context, req Req) (struct{}, error) {
			if err := handler(c, req); err != nil {
				return struct{}{}, err
			}
			return struct{}{}, nil
		}, func(c echo.Context, _ struct{}) error {
			return c.NoContent(status)
		})
	}
}
