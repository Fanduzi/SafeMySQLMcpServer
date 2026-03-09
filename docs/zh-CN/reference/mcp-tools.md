# MCP 工具参考

SafeMySQLMcpServer 提供的所有 MCP 工具完整参考。

## 工具列表

| 工具 | 说明 | 类别 |
|------|-------------|--------|
| [query](#query) | 执行 SQL 查询 | DML |
| [list_databases](#list_databases) | 列出可用数据库 | 元数据 |
| [list_tables](#list_tables) | 列出数据库中的表 | 元数据 |
| [describe_table](#describe_table) | 获取表结构 | 元数据 |
| [show_create_table](#show_create_table) | 获取 CREATE TABLE 语句 | DDL |
| [explain](#explain) | 获取查询执行计划 | 分析 |
| [search_tables](#search_tables) | 按模式搜索表 | 搜索 |

## query
tab: query

### 说明
在指定数据库上执行 SQL 查询。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| database | string | 是 | 目标数据库名 |
| sql | string | 是 | 要执行的 SQL 语句 |

### 示例请求

```json
{
  "name": "query",
  "arguments": {
    "database": "mydb",
    "sql": "SELECT * FROM users WHERE id = 123"
  }
}
```

### 示例响应

```json
{
  "columns": ["id", "name", "email", "created_at"],
  "rows": [
    [123, "John Doe", "john@example.com", "2026-01-15T10:30:00Z"]
  ],
  "rows_affected": 1
}
```

### 允许的操作

| SQL 类型 | 配置 | 默认 |
|----------|--------------|---------|
| SELECT | `allowed_dml` | ✅ 允许 |
| INSERT | `allowed_dml` | ✅ 允许 |
| UPDATE | `allowed_dml` | ✅ 允许 |
| DELETE | `allowed_dml` | ✅ 允许 |

### 限制

- 最大行数: 由 `max_rows` 配置（默认: 10,000）
- 查询超时: 由 `query_timeout` 配置（默认: 30s）
- 无 WHERE 的 UPDATE/DELETE: 自动添加 `auto_limit` 行的 LIMIT

## list_databases
tab: list_databases

### 说明
列出路由中配置的所有数据库。

### 参数
无

### 示例请求

```json
{
  "name": "list_databases",
  "arguments": {}
}
```

### 示例响应

```json
{
  "databases": ["user_db", "order_db", "analytics_db"]
}
```

### 注意事项
- 只返回 `databases` 部分配置的数据库
- 不会查询 MySQL 获取数据库列表

## list_tables
tab: list_tables

### 说明
列出指定数据库中的表。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| database | string | 是 | 数据库名 |

### 示例请求

```json
{
  "name": "list_tables",
  "arguments": {
    "database": "mydb"
  }
}
```

### 示例响应

```json
{
  "tables": [
    {"name": "users", "comment": "用户账户"},
    {"name": "orders", "comment": "客户订单"},
    {"name": "products", "comment": ""}
  ],
  "truncated": false
}
```

### 注意事项
- 返回的最大表数: 1,000
- `truncated: true` 表示还有更多表

## describe_table
tab: describe_table

### 说明
获取表的列信息。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| database | string | 是 | 数据库名 |
| table | string | 是 | 表名 |

### 示例请求

```json
{
  "name": "describe_table",
  "arguments": {
    "database": "mydb",
    "table": "users"
  }
}
```

### 示例响应

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
      "comment": "主键"
    },
    {
      "field": "email",
      "type": "varchar(255)",
      "null": "NO",
      "key": "UNI",
      "default": "",
      "extra": "",
      "comment": "用户邮箱"
    }
  ]
}
```

## show_create_table
tab: show_create_table

### 说明
获取表的 CREATE TABLE 语句。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| database | string | 是 | 数据库名 |
| table | string | 是 | 表名 |

### 示例请求

```json
{
  "name": "show_create_table",
  "arguments": {
    "database": "mydb",
    "table": "users"
  }
}
```

### 示例响应

```json
{
  "create_statement": "CREATE TABLE `users` (\n  `id` int NOT NULL AUTO_INCREMENT,\n  `email` varchar(255) NOT NULL,\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `idx_email` (`email`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
}
```

## explain
tab: explain

### 说明
获取 SQL 语句的查询执行计划。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| database | string | 是 | 数据库名 |
| sql | string | 是 | 要解释的 SQL 语句 |

### 示例请求

```json
{
  "name": "explain",
  "arguments": {
    "database": "mydb",
    "sql": "SELECT * FROM users WHERE email = 'test@example.com'"
  }
}
```

### 示例响应

```json
{
  "columns": ["id", "select_type", "table", "partitions", "type", "possible_keys", "key", "key_len", "ref", "rows", "filtered", "Extra"],
  "rows": [
    [1, "SIMPLE", "users", null, "const", "idx_email", "idx_email", "1022", "const", 1, 100.00, null]
  ]
}
```

### 支持的 SQL 类型

| SQL 类型 | 支持 |
|----------|-----------|
| SELECT | ✅ |
| INSERT | ✅ |
| UPDATE | ✅ |
| DELETE | ✅ |
| DDL | ❌ |

## search_tables
tab: search_tables

### 说明
在所有数据库中按名称模式搜索表。

### 参数

| 参数 | 类型 | 必需 | 说明 |
|-----------|------|----------|-------------|
| table_pattern | string | 是 | 搜索模式（支持 LIKE 通配符） |

### 示例请求

```json
{
  "name": "search_tables",
  "arguments": {
    "table_pattern": "user%"
  }
}
```

### 示例响应

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

### 模式语法

| 模式 | 匹配 |
|---------|---------|
| `user%` | 以 "user" 开头的表 |
| `%log%` | 包含 "log" 的表 |
| `%_backup` | 以 "_backup" 结尾的表 |

### 注意事项
- 不区分大小写的搜索
- 最大结果数: `max_rows`（默认: 10,000）
- LIKE 特殊字符（`%`, `_`）在输入中会被转义
