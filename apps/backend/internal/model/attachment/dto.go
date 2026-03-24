package attachment

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UploadTodoAttachmentPayload struct {
	TodoID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *UploadTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteTodoAttachmentPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
	TodoID uuid.UUID `param:"todoId" validate:"required,uuid"`
}

func (p *DeleteTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetAttachmentPresignedURLPayload struct {
	TodoID uuid.UUID `param:"id" validate:"required,uuid"`
	AttachmentID uuid.UUID `param:"attachmentId" validate:"required,uuid"`
}

func (p *GetAttachmentPresignedURLPayload) Validate() error {
	return validator.New().Struct(p)
}