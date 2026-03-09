# Token Generator

CLI tool for generating JWT tokens for API authentication.

## Files
| File | Responsibility |
|------|---------------|
| main.go | CLI entry point, flag parsing, token generation |
| main_test.go | Unit tests for CLI validation |

## Usage
```bash
./token -user <user_id> -email <email> -expire <duration> -secret <jwt_secret>
```

## Flags
| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| -user | Yes | - | User ID |
| -email | Yes | - | User email |
| -expire | No | 24h | Token expiry (e.g., 24h, 7d, 365d) |
| -secret | No | - | JWT secret (or use JWT_SECRET env) |

## Security Warning

**IMPORTANT**: The generated token is output to stdout. Take care to:

1. **Never log the token** - Avoid shell history logging:
   ```bash
   # GOOD: Use directly without shell history
   ./token -user admin -email admin@example.com | pbcopy

   # BAD: Token will be in shell history
   TOKEN=$(./token -user admin -email admin@example.com)
   ```

2. **Never commit tokens** to version control

3. **Use environment variables** for secrets:
   ```bash
   export JWT_SECRET="your-secret-key"
   ./token -user admin -email admin@example.com
   ```

4. **Rotate secrets** if tokens are accidentally exposed

## Dependencies
- Upstream: `internal/auth` - JWT generation
- Downstream: None (CLI tool)

## Update Rule
If CLI flags change, update this file in the same change.
