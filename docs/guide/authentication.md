# Authentication

SafeMySQLMcpServer uses JWT (JSON Web Token) authentication for all MCP requests.

## Overview

```
Client                    Server
   │                        │
   │  1. Request + Token    │
   │ ─────────────────────>│
   │                        │
   │  2. Validate Token    │
   │                        │
   │  3. Return Response │
   │ <─────────────────────│
```

## Token Generation
tab: Token Generation

### Using the Token CLI Tool

```bash
# Basic usage
./bin/mysql-mcp-token --user admin --email admin@example.com

# With custom expiration (default: 24h)
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d

# Using environment variable for secret
export JWT_SECRET="your-secret-key"
./bin/mysql-mcp-token --user admin --email admin@example.com
```

### CLI Options

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--user` | Yes | - | User identifier |
| `--email` | Yes | - | User email address |
| `--expire` | No | 24h | Token expiration duration |
| `--secret` | No | - | JWT secret (or use JWT_SECRET env) |

### Token Output

The token is output to stdout. Use it in the Authorization header:

```bash
# Capture token to clipboard (macOS)
./bin/mysql-mcp-token --user admin --email admin@example.com | pbcopy

# Use in curl
curl -H "Authorization: Bearer $(./bin/mysql-mcp-token --user admin --email admin@example.com)" ...
```

## Security Best Practices
tab: Security Best Practices

### 1. Never Log Tokens

```bash
# BAD: Token will be in shell history
TOKEN=$(./bin/mysql-mcp-token --user admin --email admin@example.com)

# GOOD: Use directly without capturing
./bin/mysql-mcp-token --user admin --email admin@example.com | pbcopy
```

### 2. Use Strong Secrets

The JWT secret must be at least 32 characters:

```bash
# Generate a strong secret
openssl rand -base64 32
```

### 3. Rotate Secrets

Periodically rotate JWT secrets:

1. Generate new secret
2. Update configuration
3. Restart server
4. Generate new tokens
5. Invalidate old tokens (they will naturally expire)

### 4. Use Environment Variables

Never hardcode secrets in configuration files:

```yaml
# GOOD
server:
  jwt_secret: ${JWT_SECRET}

# BAD
server:
  jwt_secret: "hardcoded-secret-do-not-do-this"
```

## Token Validation
tab: Token Validation

The server validates tokens on each MCP request:

```go
// Server-side validation
claims, err := validator.Validate(token)
if err != nil {
    return http.Error(401, "invalid token")
}
```

### Token Claims

| Claim | Description |
|-------|-------------|
| `sub` | User ID |
| `email` | User email |
| `iat` | Issued at timestamp |
| `exp` | Expiration timestamp |

### Error Responses

| Status Code | Error | Description |
|-------------|-------|-------------|
| 401 | missing authorization token | No token provided |
| 401 | invalid token | Token is invalid or expired |
| 401 | token has expired | Token has passed expiration time |
