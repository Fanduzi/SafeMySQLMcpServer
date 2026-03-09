# Token Generator

CLI tool for generating JWT tokens for API authentication.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Token Generator CLI                       │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Input Sources                         ││
│  │                                                          ││
│  │   Flags:    -user, -email, -expire, -secret             ││
│  │   Env:      JWT_SECRET                                  ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Validation                            ││
│  │                                                          ││
│  │   • User/Email required                                 ││
│  │   • Secret from flag or env                             ││
│  │   • Expiry duration parsing                             ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   JWT Generation                        ││
│  │                                                          ││
│  │   Header:  {"alg":"HS256","typ":"JWT"}                  ││
│  │   Payload: {"user_id":"...","user_email":"..."}         ││
│  │   Sign:    HMAC-SHA256(secret)                          ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Output                                ││
│  │                                                          ││
│  │   stdout: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...      ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| main.go | CLI entry point, flag parsing, token generation | ~80 |
| main_test.go | Unit tests for CLI validation | ~60 |

## Usage

### Basic
```bash
# With environment variable
export JWT_SECRET="your-secret-key-min-32-characters-long"
./bin/mysql-mcp-token --user admin --email admin@example.com

# With flag
./bin/mysql-mcp-token \
  --user admin \
  --email admin@example.com \
  --secret your-secret-key
```

### With Expiry
```bash
# 1 year token
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d

# 7 days token
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 168h
```

### Docker
```bash
docker exec safemysql-app /app/token \
  --user admin \
  --email admin@example.com \
  --secret your-jwt-secret
```

## Flags
| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-user` | Yes | - | User ID (stored in JWT claims) |
| `-email` | Yes | - | User email (stored in JWT claims) |
| `-expire` | No | 24h | Token expiry duration |
| `-secret` | No* | - | JWT signing secret |
| `-help` | No | - | Show help |

*If `-secret` not provided, reads from `JWT_SECRET` environment variable.

## Expiry Format
| Format | Example | Description |
|--------|---------|-------------|
| Hours | `24h`, `168h` | Hours until expiry |
| Days | `7d`, `30d`, `365d` | Days until expiry |

## Output
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4iLCJ1c2VyX2VtYWlsIjoiYWRtaW5AZXhhbXBsZS5jb20iLCJleHAiOjE2NDY4MjQ2MDAsImlhdCI6MTY0NjgyMTAwMH0.xxxxx
```

## Security Warning

**IMPORTANT**: The generated token is output to stdout. Take care to:

1. **Never log the token** - Avoid shell history logging:
   ```bash
   # GOOD: Use directly without shell history
   ./bin/mysql-mcp-token --user admin --email admin@example.com | pbcopy

   # BAD: Token will be in shell history
   TOKEN=$(./bin/mysql-mcp-token --user admin --email admin@example.com)
   ```

2. **Never commit tokens** to version control

3. **Use environment variables** for secrets:
   ```bash
   export JWT_SECRET="your-secret-key"
   ./bin/mysql-mcp-token --user admin --email admin@example.com
   ```

4. **Rotate secrets** if tokens are accidentally exposed

## Token Structure
```json
// Header
{
  "alg": "HS256",
  "typ": "JWT"
}

// Payload
{
  "user_id": "admin",
  "user_email": "admin@example.com",
  "exp": 1646824600,
  "iat": 1646821000
}
```

## Dependencies
```
Upstream:
  └── internal/auth  → GenerateToken function

Downstream:
  └── None (CLI tool)

External:
  └── github.com/golang-jwt/jwt/v5  → JWT library
```

## Update Rule
If CLI flags change, update:
1. This file
2. main.go
3. docs/guide/authentication.md
