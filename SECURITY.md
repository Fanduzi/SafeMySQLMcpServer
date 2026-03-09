# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of SafeMySQLMcpServer seriously. If you have discovered a security vulnerability, we appreciate your help in disclosing it to us in a responsible manner.

### Please DO NOT

- Open a public GitHub issue
- Discuss the vulnerability in public forums
- Exploit the vulnerability or problem you have discovered
- Reveal the problem to others before it has been resolved

### Please DO

**Report security vulnerabilities to: security@safemysql-mcp.example.com**

Include the following information:

1. **Description** of the vulnerability
2. **Steps to reproduce** the issue
3. **Potential impact** of the vulnerability
4. **Possible solutions** (if you have any)
5. **Your name/handle** (for credit in our security advisories, optional)

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours.

2. **Assessment**: We will assess the vulnerability and determine its severity within 7 days.

3. **Fix Development**: We will develop and test a fix.

4. **Disclosure**: We will coordinate with you on the disclosure timeline.

5. **Credit**: We will credit you in our security advisory (unless you prefer to remain anonymous).

## Security Best Practices

When using SafeMySQLMcpServer, please follow these security best practices:

### Authentication

- **Always use JWT authentication** in production environments
- Use strong, unique JWT secrets (minimum 32 characters)
- Rotate JWT secrets periodically
- Set appropriate token expiration times

### Database Security

- Use database users with **minimal required privileges**
- Never use `root` or admin users for the application
- Configure allowed DML/DDL operations carefully
- Enable audit logging for compliance

### Network Security

- Run behind a reverse proxy (nginx, traefik) in production
- Use TLS/HTTPS for all connections
- Implement rate limiting at the network level
- Restrict database network access

### Configuration

- Never commit secrets to version control
- Use environment variables for sensitive configuration
- Review `security.yaml` rules carefully
- Enable audit logging

### Example Secure Configuration

```yaml
# security.yaml
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
  allowed_ddl: []  # Disable DDL in production
  blocked:
    - LOAD_FILE
    - INTO OUTFILE
    - INTO DUMPFILE
  query_timeout: 30s
  max_rows: 10000
```

## Known Security Considerations

### SQL Injection Prevention

SafeMySQLMcpServer includes multiple layers of SQL injection protection:

1. **Identifier Validation**: Database and table names are validated against a strict regex pattern
2. **SQL Parsing**: All SQL statements are parsed and analyzed
3. **Security Rules**: Configurable allowlist/blocklist for SQL operations
4. **Query Rewriting**: Automatic modification of dangerous queries

However, **no protection is perfect**. Always:

- Review generated SQL before execution in critical environments
- Use parameterized queries when possible
- Keep the security rules updated

### Audit Logging

Audit logs may contain sensitive SQL statements. Ensure:

- Audit log files have restricted permissions
- Logs are rotated and archived securely
- Log retention complies with your data protection requirements

## Security Updates

Security updates will be released as:

- Patch versions (e.g., 1.0.1 → 1.0.2) for security fixes
- Documented in GitHub Security Advisories
- Announced in release notes

## Contact

For general security questions, reach out to: security@safemysql-mcp.example.com

For non-security issues, please use [GitHub Issues](https://github.com/your-org/SafeMySQLMcpServer/issues).

---

Thank you for helping keep SafeMySQLMcpServer and its users safe! 🛡️
