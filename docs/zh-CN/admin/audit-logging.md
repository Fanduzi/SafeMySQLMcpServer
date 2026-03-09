# 审计日志

SafeMySQLMcpServer 记录所有 SQL 操作，用于合规和调试。

## 日志格式
tab: 日志格式

### JSON 结构

```json
{
  "timestamp": "2026-03-09T10:30:00.123Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "client_ip": "192.168.1.100",
  "database": "mydb",
  "sql": "SELECT * FROM users WHERE id = 123",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 1,
  "duration_ms": 15,
  "block_reason": ""
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|-------|------|-------------|
| timestamp | string | ISO 8601 时间戳 |
| user_id | string | JWT 中的用户标识 |
| user_email | string | JWT 中的用户邮箱 |
| client_ip | string | 客户端 IP 地址 |
| database | string | 目标数据库名 |
| sql | string | SQL 语句（过长时截断） |
| sql_type | string | SQL 类型（SELECT, INSERT 等） |
| status | string | success, error, blocked |
| rows_affected | number | 影响/返回行数 |
| duration_ms | number | 执行时间（毫秒） |
| block_reason | string | 阻止原因（如被阻止） |

## 日志轮转
tab: 日志轮转

### 配置

```yaml
audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000      # SQL 超过 2000 字符时截断
  max_size_mb: 100          # 达到 100MB 时轮转
  max_backups: 10           # 保留 10 个备份文件
  max_age_days: 30          # 30 天后删除备份
  compress: true            # 用 gzip 压缩备份
```

### 轮转行为

1. 当日志达到 `max_size_mb` 时，进行轮转
2. 旧日志重命名为带时间戳的文件: `audit-2026-03-09T10.30.00.log`
3. 如果 `compress: true`，轮转的日志会被 gzip 压缩
4. 超过 `max_backups` 数量或 `max_age_days` 天数的旧备份会被删除

### 手动轮转

```bash
# 强制轮转（发送 USR1 信号）
kill -USR1 <pid>
```

## 日志分析
tab: 日志分析

### 按用户查询

```bash
# 查找某用户的所有查询
jq 'select(.user_id == "admin")' logs/audit.log
```

### 查找被阻止的查询

```bash
# 查找所有被阻止的查询
jq 'select(.status == "blocked")' logs/audit.log
```

### 查找慢查询

```bash
# 查找执行超过 1 秒的查询
jq 'select(.duration_ms > 1000)' logs/audit.log
```

### 按时间范围查询

```bash
# 查找特定时间范围内的查询
jq 'select(.timestamp >= "2026-03-09T10:00:00" and .timestamp < "2026-03-09T11:00:00")' logs/audit.log
```

### 统计分析

```bash
# 按用户统计查询数量
jq -r '.user_id' logs/audit.log | sort | uniq -c | sort -rn

# 按数据库统计平均查询耗时
jq -s 'group_by(.database) | .[] | {database: .[0].database, avg_duration: (map(.duration_ms) | add) / length}'' logs/audit.log
```

## 合规要求
tab: 合规要求

### 数据保留期限

根据合规要求配置保留期限:

| 法规 | 建议保留期限 |
|------------|------------------|
| SOX | 7 年 |
| GDPR | 按需，最小化数据 |
| HIPAA | 6 年 |
| PCI-DSS | 1 年 |

### PII 考虑

审计日志可能包含 PII:

| 字段 | 是否为 PII | 建议 |
|-------|------------|------|
| user_email | 是 | 考虑脱敏处理 |
| client_ip | 可能 | 考虑掩码处理 |
| sql | 可能 | 检查敏感数据 |

### 日志安全

```bash
# 设置限制性权限
chmod 640 logs/audit.log
chown safemysql:safemysql logs/audit.log

# 确保日志目录安全
chmod 750 logs/
```

## 日志转发
tab: 日志转发

### Filebeat 配置

```yaml
# /etc/filebeat/filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /opt/safemysql/logs/audit.log
    json.keys_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "safemysql-audit"
```

### Fluentd 配置

```xml
<!-- /etc/fluentd/fluent.conf -->
<source>
  @type tail
  path /opt/safemysql/logs/audit.log
  tag safemysql.audit
  <parse>
    @type json
  </parse>
</source>

<match safemysql.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
</match>
```
