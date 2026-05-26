// Package gitconfig writes the fenced agent-init ignore block to the user's
// machine-wide git excludes file (the "global-default" visibility mode, #52).
//
// This is the only place in agent-init that mutates global git configuration,
// and it does so only under the explicit `init --visibility=global-default`
// flag. The write is action-at-a-distance — it affects every repository on the
// machine — so the caller is expected to announce the edited path loudly and
// warn the user. This package keeps the global-config read/write isolated from
// the repo-local block management in internal/gitignore (which owns the block
// content via gitignore.Block); per the repo layout, the global excludes-file
// location logic lives here.
package gitconfig

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner runs a git subcommand and returns its trimmed stdout. It is the single
// seam through which this package touches the user's global git config, so
// tests inject a fake that records writes and never shells out. The production
// implementation is execRunner.
type Runner interface {
	// Get reads a global config value. ok is false when the key is unset
	// (git exits non-zero), which callers must distinguish from a hard error.
	Get(key string) (value string, ok bool, err error)
	// Set writes a global config value.
	Set(key, value string) error
}

// Env supplies the home directory and XDG config base used to resolve the
// default excludes path. It is injectable so tests never read the real HOME.
type Env interface {
	HomeDir() (string, error)
	Getenv(key string) string
}

// upsertFunc renders the managed-block-preserving content for a file. It is the
// same upsert logic gitignore uses for repo-local files; injecting it keeps the
// block content owned by internal/gitignore while the global-file location
// logic stays here.
type upsertFunc func(existing string) string

// EnsureGlobal writes the fenced ignore block to the user's machine-wide git
// excludes file and returns the absolute path it edited. It resolves the target
// from `git config --global core.excludesfile`; if that is unset it falls back
// to ${XDG_CONFIG_HOME:-~/.config}/git/ignore and sets core.excludesfile to that
// path (the single global-config write this mode authorizes). The block is
// upserted idempotently: a stale managed block is replaced in place.
func EnsureGlobal(runner Runner, env Env, upsert upsertFunc) (string, error) {
	path, setKey, err := resolveExcludesPath(runner, env)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("creating global excludes directory: %w", err)
	}
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(upsert(string(existing))), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	// Only point core.excludesfile at the fallback path when it was unset.
	// When the user already configured an excludes file we wrote to that one
	// and must not touch any other global-config key.
	if setKey {
		if err := runner.Set("core.excludesfile", path); err != nil {
			return "", fmt.Errorf("setting core.excludesfile: %w", err)
		}
	}
	return path, nil
}

// GlobalPath resolves the machine-wide excludes path without writing anything.
// It is used by the --dry-run preview so the announced path matches what
// EnsureGlobal would edit.
func GlobalPath(runner Runner, env Env) (string, error) {
	path, _, err := resolveExcludesPath(runner, env)
	return path, err
}

// resolveExcludesPath returns the excludes file path and whether
// core.excludesfile needs to be set (true only on the unset-fallback path).
func resolveExcludesPath(runner Runner, env Env) (path string, setKey bool, err error) {
	configured, ok, err := runner.Get("core.excludesfile")
	if err != nil {
		return "", false, fmt.Errorf("reading core.excludesfile: %w", err)
	}
	if ok && strings.TrimSpace(configured) != "" {
		expanded, err := expandHome(env, strings.TrimSpace(configured))
		if err != nil {
			return "", false, err
		}
		return expanded, false, nil
	}
	fallback, err := defaultExcludesPath(env)
	if err != nil {
		return "", false, err
	}
	return fallback, true, nil
}

// defaultExcludesPath returns ${XDG_CONFIG_HOME:-~/.config}/git/ignore, the
// path git itself uses when core.excludesfile is unset.
func defaultExcludesPath(env Env) (string, error) {
	if xdg := strings.TrimSpace(env.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "git", "ignore"), nil
	}
	home, err := env.HomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "git", "ignore"), nil
}

// expandHome resolves a leading "~" in a configured path against the env's home
// directory, matching how git expands core.excludesfile.
func expandHome(env Env, path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := env.HomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

// OSEnv is the production Env backed by the process environment.
type OSEnv struct{}

func (OSEnv) HomeDir() (string, error) { return os.UserHomeDir() }
func (OSEnv) Getenv(key string) string { return os.Getenv(key) }

// execRunner is the production Runner backed by the `git` binary.
type execRunner struct{}

// NewExecRunner returns a Runner that shells out to git for global config.
func NewExecRunner() Runner { return execRunner{} }

func (execRunner) Get(key string) (string, bool, error) {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		// `git config --get` exits 1 when the key is unset; that is not a
		// hard error, just an absent value.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("running git config --get %s: %w", key, err)
	}
	return strings.TrimRight(string(out), "\n"), true, nil
}

func (execRunner) Set(key, value string) error {
	if err := exec.Command("git", "config", "--global", key, value).Run(); err != nil {
		return fmt.Errorf("running git config --global %s: %w", key, err)
	}
	return nil
}
