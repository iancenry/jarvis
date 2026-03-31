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
