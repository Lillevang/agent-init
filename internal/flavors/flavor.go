package flavors

import "io/fs"

type Flavor struct {
	Name            string
	DisplayName     string
	Description     string
	Templates       fs.FS
	TemplateRoot    string
	ExecutablePaths []string
}
