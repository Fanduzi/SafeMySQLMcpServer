# 热重载配置

SafeMySQLMcpServer 支持热配置重载，允许你在不重启服务的情况下更新设置。

## 工作原理

```
┌───────────────────┐     ┌───────────────────┐
│  配置文件         │     │  安全配置文件      │
│  config.yaml      │     │  security.yaml   │
└───────────────────┘     └───────────────────┘
         │                        │
         │    fsnotify 监听        │
         ▼                        ▼
┌─────────────────────────────────────────────────┐
│              配置监视器                          │
│  - 检测文件变化                                 │
│  - 防抖快速连续变化                               │
│  - 解析新配置                                   │
└─────────────────────────────────────────────────┘
         │
         │ OnChange 回调
         ▼
┌─────────────────────────────────────────────────┐
│              服务器更新                          │
│  - 更新连接池                                   │
│  - 更新安全规则                                 │
│  - 更新审计设置                                 │
└─────────────────────────────────────────────────┘
```

## 配置文件
tab: 配置文件

### 主配置 (config.yaml)

触发重载的变化:

| 部分 | 热重载 | 说明 |
|---------|---------------|-------|
| `server` | 部分 | 端口/主机需要重启 |
| `clusters` | 是 | 连接优雅更新 |
| `databases` | 是 | 路由立即更新 |
| `audit` | 是 | 日志记录器设置更新 |
| `rate_limit` | 是 | 速率限制更新 |

### 安全配置 (security.yaml)

所有设置都支持热重载:

| 设置 | 热重载 |
|---------|---------------|
| `allowed_dml` | 是 |
| `allowed_ddl` | 是 |
| `blocked` | 是 |
| `auto_limit` | 是 |
| `query_timeout` | 是 |
| `max_rows` | 是 |

## 更新行为
tab: 更新行为

### 连接池更新

当集群配置变化时:

1. 新连接使用新设置
2. 现有连接完成其查询
3. 空闲连接优雅关闭

```go
func (p *Pool) UpdateConfig(clusters ClustersConfig) error {
    // 为新集群创建新连接
    // 更新现有集群的设置
    // 标记要移除的集群以进行清理
}
```

### 安全规则更新

当安全配置变化时:

1. 新规则立即应用
2. 进行中的查询继续使用旧规则
3. 新查询使用新规则

```go
func (rc *ReloadableConfig) Update(cfg *Config, security *SecurityConfig) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.config = cfg
    rc.security = security
}
```

## 示例
tab: 示例

### 添加新数据库

```yaml
# config/config.yaml
databases:
  user_db:
    cluster: dev-cluster-1
  order_db:
    cluster: dev-cluster-1
  # 添加新数据库 - 热重载会立即使其可用
  analytics_db:
    cluster: dev-cluster-1
```

```bash
# 保存文件 - 变化会自动检测
# 无需重启！
# 新数据库立即可用
```

### 收紧安全规则

```yaml
# config/security.yaml
security:
  allowed_dml:
    - SELECT
    # 移除 INSERT/UPDATE/DELETE - 热重载立即应用
```

```bash
# 保存文件
# 新查询立即被阻止
# 进行中的查询正常完成
```

## 限制
tab: 限制

### 需要重启的更改

这些更改需要服务器重启。

| 更改 | 原因 |
|-------|--------|
| 服务器端口 | 需要新的监听器 |
| 服务器主机 | 需要新的监听器 |
| JWT 密钥 | 使所有 token 失效 |

### 优雅的连接处理

当集群设置变化时:
- 现有查询正常完成
- 新查询使用新设置
- 空闲连接被更新
- 活跃连接保留直到空闲
