package attachment

import "github.com/go-playground/validator/v10"

type UploadTodoAttachmentPayload struct {
	TodoID string `param:"id" validate:"required,uuid"`
}

func (p *UploadTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteTodoAttachmentPayload struct {
	ID string `param:"id" validate:"required,uuid"`
	TodoID string `param:"todoId" validate:"required,uuid"`
}

func (p *DeleteTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetAttachmentPresignedURLPayload struct {
	TodoID string `param:"id" validate:"required,uuid"`
	AttachmentID string `param:"attachmentId" validate:"required,uuid"`
}

func (p *GetAttachmentPresignedURLPayload) Validate() error {
	return validator.New().Struct(p)
}