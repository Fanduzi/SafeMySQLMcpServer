# Constants Module

Shared constants used across the application.

## Files
| File | Responsibility |
|------|---------------|
| constants.go | Application-wide constants |

## Exports
- `Version` - Application version
- `AppName` - Application name
- `DefaultPort` - Default HTTP port (8080)
- `DefaultConfigPath` - Default config file path

## Dependencies
- Upstream: None
- Downstream: All modules use these constants

## Update Rule
If constants change, update this file in the same change.
