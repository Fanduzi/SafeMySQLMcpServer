# Security Module

SQL parsing, validation, and rewriting for SQL injection prevention.

## Files
| File | Responsibility |
|------|---------------|
| parser.go | Parse SQL using TiDB parser |
| checker.go | Check SQL against security rules |
| rewriter.go | Rewrite dangerous SQL (add LIMIT, etc.) |
| doc.go | Package documentation |
| *_test.go | Unit tests |

## Exports
### Parser
- `Parser` - SQL parser wrapper
- `NewParser() *Parser` - Create parser
- `Parse(sql string) (*ParsedSQL, error)` - Parse SQL
- `ParsedSQL` - Parsed result with type, tables, etc.

### Checker
- `Checker` - Security checker
- `NewChecker(rules *config.SecurityRules) *Checker` - Create checker
- `Check(parsed *ParsedSQL) *CheckResult` - Check SQL

### Rewriter
- `Rewriter` - SQL rewriter
- `NewRewriter(rules *config.SecurityRules) *Rewriter` - Create rewriter
- `Rewrite(parsed *ParsedSQL) *RewriteResult` - Rewrite SQL

## Security Checks
1. Parse SQL to AST
2. Check against allowed DML/DDL
3. Check against blocked operations
4. Auto-add LIMIT for UPDATE/DELETE without WHERE
5. Cap LIMIT to max_limit

## Dependencies
- Upstream: `internal/config` - Security rules
- Downstream: `internal/mcp` - Uses checker/rewriter

## Update Rule
If security logic changes, update this file in the same change.
