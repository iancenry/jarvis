package handler

import (
	"net/http"

	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/iancenry/jarvis/internal/service"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	Handler
	categoryService *service.CategoryService
}

func NewCategoryHandler(s *server.Server, categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		Handler:         NewHandler(s),
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *category.CreateCategoryPayload) (*category.Category, error) {
			userID := middleware.GetUserID(c)
			return h.categoryService.CreateCategory(c, userID, payload)
		},
		http.StatusCreated,
		&category.CreateCategoryPayload{},
	)(c)
}

func (h *CategoryHandler) GetCategories(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, query *category.GetCategoriesQuery) (
			*model.PaginatedResponse[category.Category], error,
		) {
			userID := middleware.GetUserID(c)
			return h.categoryService.GetCategories(c, userID, query)
		},
		http.StatusOK,
		&category.GetCategoriesQuery{},
	)(c)
}

func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *category.UpdateCategoryPayload) (*category.Category, error) {
			userID := middleware.GetUserID(c)
			return h.categoryService.UpdateCategory(c, userID, payload.ID, payload)
		},
		http.StatusOK,
		&category.UpdateCategoryPayload{},
	)(c)
}

func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *category.DeleteCategoryByIDQuery) error {
			userID := middleware.GetUserID(c)
			return h.categoryService.DeleteCategory(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&category.DeleteCategoryByIDQuery{},
	)(c)
}