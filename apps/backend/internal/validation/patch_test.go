package validation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type patchFieldTestPayload struct {
	Description PatchField[string] `json:"description"`
}

func TestPatchFieldDistinguishesOmittedNullAndValue(t *testing.T) {
	t.Run("omitted field", func(t *testing.T) {
		var payload patchFieldTestPayload

		err := json.Unmarshal([]byte(`{}`), &payload)
		require.NoError(t, err)
		assert.False(t, payload.Description.IsSet())
		assert.False(t, payload.Description.IsNull())
	})

	t.Run("explicit null", func(t *testing.T) {
		var payload patchFieldTestPayload

		err := json.Unmarshal([]byte(`{"description":null}`), &payload)
		require.NoError(t, err)
		assert.True(t, payload.Description.IsSet())
		assert.True(t, payload.Description.IsNull())
	})

	t.Run("explicit value", func(t *testing.T) {
		var payload patchFieldTestPayload

		err := json.Unmarshal([]byte(`{"description":"kept"}`), &payload)
		require.NoError(t, err)
		assert.True(t, payload.Description.IsSet())
		assert.False(t, payload.Description.IsNull())
		assert.Equal(t, "kept", payload.Description.Value())
	})
}
