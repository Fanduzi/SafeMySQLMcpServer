# MCP Tools Reference

Complete reference for all MCP tools provided by SafeMySQLMcpServer.

## Tool List

| Tool | Description | Category |
|------|-------------|----------|
| [query](#query) | Execute SQL query | DML |
| [list_databases](#list_databases) | List available databases | Metadata |
| [list_tables](#list_tables) | List tables in database | Metadata |
| [describe_table](#describe_table) | Get table structure | Metadata |
| [show_create_table](#show_create_table) | Get CREATE TABLE statement | DDL |
| [explain](#explain) | Get query execution plan | Analysis |
| [search_tables](#search_tables) | Search tables by pattern | Search |

## query
tab: query

### Description
Execute SQL queries on specified database.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| database | string | Yes | Target database name |
| sql | string | Yes | SQL statement to execute |

### Example Request

```json
{
  "name": "query",
  "arguments": {
    "database": "mydb",
    "sql": "SELECT * FROM users WHERE id = 123"
  }
}
```

### Example Response

```json
{
  "columns": ["id", "name", "email", "created_at"],
  "rows": [
    [123, "John Doe", "john@example.com", "2026-01-15T10:30:00Z"]
  ],
  "rows_affected": 1
}
```

### Allowed Operations

| SQL Type | Configuration | Default |
|----------|--------------|---------|
| SELECT | `allowed_dml` | ✅ Allowed |
| INSERT | `allowed_dml` | ✅ Allowed |
| UPDATE | `allowed_dml` | ✅ Allowed |
| DELETE | `allowed_dml` | ✅ Allowed |

### Limitations

- Max rows: configured by `max_rows` (default: 10,000)
- Query timeout: configured by `query_timeout` (default: 30s)
- UPDATE/DELETE without WHERE: auto-limited to `auto_limit` rows

## list_databases
tab: list_databases

### Description
List all databases configured in the routing.

### Parameters
None

### Example Request

```json
{
  "name": "list_databases",
  "arguments": {}
}
```

### Example Response

```json
{
  "databases": ["user_db", "order_db", "analytics_db"]
}
```

### Notes
- Only returns databases configured in `databases` section
- Does not query MySQL for database list

## list_tables
tab: list_tables

### Description
List tables in a specified database.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| database | string | Yes | Database name |

### Example Request

```json
{
  "name": "list_tables",
  "arguments": {
    "database": "mydb"
  }
}
```

### Example Response

```json
{
  "tables": [
    {"name": "users", "comment": "User accounts"},
    {"name": "orders", "comment": "Customer orders"},
    {"name": "products", "comment": ""}
  ],
  "truncated": false
}
```

### Notes
- Max tables returned: 1,000
- `truncated: true` indicates more tables exist

## describe_table
tab: describe_table

### Description
Get column information for a table.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| database | string | Yes | Database name |
| table | string | Yes | Table name |

### Example Request

```json
{
  "name": "describe_table",
  "arguments": {
    "database": "mydb",
    "table": "users"
  }
}
```

### Example Response

```json
{
  "columns": [
    {
      "field": "id",
      "type": "int(11)",
      "null": "NO",
      "key": "PRI",
      "default": "",
      "extra": "auto_increment",
      "comment": "Primary key"
    },
    {
      "field": "email",
      "type": "varchar(255)",
      "null": "NO",
      "key": "UNI",
      "default": "",
      "extra": "",
      "comment": "User email"
    }
  ]
}
```

## show_create_table
tab: show_create_table

### Description
Get the CREATE TABLE statement for a table.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| database | string | Yes | Database name |
| table | string | Yes | Table name |

### Example Request

```json
{
  "name": "show_create_table",
  "arguments": {
    "database": "mydb",
    "table": "users"
  }
}
```

### Example Response

```json
{
  "create_statement": "CREATE TABLE `users` (\n  `id` int NOT NULL AUTO_INCREMENT,\n  `email` varchar(255) NOT NULL,\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `idx_email` (`email`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
}
```

## explain
tab: explain

### Description
Get query execution plan for a SQL statement.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| database | string | Yes | Database name |
| sql | string | Yes | SQL statement to explain |

### Example Request

```json
{
  "name": "explain",
  "arguments": {
    "database": "mydb",
    "sql": "SELECT * FROM users WHERE email = 'test@example.com'"
  }
}
```

### Example Response

```json
{
  "columns": ["id", "select_type", "table", "partitions", "type", "possible_keys", "key", "key_len", "ref", "rows", "filtered", "Extra"],
  "rows": [
    [1, "SIMPLE", "users", null, "const", "idx_email", "idx_email", "1022", "const", 1, 100.00, null]
  ]
}
```

### Supported SQL Types

| SQL Type | Supported |
|----------|-----------|
| SELECT | ✅ |
| INSERT | ✅ |
| UPDATE | ✅ |
| DELETE | ✅ |
| DDL | ❌ |

## search_tables
tab: search_tables

### Description
Search for tables by name pattern across all databases.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| table_pattern | string | Yes | Pattern to search (supports LIKE wildcards) |

### Example Request

```json
{
  "name": "search_tables",
  "arguments": {
    "table_pattern": "user%"
  }
}
```

### Example Response

```json
{
  "matches": [
    {"database": "user_db", "table": "users"},
    {"database": "user_db", "table": "user_profiles"},
    {"database": "order_db", "table": "user_orders"}
  ],
  "truncated": false
}
```

### Pattern Syntax

| Pattern | Matches |
|---------|---------|
| `user%` | Tables starting with "user" |
| `%log%` | Tables containing "log" |
| `%_backup` | Tables ending with "_backup" |

### Notes
- Case-insensitive search
- Max results: `max_rows` (default: 10,000)
- LIKE special characters (`%`, `_`) are escaped in input
