package gobackend

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

func ExecutablePaths() []string {
	return []string{
		".devcontainer/post-create.sh",
	}
}
