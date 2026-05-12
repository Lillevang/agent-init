package fullstack

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

func ExecutablePaths() []string {
	return []string{
		".agent/scripts/record-feature.sh",
		".devcontainer/post-create.sh",
	}
}
