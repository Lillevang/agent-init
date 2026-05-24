// Package gitignore manages the fenced "agent-init" ignore block that hides the
// scaffold's agentic envelope from version control.
//
// The block is a fixed list of paths (the envelope) wrapped in start/end
// markers so it can be found, replaced in place (idempotent re-runs), and
// removed by hand to undo. The same block is written to different target files
// depending on the chosen visibility mode. This package owns the block itself
// and the repo-local targets: the committed .gitignore for "local" (shipped)
// and, when it lands, .git/info/exclude for "hidden" (#53). The machine-wide
// "global-default" mode (#52) mutates the user's global git excludes and
// belongs in internal/gitconfig/, not here, per the repo conventions.
package gitignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// blockStart and blockEnd delimit the managed block. They must stay stable:
	// idempotent re-runs find the existing block by these markers and replace
	// it in place. The "(private)" tag is part of the agreed block content.
	blockStart = "# >>> agent-init (private) >>>"
	blockEnd   = "# <<< agent-init <<<"
)

// envelope is the set of scaffold paths the ignore block covers: the agentic
// envelope every code flavor ships. The list is fixed and identical across all
// visibility modes, so a block written by one mode reads the same as another.
// Entries are reproduced verbatim from the agreed block content (the leading
// "/" on the top-level files anchors them to the repo root; the directory
// entries match those dirs anywhere, which is intended for the scaffold's
// nested copies).
var envelope = []string{
	".agent/",
	"/AGENTS.md",
	"/CLAUDE.md",
	".devcontainer/",
	"/Justfile",
	".pre-commit-config.yaml",
}

// Block returns the fenced ignore block, marker lines included, terminated by a
// trailing newline. The content is identical regardless of which file it is
// written to.
func Block() string {
	var b strings.Builder
	b.WriteString(blockStart)
	b.WriteByte('\n')
	for _, p := range envelope {
		b.WriteString(p)
		b.WriteByte('\n')
	}
	b.WriteString(blockEnd)
	b.WriteByte('\n')
	return b.String()
}

// LocalPath returns the absolute path of the committed .gitignore that
// EnsureLocal manages for the given target directory.
func LocalPath(target string) (string, error) {
	abs, err := filepath.Abs(filepath.Join(target, ".gitignore"))
	if err != nil {
		return "", fmt.Errorf("resolving .gitignore path: %w", err)
	}
	return abs, nil
}

// EnsureLocal appends the fenced ignore block to the committed .gitignore in
// target, creating the file if absent. If a block with the same markers is
// already present, it is replaced in place so re-runs never duplicate it. It
// returns the absolute path of the file it wrote.
func EnsureLocal(target string) (string, error) {
	path, err := LocalPath(target)
	if err != nil {
		return "", err
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	updated := upsertBlock(string(existing))
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	return path, nil
}

// upsertBlock returns content with the managed block present exactly once. An
// existing block (matched by markers) is replaced in place; otherwise the block
// is appended, separated from prior content by a blank line.
func upsertBlock(content string) string {
	if start, end, ok := findBlock(content); ok {
		return content[:start] + Block() + content[end:]
	}
	if content == "" {
		return Block()
	}
	sep := "\n"
	if strings.HasSuffix(content, "\n") {
		sep = ""
	}
	return content + sep + "\n" + Block()
}

// findBlock locates the managed block in content. It returns the byte offset of
// the start marker and the offset just past the end marker's line (including its
// trailing newline, if any), and whether a block was found.
func findBlock(content string) (start, end int, ok bool) {
	start = strings.Index(content, blockStart)
	if start < 0 {
		return 0, 0, false
	}
	endMarker := strings.Index(content[start:], blockEnd)
	if endMarker < 0 {
		return 0, 0, false
	}
	end = start + endMarker + len(blockEnd)
	if nl := strings.IndexByte(content[end:], '\n'); nl >= 0 {
		end += nl + 1
	}
	return start, end, true
}
