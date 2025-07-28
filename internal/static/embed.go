package static

import "embed"

//go:embed all:client
var StaticFiles embed.FS
