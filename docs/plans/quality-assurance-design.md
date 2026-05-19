# Quality Assurance Design Document

> Status: APPROVED
> Date: 2026-05-19
> Author: Claude + Fan

## 1. Problem Statement

当前项目测试分布严重不均：

- **高覆盖包**（validation 100%, metrics 95%, audit 88%, auth 89%, security 87%）— 辅助层，出问题不致命
- **低覆盖包**（mcp 14%, database 26%, server 37%）— 核心业务路径，出问题直接影响用户
- **零覆盖**（cmd/server 0%, pkg/token 0%）— 入口和工具

结果：每个面向用户的 bug（DATABASE() 没切库、热加载删连接不重建、token 365d 报错）都在核心路径上，而核心路径恰好没有测试。

## 2. Goals

| 目标 | 指标 |
|------|------|
| 核心包覆盖 | mcp ≥ 70%, database ≥ 70%, server ≥ 60% |
| E2E 覆盖 | 所有 MCP 工具至少 1 个端到端测试 |
| CI gate | PR 合并前必须全绿（lint + unit + integration） |
| 回归防护 | 每次发版前跑 E2E，Docker 镜像 health check |

## 3. Decisions

| 决策 | 选择 | 原因 |
|------|------|------|
| Mock 策略 | **sqlmock** | 不改生产代码，只加测试代码，风险最小。等痛点明显了再重构接口 |
| 执行顺序 | **按 ROI 排优先级**，不按 Phase 线性推进 | P0 先抓已知 bug 的回归测试 |
| 覆盖率门槛 | **渐进式** | 初期只报告不拦截，核心包 ≥ 50% 后加 CI 门槛 |

## 4. Test Architecture

```
        ┌──────────┐
        │   E2E    │  ← HTTP → Auth → MCP → Real MySQL
        │  (few)   │  ← 验证完整链路
        ├──────────┤
        │Integration│  ← MCP handler + 真实 MySQL
        │ (some)    │  ← 验证 execute* 函数实际查询结果
        ├──────────┤
        │  Unit    │  ← sqlmock mock *sql.DB
        │ (many)   │  ← 验证边界条件、错误处理、纯函数
        └──────────┘
```

### 4.1 Mock Strategy: sqlmock

- 用 `github.com/DATA-DOG/go-sqlmock` mock `*sql.DB`
- 不修改任何生产代码的签名
- MCP handler 测试：mock `*sql.DB` 的 QueryContext 返回预设行
- Router 测试：mock `*sql.DB` 的 Conn + ExecContext(USE) + QueryContext
- Pool 测试：mock `sql.Open` 或直接注入 mock DB

### 4.2 Integration Tests: 真实 MySQL

- 复用 CI 已有的 MySQL service container
- 新建 `tests/integration/` 目录，build tag `//go:build integration`
- 直接调用 MCP handler 的 execute* 函数（不经过 HTTP），连接真实 MySQL
- 这是最快覆盖 mcp 14% → 50%+ 的方式

### 4.3 E2E Tests: 完整链路

- 新建 `tests/e2e/`
- 启动真实 HTTP server + MySQL
- 通过 HTTP 调用 MCP 工具
- 覆盖认证、限流、安全拦截

## 5. Execution Plan（按 ROI 排序）

### P0：快速胜利 + 核心回归（本次执行）

| # | 工作项 | 预期收益 | 预估 |
|---|--------|----------|------|
| 1 | `pkg/token` — parseDuration 单元测试 | token 0% → 60%+ | 15min |
| 2 | 安装 sqlmock 依赖 | 为后续测试铺路 | 5min |
| 3 | MCP 集成测试 — 新建 `tests/integration/mcp_tools_test.go` | mcp 14% → 50%+ | 2h |
|   | - executeListTables（回归 DATABASE() bug） | | |
|   | - executeDescribeTable | | |
|   | - executeQuery | | |
|   | - executeShowCreateTable | | |
|   | - executeExplain | | |
|   | - executeSearchTables | | |
|   | - 安全拦截（DROP 被 block） | | |
| 4 | Router 集成测试 — 数据库切换 | 回归 USE db bug | 30min |

### P1：单元测试补强（下次执行）

| # | 工作项 | 预期收益 | 预估 |
|---|--------|----------|------|
| 5 | `internal/database` — Router.Query/Exec 单元测试（sqlmock） | database 26% → 50%+ | 1h |
| 6 | `internal/database` — Pool.UpdateConfig 单元测试 | 回归热加载 bug | 1h |
| 7 | `internal/mcp` — handler 校验逻辑单元测试（sqlmock） | mcp 边界条件 | 1h |

### P2：E2E + CI Gate（稳定后执行）

| # | 工作项 | 预期收益 | 预估 |
|---|--------|----------|------|
| 8 | 新建 `tests/e2e/e2e_test.go` | 完整链路覆盖 | 3h |
| 9 | CI 加 E2E job | 自动化 | 30min |
| 10 | CI 覆盖率报告 | 可观测 | 30min |
| 11 | GitHub branch protection | 强制 gate | 15min |

### P3：覆盖率达标（持续改进）

| # | 工作项 |
|---|--------|
| 12 | mcp → 70%, database → 70%, server → 60% |
| 13 | CI 覆盖率门槛（核心包 ≥ 50%） |
| 14 | Docker build + health check in CI |

## 6. CI Gate Design

### 当前 CI 流程

```
push → Build (Go 1.22/1.23/1.24) → Unit Test → Lint → Security Scan
     → Integration Test (MySQL service container, 只测 database 层)
```

### 改进后

```
PR / Push
  ├─ Job 1: Lint (golangci-lint)
  ├─ Job 2: Security (gosec)
  ├─ Job 3: Unit Test (go test -short -race -cover)
  ├─ Job 4: Integration Test (MySQL container)
  │     - database 层（已有）
  │     - MCP 工具集成测试（P0 新增）
  │     - 覆盖率报告（P2 新增）
  └─ Job 5: E2E Test (MySQL + full server)  [P2]
```

### Coverage Gate（P3 启用）

初期只报告不拦截。核心包 ≥ 50% 后在 CI 中加最低门槛。

## 7. Risk & Trade-offs

| 决策 | 权衡 |
|------|------|
| sqlmock 不改生产代码 | sqlmock mock 规则较复杂（特别是 Conn + USE 链路），但风险可控 |
| 集成测试优先于单元测试 | 集成测试 ROI 更高（一个测试覆盖整条链路），但 CI 时间稍长 |
| 渐进式覆盖率门槛 | 避免阻塞开发，但可能长期不达标 |
| E2E 放 P2 | 先用集成测试快速补覆盖，E2E 等集成测试稳定后再做 |

## 8. Success Criteria

- [ ] token 覆盖率 ≥ 60%
- [ ] mcp 覆盖率 ≥ 70%
- [ ] database 覆盖率 ≥ 70%
- [ ] server 覆盖率 ≥ 60%
- [ ] 所有 MCP 工具有集成测试（P0 完成）
- [ ] Router.Query USE db 逻辑有回归测试
- [ ] Pool.UpdateConfig 有回归测试
- [ ] 所有 MCP 工具有 E2E 测试（P2 完成）
- [ ] CI 全绿才能合并 PR（P3 完成）
- [ ] Docker build + health check 在 CI 中验证（P3 完成）
