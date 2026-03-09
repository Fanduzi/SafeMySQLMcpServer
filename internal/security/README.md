# Security Module

SQL parsing, validation, and rewriting for safe database operations.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Pipeline                         │
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐              │
│  │  Parser  │───▶│ Checker  │───▶│ Rewriter │              │
│  │ (vitess) │    │  (rules) │    │  (LIMIT) │              │
│  └──────────┘    └──────────┘    └──────────┘              │
│       │              │               │                      │
│       ▼              ▼               ▼                      │
│  ParsedSQL      CheckResult    RewriteResult               │
│  • Type         • Allowed      • Modified                  │
│  • Tables       • Reason       • SQL                       │
│  • AST          • Violations                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| parser.go | Parse SQL using vitess parser | ~100 |
| checker.go | Check SQL against security rules | ~150 |
| rewriter.go | Rewrite dangerous SQL (add LIMIT) | ~100 |
| doc.go | Package documentation | ~20 |
| parser_test.go | Parser unit tests | ~80 |
| checker_security_test.go | Security checker tests | ~150 |
| rewriter_security_test.go | Rewriter tests | ~100 |

## Test Coverage
```
Coverage: ~90%
- SQL parsing for all statement types
- DML/DDL allowlist checking
- Blocked operation detection
- Auto-LIMIT for unsafe operations
- LIMIT capping to max_limit
```

## Exports

### Parser
```go
type Parser struct {
    parser *sqlparser.Parser
}

func NewParser() *Parser
func (p *Parser) Parse(sql string) (*ParsedSQL, error)

type ParsedSQL struct {
    Type      sqlparser.StatementType  // SELECT, INSERT, UPDATE, DELETE, DDL
    Tables    []string                 // Tables involved
    Statement sqlparser.Statement      // AST
    RawSQL    string                   // Original SQL
}
```

### Checker
```go
type Checker struct {
    rules *config.SecurityRules
}

func NewChecker(rules *config.SecurityRules) *Checker
func (c *Checker) Check(parsed *ParsedSQL) *CheckResult

type CheckResult struct {
    Allowed     bool     // Is operation allowed?
    Reason      string   // Why blocked (if blocked)
    Violations  []string // List of violations
}
```

### Rewriter
```go
type Rewriter struct {
    rules *config.SecurityRules
}

func NewRewriter(rules *config.SecurityRules) *Rewriter
func (r *Rewriter) Rewrite(parsed *ParsedSQL) *RewriteResult

type RewriteResult struct {
    Modified bool   // Was SQL modified?
    SQL      string // Rewritten SQL
    Reason   string // Why modified
}
```

## Security Rules

### Check Pipeline
```
1. Parse SQL → AST
2. Check DML allowlist (SELECT, INSERT, UPDATE, DELETE)
3. Check DDL allowlist (CREATE_TABLE, CREATE_INDEX, ALTER_TABLE)
4. Check blocked operations (DROP, TRUNCATE)
5. Return CheckResult
```

### Rewrite Pipeline
```
1. Check if UPDATE/DELETE without WHERE
2. If yes, add LIMIT auto_limit
3. Check if LIMIT > max_limit
4. If yes, cap to max_limit
5. Return RewriteResult
```

### Default Rules (security.yaml)
| Rule | Default | Description |
|------|---------|-------------|
| allowed_dml | [SELECT, INSERT, UPDATE, DELETE] | Allowed DML ops |
| allowed_ddl | [CREATE_TABLE, CREATE_INDEX, ALTER_TABLE] | Allowed DDL ops |
| blocked | [DROP, TRUNCATE] | Always blocked |
| auto_limit | 1000 | LIMIT for unsafe ops |
| max_limit | 10000 | Maximum LIMIT |
| query_timeout | 30s | Query timeout |
| max_rows | 10000 | Max rows returned |

## Dependencies
```
Upstream:
  └── internal/config  → SecurityRules

Downstream:
  └── internal/mcp     → Uses Parser, Checker, Rewriter

External:
  └── github.com/mdibaiee/vitess-go/sqlparser  → SQL parsing
```

## Update Rule
If security logic changes, update:
1. This file (rules/pipeline)
2. Relevant .go file (implementation)
3. *_test.go (tests)
4. docs/reference/security-config.md
