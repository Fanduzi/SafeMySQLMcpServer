# 安全配置

SafeMySQLMcpServer 提供全面的 SQL 安全层来防止危险操作。

## 安全层

```
SQL 请求
     │
     ▼
┌─────────────────────┐
│ 输入验证           │  验证标识符
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ SQL 解析            │  解析 SQL 为 AST
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 安全检查            │  检查规则
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ SQL 重写            │  添加 LIMIT
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│ 查询执行            │  巻加超时
└─────────────────────┘
```

## 配置文件
tab: 配置文件

安全配置保存在 `config/security.yaml` 中：

```yaml
security:
  # 允许的 DML 操作
  # 选项: SELECT, INSERT, UPDATE, DELETE
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE

  # 允许的 DDL 操作
  # 选项: CREATE_TABLE, CREATE_INDEX, ALTER_TABLE, DROP_TABLE
  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE

  # 始终阻止的操作（即使在允许列表中）
  # 选项: DROP, TRUNCATE, ALTER_USER, GRANT, REVOKE
  blocked:
    - DROP
    - TRUNCATE
    - GRANT
    - REVOKE

  # 为没有 WHERE 的 UPDATE/DELETE 自动添加 LIMIT
  auto_limit: 1000

  # 允许的最大 LIMIT 值
  max_limit: 10000

  # 查询超时
  query_timeout: 30s

  # 返回的最大行数
  max_rows: 10000

  # 最大 SQL 长度（字符数）
  max_sql_length: 100000
```

## 安全规则
tab: 安全规则

### DML 允许列表

控制允许的数据操作操作：

| 操作 | 默认 | 说明 |
|-----------|---------|-------------|
| SELECT | ✅ 允许 | 读取数据 |
| INSERT | ✅ 允许 | 插入数据 |
| UPDATE | ✅ 允许 | 更新数据 |
| DELETE | ✅ 允许 | 删除数据 |

### DDL 允许列表

控制允许的数据定义操作。

| 操作 | 默认 | 说明 |
|-----------|---------|-------------|
| CREATE_TABLE | ✅ 允许 | 创建表 |
| CREATE_INDEX | ✅ 允许 | 创建索引 |
| ALTER_TABLE | ✅ 允许 | 修改表结构 |
| DROP_TABLE | ❌ 阻止 | 删除表 |

### 阻止的操作
始终阻止，即使在允许列表中。

| 操作 | 原因 |
|-----------|--------|
| DROP | 可能破坏数据 |
| TRUNCATE | 可能破坏数据 |
| GRANT | 安全风险 |
| REVOKE | 安全风险 |

### 自动 LIMIT

自动为危险操作添加 LIMIT:

| 条件 | 行为 |
|--------------|--------|
| UPDATE 无 WHERE | 添加 LIMIT 1000 |
| DELETE 无 WHERE | 添加 LIMIT 1000 |

### 查询限制
tab: 查询限制

| 设置 | 默认值 | 说明 |
|---------|---------|-------------|
| `query_timeout` | 30s | 最大查询执行时间 |
| `max_rows` | 10,000 | 返回的最大行数 |
| `max_limit` | 10,000 | 允许的最大 LIMIT 值 |
| `max_sql_length` | 100,000 | 最大 SQL 长度 |

## 示例
tab: 示例

### 仅允许 SELECT

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

### 允许读写，禁止 DDL

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

### 完全访问（生产环境不推荐）
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
