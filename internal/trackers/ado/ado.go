// Package ado holds the Azure DevOps tracker integration.
//
// MCP server: @azure-devops/mcp (Microsoft official; verify the package
// name and shape — the Azure DevOps MCP ecosystem is the least mature of
// the three trackers and naming has shifted).
package ado

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

// MCPServer returns the .mcp.json entry to merge under mcpServers["azure-devops"].
// The default points at the official Microsoft Azure DevOps MCP server.
// **Verify the package name and arg shape against the upstream README
// before activating** — Microsoft has shipped multiple ADO MCP attempts
// under different package names; the name below may be outdated.
func MCPServer() map[string]any {
	return map[string]any{
		"command": "npx",
		"args":    []any{"-y", "@azure-devops/mcp"},
		"env": map[string]any{
			"ADO_ORG_URL": "https://dev.azure.com/yourorg",
			"ADO_PROJECT": "",
			"ADO_PAT":     "",
		},
	}
}
