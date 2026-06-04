package templates

import "embed"

//go:embed *.html partials/*.html
var TemplateFS embed.FS
