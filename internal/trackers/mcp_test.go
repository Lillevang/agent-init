package trackers_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lillevang/agent-init/internal/trackers"
)

func TestMergeMCPServerAddsNewEntryToEmptyConfig(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	writeMCP(t, target, `{"mcpServers": {}}`)

	changed, err := trackers.MergeMCPServer(target, "github", map[string]any{
		"command": "npx",
		"args":    []any{"-y", "@modelcontextprotocol/server-github"},
	})
	if err != nil {
		t.Fatalf("MergeMCPServer error = %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true on first merge")
	}
	got := readMCP(t, target)
	servers := got["mcpServers"].(map[string]any)
	if _, ok := servers["github"]; !ok {
		t.Fatalf("github not added to mcpServers; got %v", servers)
	}
}

func TestMergeMCPServerIsIdempotent(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	writeMCP(t, target, `{"mcpServers":{"github":{"command":"npx"}}}`)

	changed, err := trackers.MergeMCPServer(target, "github", map[string]any{
		"command": "npx",
		"args":    []any{"different"},
	})
	if err != nil {
		t.Fatalf("MergeMCPServer error = %v", err)
	}
	if changed {
		t.Fatal("expected changed = false when entry already present")
	}
	// Verify the existing entry was not overwritten with the new (different) config.
	got := readMCP(t, target)
	servers := got["mcpServers"].(map[string]any)
	existing := servers["github"].(map[string]any)
	if existing["command"] != "npx" {
		t.Fatalf("existing entry mutated; command = %v", existing["command"])
	}
	if _, ok := existing["args"]; ok {
		t.Fatal("existing entry mutated: 'args' got merged in")
	}
}

func TestMergeMCPServerCreatesFileIfMissing(t *testing.T) {
	t.Parallel()
	target := t.TempDir()

	changed, err := trackers.MergeMCPServer(target, "github", map[string]any{
		"command": "npx",
	})
	if err != nil {
		t.Fatalf("MergeMCPServer error = %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true when creating file")
	}
	if _, err := os.Stat(filepath.Join(target, ".mcp.json")); err != nil {
		t.Fatalf("expected .mcp.json to exist: %v", err)
	}
}

func TestMergeMCPServerErrorsOnMalformedJSON(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	writeMCP(t, target, "{this is not json}")

	_, err := trackers.MergeMCPServer(target, "github", map[string]any{})
	if err == nil {
		t.Fatal("expected error on malformed .mcp.json")
	}
}

func TestMergeMCPServerCreatesMissingMcpServersKey(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	// File exists but has no mcpServers key at all.
	writeMCP(t, target, `{"otherKey": "value"}`)

	changed, err := trackers.MergeMCPServer(target, "github", map[string]any{
		"command": "npx",
	})
	if err != nil {
		t.Fatalf("MergeMCPServer error = %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true when mcpServers missing")
	}
	got := readMCP(t, target)
	if got["otherKey"] != "value" {
		t.Fatalf("unrelated keys lost; otherKey = %v", got["otherKey"])
	}
	servers := got["mcpServers"].(map[string]any)
	if _, ok := servers["github"]; !ok {
		t.Fatal("github not added")
	}
}

func TestDefaultRegistryListsAllShippedTrackers(t *testing.T) {
	t.Parallel()
	got := trackers.DefaultRegistry().List()
	want := map[string]bool{"ado": false, "gh": false, "jira": false}
	for _, tr := range got {
		if _, ok := want[tr.Name]; ok {
			want[tr.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("tracker %q not in DefaultRegistry()", name)
		}
	}
}

func TestRegistryGetReturnsKnownTrackersInError(t *testing.T) {
	t.Parallel()
	_, err := trackers.DefaultRegistry().Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown tracker")
	}
	if msg := err.Error(); !contains(msg, "gh") || !contains(msg, "jira") || !contains(msg, "ado") {
		t.Fatalf("error message should list known trackers; got: %s", msg)
	}
}

func writeMCP(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write .mcp.json: %v", err)
	}
}

func readMCP(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	return got
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
