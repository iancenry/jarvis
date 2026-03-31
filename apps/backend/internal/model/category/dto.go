package category

import (
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/validation"
)

// CreateCategoryPayload defines the structure for creating a new category
type CreateCategoryPayload struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Color       string  `json:"color" validate:"required,hexcolor|rgb|rgba|hsl|hsla"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

func (p *CreateCategoryPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// UpdateCategoryPayload defines the structure for updating an existing category
type UpdateCategoryPayload struct {
	ID          uuid.UUID                     `param:"id" validate:"required,uuid"`
	Name        validation.PatchField[string] `json:"name"`
	Color       validation.PatchField[string] `json:"color"`
	Description validation.PatchField[string] `json:"description"`
}

func (p *UpdateCategoryPayload) Validate() error {
	validate := validator.New()
	var fieldErrors validation.CustomValidationErrors

	if p.ID == uuid.Nil {
		fieldErrors = append(fieldErrors, validation.CustomValidationError{
			Field:   "id",
			Message: "is required",
		})
	}

	validateRequiredCategoryStringField(&fieldErrors, "name", p.Name, 255)
	validateRequiredCategoryColorField(&fieldErrors, validate, p.Color)
	validateNullableCategoryStringField(&fieldErrors, "description", p.Description, 1000)

	if len(fieldErrors) > 0 {
		return fieldErrors
	}

	return nil
}

// GetCategoriesQuery defines the structure for querying categories with pagination options
type GetCategoriesQuery struct {
	Page   *int    `query:"page" validate:"omitempty,min=1"`
	Limit  *int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort   *string `query:"sort" validate:"omitempty,oneof=created_at updated_at name"`
	Order  *string `query:"order" validate:"omitempty,oneof=asc desc"`
	Search *string `query:"search" validate:"omitempty,max=255"`
}

func (q *GetCategoriesQuery) Validate() error {
	if err := validator.New().Struct(q); err != nil {
		return err
	}

	// set defaults before fowarding to repository layer
	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil {
		defaultLimit := 50
		q.Limit = &defaultLimit
	}
	if q.Sort == nil {
		defaultSort := "name"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "asc"
		q.Order = &defaultOrder
	}

	return nil
}

type DeleteCategoryByIDQuery struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (q *DeleteCategoryByIDQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)

}

func validateRequiredCategoryStringField(fieldErrors *validation.CustomValidationErrors, field string, value validation.PatchField[string], maxLen int) {
	if !value.IsSet() {
		return
	}

	if value.IsNull() {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   field,
			Message: "cannot be null",
		})
		return
	}

	length := utf8.RuneCountInString(value.Value())
	switch {
	case length < 1:
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   field,
			Message: "must be at least 1 characters",
		})
	case length > maxLen:
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   field,
			Message: "must not exceed 255 characters",
		})
	}
}

func validateRequiredCategoryColorField(fieldErrors *validation.CustomValidationErrors, validate *validator.Validate, value validation.PatchField[string]) {
	if !value.IsSet() {
		return
	}

	if value.IsNull() {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "color",
			Message: "cannot be null",
		})
		return
	}

	if err := validate.Var(value.Value(), "hexcolor|rgb|rgba|hsl|hsla"); err != nil {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "color",
			Message: "must be a valid color value",
		})
	}
}

func validateNullableCategoryStringField(fieldErrors *validation.CustomValidationErrors, field string, value validation.PatchField[string], maxLen int) {
	if !value.IsSet() || value.IsNull() {
		return
	}

	if utf8.RuneCountInString(value.Value()) > maxLen {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   field,
			Message: "must not exceed 1000 characters",
		})
	}
}
