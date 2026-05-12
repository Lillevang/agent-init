package flavors

import "io/fs"

// Flavor describes a project scaffold: metadata, a template tree, and the
// subset of paths that must end up executable.
//
// Templates are walked first. CommonTemplates (optional) is walked second;
// any relative path that the flavor already produced is skipped, so flavors
// always win conflicts with the common overlay.
type Flavor struct {
	Name            string
	DisplayName     string
	Description     string
	Templates       fs.FS
	TemplateRoot    string
	CommonTemplates fs.FS
	CommonRoot      string
	ExecutablePaths []string
}
