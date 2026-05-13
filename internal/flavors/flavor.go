package flavors

import "io/fs"

// Flavor describes a project scaffold: metadata, a template tree, and the
// subset of paths that must end up executable.
//
// Templates are walked first. CommonTemplates (optional) is walked second;
// any relative path that the flavor already produced is skipped, so flavors
// always win conflicts with the common overlay.
//
// Symlinks are created after templates are written. Each Symlink declares
// a relative Path that should be a symlink pointing at Target. Flavors that
// don't use symlinks (e.g. doc-collaboration scaffolds where the target
// filesystem may not support symlinks reliably) leave this nil.
type Flavor struct {
	Name            string
	DisplayName     string
	Description     string
	Templates       fs.FS
	TemplateRoot    string
	CommonTemplates fs.FS
	CommonRoot      string
	ExecutablePaths []string
	Symlinks        []Symlink
	// NextSteps optionally returns the "what to do next" message printed
	// after a successful scaffold. If nil, the engine prints its default
	// code-project message (devcontainer + just check). Doc-collaboration
	// flavors should override this.
	NextSteps func(target string) string
}

// Symlink describes a symlink the scaffold engine should create after
// writing template files. Path is relative to the scaffold target. Target
// is the symlink's destination, written verbatim (so relative-path conventions
// like ".agent/AGENTS.md" survive into the scaffolded tree).
type Symlink struct {
	Path   string
	Target string
}
