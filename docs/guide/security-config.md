# Security Configuration

SafeMySQLMcpServer provides a comprehensive SQL security layer to prevent dangerous operations.

## Security Layers

```
SQL Request
     │
     ▼
┌─────────────────────┐
│ Input Validation │  Validate identifiers
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ SQL Parsing      │  Parse SQL into AST
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ Security Check    │  Check against rules
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ SQL Rewriting    │  Add LIMIT if needed
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ Query Execution │  Execute with timeout
└─────────────────────┘
```

## Configuration File

tab: Configuration File

The security configuration is stored in `config/security.yaml`:

```yaml
security:
  # Allowed DML operations
  # Options: SELECT, INSERT, UPDATE, DELETE
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE

  # Allowed DDL operations
  # Options: CREATE_TABLE, CREATE_INDEX, ALTER_TABLE, DROP_TABLE
  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE

  # Always blocked operations (even if in allowlist)
  # Options: DROP, TRUNCATE, ALTER_USER, GRANT, REVOKE
  blocked:
    - DROP
    - TRUNCATE
    - GRANT
    - REVOKE

  # Auto-add LIMIT to UPDATE/DELETE without WHERE
  auto_limit: 1000

  # Maximum allowed LIMIT value
  max_limit: 10000

  # Query timeout
  query_timeout: 30s

  # Maximum rows returned
  max_rows: 10000

  # Maximum SQL length (characters)
  max_sql_length: 100000
```

## Security Rules
tab: Security Rules

### DML Allowlist

Controls which data manipulation operations are allowed:

| Operation | Default | Description |
|-----------|---------|-------------|
| SELECT | ✅ Allowed | Read data |
| INSERT | ✅ Allowed | Insert data |
| UPDATE | ✅ Allowed | Update data |
| DELETE | ✅ Allowed | Delete data |

### DDL Allowlist

Controls which data definition operations are allowed.

| Operation | Default | Description |
|-----------|---------|-------------|
| CREATE_TABLE | ✅ Allowed | Create tables |
| CREATE_INDEX | ✅ Allowed | Create indexes |
| ALTER_TABLE | ✅ Allowed | Modify tables |
| DROP_TABLE | ❌ Blocked | Drop tables |

### Blocked Operations
Always blocked regardless of allowlist.

| Operation | Reason |
|-----------|--------|
| DROP | Can destroy data |
| TRUNCATE | Can destroy data |
| GRANT | Security risk |
| REVOKE | Security risk |

### Auto LIMIT

Automatically adds LIMIT to dangerous operations:

| Condition | Action |
|-----------|--------|
| UPDATE without WHERE | Add LIMIT 1000 |
| DELETE without WHERE | Add LIMIT 1000 |

### Query Limits

tab: Query Limits

| Setting | Default | Description |
|---------|---------|-------------|
| `query_timeout` | 30s | Maximum query execution time |
| `max_rows` | 10,000 | Maximum rows returned |
| `max_limit` | 10,000 | Maximum LIMIT value |
| `max_sql_length` | 100,000 | Maximum SQL length |

## Examples
tab: Examples

### Allow Only SELECT

```yaml
security:
  allowed_dml:
    - SELECT
  allowed_ddl: []
  blocked:
    - DROP
    - TRUNCATE
    - INSERT
    - UPDATE
    - DELETE
```

### Allow Read-Write, No DDL

```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE
  allowed_ddl: []
  blocked:
    - DROP
    - TRUNCATE
```

### Full Access (Not Recommended for Production)
```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE
  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE
  blocked: []
  auto_limit: 1000
```
