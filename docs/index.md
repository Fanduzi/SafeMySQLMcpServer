# SafeMySQLMcpServer Documentation

Welcome to the SafeMySQLMcpServer documentation.

## Overview

SafeMySQLMcpServer is a secure MySQL MCP (Model Context Protocol) server that enables AI tools like Claude Code to safely operate MySQL databases in development and testing environments.

## Key Features

- **MCP Protocol Support**: Full MCP protocol implementation with HTTP transport
- **JWT Authentication**: Secure token-based authentication with configurable expiration
- **SQL Security**: SQL parsing, validation, and automatic rewriting for dangerous operations
- **Audit Logging**: Complete audit trail with JSON format and log rotation
- **Hot Configuration Reload**: Update settings without server restart
- **Prometheus Metrics**: Comprehensive observability support

## Documentation Sections

| Section | Audience | Description |
|---------|----------|-------------|
| [Guide](./guide) | All Users | Getting started tutorials |
| [Concepts](./concepts) | Developers | Architecture and design concepts |
| [Dev](./dev) | Developers | Development guides and best practices |
| [Admin](./admin) | Operators | Installation and configuration |
| [Reference](./reference) | All Users | API and configuration reference |

## Quick Links

- [Getting Started](./guide/getting-started.md)
- [Configuration Reference](./reference/configuration.md)
- [MCP Tools Reference](./reference/mcp-tools.md)
- [Architecture Overview](./concepts/architecture.md)

## Version

Current version: **v1.0.0**

See [CHANGELOG](../CHANGELOG.md) for version history.
