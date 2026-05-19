# Quality Assurance P0 执行报告

> 日期：2026-05-19
> 作者：Fan + Claude
> 状态：P0 已完成，待 review

---

## 1. Why — 为什么要做这件事

### 背景

SafeMySQLMcpServer 是一个 MCP（Model Context Protocol）服务器，让 AI 工具（如 Claude Code）能安全地操作 MySQL 数据库。项目有 12 个 Go 包，单元测试覆盖率如下：

| 包 | 覆盖率 | 角色 |
|---|--------|------|
| validation | 100% | 输入校验 |
| metrics | 95.9% | Prometheus 指标 |
| audit | 88.6% | 审计日志 |
| auth | 89.8% | JWT 认证 |
| security | 87.3% | SQL 安全检查 |
| config | 82.5% | 配置管理 |
| server | 37.7% | HTTP 服务 |
| database | 26.8% | 数据库连接和路由 |
| mcp | 14.2% | MCP 工具（核心业务） |
| pkg/token | 0% → 20% | Token 生成 CLI |
| cmd/server | 0% | 入口 |

**核心矛盾**：高覆盖率的包都是辅助层（出问题不致命），而核心业务路径（mcp、database）覆盖率极低。

### 直接触发事件

近期在真实使用中发现了多个面向用户的 bug：

1. **`list_tables` 返回错误的表** — `DATABASE()` 函数返回 NULL，因为没有执行 `USE database` 切换数据库，导致查到的是其他库的表
2. **热加载配置后连接池不重建** — 修改配置后旧连接被删除但没有创建新连接
3. **`token --expire 365d` 报错** — Go 标准库的 `time.ParseDuration` 不支持 `d`（天）后缀

这些 bug 全部在核心路径上，而核心路径恰好没有测试。

---

## 2. What — 做了什么

### 产出物

| # | 文件 | 类型 | 说明 |
|---|------|------|------|
| 1 | `docs/plans/quality-assurance-design.md` | 设计文档 | 完整的质量保障方案，包含架构、优先级、ROI 排序 |
| 2 | `pkg/token/main.go` | Bug 修复 | 新增 `parseDuration` 函数支持 `d`（天）后缀 |
| 3 | `pkg/token/main_test.go` | 单元测试 | 15 个测试用例覆盖 parseDuration（正负天数、小数、非法输入） |
| 4 | `internal/mcp/integration_test.go` | 集成测试 | 11 个测试覆盖全部 7 个 MCP 工具 + 安全拦截 + DATABASE() 回归 |
| 5 | `internal/database/integration_test.go` | 集成测试 | 新增数据库切换回归测试（验证 `USE database` 正确执行） |

### 测试清单（12 个测试）

**MCP 工具集成测试**（`internal/mcp/integration_test.go`，需要真实 MySQL）：

| 测试名 | 验证内容 |
|--------|----------|
| `TestIntegrationListTables` | `executeListTables` 返回指定数据库的表 |
| `TestIntegrationListTables_DatabaseRegression` | **回归测试**：验证 DATABASE() 不再返回 NULL |
| `TestIntegrationDescribeTable` | `executeDescribeTable` 返回正确的列信息 |
| `TestIntegrationQuery_Select` | SELECT 查询返回正确的数据 |
| `TestIntegrationQuery_InsertUpdateDelete` | INSERT/UPDATE/DELETE 完整生命周期 |
| `TestIntegrationQuery_SecurityBlock` | DROP 语句被安全拦截 |
| `TestIntegrationShowCreateTable` | `executeShowCreateTable` 返回 CREATE 语句 |
| `TestIntegrationExplain` | `executeExplain` 返回执行计划 |
| `TestIntegrationExplain_RejectsDrop` | EXPLAIN DROP 被拒绝 |
| `TestIntegrationSearchTables` | `executeSearchTables` 跨库搜索表名 |
| `TestIntegrationListDatabases` | `executeListDatabases` 返回配置的数据库列表 |

**数据库路由测试**（`internal/database/integration_test.go`）：

| 测试名 | 验证内容 |
|--------|----------|
| `TestIntegrationDatabaseSwitching` | **回归测试**：`SELECT DATABASE()` 返回正确的库名 |

**parseDuration 单元测试**（`pkg/token/main_test.go`）：

| 测试名 | 验证内容 |
|--------|----------|
| `TestParseDuration` | 15 个用例：`1d`/`7d`/`365d`/`0d`/`1h`/`30m`/`90s`/`500ms`/`1h30m`/`24h`/`-1d`/`1.5d`/非法输入 |

### Bug 修复

在写 `parseDuration` 测试时，发现并修复了两个额外 bug：
- `-1d`（负天数）解析失败 — 原实现 `fmt.Sscanf` 不支持负数
- `1.5d`（小数天数）解析失败 — 原实现不支持浮点数
- 修复方式：将 `fmt.Sscanf` 替换为 `strconv.ParseFloat`

---

## 3. How — 怎么做的

### 设计方法

先写设计文档（`docs/plans/quality-assurance-design.md`），按 ROI（投入产出比）排序优先级，而不是按包线性推进：

```
P0：快速胜利 + 核心回归（本次完成）
  ├── parseDuration 单元测试
  └── MCP 工具集成测试 + Router 切换测试

P1：单元测试补强（下次执行）
  ├── database Router 单元测试（sqlmock）
  ├── Pool.UpdateConfig 回归测试
  └── MCP handler 校验逻辑测试

P2：E2E + CI Gate
P3：覆盖率达标
```

### 测试策略

采用三层测试金字塔：

```
        ┌──────────┐
        │   E2E    │  ← P2，完整 HTTP 链路
        ├──────────┤
        │Integration│  ← P0 本次完成，真实 MySQL
        ├──────────┤
        │  Unit    │  ← P0 完成 parseDuration，P1 继续
        └──────────┘
```

**集成测试**选择直接调用内部函数（`executeQuery`、`executeListTables` 等），而非走 HTTP，原因：
- 一个测试覆盖整条链路（MCP handler → Security → Router → MySQL），ROI 最高
- 不需要启动完整服务器，运行更快
- 能直接访问函数返回的强类型结构体，断言更精确

**构建标签**：使用 `//go:build integration` 标签隔离，日常 `go test ./...` 不会触发集成测试，CI 中单独运行。

---

## 4. Who — 谁做的

- **Fan**：项目 owner，发现所有 bug，提供需求和技术方向，review 产出
- **Claude (AI)**：编写设计文档、实现代码、编写测试，执行测试并修复问题

---

## 5. Where — 影响范围

### 代码变更

| 文件 | 变更类型 | 行数 |
|------|----------|------|
| `internal/mcp/integration_test.go` | 新增 | +417 行 |
| `internal/database/integration_test.go` | 修改（新增测试） | +63 行 |
| `pkg/token/main.go` | 修改（bug 修复） | +7/-3 行 |
| `pkg/token/main_test.go` | 新增 | +44 行 |
| `docs/plans/quality-assurance-design.md` | 新增 | +161 行 |
| `tests/integration/mcp_tools_test.go` | 新增（占位） | +6 行 |

**没有修改任何生产代码的签名或行为**（除了 `parseDuration` 的 bug 修复）。所有新增代码都是测试和文档。

### 运行验证

```
# 单元测试（全部通过）
$ go test -short ./...
ok  12 packages, 0 failures

# 集成测试（全部通过，需要 MySQL）
$ MYSQL_HOST=127.0.0.1 ... go test -tags integration -v ./internal/mcp/... -run TestIntegration
=== RUN   TestIntegrationListTables              --- PASS
=== RUN   TestIntegrationListTables_Database...   --- PASS
=== RUN   TestIntegrationDescribeTable            --- PASS
=== RUN   TestIntegrationQuery_Select             --- PASS
=== RUN   TestIntegrationQuery_InsertUpdate...    --- PASS
=== RUN   TestIntegrationQuery_SecurityBlock      --- PASS
=== RUN   TestIntegrationShowCreateTable          --- PASS
=== RUN   TestIntegrationExplain                  --- PASS
=== RUN   TestIntegrationExplain_RejectsDrop      --- PASS
=== RUN   TestIntegrationSearchTables             --- PASS
=== RUN   TestIntegrationListDatabases            --- PASS
PASS

$ go test -tags integration -v ./internal/database/... -run TestIntegrationDatabaseSwitching
=== RUN   TestIntegrationDatabaseSwitching        --- PASS
PASS
```

---

## 6. When — 时间线

| 时间 | 事件 |
|------|------|
| 近期 | 用户在使用中发现 DATABASE() bug、热加载 bug、token 365d bug |
| 2026-05-19 | 讨论并确认需要建立质量保障体系 |
| 2026-05-19 | 编写 QA 设计文档，确定 P0-P3 分期计划 |
| 2026-05-19 | 执行 P0：parseDuration 测试 → 发现并修复 2 个额外 bug |
| 2026-05-19 | 执行 P0：MCP 工具集成测试（11 个） + Router 切换测试（1 个） |
| 待定 | P1：sqlmock 单元测试补强 |
| 待定 | P2：E2E 测试 + CI Gate |
| 待定 | P3：覆盖率达标（mcp ≥ 70%, database ≥ 70%） |

---

## 7. 后续计划

P0 完成后的下一步（P1）：

1. **`internal/database` Router.Query/Exec 单元测试**（sqlmock） — 不需要 MySQL，覆盖边界条件
2. **`internal/database` Pool.UpdateConfig 单元测试** — 回归热加载 bug
3. **`internal/mcp` handler 校验逻辑单元测试** — 覆盖输入校验、错误路径

预期 P1 完成后覆盖率变化：
- mcp: 14.2% → 50%+
- database: 26.8% → 50%+

---

## 8. Review 检查清单

请 reviewer 关注以下几点：

- [ ] 设计文档的优先级排序是否合理
- [ ] 集成测试是否覆盖了所有关键场景
- [ ] `parseDuration` 的 bug 修复是否正确（特别是 `-1d`、`1.5d` 的处理）
- [ ] 测试代码是否遵循了项目现有的测试风格
- [ ] 是否有遗漏的回归测试场景
- [ ] 是否可以提交并继续 P1
