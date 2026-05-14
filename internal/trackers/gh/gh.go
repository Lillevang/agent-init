// Package gh holds the GitHub Issues tracker integration.
//
// MCP server: @modelcontextprotocol/server-github. Configure
// GITHUB_PERSONAL_ACCESS_TOKEN (or set GITHUB_TOKEN in the host env and
// keep the ${env:GITHUB_TOKEN} interpolation).
package gh

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// MCPServer returns the .mcp.json entry to merge under mcpServers["github"].
// The shape matches the @modelcontextprotocol/server-github project's
// documented config. Verify against the upstream README before activating —
// server names and arg conventions change.
func MCPServer() map[string]any {
	return map[string]any{
		"command": "npx",
		"args":    []any{"-y", "@modelcontextprotocol/server-github"},
		"env": map[string]any{
			"GITHUB_PERSONAL_ACCESS_TOKEN": "${env:GITHUB_TOKEN}",
		},
	}
}
