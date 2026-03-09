# Getting Started

This guide walks you installing and running SafeMySQLMcpServer.

## Installation

tab: Installation

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer

# Start with docker-compose (includes MySQL)
docker-compose up -d

# Check server health
curl http://localhost:8080/health
```

### Option 2: Binary Installation

```bash
# Build from source
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer
make build

# The binary will be at ./bin/safe-mysql-mcp
```

### Option 3: Go Install

```bash
go install github.com/YOUR_USERNAME/safe-mysql-mcp/cmd/server@latest
```

## Configuration
tab: Configuration

### 1. Create Configuration Files

```bash
# Copy example configurations
cp -r examples/config config

# Edit configuration
vim config/config.yaml
vim config/security.yaml
```

### 2. Set Environment Variables

```bash
# Required: JWT secret (minimum 32 characters)
export JWT_SECRET="your-secret-key-min-32-characters-long"

# Database credentials
export DEV_DB_USER="your_db_user"
export DEV_DB_PASSWORD="your_db_password"
```

### 3. Generate Authentication Token

```bash
# Generate a token for your user
./bin/mysql-mcp-token --user admin --email admin@example.com

# Or with custom expiration
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d
```

## Running the Server
tab: Running

### Start the Server

```bash
./bin/safe-mysql-mcp -config config/config.yaml
```

### Verify the Server is Running

```bash
# Health check
curl http://localhost:8080/health
# Expected: OK

# Metrics endpoint
curl http://localhost:8080/metrics
# Expected: Prometheus metrics output
```

### Test Authentication

```bash
# Replace YOUR_TOKEN with the token generated earlier
curl -X POST http://localhost:8080/mcp \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## Connecting Claude Code
tab: Claude Code

Add to your Claude Code configuration:

```json
{
  "mcpServers": {
    "mysql-dev": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

See [Integration with Claude Code](./claude-code.md) for detailed instructions.

## Next Steps

- [Configure Security Rules](./security-config.md)
- [Set up Audit Logging](../admin/audit-logging.md)
- [Monitor with Prometheus](../admin/monitoring.md)
