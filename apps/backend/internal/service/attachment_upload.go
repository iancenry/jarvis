package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/errs"
)

const (
	MaxAttachmentSizeBytes              int64 = 10 << 20
	MaxAttachmentUploadRequestSizeBytes int64 = MaxAttachmentSizeBytes + (1 << 20)
	maxAttachmentFileNameLength               = 120
)

var allowedAttachmentContentTypes = map[string]struct{}{
	"application/json": {},
	"application/pdf":  {},
	"application/zip":  {},
}

func validateAttachmentSize(fileSize int64) error {
	switch {
	case fileSize <= 0:
		return newAttachmentBadRequest("attachment file is empty", "ATTACHMENT_FILE_EMPTY")
	case fileSize > MaxAttachmentSizeBytes:
		return newAttachmentBadRequest(
			fmt.Sprintf("attachment file exceeds %d MB limit", MaxAttachmentSizeBytes/(1<<20)),
			"ATTACHMENT_FILE_TOO_LARGE",
		)
	default:
		return nil
	}
}

func sanitizeAttachmentFileName(fileName string) (string, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(fileName, "\\", "/"))
	normalized = path.Base(normalized)
	normalized = strings.TrimSpace(normalized)

	if normalized == "" || normalized == "." || normalized == "/" {
		return "", newAttachmentBadRequest("attachment filename is invalid", "ATTACHMENT_FILE_NAME_INVALID")
	}

	originalExt := path.Ext(normalized)
	base := sanitizeAttachmentNameSegment(strings.TrimSuffix(normalized, originalExt))
	ext := sanitizeAttachmentExtension(originalExt)

	if base == "" {
		base = "attachment"
	}

	maxBaseLength := maxAttachmentFileNameLength - len(ext)
	if maxBaseLength < 1 {
		maxBaseLength = maxAttachmentFileNameLength
	}

	base = truncateRunes(base, maxBaseLength)
	sanitized := base + ext
	if sanitized == "" {
		return "", newAttachmentBadRequest("attachment filename is invalid", "ATTACHMENT_FILE_NAME_INVALID")
	}

	return sanitized, nil
}

func detectAttachmentContentType(src io.Reader) (string, io.Reader, error) {
	header := make([]byte, 512)

	bytesRead, err := io.ReadFull(src, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", nil, fmt.Errorf("failed to read attachment header: %w", err)
	}

	if bytesRead == 0 {
		return "", nil, newAttachmentBadRequest("attachment file is empty", "ATTACHMENT_FILE_EMPTY")
	}

	detectedType := http.DetectContentType(header[:bytesRead])
	contentType, _, parseErr := mime.ParseMediaType(detectedType)
	if parseErr != nil {
		contentType = detectedType
	}

	if err := validateAttachmentContentType(contentType); err != nil {
		return "", nil, err
	}

	return contentType, io.MultiReader(bytes.NewReader(header[:bytesRead]), src), nil
}

func validateAttachmentContentType(contentType string) error {
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "text/") {
		return nil
	}

	if _, ok := allowedAttachmentContentTypes[contentType]; ok {
		return nil
	}

	return newAttachmentBadRequest(
		fmt.Sprintf("attachment file type %q is not supported", contentType),
		"ATTACHMENT_FILE_TYPE_NOT_ALLOWED",
	)
}

func buildAttachmentStorageKey(fileName string) string {
	return fmt.Sprintf("todos/attachments/%s%s", uuid.NewString(), sanitizeAttachmentExtension(path.Ext(fileName)))
}

func sanitizeAttachmentNameSegment(segment string) string {
	segment = strings.TrimSpace(segment)

	var builder strings.Builder
	lastSeparator := false

	for _, r := range segment {
		switch {
		case isSafeAttachmentRune(r):
			builder.WriteRune(r)
			lastSeparator = false
		case unicode.IsSpace(r):
			if builder.Len() > 0 && !lastSeparator {
				builder.WriteByte('-')
				lastSeparator = true
			}
		default:
			if builder.Len() > 0 && !lastSeparator {
				builder.WriteByte('-')
				lastSeparator = true
			}
		}
	}

	return strings.Trim(builder.String(), ".-_")
}

func sanitizeAttachmentExtension(ext string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}

	cleaned := builder.String()
	if cleaned == "" {
		return ""
	}

	return "." + truncateRunes(cleaned, 10)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}

func isSafeAttachmentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '.' ||
		r == '_' ||
		r == '-'
}

func newAttachmentBadRequest(message string, code string) error {
	return errs.NewBadRequestError(message, false, &code, nil, nil)
}
