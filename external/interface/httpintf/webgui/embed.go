package webgui

import (
	"embed"
	"io/fs"
)

//go:embed assets
var assetRootFS embed.FS
var assetFS, _ = fs.Sub(assetRootFS, "assets")
