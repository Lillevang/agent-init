package gitconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeRunner stands in for the git binary so tests never read or mutate the
// real global config. It records the last Set call so tests can assert that
// core.excludesfile is written only on the unset-fallback path.
type fakeRunner struct {
	value  string // current core.excludesfile value
	hasKey bool   // whether the key is set
	getErr error  // injected hard error from Get

	setKey   string
	setValue string
	setCalls int
}

func (r *fakeRunner) Get(key string) (string, bool, error) {
	if r.getErr != nil {
		return "", false, r.getErr
	}
	if key == "core.excludesfile" && r.hasKey {
		return r.value, true, nil
	}
	return "", false, nil
}

func (r *fakeRunner) Set(key, value string) error {
	r.setKey = key
	r.setValue = value
	r.setCalls++
	r.value = value
	r.hasKey = true
	return nil
}

// fakeEnv resolves HOME and XDG_CONFIG_HOME from test-controlled values so the
// real environment is never consulted. The CLAUDE.md testing checklist for
// internal/gitconfig requires a fake HOME; this is that seam.
type fakeEnv struct {
	home string
	vars map[string]string
}

func (e fakeEnv) HomeDir() (string, error) { return e.home, nil }
func (e fakeEnv) Getenv(key string) string { return e.vars[key] }

// block is a stand-in managed block; the real one comes from gitignore.Block.
const block = "# >>> agent-init (private) >>>\n.agent/\n/Justfile\n# <<< agent-init <<<\n"

// upsert mimics the gitignore upsert: replace an existing block in place, else
// append. Kept minimal — gitignore owns the real logic and is tested there.
func upsert(existing string) string {
	const start = "# >>> agent-init (private) >>>"
	const end = "# <<< agent-init <<<"
	if i := strings.Index(existing, start); i >= 0 {
		if j := strings.Index(existing[i:], end); j >= 0 {
			tail := existing[i+j+len(end):]
			tail = strings.TrimPrefix(tail, "\n")
			return existing[:i] + block + tail
		}
	}
	if existing == "" {
		return block
	}
	sep := "\n"
	if strings.HasSuffix(existing, "\n") {
		sep = ""
	}
	return existing + sep + "\n" + block
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// withFakeHome sets HOME and XDG_CONFIG_HOME to a temp dir so a stray real
// implementation (OSEnv) would also be sandboxed. The injected fakeEnv is what
// the code under test actually uses; this is belt-and-suspenders.
func withFakeHome(t *testing.T) (home string, env fakeEnv) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	return home, fakeEnv{home: home, vars: map[string]string{}}
}

func TestEnsureGlobalCreatesDefaultPathAndSetsKey(t *testing.T) {
	home, env := withFakeHome(t)
	runner := &fakeRunner{} // core.excludesfile unset

	got, err := EnsureGlobal(runner, env, upsert)
	if err != nil {
		t.Fatalf("EnsureGlobal() error = %v", err)
	}

	want := filepath.Join(home, ".config", "git", "ignore")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if content := read(t, got); !strings.Contains(content, block) {
		t.Errorf("block not written:\n%s", content)
	}
	// On the unset-fallback path, core.excludesfile must be set to the file.
	if runner.setCalls != 1 || runner.setKey != "core.excludesfile" || runner.setValue != want {
		t.Errorf("Set = (%d, %q, %q); want one set of core.excludesfile to %q",
			runner.setCalls, runner.setKey, runner.setValue, want)
	}
}

func TestEnsureGlobalHonorsXDGConfigHome(t *testing.T) {
	_, env := withFakeHome(t)
	xdg := t.TempDir()
	env.vars["XDG_CONFIG_HOME"] = xdg
	runner := &fakeRunner{}

	got, err := EnsureGlobal(runner, env, upsert)
	if err != nil {
		t.Fatalf("EnsureGlobal() error = %v", err)
	}
	want := filepath.Join(xdg, "git", "ignore")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("excludes file not created at XDG path: %v", err)
	}
}

func TestEnsureGlobalHonorsConfiguredExcludesFileAndDoesNotSetKey(t *testing.T) {
	_, env := withFakeHome(t)
	custom := filepath.Join(t.TempDir(), "unusual", "place", "myexcludes")
	runner := &fakeRunner{value: custom, hasKey: true}

	got, err := EnsureGlobal(runner, env, upsert)
	if err != nil {
		t.Fatalf("EnsureGlobal() error = %v", err)
	}
	if got != custom {
		t.Errorf("path = %q, want configured %q", got, custom)
	}
	if content := read(t, custom); !strings.Contains(content, block) {
		t.Errorf("block not written to configured path:\n%s", content)
	}
	// When core.excludesfile is already configured, no global-config key may
	// be mutated.
	if runner.setCalls != 0 {
		t.Errorf("Set called %d times; want 0 when core.excludesfile is configured", runner.setCalls)
	}
}

func TestEnsureGlobalExpandsTildeInConfiguredPath(t *testing.T) {
	home, env := withFakeHome(t)
	runner := &fakeRunner{value: "~/my-global-ignore", hasKey: true}

	got, err := EnsureGlobal(runner, env, upsert)
	if err != nil {
		t.Fatalf("EnsureGlobal() error = %v", err)
	}
	want := filepath.Join(home, "my-global-ignore")
	if got != want {
		t.Errorf("path = %q, want %q (tilde expanded)", got, want)
	}
}

func TestEnsureGlobalAppendsToExistingFileWithoutOurBlock(t *testing.T) {
	_, env := withFakeHome(t)
	custom := filepath.Join(t.TempDir(), "excludes")
	if err := os.WriteFile(custom, []byte("*.log\n.DS_Store\n"), 0o644); err != nil {
		t.Fatalf("seed excludes: %v", err)
	}
	runner := &fakeRunner{value: custom, hasKey: true}

	if _, err := EnsureGlobal(runner, env, upsert); err != nil {
		t.Fatalf("EnsureGlobal() error = %v", err)
	}
	content := read(t, custom)
	for _, want := range []string{"*.log", ".DS_Store", ".agent/"} {
		if !strings.Contains(content, want) {
			t.Errorf("content missing %q:\n%s", want, content)
		}
	}
	if n := strings.Count(content, "agent-init (private)"); n != 1 {
		t.Errorf("got %d blocks, want 1:\n%s", n, content)
	}
}

func TestEnsureGlobalIsIdempotent(t *testing.T) {
	_, env := withFakeHome(t)
	custom := filepath.Join(t.TempDir(), "excludes")
	if err := os.WriteFile(custom, []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("seed excludes: %v", err)
	}
	runner := &fakeRunner{value: custom, hasKey: true}

	if _, err := EnsureGlobal(runner, env, upsert); err != nil {
		t.Fatalf("first EnsureGlobal() error = %v", err)
	}
	first := read(t, custom)
	if _, err := EnsureGlobal(runner, env, upsert); err != nil {
		t.Fatalf("second EnsureGlobal() error = %v", err)
	}
	second := read(t, custom)

	if first != second {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if n := strings.Count(second, "agent-init (private)"); n != 1 {
		t.Errorf("re-run produced %d blocks, want 1:\n%s", n, second)
	}
}

func TestGlobalPathDoesNotWriteOrSet(t *testing.T) {
	home, env := withFakeHome(t)
	runner := &fakeRunner{}

	got, err := GlobalPath(runner, env)
	if err != nil {
		t.Fatalf("GlobalPath() error = %v", err)
	}
	want := filepath.Join(home, ".config", "git", "ignore")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if _, err := os.Stat(want); !os.IsNotExist(err) {
		t.Errorf("GlobalPath wrote a file, stat err = %v", err)
	}
	if runner.setCalls != 0 {
		t.Errorf("GlobalPath set config %d times; want 0", runner.setCalls)
	}
}
