package fullstack

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

func ExecutablePaths() []string {
	return []string{
		".agent/scripts/check.sh",
		".agent/scripts/gen-codemap.sh",
		".agent/scripts/record-feature.sh",
		".agent/scripts/review.sh",
		".devcontainer/post-create.sh",
	}
}
