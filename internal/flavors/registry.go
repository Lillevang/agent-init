package flavors

import (
	"fmt"
	"sort"

	"github.com/Lillevang/agent-init/internal/flavors/claudecowork"
	"github.com/Lillevang/agent-init/internal/flavors/common"
	"github.com/Lillevang/agent-init/internal/flavors/fullstack"
	"github.com/Lillevang/agent-init/internal/flavors/gobackend"
	"github.com/Lillevang/agent-init/internal/flavors/gocli"
	"github.com/Lillevang/agent-init/internal/flavors/iac"
	"github.com/Lillevang/agent-init/internal/flavors/projectmgmt"
)

type Registry struct {
	byName map[string]Flavor
}

func DefaultRegistry() Registry {
	commonTemplates := common.Templates()
	commonExec := common.ExecutablePaths()
	return NewRegistry(
		Flavor{
			Name:            "fullstack",
			DisplayName:     "Fullstack",
			Description:     "TypeScript/Node frontend and backend scaffold with Playwright recording and OpenAPI client hooks.",
			Templates:       fullstack.Templates(),
			TemplateRoot:    "templates",
			CommonTemplates: commonTemplates,
			CommonRoot:      "templates",
			ExecutablePaths: append(commonExec, fullstack.ExecutablePaths()...),
			Symlinks:        codeFlavorSymlinks(),
			// --agents-only mode: skip the OpenAPI client-generation
			// placeholders. The shipped Justfile is already layout-agnostic
			// (every recipe no-ops gracefully when package.json / tsconfig
			// are absent), so no variant is needed.
			SupportsAgentsOnly: true,
			FreshOnlyPaths: []string{
				"apis/README.md",
				"clients/README.md",
			},
		},
		Flavor{
			Name:            "go-cli",
			DisplayName:     "Go CLI",
			Description:     "Go command-line tool scaffold with cmd/, internal/, cross-build via Justfile, and golangci-lint.",
			Templates:       gocli.Templates(),
			TemplateRoot:    "templates",
			CommonTemplates: commonTemplates,
			CommonRoot:      "templates",
			ExecutablePaths: append(commonExec, gocli.ExecutablePaths()...),
			Symlinks:        codeFlavorSymlinks(),
			// --agents-only mode: skip the Go bootstrap files (entry point,
			// go.mod, version package) and use Justfile.agents-only.tmpl
			// which omits the layout-specific build/cross-build recipes.
			SupportsAgentsOnly: true,
			FreshOnlyPaths: []string{
				"cmd/{{.ProjectName}}/main.go",
				"go.mod",
				"internal/version/version.go",
				"internal/version/version_test.go",
			},
		},
		Flavor{
			Name:            "go-backend",
			DisplayName:     "Go Backend",
			Description:     "Go HTTP backend scaffold with cmd/server, internal/api router, /healthz handler, and run-dev/cross-build Justfile recipes.",
			Templates:       gobackend.Templates(),
			TemplateRoot:    "templates",
			CommonTemplates: commonTemplates,
			CommonRoot:      "templates",
			ExecutablePaths: append(commonExec, gobackend.ExecutablePaths()...),
			Symlinks:        codeFlavorSymlinks(),
			// --agents-only mode: skip the Go bootstrap (entry point, go.mod,
			// example router + tests) and use Justfile.agents-only.tmpl,
			// which omits run-dev/build/cross-build because they hardcode
			// ./cmd/server.
			SupportsAgentsOnly: true,
			FreshOnlyPaths: []string{
				"cmd/server/main.go",
				"go.mod",
				"internal/api/handlers.go",
				"internal/api/handlers_test.go",
			},
		},
		Flavor{
			Name:            "iac",
			DisplayName:     "Infrastructure as Code",
			Description:     "Combined Terraform + Ansible scaffold. Devcontainer ships terraform, tflint, tfsec, trivy, ansible-core, ansible-lint, yamllint. Justfile recipes auto-detect which toolchain is in use; same scaffold works in Terraform-only, Ansible-only, or mixed repos.",
			Templates:       iac.Templates(),
			TemplateRoot:    "templates",
			CommonTemplates: commonTemplates,
			CommonRoot:      "templates",
			ExecutablePaths: append(commonExec, iac.ExecutablePaths()...),
			Symlinks:        codeFlavorSymlinks(),
			// --agents-only mode: skip the Terraform/Ansible boilerplate
			// (root modules, example playbooks). Lint configs are kept:
			// they're inert until the user adds matching files, and existing
			// configs are skipped via the engine's exists-check anyway.
			// The shipped Justfile is already auto-detecting, so no variant
			// is needed.
			SupportsAgentsOnly: true,
			FreshOnlyPaths: []string{
				"terraform/main.tf",
				"terraform/outputs.tf",
				"terraform/variables.tf",
				"terraform/versions.tf",
				"terraform/modules/.gitkeep",
				"ansible/inventory/hosts.yml.example",
				"ansible/playbooks/site.yml",
				"ansible/requirements.yml",
				"ansible/roles/.gitkeep",
				"ansible.cfg",
			},
		},
		Flavor{
			Name:            "claude-cowork",
			DisplayName:     "Claude Cowork",
			Description:     "Shared document-collaboration folder for OneDrive-backed Claude Cowork sessions. AGENTS.md, decisions.md, corrections.md, and reference/templates/archive/ at the workspace root. No devcontainer, no Justfile, no symlinks (OneDrive sync hates them).",
			Templates:       claudecowork.Templates(),
			TemplateRoot:    "templates",
			ExecutablePaths: claudecowork.ExecutablePaths(),
			NextSteps:       claudecowork.NextSteps,
		},
		Flavor{
			Name:            "project-management",
			DisplayName:     "Project Management",
			Description:     "Project-management workspace: epics, meetings, decisions, stakeholders, open questions, time plans. Multi-skill (intake-meeting, break-down-epic, log-decision, track-stakeholder, sync-tracker). Extendable via `agent-init add-tracker {jira|ado|gh}` for tracker integrations.",
			Templates:       projectmgmt.Templates(),
			TemplateRoot:    "templates",
			ExecutablePaths: projectmgmt.ExecutablePaths(),
			NextSteps:       projectmgmt.NextSteps,
		},
	)
}

// codeFlavorSymlinks returns the AGENTS.md/CLAUDE.md symlink trio that every
// code-project flavor ships. Doc-collaboration flavors omit these because
// OneDrive/Dropbox-style sync clients don't reliably preserve symlinks.
func codeFlavorSymlinks() []Symlink {
	return []Symlink{
		{Path: "AGENTS.md", Target: ".agent/AGENTS.md"},
		{Path: "CLAUDE.md", Target: ".agent/CLAUDE.md"},
		{Path: ".agent/CLAUDE.md", Target: "AGENTS.md"},
	}
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
