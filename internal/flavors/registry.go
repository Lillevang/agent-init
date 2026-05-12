package flavors

import (
	"fmt"
	"sort"

	"github.com/mikeschinkel/agent-init/internal/flavors/fullstack"
)

type Registry struct {
	byName map[string]Flavor
}

func DefaultRegistry() Registry {
	return NewRegistry(Flavor{
		Name:            "fullstack",
		DisplayName:     "Fullstack",
		Description:     "TypeScript/Node frontend and backend scaffold with Playwright recording and OpenAPI client hooks.",
		Templates:       fullstack.Templates(),
		TemplateRoot:    "templates",
		ExecutablePaths: fullstack.ExecutablePaths(),
	})
}

func NewRegistry(items ...Flavor) Registry {
	byName := make(map[string]Flavor, len(items))
	for _, item := range items {
		byName[item.Name] = item
	}
	return Registry{byName: byName}
}

func (r Registry) Get(name string) (Flavor, error) {
	flavor, ok := r.byName[name]
	if !ok {
		return Flavor{}, fmt.Errorf("unknown flavor %q", name)
	}
	return flavor, nil
}

func (r Registry) List() []Flavor {
	items := make([]Flavor, 0, len(r.byName))
	for _, item := range r.byName {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}
