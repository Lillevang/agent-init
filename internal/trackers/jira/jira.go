// Package jira holds the Jira (Atlassian) tracker integration.
//
// MCP server: mcp-atlassian (community-maintained). Distributed via PyPI,
// runnable through uvx. Credentials (JIRA_USERNAME, JIRA_API_TOKEN) are read
// from the environment via ${env:...} references so they never land in the
// tracked .mcp.json; see integrations/jira/.env.example.
package jira

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// MCPServer returns the .mcp.json entry to merge under mcpServers["atlassian"].
// The default points at the community mcp-atlassian project (sooperset/mcp-atlassian
// or similar). Verify against the upstream README before activating —
// the Atlassian MCP ecosystem has multiple competing servers with
// different config shapes.
func MCPServer() map[string]any {
	return map[string]any{
		"command": "uvx",
		"args":    []any{"--from", "mcp-atlassian", "mcp-atlassian"},
		"env": map[string]any{
			"JIRA_URL":       "${env:JIRA_URL}",
			"JIRA_USERNAME":  "${env:JIRA_USERNAME}",
			"JIRA_API_TOKEN": "${env:JIRA_API_TOKEN}",
		},
	}
}
