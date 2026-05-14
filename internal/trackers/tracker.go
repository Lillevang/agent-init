// Package trackers describes work-tracker integrations (Jira, Azure DevOps,
// GitHub) that can be layered onto an existing project-management scaffold
// via `agent-init add-tracker`.
//
// Each integration ships its own template tree (terminology cheatsheets,
// setup notes) and an MCP server config that gets merged into the target's
// .mcp.json. Trackers are additive: multiple can be active on a single
// workspace simultaneously, which is useful during migrations between
// tracker systems.
package trackers

import "io/fs"

// Tracker describes one work-tracker integration.
type Tracker struct {
	// Name is the kebab-case identifier passed to `agent-init add-tracker`.
	Name string
	// DisplayName is the human-readable name shown by `list-trackers`.
	DisplayName string
	// Description is the one-line summary shown by `list-trackers`.
	Description string
	// Templates is the embedded FS holding integration-specific files
	// (typically a single README.md describing the terminology and setup).
	Templates fs.FS
	// TemplateRoot is the directory inside Templates to walk (usually "templates").
	TemplateRoot string
	// MCPServerKey is the key under .mcp.json's mcpServers map for this tracker.
	MCPServerKey string
	// MCPServer is the value to insert (or merge as) under mcpServers[MCPServerKey].
	// Conventionally a map with "command", "args", and "env" fields, matching
	// the MCP server config shape.
	MCPServer map[string]any
}
