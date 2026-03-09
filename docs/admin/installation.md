# Installation

This guide covers production deployment options for SafeMySQLMcpServer.

## Docker Deployment
tab: Docker Deployment

### Using Docker Compose (Recommended)

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

### Start Services

```bash
# Create .env file
cat > .env << EOF
JWT_SECRET=your-secret-key-min-32-characters-long
MYSQL_ROOT_PASSWORD=your-mysql-root-password
DEV_DB_USER=appuser
DEV_DB_PASSWORD=appuser-password
EOF

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f safemysql
```

## Kubernetes Deployment
tab: Kubernetes Deployment

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

## Systemd Service
tab: Systemd Service

### Service File

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

# Security
NoNewPrivileges=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
```

### Enable Service

```bash
# Copy binary
sudo cp bin/safe-mysql-mcp /opt/safemysql/bin/

# Copy config
sudo cp -r config / /opt/safemysql/config/

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable safemysql
sudo systemctl start safemysql

# Check status
sudo systemctl status safemysql
```

## Health Checks
tab: Health Checks

### HTTP Health Check

```bash
# Basic health check
curl http://localhost:8080/health
# Expected: OK

# Detailed health check (if implemented)
curl http://localhost:8080/health/ready
```

### Kubernetes Probes

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

## Resource Requirements
tab: Resource Requirements

### Minimum Requirements

| Resource | Value |
|----------|-------|
| CPU | 100m |
| Memory | 128Mi |
| Disk | 100MB |

### Recommended for Production

| Resource | Value |
|----------|-------|
| CPU | 500m - 1000m |
| Memory | 256Mi - 512Mi |
| Disk | 1GB (for audit logs) |

### Scaling Considerations

| Scenario | Action |
|----------|--------|
| High query volume | Add replicas |
| Large result sets | Increase memory |
| Many databases | Increase connection pool |
