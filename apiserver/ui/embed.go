package ui

import (
	"embed"
	"io/fs"
)

//go:embed static/*
var embeddedFiles embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "static")
}
