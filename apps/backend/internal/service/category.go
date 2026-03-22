package service

import (
	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/labstack/echo/v4"
)

type CategoryService struct {
	server *server.Server
	categoryRepo *repository.CategoryRepository
	
}

func NewCategoryService(s *server.Server, categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		server: s,
		categoryRepo: categoryRepo,
	}
}

func (cs *CategoryService) CreateCategory(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	logger := middleware.GetLogger(ctx)

	categoryItem, err := cs.categoryRepo.CreateCategory(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create category")
		return nil, err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "category_created").
		Str("category_id", categoryItem.ID.String()).
		Msg("Category created")

	return categoryItem, nil
}

func (cs *CategoryService) GetCategories(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	logger := middleware.GetLogger(ctx)

	categories, err := cs.categoryRepo.GetCategories(ctx.Request().Context(), userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get categories")
		return nil, err
	}

	return categories, nil
}

func (cs *CategoryService) GetCategoryByID(ctx echo.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	logger := middleware.GetLogger(ctx)

	categoryItem, err := cs.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get category by id")
		return nil, err
	}

	return categoryItem, nil
}

func (cs *CategoryService) UpdateCategory(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	logger := middleware.GetLogger(ctx)
	
	// Check if category exists and belongs to the user
	_, err := cs.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("category validation failed for update")
		return nil, err
	}

	categoryItem, err := cs.categoryRepo.UpdateCategory(ctx.Request().Context(), userID, categoryID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update category")
		return nil, err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "category_updated").
		Str("category_id", categoryItem.ID.String()).
		Msg("Category updated")

	return categoryItem, nil
}

func (cs *CategoryService) DeleteCategory(ctx echo.Context, userID string, categoryID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	// Check if category exists and belongs to the user
	_, err := cs.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("category validation failed for deletion")
		return err
	}

	err = cs.categoryRepo.DeleteCategory(ctx.Request().Context(), userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete category")	
		return err
	}

	// business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "category_deleted").
		Str("category_id", categoryID.String()).
		Msg("Category deleted")

	return nil	
}