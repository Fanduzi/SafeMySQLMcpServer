# Contributing to SafeMySQLMcpServer

Thank you for your interest in contributing to SafeMySQLMcpServer! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the maintainers.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- MySQL 8.0 (for integration testing)
- Docker and Docker Compose (optional, for containerized development)
- Make (optional, for using Makefile commands)

### Development Setup

1. **Fork and Clone**

   ```bash
   git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
   cd SafeMySQLMcpServer
   ```

2. **Install Dependencies**

   ```bash
   go mod download
   ```

3. **Setup Local Environment**

   ```bash
   # Copy example configs
   cp config/config.yaml.example config/config.yaml
   cp config/security.yaml.example config/security.yaml

   # Set environment variables
   export JWT_SECRET=your-development-secret-min-32-characters
   export DEV_DB_USER=root
   export DEV_DB_PASSWORD=password
   ```

4. **Start MySQL (Docker)**

   ```bash
   docker-compose up -d mysql
   ```

5. **Run Tests**

   ```bash
   # Unit tests
   go test ./... -short -race

   # With coverage
   go test ./... -cover -short

   # Integration tests (requires MySQL)
   go test ./... -run Integration
   ```

6. **Build**

   ```bash
   make build
   # or
   go build -o bin/safe-mysql-mcp ./cmd/server
   ```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed and expected**
- **Include logs, screenshots, or screen recordings if helpful**
- **Include your environment details** (OS, Go version, MySQL version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the steps**
- **Describe the current behavior and explain the expected behavior**
- **Explain why this enhancement would be useful**

### Pull Requests

- Fill in the required template
- Do not include issue numbers in the PR title
- Include screenshots and animated GIFs in your pull request whenever possible
- Follow the coding standards
- Include tests for new functionality
- Update documentation for changed functionality
- End all files with a newline

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Run `go fmt` before committing
- Run `go vet` and fix any issues
- Use meaningful variable and function names
- Write comments for exported functions and types

### Testing

- Write unit tests for new functionality
- Maintain test coverage above 80% for new code
- Use table-driven tests for multiple test cases
- Include both positive and negative test cases
- Mock external dependencies (database, network)

### Example Test Structure

```go
func TestValidateDatabaseName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "mydb", false},
        {"empty name", "", true},
        {"invalid chars", "my-db", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateDatabaseName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateDatabaseName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
        })
    }
}
```

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools
- `ci`: Changes to CI configuration

### Examples

```
feat(auth): add JWT refresh token support
fix(security): prevent SQL injection in table names
docs(readme): add installation instructions
test(database): add integration tests for connection pool
```

## Pull Request Process

1. **Create a Branch**

   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make Your Changes**

   - Write clean, well-documented code
   - Add/update tests
   - Update documentation

3. **Run Checks**

   ```bash
   # Format code
   go fmt ./...

   # Run linter
   go vet ./...

   # Run tests
   go test ./... -race -short

   # Check coverage
   go test ./... -cover
   ```

4. **Commit Your Changes**

   ```bash
   git add .
   git commit -m "feat: your feature description"
   ```

5. **Push and Create PR**

   ```bash
   git push origin feature/your-feature-name
   ```

   Then create a Pull Request on GitHub.

6. **Code Review**

   - Respond to all review comments
   - Make requested changes in new commits
   - Mark conversations as resolved

7. **Merge Requirements**

   - All CI checks must pass
   - At least one approval from a maintainer
   - No merge conflicts
   - Coverage must not decrease significantly

## Questions?

Feel free to open an issue with the "question" label or reach out to the maintainers.

Thank you for contributing! 🎉
