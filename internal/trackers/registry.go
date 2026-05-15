package trackers

import (
	"fmt"
	"sort"

	"github.com/Lillevang/agent-init/internal/trackers/ado"
	"github.com/Lillevang/agent-init/internal/trackers/gh"
	"github.com/Lillevang/agent-init/internal/trackers/jira"
)

// Registry holds the set of known trackers, looked up by name.
type Registry struct {
	byName map[string]Tracker
}

// DefaultRegistry returns the trackers shipped with agent-init.
func DefaultRegistry() Registry {
	return NewRegistry(
		Tracker{
			Name:         "gh",
			DisplayName:  "GitHub Issues",
			Description:  "GitHub Issues (flat or grouped via labels/milestones). MCP server: @modelcontextprotocol/server-github.",
			Templates:    gh.Templates(),
			TemplateRoot: "templates",
			MCPServerKey: "github",
			MCPServer:    gh.MCPServer(),
		},
		Tracker{
			Name:         "jira",
			DisplayName:  "Jira",
			Description:  "Jira (Epic → Feature → User Story). MCP server: mcp-atlassian (community).",
			Templates:    jira.Templates(),
			TemplateRoot: "templates",
			MCPServerKey: "atlassian",
			MCPServer:    jira.MCPServer(),
		},
		Tracker{
			Name:         "ado",
			DisplayName:  "Azure DevOps",
			Description:  "Azure DevOps (Epic → Feature → PBI). MCP server: @azure-devops/mcp (official; verify package name before activating).",
			Templates:    ado.Templates(),
			TemplateRoot: "templates",
			MCPServerKey: "azure-devops",
			MCPServer:    ado.MCPServer(),
		},
	)
}

// NewRegistry builds a registry from an explicit set of trackers — useful in tests.
func NewRegistry(items ...Tracker) Registry {
	byName := make(map[string]Tracker, len(items))
	for _, item := range items {
		byName[item.Name] = item
	}
	return Registry{byName: byName}
}

// Get returns the tracker with the given name, or an error listing known
// trackers if the name doesn't match.
func (r Registry) Get(name string) (Tracker, error) {
	t, ok := r.byName[name]
	if !ok {
		known := make([]string, 0, len(r.byName))
		for k := range r.byName {
			known = append(known, k)
		}
		sort.Strings(known)
		return Tracker{}, fmt.Errorf("unknown tracker %q (known: %v)", name, known)
	}
	return t, nil
}

// List returns trackers sorted by name.
func (r Registry) List() []Tracker {
	items := make([]Tracker, 0, len(r.byName))
	for _, item := range r.byName {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}
