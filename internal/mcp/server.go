// Package mcp handles MCP protocol implementation
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
