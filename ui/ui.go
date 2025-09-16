package ui

import (
	"embed"
	"io/fs"
)

//go:embed index.html
//go:embed objects.html
//go:embed js/*.js
//go:embed css/*.css
var files embed.FS

func FileSystem() (fs.FS, error) {
	return fs.Sub(files, ".")
}
