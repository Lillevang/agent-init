// Package common holds template files shared across every flavor.
//
// The scaffold engine walks the flavor's own templates first, then walks
// common as a fallback layer — so any flavor can override a common file by
// shipping its own copy at the same relative path.
package common

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// ExecutablePaths is the subset of common template paths that must be
// installed with executable mode. Flavor-specific executables are declared
// alongside the flavor.
func ExecutablePaths() []string {
	return []string{
		".agent/scripts/check.sh",
		".agent/scripts/gen-codemap.sh",
		".agent/scripts/review.sh",
	}
}
