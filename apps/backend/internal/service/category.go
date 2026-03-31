package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/middleware"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/repository"
	"github.com/iancenry/jarvis/internal/server"
)

type CategoryService struct {
	server       *server.Server
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(s *server.Server, categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		server:       s,
		categoryRepo: categoryRepo,
	}
}

func (cs *CategoryService) CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	logger := middleware.LoggerFromContext(ctx)

	categoryItem, err := cs.categoryRepo.CreateCategory(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create category")
		return nil, err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "category_created").
		Str("category_id", categoryItem.ID.String()).
		Msg("Category created")

	return categoryItem, nil
}

func (cs *CategoryService) GetCategories(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	logger := middleware.LoggerFromContext(ctx)

	categories, err := cs.categoryRepo.GetCategories(ctx, userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get categories")
		return nil, err
	}

	return categories, nil
}

func (cs *CategoryService) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	logger := middleware.LoggerFromContext(ctx)

	categoryItem, err := cs.categoryRepo.GetCategoryByID(ctx, userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get category by id")
		return nil, err
	}

	return categoryItem, nil
}

func (cs *CategoryService) UpdateCategory(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	logger := middleware.LoggerFromContext(ctx)

	categoryItem, err := cs.categoryRepo.UpdateCategory(ctx, userID, categoryID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update category")
		return nil, err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "category_updated").
		Str("category_id", categoryItem.ID.String()).
		Msg("Category updated")

	return categoryItem, nil
}

func (cs *CategoryService) DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) error {
	logger := middleware.LoggerFromContext(ctx)

	deletedCategory, err := cs.categoryRepo.DeleteCategory(ctx, userID, categoryID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete category")
		return err
	}

	// business event log
	eventLogger := middleware.LoggerFromContext(ctx)
	eventLogger.Info().
		Str("event", "category_deleted").
		Str("category_id", deletedCategory.ID.String()).
		Msg("Category deleted")

	return nil
}
