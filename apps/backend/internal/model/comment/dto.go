package comment

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateCommentPayload defines the structure for creating a new comment
type CreateCommentPayload struct {
	TodoID  uuid.UUID `param:"id" validate:"required,uuid"`
	Content string    `json:"content" validate:"required,min=1,max=1000"`
}

func (p *CreateCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// GetCommentsByTodoIDQuery defines the structure for querying comments by todo ID
type GetCommentsByTodoIDQuery struct {
	TodoID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (q *GetCommentsByTodoIDQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}

// UpdateCommentPayload defines the structure for updating an existing comment
type UpdateCommentPayload struct {
	ID      uuid.UUID `param:"id" validate:"required,uuid"`
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

func (p *UpdateCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// DeleteCommentByIDQuery defines the structure for deleting a comment
type DeleteCommentByIDQuery struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteCommentByIDQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
