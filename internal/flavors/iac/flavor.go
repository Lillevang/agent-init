package iac

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// ExecutablePaths lists flavor-specific scripts that must be marked executable
// on write. The common scripts (check.sh, gen-codemap.sh as fallback, review.sh)
// are listed in common.ExecutablePaths() and prepended at registration. The
// iac flavor ships its own gen-codemap.sh that overrides the common one so
// the codemap surfaces Terraform modules and Ansible roles.
func ExecutablePaths() []string {
	return []string{
		".agent/scripts/gen-codemap.sh",
		".devcontainer/post-create.sh",
	}
}
