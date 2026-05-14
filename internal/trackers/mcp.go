package trackers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MergeMCPServer reads <target>/.mcp.json, ensures the mcpServers map exists,
// and adds an entry under serverKey with serverConfig. It is idempotent: if
// serverKey already exists, the function returns (false, nil) without
// touching the file. The caller can present that as "already configured".
//
// Returns (changed, error). changed=true means the file was rewritten;
// changed=false means the entry was already present.
//
// If .mcp.json doesn't exist, the function creates it with a minimal shape.
// If .mcp.json is malformed JSON, it returns an error rather than corrupting
// the file further.
func MergeMCPServer(target, serverKey string, serverConfig map[string]any) (bool, error) {
	path := filepath.Join(target, ".mcp.json")

	root, err := readMCPFile(path)
	if err != nil {
		return false, err
	}

	servers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
	}
	if _, exists := servers[serverKey]; exists {
		return false, nil
	}
	servers[serverKey] = serverConfig
	root["mcpServers"] = servers

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshalling .mcp.json: %w", err)
	}
	// json.MarshalIndent doesn't emit a trailing newline; add one for POSIX hygiene.
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	return true, nil
}

func readMCPFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"mcpServers": map[string]any{}}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing %s: %w (check for malformed JSON)", path, err)
	}
	if root == nil {
		root = map[string]any{}
	}
	return root, nil
}
