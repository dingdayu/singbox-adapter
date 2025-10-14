// Package assets embeds static resources for distribution.
package assets

import "embed"

// IndexFS contains the embedded index.html.
//
//go:embed index.html
var IndexFS embed.FS

// DistFS contains the built frontend static assets.
//
//go:embed all:dist
var DistFS embed.FS
