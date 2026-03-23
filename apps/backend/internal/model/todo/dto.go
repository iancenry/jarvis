package todo

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateTodoPayload defines the structure for creating a new todo item
type CreateTodoPayload struct {
	Title       string   `json:"title" validate:"required,min=1,max=255"`
	Description *string   `json:"description" validate:"omitempty,max=1000"`
	Priority    *Priority `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate 	*time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"` 
	CategoryID *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"` 
	Metadata *Metadata `json:"metadata"`
}

func (p *CreateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p) 
}

// UpdateTodoPayload defines the structure for updating an existing todo item
type UpdateTodoPayload struct {
	ID 		uuid.UUID `param:"id" validate:"required,uuid"`
	Title       *string   `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string   `json:"description" validate:"omitempty,max=1000"`
	Status      *Status   `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority    *Priority `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate 	*time.Time `json:"dueDate,omitempty"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata *Metadata `json:"metadata"`
}

func (p *UpdateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// GetTodosQuery defines the structure for querying todo items with various filters and pagination options
type GetTodosQuery struct {
	Page *int `query:"page" validate:"omitempty,min=1"`
	Limit *int `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort *string `query:"sort" validate:"omitempty,oneof=created_at updated_at due_date priority"`
	Order *string `query:"order" validate:"omitempty,oneof=asc desc"`
	Search *string `query:"search" validate:"omitempty,max=255"`
	Status *Status `query:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority *Priority `query:"priority" validate:"omitempty,oneof=low medium high"`
	CategoryID *uuid.UUID `query:"categoryId" validate:"omitempty,uuid"`
	ParentTodoID *uuid.UUID `query:"parentTodoId" validate:"omitempty,uuid"`
	DueFrom *time.Time `query:"dueFrom" validate:"omitempty"`
	DueTo *time.Time `query:"dueTo" validate:"omitempty,gtfield=DueFrom"`
	Overdue *bool `query:"overdue" validate:"omitempty"`
	Completed *bool `query:"completed" validate:"omitempty"`
}

func (q *GetTodosQuery) Validate() error {
	validate := validator.New()
	if err := validate.Struct(q); err != nil {
		return err
	}

	// set defaults before fowarding to repository layer
	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil {
		defaultLimit := 20
		q.Limit = &defaultLimit
	}
	if q.Sort == nil {
		defaultSort := "created_at"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "desc"
		q.Order = &defaultOrder
	}

	return nil
}

// GetTodoByIDQuery defines the structure for querying a single todo item by its ID
type GetTodoByIDQuery struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (q *GetTodoByIDQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}

// DeleteTodoByIDQuery defines the structure for deleting a single todo item by its ID
type DeleteTodoByIDQuery struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (q *DeleteTodoByIDQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}

type GetTodoStatsQuery struct {}

func (q *GetTodoStatsQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}