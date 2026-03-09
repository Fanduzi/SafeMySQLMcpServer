# SafeMySQLMcpServer 文档

欢迎使用 SafeMySQLMcpServer 文档。

## 概述

SafeMySQLMcpServer 是一个安全的 MySQL MCP (Model Context Protocol) 服务器，让 Claude Code 等 AI 工具能够在开发和测试环境中安全地操作 MySQL 数据库。

## 核心特性

- **MCP 协议支持**: 完整的 MCP 协议实现，支持 HTTP 传输
- **JWT 认证**: 安全的 token 认证，支持可配置过期时间
- **SQL 安全层**: SQL 解析、验证和自动改写
- **审计日志**: 完整的操作审计，支持日志轮转
- **热配置重载**: 无需重启即可更新配置
- **Prometheus 指标**: 全面的可观测性支持

## 文档目录

| 章节 | 读者 | 说明 |
|------|------|------|
| [入门指南](./guide) | 所有用户 | 快速开始教程 |
| [核心概念](./concepts) | 开发者 | 架构和设计概念 |
| [开发指南](./dev) | 贡献者 | 开发指南和最佳实践 |
| [运维指南](./admin) | 运维人员 | 安装和配置管理 |
| [API 参考](./reference) | 所有用户 | API 和配置参考 |

## 快速链接

- [快速开始](./guide/getting-started.md)
- [配置参考](./reference/configuration.md)
- [MCP 工具参考](./reference/mcp-tools.md)
- [架构概览](./concepts/architecture.md)

## 版本

当前版本: **v1.0.0**

查看 [更新日志](../../CHANGELOG.md) 了解版本历史。
