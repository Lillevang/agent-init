package flavors_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/Lillevang/agent-init/internal/flavors"
)

// TestAllTemplatesParse walks every registered flavor's Templates and
// CommonTemplates and confirms each `.tmpl` file is a syntactically valid
// Go text/template. Catches broken `{{...}}` syntax before the smoke test
// renders the file, and catches malformed templates in code paths the
// smoke test happens not to exercise.
func TestAllTemplatesParse(t *testing.T) {
	for _, flavor := range flavors.DefaultRegistry().List() {
		flavor := flavor
		t.Run(flavor.Name, func(t *testing.T) {
			checkLayerParses(t, flavor.Templates, flavor.TemplateRoot, "flavor")
			if flavor.CommonTemplates != nil {
				checkLayerParses(t, flavor.CommonTemplates, flavor.CommonRoot, "common")
			}
		})
	}
}

func checkLayerParses(t *testing.T, fsys fs.FS, root, label string) {
	t.Helper()
	if fsys == nil {
		return
	}
	sub, err := fs.Sub(fsys, root)
	if err != nil {
		t.Fatalf("fs.Sub(%s, %q): %v", label, root, err)
	}
	err = fs.WalkDir(sub, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		content, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		if _, err := template.New(filepath.Base(path)).Parse(string(content)); err != nil {
			t.Errorf("[%s layer] %s fails to parse as text/template: %v", label, path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s layer: %v", label, err)
	}
}
