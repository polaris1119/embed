package embed

import (
	"embed"

	"github.com/benbjohnson/hashfs"
)

//go:embed static
var embedFS embed.FS

// 带 hash 功能的 fs.FS
var Fsys = hashfs.NewFS(embedFS)
