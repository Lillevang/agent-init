package claudecowork

import (
	"embed"
	"fmt"
)

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// ExecutablePaths is empty because this flavor ships no scripts — the
// scaffold is a document-collaboration folder, not a code project. No
// done-gate, no codemap regeneration.
func ExecutablePaths() []string {
	return nil
}

// NextSteps returns the post-scaffold message tailored to a doc-collab
// workspace. No devcontainer, no `just check` — the next moves are to fill
// in the agent instructions and load the folder into Claude Cowork.
func NextSteps(target string) string {
	return fmt.Sprintf(`
Done.

Next steps:
  1. Edit %s/AGENTS.md — replace the "What this workspace is" paragraph
     with one or two sentences describing what you and your coworkers do here.
  2. Create CLAUDE.md alongside AGENTS.md (run inside %s):
       - Linux/macOS:  ln -s AGENTS.md CLAUDE.md
       - Windows:      copy AGENTS.md CLAUDE.md   (OneDrive sync breaks
                       symlinks; keep the two in sync manually)
  3. Drop source materials into %s/reference/ and reusable templates into
     %s/templates/.
  4. Share the OneDrive folder link with coworkers; ask them to skim
     AGENTS.md so they know what Claude is allowed to do.
  5. Open Claude Cowork (https://cowork.claude.com or your entry point)
     and load this folder.
`, target, target, target, target)
}
