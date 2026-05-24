package projectmgmt

import (
	"embed"
	"fmt"
)

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// ExecutablePaths is empty — this flavor ships markdown and JSON, no scripts.
func ExecutablePaths() []string {
	return nil
}

// NextSteps tailors the post-scaffold message for a PM workspace. No
// devcontainer, no `just check`. The next move is to wire a tracker via
// `agent-init add-tracker`, then start filling in stakeholders/decisions.
func NextSteps(target string) string {
	return fmt.Sprintf(`
Done.

Next steps:
  1. Edit %s/AGENTS.md — replace the "Project context" paragraph and
     the "Active trackers" line (initially blank).
  2. Wire a tracker (one or more — re-run for each):
       agent-init add-tracker gh    %s
       agent-init add-tracker jira  %s
       agent-init add-tracker ado   %s
     Each call writes integrations/<tracker>/ and adds an entry to .mcp.json.
     Credentials are read from the environment via ${env:...} references — set
     them in your shell or a gitignored .env (see integrations/<tracker>/.env.example),
     never in the tracked .mcp.json. Restart the MCP client after setting them.
  3. Edit %s/stakeholders.md — list the people who can make decisions on
     this project and what kinds of decisions each can authorize.
  4. Open the folder in Claude Code (or load it into Claude Cowork). The
     skills under .claude/skills/ are now invokable:
       /intake-meeting       /break-down-epic
       /log-decision         /track-stakeholder
       /sync-tracker
  5. (Optional, Linux/macOS only) Create CLAUDE.md as a symlink:
       cd %s && ln -s AGENTS.md CLAUDE.md
     Windows: copy AGENTS.md CLAUDE.md (OneDrive eats symlinks).
`, target, target, target, target, target, target)
}
