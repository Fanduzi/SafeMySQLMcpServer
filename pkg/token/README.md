# Token Generator

CLI tool for generating JWT tokens for API authentication.

## Files
| File | Responsibility |
|------|---------------|
| main.go | CLI entry point, flag parsing, token generation |

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

## Dependencies
- Upstream: `internal/auth` - JWT generation
- Downstream: None (CLI tool)

## Update Rule
If CLI flags change, update this file in the same change.
