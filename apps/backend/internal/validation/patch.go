package validation

import (
	"bytes"
	"encoding/json"
)

// PatchField preserves the difference between an omitted JSON field and a field
// explicitly set to null so PATCH handlers can distinguish "no change" from
// "clear this value".
type PatchField[T any] struct {
	set   bool
	null  bool
	value T
}

func (f *PatchField[T]) UnmarshalJSON(data []byte) error {
	f.set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.null = true
		var zero T
		f.value = zero
		return nil
	}

	f.null = false

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.value = value
	return nil
}

func (f PatchField[T]) IsSet() bool {
	return f.set
}

func (f PatchField[T]) IsNull() bool {
	return f.set && f.null
}

func (f PatchField[T]) Value() T {
	return f.value
}

func NewPatchValue[T any](value T) PatchField[T] {
	return PatchField[T]{
		set:   true,
		value: value,
	}
}

func NewPatchNull[T any]() PatchField[T] {
	return PatchField[T]{
		set:  true,
		null: true,
	}
}
