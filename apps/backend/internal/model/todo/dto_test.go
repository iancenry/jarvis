package todo

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTodoPayloadValidate(t *testing.T) {
	t.Run("rejects explicit null for required fields", func(t *testing.T) {
		payload := &UpdateTodoPayload{
			ID:       uuid.New(),
			Title:    validation.NewPatchNull[string](),
			Status:   validation.NewPatchNull[Status](),
			Priority: validation.NewPatchNull[Priority](),
		}

		err := payload.Validate()
		require.Error(t, err)

		var fieldErrors validation.CustomValidationErrors
		require.True(t, errors.As(err, &fieldErrors))
		assert.Contains(t, fieldErrors, validation.CustomValidationError{Field: "title", Message: "cannot be null"})
		assert.Contains(t, fieldErrors, validation.CustomValidationError{Field: "status", Message: "cannot be null"})
		assert.Contains(t, fieldErrors, validation.CustomValidationError{Field: "priority", Message: "cannot be null"})
	})

	t.Run("allows clearing nullable fields", func(t *testing.T) {
		payload := &UpdateTodoPayload{
			ID:           uuid.New(),
			Description:  validation.NewPatchNull[string](),
			DueDate:      validation.NewPatchNull[time.Time](),
			ParentTodoID: validation.NewPatchNull[uuid.UUID](),
			CategoryID:   validation.NewPatchNull[uuid.UUID](),
			Metadata:     validation.NewPatchNull[Metadata](),
		}

		require.NoError(t, payload.Validate())
	})
}
