package category

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateCategoryPayload defines the structure for creating a new category
type CreateCategoryPayload struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Color string `json:"color" validate:"required,hexcolor|rgb|rgba|hsl|hsla"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

func (p *CreateCategoryPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// UpdateCategoryPayload defines the structure for updating an existing category
type UpdateCategoryPayload struct {
	ID   uuid.UUID  `param:"id" validate:"required,uuid"`
	Name *string `json:"name" validate:"omitempty,min=1,max=255"`
	Color *string `json:"color" validate:"omitempty,hexcolor|rgb|rgba|hsl|hsla"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

func (p *UpdateCategoryPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// GetCategoriesQuery defines the structure for querying categories with pagination options
type GetCategoriesQuery struct {
	Page *int `query:"page" validate:"omitempty,min=1"`
	Limit *int `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort *string `query:"sort" validate:"omitempty,oneof=created_at updated_at name"`
	Order *string `query:"order" validate:"omitempty,oneof=asc desc"`
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