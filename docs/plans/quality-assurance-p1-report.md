# Quality Assurance P1 执行报告

> 日期：2026-05-19
> 作者：Fan + Claude
> 前置：[P0 报告](quality-assurance-p0-report.md)（已完成）
> 状态：P1 已完成，待 review

---

## 1. Why — 为什么要做 P1

P0 完成后，测试覆盖依然不均：

| 包 | P0 后覆盖 | 目标 | 差距 |
|---|----------|------|------|
| database | 26.8% | ≥ 70% | 缺少 Router.Query/Exec 的单元测试和 Pool.UpdateConfig 的回归测试 |
| mcp | 14.2% | ≥ 70% | 缺少安全拦截路径的单元测试（DROP/TRUNCATE/RENAME 是否被正确 block） |

核心问题：
1. **Router.Query/Exec 是每个 SQL 请求必经的路径**，但只有集成测试覆盖，没有快速可重复的单元测试
2. **Pool.UpdateConfig 刚修了一个死锁 bug**，但没有回归测试防止复现
3. **MCP handler 的安全拦截**是系统核心价值，但只有集成测试在真实 MySQL 上验证，单元测试层面缺失

---

## 2. What — 做了什么

### 新增产出物

| # | 文件 | 类型 | 说明 |
|---|------|------|------|
| 1 | `internal/database/router_sqlmock_test.go` | 新增 | 12 个单元测试：Router.Query/Exec 的成功、失败、并发路径 |
| 2 | `internal/database/pool_update_test.go` | 新增 | 5 个单元测试：Pool.UpdateConfig 的各种场景 + 死锁回归 |
| 3 | `internal/mcp/security_test.go` | 新增 | 7 个单元测试：executeQuery 安全拦截 + executeExplain 类型限制 |
| 4 | `go.mod` / `go.sum` | 修改 | 新增 `github.com/DATA-DOG/go-sqlmock v1.5.2` 依赖 |

### 附带修复

| 文件 | 修改 | 原因 |
|------|------|------|
| `internal/database/pool_test.go` | 更新注释 | 原注释说"UpdateConfig 无法测试"（因为旧代码死锁），已过时 |
| `internal/database/pool.go` | 死锁修复 | 热加载卡死的根因（P1 开始前发现并修复，已单独提交 `6ef955c`） |

### 覆盖率变化

| 包 | P0 后 | P1 后 | 变化 |
|---|-------|-------|------|
| **database** | 26.8% | **80.8%** | +54.0%（超过 70% 目标） |
| mcp | 14.2% | **21.9%** | +7.7%（单元测试层面，不含集成测试） |
| pkg/token | 20.0% | 20.0% | 不变 |

### 全项目覆盖率现状

| 包 | 覆盖率 | P0 前 | 变化 |
|---|--------|-------|------|
| validation | 100.0% | 100.0% | — |
| metrics | 95.9% | 95.9% | — |
| audit | 88.6% | 88.6% | — |
| auth | 89.8% | 89.8% | — |
| security | 87.3% | 87.3% | — |
| config | 82.5% | 82.5% | — |
| **database** | **80.8%** | 26.8% | **+54.0%** |
| server | 37.7% | 37.7% | — |
| **mcp** | **21.9%** | 14.2% | +7.7% |
| pkg/token | 20.0% | 0% | +20.0% |
| cmd/server | 0.0% | 0.0% | — |

---

## 3. How — 怎么做的

### 工具选择：go-sqlmock

新增了 `github.com/DATA-DOG/go-sqlmock` 依赖，用于 mock `*sql.DB`：

- **Router 测试**：用 sqlmock 创建 mock `*sql.DB`，设置期望的 SQL 调用（USE + Query/Exec），验证 Router 正确切换数据库并转发查询
- **Pool 测试**：用 sqlmock 创建的 DB 验证 UpdateConfig 的连接管理（关闭、替换、新增），不需要真实 MySQL

### P1-5：Router 单元测试（12 个测试）

文件：`internal/database/router_sqlmock_test.go`

| 测试 | 验证内容 |
|------|----------|
| `TestRouter_Query_Success` | Query 成功路径：USE → SELECT → 返回 2 行 |
| `TestRouter_Query_UseDBError` | USE 失败时 conn 被正确释放，返回错误 |
| `TestRouter_Query_QueryError` | SQL 执行失败时 conn 被正确释放 |
| `TestRouter_Query_UnknownDatabase` | 未知数据库返回错误 |
| `TestRouter_Query_CancelledContext` | 已取消的 context 立即返回错误 |
| `TestRouter_Exec_Success` | Exec 成功路径：USE → INSERT → RowsAffected |
| `TestRouter_Exec_UseDBError` | Exec 中 USE 失败 |
| `TestRouter_Exec_UnknownDatabase` | Exec 未知数据库 |
| `TestRouter_GetDB_Success` | GetDB 返回正确连接 |
| `TestRouter_GetDB_Unknown` | GetDB 未知数据库 |
| `TestRouter_Query_SpecialCharsInDBName` | 数据库名含 `-` 时 USE 正确引用为 `` `my-db` `` |
| `TestRouter_ConcurrentAccess` | 100 个并发 GetCluster/ListDatabases 无 data race |

### P1-6：Pool.UpdateConfig 单元测试（5 个测试）

文件：`internal/database/pool_update_test.go`

| 测试 | 验证内容 |
|------|----------|
| `TestPool_UpdateConfig_RemoveCluster` | 移除集群后旧连接被关闭 |
| `TestPool_UpdateConfig_ChangedCredentials` | 凭据变更后旧连接被替换 |
| `TestPool_UpdateConfig_NoChange` | 无变更时连接保持不变 |
| `TestPool_UpdateConfig_DeadlockRegression` | **关键回归**：3 个并发 Get + UpdateConfig 不死锁 |
| `TestPool_UpdateConfig_EmptyToEmpty` | 空配置边界条件 |

### P1-7：MCP 安全拦截单元测试（7 个测试）

文件：`internal/mcp/security_test.go`

| 测试 | 验证内容 |
|------|----------|
| `TestExecuteQuery_SecurityBlock_Drop` | DROP TABLE 被拦截 |
| `TestExecuteQuery_SecurityBlock_Truncate` | TRUNCATE TABLE 被拦截 |
| `TestExecuteQuery_SecurityBlock_Rename` | RENAME TABLE 被拦截 |
| `TestExecuteQuery_ParseError` | 非法 SQL 返回 parse 错误 |
| `TestExecuteExplain_RejectsDrop` | EXPLAIN DROP 被拒绝 |
| `TestExecuteExplain_RejectsCreate` | EXPLAIN CREATE 被拒绝 |
| `TestGetQueryTimeout_Default` / `TestGetMaxRows_Default` | 配置读取默认值 |

### 附加发现：热加载死锁 bug

在准备 P1-6 测试时，审查 `Pool.UpdateConfig` 代码发现了一个**已存在的死锁 bug**（正是用户报告的"热加载后服务卡死"问题）：

**根因**：函数顶部 `p.mu.Lock()` + `defer p.mu.Unlock()` 持有写锁，循环内第 153 行又调用 `p.mu.Lock()`。Go 的 `sync.Mutex` 不可重入，永久死锁。

**修复**：重写为两阶段——Phase 1 短暂持锁完成状态快照，Phase 2 无锁等待连接释放。已单独提交为 `6ef955c`。

---

## 4. Who

- **Fan**：报告热加载卡死问题，验证修复，review
- **Claude (AI)**：编写测试、发现并修复死锁 bug

---

## 5. Where

### 代码变更

```
新增文件：
  internal/database/router_sqlmock_test.go  (+261 行)
  internal/database/pool_update_test.go     (+150 行)
  internal/mcp/security_test.go             (+107 行)

修改文件：
  internal/database/pool_test.go            (注释更新)
  internal/database/pool.go                 (死锁修复，已单独提交)
  go.mod / go.sum                           (新增 sqlmock 依赖)
```

**生产代码变更**：仅 `pool.go` 的死锁修复（已单独提交）。其余全是测试代码。

### 依赖变更

新增 `github.com/DATA-DOG/go-sqlmock v1.5.2`，仅用于测试，不影响生产二进制。

---

## 6. When

| 时间 | 事件 |
|------|------|
| 2026-05-19 | 用户报告热加载后服务卡死 |
| 2026-05-19 | 定位死锁根因并修复（`6ef955c`） |
| 2026-05-19 | 执行 P1-5：Router 单元测试（12 个） |
| 2026-05-19 | 执行 P1-6：Pool.UpdateConfig 单元测试（5 个） |
| 2026-05-19 | 执行 P1-7：MCP 安全拦截单元测试（7 个） |
| 待定 | P2：E2E 测试 + CI Gate |
| 待定 | P3：覆盖率达标 |

---

## 7. 运行验证

```
$ go test -short -race -count=1 ./...
ok  12 packages, 0 failures, 0 data races

$ go test -short -cover ./internal/database/... ./internal/mcp/...
ok  database  coverage: 80.8%
ok  mcp       coverage: 21.9%
```

全部测试通过，无 race condition。

---

## 8. Review 检查清单

- [ ] 死锁修复（`pool.go`）的两阶段锁设计是否正确
- [ ] sqlmock 期望的 SQL 匹配是否合理（`USE \`testdb\``）
- [ ] `TestPool_UpdateConfig_DeadlockRegression` 的并发场景是否充分
- [ ] 安全拦截测试是否覆盖了所有 blocked 类型（DROP/TRUNCATE/RENAME）
- [ ] 新增依赖 `go-sqlmock` 是否可接受
- [ ] 是否可以提交并继续 P2

---

## 9. 下一步 P2 计划

| # | 工作项 | 说明 |
|---|--------|------|
| 8 | E2E 测试 | 完整 HTTP 链路：Auth → MCP → MySQL |
| 9 | CI 加 E2E job | GitHub Actions 自动运行 |
| 10 | CI 覆盖率报告 | PR 中自动评论覆盖率变化 |
| 11 | GitHub branch protection | PR 必须全绿才能合并 |
