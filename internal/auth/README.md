# Auth Module

JWT authentication and token generation for API security.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    JWT Authentication                        │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Token Generation                       ││
│  │                                                          ││
│  │   mysql-mcp-token --user admin --email a@b.com          ││
│  │                          │                               ││
│  │                          ▼                               ││
│  │   JWT: header.payload.signature                         ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Token Validation                       ││
│  │                                                          ││
│  │   Authorization: Bearer <token>                         ││
│  │                          │                               ││
│  │                          ▼                               ││
│  │   Validate signature → Check expiry → Extract claims    ││
│  │                          │                               ││
│  │                          ▼                               ││
│  │   context.WithValue(userID, userEmail)                  ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| jwt.go | JWT token validation and generation | ~100 |
| jwt_test.go | Unit tests for JWT operations | ~80 |

## Test Coverage
```
Coverage: ~90%
- Token generation with various expiry times
- Token validation (valid, expired, invalid signature)
- Claims extraction
- Context operations
- Bearer token extraction
```

## Exports

### Validator
```go
type Validator struct {
    secret []byte
}

func NewValidator(secret string) *Validator
func NewValidatorFromEnv() (*Validator, error)  // Reads JWT_SECRET
func (v *Validator) Validate(tokenString string) (*Claims, error)
```

### Claims
```go
type Claims struct {
    UserID    string `json:"user_id"`
    UserEmail string `json:"user_email"`
    jwt.RegisteredClaims
}
```

### Token Generation
```go
func GenerateToken(userID, email string, secret string, expiry time.Duration) (string, error)
```

### Context Operations
```go
func ContextWithUser(ctx context.Context, userID, email string) context.Context
func GetUserID(ctx context.Context) string
func GetUserEmail(ctx context.Context) string
```

### Token Extraction
```go
func ExtractToken(authHeader string) string  // "Bearer xxx" → "xxx"
```

## JWT Token Structure
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "admin",
    "user_email": "admin@example.com",
    "exp": 1646824600,
    "iat": 1646821000
  },
  "signature": "..."
}
```

## Configuration
| Env Variable | Required | Description |
|--------------|----------|-------------|
| `JWT_SECRET` | Yes | Signing secret (min 32 chars) |

## Usage Example
```bash
# Generate token
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d

# Use token
curl -H "Authorization: Bearer <token>" http://localhost:8080/mcp
```

## Dependencies
```
Upstream: None

Downstream:
  └── internal/server  → Auth middleware

External:
  └── github.com/golang-jwt/jwt/v5  → JWT library
```

## Update Rule
If JWT handling changes, update:
1. This file
2. jwt.go
3. jwt_test.go
4. docs/guide/authentication.md
