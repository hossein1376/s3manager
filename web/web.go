package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed template
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

func Load() (templates fs.FS, statics fs.FS, err error) {
	templates, err = fs.Sub(templateFS, "template")
	if err != nil {
		err = fmt.Errorf("loading templates: %w", err)
		return
	}
	statics, err = fs.Sub(staticFS, "static")
	if err != nil {
		err = fmt.Errorf("loading statics: %w", err)
		return
	}

	return
}
