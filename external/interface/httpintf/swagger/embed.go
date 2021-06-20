package swagger

import (
	"embed"
	_ "embed"
	"io/fs"
)

//= embed swagger/openapi config
//go:generate rm -f swagger.json
//go:generate swagger generate spec --work-dir ../httpapi --output swagger.json
//go:generate swagger validate swagger.json
//go:embed swagger.json
var configJSON []byte

//= go swagger client
//go:generate rm -rf lib
//go:generate mkdir -p lib
//go:generate swagger generate client --quiet --spec swagger.json --target lib

//= swagger-ui
//go:embed assets
var assetFS embed.FS
var uiFS, _ = fs.Sub(assetFS, "assets/swagger-ui")
