package category

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCategoryPayloadValidate(t *testing.T) {
	t.Run("rejects explicit null for required fields", func(t *testing.T) {
		payload := &UpdateCategoryPayload{
			ID:    uuid.New(),
			Name:  validation.NewPatchNull[string](),
			Color: validation.NewPatchNull[string](),
		}

		err := payload.Validate()
		require.Error(t, err)

		var fieldErrors validation.CustomValidationErrors
		require.True(t, errors.As(err, &fieldErrors))
		assert.Contains(t, fieldErrors, validation.CustomValidationError{Field: "name", Message: "cannot be null"})
		assert.Contains(t, fieldErrors, validation.CustomValidationError{Field: "color", Message: "cannot be null"})
	})

	t.Run("allows clearing nullable description", func(t *testing.T) {
		payload := &UpdateCategoryPayload{
			ID:          uuid.New(),
			Description: validation.NewPatchNull[string](),
		}

		require.NoError(t, payload.Validate())
	})
}
