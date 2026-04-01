package assets

import "embed"

const (
	OpenAPIUIPath   = "openapi/openapi.html"
	OpenAPISpecPath = "openapi/openapi.json"
)

// Files contains runtime assets embedded into the backend binary.
//
//go:embed emails/*.html openapi/*.html openapi/*.json
var Files embed.FS

func EmailTemplatePath(name string) string {
	return "emails/" + name + ".html"
}
