// Package mcp handles MCP protocol implementation.
// input: MCP SDK configuration
// output: *mcp.Server instance, tool registration
// pos: MCP layer entry point, creates SDK server for tool registration
// note: if this file changes, update header and internal/mcp/README.md
package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewMCPServer creates a new MCP SDK server instance
func NewMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "safe-mysql-mcp",
		Version: "1.0.0",
	}, nil)

	return server
}
