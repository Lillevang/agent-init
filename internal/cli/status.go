package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Lillevang/agent-init/internal/gitconfig"
	"github.com/Lillevang/agent-init/internal/gitignore"
)

// runStatus reports how the scaffold's agentic envelope is being tracked (or
// ignored) for the target repository. It is strictly read-only: no file on
// disk and no git config is mutated. The acceptance criteria come from
// issue #60: print the current visibility mode, the absolute path of the file
// carrying the agent-init ignore block, the exact undo command, and — when the
// scaffold is committed locally but a global-default ignore exists — the
// force-add hint explaining why files are still tracked.
func (a App) runStatus(args []string) error {
	if wantsHelp(args) {
		cmd, _ := lookupCommand("status")
		a.printCommandHelp(cmd)
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("usage: agent-init status [target]\nRun 'agent-init status --help' for usage")
	}
	target := "."
	if len(args) == 1 {
		target = args[0]
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolving target path: %w", err)
	}
	return a.reportStatus(absTarget, gitconfig.NewExecRunner(), gitconfig.OSEnv{})
}

// statusMode names the four visibility states `status` reports. Kept separate
// from the writer-side `visibility` type because the reader has one extra
// observable state (`shadowed-by-global`) that no `--visibility` flag value
// maps to: it is the combination of "no repo-local block here" plus "a global
// block does exist".
type statusMode string

const (
	statusModeShared   statusMode = "shared"
	statusModeLocal    statusMode = "local"
	statusModeHidden   statusMode = "hidden"
	statusModeShadowed statusMode = "shadowed-by-global"
)

// statusFinding bundles what `status` discovered about the target repo so the
// reporter is a pure formatter — easy to test, easy to extend with new fields.
type statusFinding struct {
	mode    statusMode
	target  string // absolute path of the target directory
	carrier string // absolute path of the file holding the ignore block, "" if none
	// machineWide is true when the carrier is the global excludes file, so the
	// printed path is annotated accordingly and the force-add hint is shown.
	machineWide bool
}

// reportStatus runs the read-only detection and writes the human-readable
// report to a.out. The runner/env seam matches applyGlobalVisibility so the
// global-excludes path resolution is testable via GIT_CONFIG_GLOBAL plus a
// fake HOME (see internal/cli/cli_test.go:isolateGlobalGitConfig).
func (a App) reportStatus(absTarget string, runner gitconfig.Runner, env gitconfig.Env) error {
	finding, err := detectStatus(absTarget, runner, env)
	if err != nil {
		return err
	}
	writeStatusReport(a.out, finding)
	return nil
}

// detectStatus probes the three carrier files in git's precedence order — the
// committed .gitignore beats .git/info/exclude beats the global excludes file —
// and returns the most-local hit. When no repo-local block is found but a
// block sits in the machine-wide excludes, the finding is `shadowed-by-global`
// rather than `shared`, matching the issue-60 spec.
func detectStatus(absTarget string, runner gitconfig.Runner, env gitconfig.Env) (statusFinding, error) {
	finding := statusFinding{target: absTarget, mode: statusModeShared}

	localPath, err := gitignore.LocalPath(absTarget)
	if err != nil {
		return statusFinding{}, fmt.Errorf("resolving .gitignore path: %w", err)
	}
	if has, err := fileHasBlock(localPath); err != nil {
		return statusFinding{}, err
	} else if has {
		finding.mode = statusModeLocal
		finding.carrier = localPath
		return finding, nil
	}

	hiddenPath, err := gitignore.HiddenPath(absTarget)
	if err != nil {
		return statusFinding{}, fmt.Errorf("resolving .git/info/exclude path: %w", err)
	}
	if has, err := fileHasBlock(hiddenPath); err != nil {
		return statusFinding{}, err
	} else if has {
		finding.mode = statusModeHidden
		finding.carrier = hiddenPath
		return finding, nil
	}

	// Only consult the global excludes file when no repo-local block carries
	// the scaffold. status is read-only, so unlike applyGlobalVisibility it
	// must not write or set core.excludesfile — GlobalPath does neither.
	globalPath, err := gitconfig.GlobalPath(runner, env)
	if err != nil {
		// A failure to resolve the global path (e.g. no HOME) is not fatal
		// for status — we still want to report the local state. Treat it as
		// "no global carrier" so the user sees `shared` rather than an error.
		return finding, nil
	}
	if has, err := fileHasBlock(globalPath); err != nil {
		return statusFinding{}, err
	} else if has {
		finding.mode = statusModeShadowed
		finding.carrier = globalPath
		finding.machineWide = true
	}
	return finding, nil
}

// fileHasBlock reads path and returns whether the managed agent-init block is
// present. A missing file is not an error (it is the normal `shared` case);
// any other read error is propagated so the user sees why detection failed.
func fileHasBlock(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("reading %s: %w", path, err)
	}
	return gitignore.HasBlock(string(content)), nil
}

// writeStatusReport prints the finding in a short, scannable layout matching
// the style of `init`'s output: labeled lines, no decorative borders, no
// emojis. The exact undo command uses a portable sed invocation that works on
// both GNU and BSD sed; the marker strings come from the gitignore package so
// they cannot drift from what is on disk.
func writeStatusReport(w io.Writer, f statusFinding) {
	_, _ = fmt.Fprintf(w, "mode:   %s\n", f.mode)
	_, _ = fmt.Fprintf(w, "target: %s\n", f.target)
	if f.carrier == "" {
		_, _ = fmt.Fprintln(w, "no agent-init ignore block found")
		return
	}
	if f.machineWide {
		_, _ = fmt.Fprintf(w, "ignore: %s (machine-wide)\n", f.carrier)
	} else {
		_, _ = fmt.Fprintf(w, "ignore: %s\n", f.carrier)
	}
	_, _ = fmt.Fprintf(w, "undo:   sed -i.bak '/%s/,/%s/d' %s\n",
		gitignore.MarkerStart, gitignore.MarkerEnd, f.carrier)
	_, _ = fmt.Fprintf(w, "        (or open %s and delete the lines from\n", f.carrier)
	_, _ = fmt.Fprintf(w, "        '%s' through '%s' inclusive)\n",
		gitignore.MarkerStart, gitignore.MarkerEnd)
	if f.mode == statusModeShadowed {
		_, _ = fmt.Fprintln(w, "note:   The scaffold may still be tracked in this repo: git ignores rules")
		_, _ = fmt.Fprintln(w, "        for files that are already in the index. To stop tracking the scaffold")
		_, _ = fmt.Fprintln(w, "        in this repo, remove it from the index (git rm --cached ...). To commit")
		_, _ = fmt.Fprintln(w, "        the scaffold openly in a repo where it is newly being added, force-add it:")
		_, _ = fmt.Fprintln(w, "          git add -f .agent AGENTS.md CLAUDE.md .devcontainer Justfile .pre-commit-config.yaml")
	}
}
