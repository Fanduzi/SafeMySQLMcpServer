# 代码规范

本指南定义 SafeMySQLMcpServer 的编码标准。

## 通用原则

### 1. 简洁性
- 优先选择简单、可读的代码而非"聪明"的代码
- 避免过早的抽象
- 保持函数专注且短小（<50 行）

### 2. 一致性
- 遵循代码库中现有的模式
- 使用一致的命名约定
- 保持一致的错误处理

### 3. 文档
- 为公共 API 编写文档
- 解释"为什么"而非"是什么"
- 保持注释更新

## Go 约定
tab: Go 约定

### 命名

| 元素 | 约定 | 示例 |
|---------|------------|---------|
| Package | 小写 | `security`, `database` |
| 导出 | PascalCase | `NewValidator`, `ValidateToken` |
| 未导出 | camelCase | `validateInput`, `parseSQL` |
| 常量 | 大写下划线 或 PascalCase | `MaxRows`, `QUERY_TIMEOUT` |
| 接口 | -er 或 -or 后缀 | `Validator`, `Parser` |

### 错误处理

```go
// 好的做法: 包装错误并添加上下文
if err != nil {
    return fmt.Errorf("validate token: %w", err)
}

// 不好的做法: 丢失错误上下文
if err != nil {
    return errors.New("validation failed")
}
```

### Context 使用

```go
// 好的做法: 将 context 作为第一个参数传递
func (h *Handler) Query(ctx context.Context, db, sql string) (*Result, error) {
    // ...
}

// 不好的做法: 忽略 context
func (h *Handler) Query(db, sql string) (*Result, error) {
    // ...
}
```

## 文件组织
tab: 文件组织

### 包结构

```
internal/security/
├── checker.go        # 安全规则检查器
├── checker_test.go   # 检查器测试
├── parser.go         # SQL 解析器
├── parser_test.go    # 解析器测试
├── rewriter.go       # SQL 重写器
├── rewriter_test.go  # 重写器测试
├── doc.go            # 包文档
└── README.md         # 模块文档
```

### 文件头部注释

每个 Go 文件应以头部注释开始:

```go
// Package security 处理 SQL 解析和安全验证。
// input: SQL 语句、安全规则配置
// output: 解析后的 SQL、安全检查结果、重写后的 SQL
// pos: 安全部，位于输入验证和数据库执行之间
// note: 如果此文件发生变化，请同时更新此头部和 internal/security/README.md
package security
```

### README 结构

每个模块应有包含以下内容的 README:

```markdown
# 模块名称

模块的简要描述。

## 文件
| 文件 | 职责 |
|------|---------------|
| file.go | 功能说明 |

## 导出
- `Function()` - 描述

## 依赖
- Upstream: 它依赖的包
- Downstream: 依赖它的包

## 更新规则
如果 X 发生变化，请在同一变更中更新此文件。
```

## 测试标准
tab: 测试标准

### 测试文件命名

```
checker.go         -> checker_test.go
parser.go          -> parser_test.go
integration_test.go -> 集成测试后缀
```

### 测试函数命名

```go
func TestFunctionName(t *testing.T)                    // 基本测试
func TestFunctionName_Scenario(t *testing.T)          // 特定场景
func TestIntegration_FunctionName(t *testing.T)       // 集成测试
```

### 表驱动测试

```go
func TestValidateDatabaseName(t *testing.T) {
    tests := []struct {
        name string
        input string
        wantErr bool
    }{
        {"valid lowercase", "mydb", false},
        {"valid with underscore", "my_db", false},
        {"invalid with dash", "my-db", true},
        {"invalid with space", "my db", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validation.ValidateDatabaseName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
        })
    }
}
```

## 代码审查清单
tab: 代码审查清单

提交 PR 前，请验证:

- [ ] 代码编译无警告
- [ ] 测试在本地通过
- [ ] 新代码有测试
- [ ] 测试覆盖率保持或提升
- [ ] 文件头部已更新
- [ ] 模块 README 已更新
- [ ] 无硬编码的密钥
- [ ] 错误处理完整
- [ ] Context 正确传递
- [ ] 文档清晰准确
