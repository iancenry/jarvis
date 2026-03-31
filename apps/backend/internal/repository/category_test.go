package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/repository"
	testing_pkg "github.com/iancenry/jarvis/internal/testing"
	"github.com/iancenry/jarvis/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_UpdateCategory(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)
	userID := uuid.New().String()

	categoryItem, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
		Name:        "Work",
		Color:       "#ff0000",
		Description: testing_pkg.Ptr("Initial description"),
	})
	require.NoError(t, err)

	t.Run("update category name successfully", func(t *testing.T) {
		payload := &category.UpdateCategoryPayload{
			ID:   categoryItem.ID,
			Name: validation.NewPatchValue("Personal"),
		}

		result, err := categoryRepo.UpdateCategory(ctx, userID, categoryItem.ID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Personal", result.Name)
	})

	t.Run("clear description with explicit null", func(t *testing.T) {
		payload := &category.UpdateCategoryPayload{
			ID:          categoryItem.ID,
			Description: validation.NewPatchNull[string](),
		}

		result, err := categoryRepo.UpdateCategory(ctx, userID, categoryItem.ID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.Description)
	})
}

func TestCategoryRepository_DomainErrors(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)
	userID := uuid.New().String()

	t.Run("create duplicate category name returns conflict", func(t *testing.T) {
		_, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
			Name:  "Work",
			Color: "#ff0000",
		})
		require.NoError(t, err)

		result, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
			Name:  "Work",
			Color: "#00ff00",
		})
		require.Nil(t, result)
		assertRepositoryConflictError(t, err, "CATEGORY_ALREADY_EXISTS", "category with this name already exists")
	})

	t.Run("get missing category returns not found", func(t *testing.T) {
		result, err := categoryRepo.GetCategoryByID(ctx, userID, uuid.New())
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "CATEGORY_NOT_FOUND", "category not found")
	})

	t.Run("update conflicting category name returns conflict", func(t *testing.T) {
		first, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
			Name:  "Personal",
			Color: "#111111",
		})
		require.NoError(t, err)

		second, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
			Name:  "Errands",
			Color: "#222222",
		})
		require.NoError(t, err)

		result, err := categoryRepo.UpdateCategory(ctx, userID, second.ID, &category.UpdateCategoryPayload{
			ID:   second.ID,
			Name: validation.NewPatchValue(first.Name),
		})
		require.Nil(t, result)
		assertRepositoryConflictError(t, err, "CATEGORY_ALREADY_EXISTS", "category with this name already exists")
	})

	t.Run("update missing category returns not found", func(t *testing.T) {
		categoryID := uuid.New()
		result, err := categoryRepo.UpdateCategory(ctx, userID, categoryID, &category.UpdateCategoryPayload{
			ID:   categoryID,
			Name: validation.NewPatchValue("Updated"),
		})
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "CATEGORY_NOT_FOUND", "category not found")
	})

	t.Run("delete missing category returns not found", func(t *testing.T) {
		deletedCategory, err := categoryRepo.DeleteCategory(ctx, userID, uuid.New())
		assert.Nil(t, deletedCategory)
		assertRepositoryNotFoundError(t, err, "CATEGORY_NOT_FOUND", "category not found")
	})
}

func TestCategoryRepository_DeleteCategory(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)
	userID := uuid.New().String()

	categoryItem, err := categoryRepo.CreateCategory(ctx, userID, &category.CreateCategoryPayload{
		Name:  "Work",
		Color: "#ff0000",
	})
	require.NoError(t, err)

	deletedCategory, err := categoryRepo.DeleteCategory(ctx, userID, categoryItem.ID)
	require.NoError(t, err)
	require.NotNil(t, deletedCategory)
	assert.Equal(t, categoryItem.ID, deletedCategory.ID)
	assert.Equal(t, categoryItem.Name, deletedCategory.Name)

	result, err := categoryRepo.GetCategoryByID(ctx, userID, categoryItem.ID)
	assert.Nil(t, result)
	assertRepositoryNotFoundError(t, err, "CATEGORY_NOT_FOUND", "category not found")
}
