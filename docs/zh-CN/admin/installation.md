# 生产安装

本指南涵盖 SafeMySQLMcpServer 的生产环境部署选项。

## Docker 部署
tab: Docker 部署

### 使用 Docker Compose（推荐）

```yaml
# docker-compose.yml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: appdb
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  safemysql:
    build: .
    ports:
      - "8080:8080"
    environment:
      JWT_SECRET: ${JWT_SECRET}
      DEV_DB_USER: ${DEV_DB_USER}
      DEV_DB_PASSWORD: ${DEV_DB_PASSWORD}
    depends_on:
      mysql:
        condition: service_healthy
    volumes:
      - ./config:/app/config:ro
      - audit_logs:/app/logs

volumes:
  mysql_data:
  audit_logs:
```

### 启动服务

```bash
# 创建 .env 文件
cat > .env << EOF
JWT_SECRET=your-secret-key-min-32-characters-long
MYSQL_ROOT_PASSWORD=your-mysql-root-password
DEV_DB_USER=appuser
DEV_DB_PASSWORD=appuser-password
EOF

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f safemysql
```

## Kubernetes 部署
tab: Kubernetes 部署

### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: safemysql
spec:
  replicas: 3
  selector:
    matchLabels:
      app: safemysql
  template:
    metadata:
      labels:
        app: safemysql
    spec:
      containers:
      - name: safemysql
        image: safemysql:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: safemysql-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: safemysql
spec:
  selector:
    app: safemysql
  ports:
  - port: 8080
    targetPort: 8080
```

### ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: safemysql-config
data:
  config.yaml: |
    server:
      port: 8080
    clusters:
      primary:
        host: mysql-service
        port: 3306
    databases:
      appdb:
        cluster: primary
```

## Systemd 服务
tab: Systemd 服务

### 服务文件

```ini
# /etc/systemd/system/safemysql.service
[Unit]
Description=SafeMySQLMcpServer
After=network.target mysql.service

[Service]
Type=simple
User=safemysql
Group=safemysql
WorkingDirectory=/opt/safemysql
ExecStart=/opt/safemysql/bin/safe-mysql-mcp -config /opt/safemysql/config/config.yaml
Restart=on-failure
RestartSec=5s

# 安全设置
NoNewPrivileges=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
```

### 启用服务

```bash
# 复制二进制文件
sudo cp bin/safe-mysql-mcp /opt/safemysql/bin/

# 复制配置
sudo cp -r config / /opt/safemysql/config/

# 启用并启动
sudo systemctl daemon-reload
sudo systemctl enable safemysql
sudo systemctl start safemysql

# 检查状态
sudo systemctl status safemysql
```

## 健康检查
tab: 健康检查

### HTTP 健康检查

```bash
# 基本健康检查
curl http://localhost:8080/health
# 预期: OK

# 详细健康检查（如实现）
curl http://localhost:8080/health/ready
```

### Kubernetes 探针

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

## 资源需求
tab: 资源需求

### 最低需求

| 资源 | 值 |
|----------|-------|
| CPU | 100m |
| Memory | 128Mi |
| Disk | 100MB |

### 生产环境推荐

| 资源 | 值 |
|----------|-------|
| CPU | 500m - 1000m |
| Memory | 256Mi - 512Mi |
| Disk | 1GB（用于审计日志） |

### 扩展考虑

| 场景 | 操作 |
|----------|--------|
| 高查询量 | 添加副本 |
| 大结果集 | 增加内存 |
| 多数据库 | 增加连接池 |
