package todo

import (
	"time"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/validation"
)

// CreateTodoPayload defines the structure for creating a new todo item
type CreateTodoPayload struct {
	Title        string     `json:"title" validate:"required,min=1,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=1000"`
	Priority     *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata  `json:"metadata"`
}

func (p *CreateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// UpdateTodoPayload defines the structure for updating an existing todo item
type UpdateTodoPayload struct {
	ID           uuid.UUID                        `param:"id" validate:"required,uuid"`
	Title        validation.PatchField[string]    `json:"title"`
	Description  validation.PatchField[string]    `json:"description"`
	Status       validation.PatchField[Status]    `json:"status"`
	Priority     validation.PatchField[Priority]  `json:"priority"`
	DueDate      validation.PatchField[time.Time] `json:"dueDate"`
	ParentTodoID validation.PatchField[uuid.UUID] `json:"parentTodoId"`
	CategoryID   validation.PatchField[uuid.UUID] `json:"categoryId"`
	Metadata     validation.PatchField[Metadata]  `json:"metadata"`
}

func (p *UpdateTodoPayload) Validate() error {
	var fieldErrors validation.CustomValidationErrors

	if p.ID == uuid.Nil {
		fieldErrors = append(fieldErrors, validation.CustomValidationError{
			Field:   "id",
			Message: "is required",
		})
	}

	validateRequiredTodoStringField(&fieldErrors, "title", p.Title, 255)
	validateNullableTodoStringField(&fieldErrors, "description", p.Description, 1000)
	validateRequiredTodoStatusField(&fieldErrors, p.Status)
	validateRequiredTodoPriorityField(&fieldErrors, p.Priority)

	if len(fieldErrors) > 0 {
		return fieldErrors
	}

	return nil
}

// GetTodosQuery defines the structure for querying todo items with various filters and pagination options
type GetTodosQuery struct {
	Page         *int       `query:"page" validate:"omitempty,min=1"`
	Limit        *int       `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort         *string    `query:"sort" validate:"omitempty,oneof=created_at updated_at due_date priority"`
	Order        *string    `query:"order" validate:"omitempty,oneof=asc desc"`
	Search       *string    `query:"search" validate:"omitempty,max=255"`
	Status       *Status    `query:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority  `query:"priority" validate:"omitempty,oneof=low medium high"`
	CategoryID   *uuid.UUID `query:"categoryId" validate:"omitempty,uuid"`
	ParentTodoID *uuid.UUID `query:"parentTodoId" validate:"omitempty,uuid"`
	DueFrom      *time.Time `query:"dueFrom" validate:"omitempty"`
	DueTo        *time.Time `query:"dueTo" validate:"omitempty,gtfield=DueFrom"`
	Overdue      *bool      `query:"overdue" validate:"omitempty"`
	Completed    *bool      `query:"completed" validate:"omitempty"`
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

type GetTodoStatsQuery struct{}

func (q *GetTodoStatsQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}

func validateRequiredTodoStringField(fieldErrors *validation.CustomValidationErrors, field string, value validation.PatchField[string], maxLen int) {
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

func validateNullableTodoStringField(fieldErrors *validation.CustomValidationErrors, field string, value validation.PatchField[string], maxLen int) {
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

func validateRequiredTodoStatusField(fieldErrors *validation.CustomValidationErrors, value validation.PatchField[Status]) {
	if !value.IsSet() {
		return
	}

	if value.IsNull() {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "status",
			Message: "cannot be null",
		})
		return
	}

	switch value.Value() {
	case StatusDraft, StatusActive, StatusCompleted, StatusArchived:
		return
	default:
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "status",
			Message: "must be one of: draft active completed archived",
		})
	}
}

func validateRequiredTodoPriorityField(fieldErrors *validation.CustomValidationErrors, value validation.PatchField[Priority]) {
	if !value.IsSet() {
		return
	}

	if value.IsNull() {
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "priority",
			Message: "cannot be null",
		})
		return
	}

	switch value.Value() {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return
	default:
		*fieldErrors = append(*fieldErrors, validation.CustomValidationError{
			Field:   "priority",
			Message: "must be one of: low medium high",
		})
	}
}
