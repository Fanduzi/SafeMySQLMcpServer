# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- OpenAPI specification documentation (`docs/openapi.yaml`)
- Three-level documentation system (L1: Root README, L2: Module READMEs, L3: File headers)
- Documentation validation script (`scripts/check-docs.sh`)

### Changed

- Improved test coverage across multiple packages

## [1.0.0] - 2024-01-15

### Added

#### Core Features
- MCP (Model Context Protocol) server implementation with HTTP transport
- JWT authentication with configurable expiration
- SQL security layer with parsing, validation, and rewriting
- Audit logging with rotation and compression support
- Hot configuration reload without server restart
- Prometheus metrics for observability

#### MCP Tools
- `query` - Execute SQL queries on specified databases
- `list_databases` - List all available databases
- `list_tables` - List tables in a database
- `describe_table` - Get table structure information
- `show_create_table` - Get CREATE TABLE statement
- `explain` - Get SQL execution plan
- `search_tables` - Search tables by name pattern across databases

#### Security Features
- SQL injection prevention via identifier validation
- Configurable DML/DDL allowlists
- SQL statement parsing and analysis
- Auto LIMIT for dangerous operations
- Query timeout and row limits
- Blocked operations list

#### Infrastructure
- Docker and Docker Compose support
- GitHub Actions CI/CD pipeline
- Multi-version Go testing (1.22, 1.23, 1.24)
- golangci-lint integration
- Security scanning with gosec
- MySQL integration tests

#### Observability
- Prometheus metrics endpoint (`/metrics`)
- Request duration tracking
- Query count and error rate metrics
- Rate limiting metrics

### Security

- Input validation for all database and table identifiers
- SQL parsing to detect potentially dangerous operations
- Audit logging of all SQL operations
- JWT token validation with configurable secret

## [0.1.0] - 2024-01-01

### Added

- Initial project structure
- Basic HTTP server setup
- Configuration loading from YAML
- Database connection pooling
- JWT token generation utility

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 1.0.0 | 2024-01-15 | First stable release |
| 0.1.0 | 2024-01-01 | Initial development version |

---

[Unreleased]: https://github.com/your-org/SafeMySQLMcpServer/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/your-org/SafeMySQLMcpServer/releases/tag/v1.0.0
[0.1.0]: https://github.com/your-org/SafeMySQLMcpServer/releases/tag/v0.1.0
