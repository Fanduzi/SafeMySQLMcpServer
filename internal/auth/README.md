# Auth Module

JWT authentication and token generation for API security.

## Files
| File | Responsibility |
|------|---------------|
| jwt.go | JWT token validation and generation |
| jwt_test.go | Unit tests for JWT operations |

## Exports
- `Validator` - JWT token validator
- `NewValidator(secret string) *Validator` - Create validator
- `NewValidatorFromEnv(secret string) (*Validator, error)` - Create validator from env
- `Validate(token string) (*Claims, error)` - Validate and parse token
- `GenerateToken(userID, email string, expiry time.Duration) (string, error)` - Generate token
- `ExtractToken(authHeader string) string` - Extract Bearer token
- `ContextWithUser(ctx, userID, email) context.Context` - Add user to context
- `GetUserID(ctx) string` - Get user ID from context
- `GetUserEmail(ctx) string` - Get user email from context

## Dependencies
- Upstream: None
- Downstream: `internal/server` - Uses auth middleware

## Update Rule
If JWT handling changes, update this file in the same change.
