# 部署指南

本指南涵盖在生产环境中部署 Anyi 应用程序的各种部署策略和最佳实践。

## 概述

Anyi 应用程序可以通过多种方式部署：

- **独立应用程序**：单一二进制部署
- **容器化部署**：Docker 和 Kubernetes
- **无服务器函数**：AWS Lambda、Google Cloud Functions
- **Web 服务**：HTTP API 服务器
- **微服务**：分布式服务架构

## 独立应用程序部署

### 生产构建

为生产环境创建优化的构建：

```bash
# 带优化的构建
go build -ldflags="-s -w" -o anyi-app ./cmd/main.go

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-linux ./cmd/main.go
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-windows.exe ./cmd/main.go
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-macos ./cmd/main.go
```

### 配置管理

使用特定环境的配置：

```go
// config/config.go
package config

import (
    "os"
    "github.com/jieliu2000/anyi"
)

type Config struct {
    Environment string
    LogLevel    string
    ConfigFile  string
    Port        string
}

func Load() (*Config, error) {
    config := &Config{
        Environment: getEnv("ENVIRONMENT", "production"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
        ConfigFile:  getEnv("CONFIG_FILE", "config.yaml"),
        Port:        getEnv("PORT", "8080"),
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// 加载 Anyi 配置
func LoadAnyiConfig(configFile string) error {
    return anyi.ConfigFromFile(configFile)
}
```

### Systemd 服务 (Linux)

创建 systemd 服务文件：

```ini
# /etc/systemd/system/anyi-app.service
[Unit]
Description=Anyi AI 应用程序
After=network.target

[Service]
Type=simple
User=anyi
Group=anyi
WorkingDirectory=/opt/anyi
ExecStart=/opt/anyi/anyi-app
Restart=always
RestartSec=10
Environment=ENVIRONMENT=production
Environment=CONFIG_FILE=/opt/anyi/config.yaml
EnvironmentFile=/opt/anyi/.env

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/anyi/logs

[Install]
WantedBy=multi-user.target
```

启用和启动服务：

```bash
# 复制二进制文件和配置
sudo mkdir -p /opt/anyi
sudo cp anyi-app /opt/anyi/
sudo cp config.yaml /opt/anyi/
sudo cp .env /opt/anyi/

# 创建用户
sudo useradd -r -s /bin/false anyi
sudo chown -R anyi:anyi /opt/anyi

# 安装和启动服务
sudo systemctl daemon-reload
sudo systemctl enable anyi-app
sudo systemctl start anyi-app

# 检查状态
sudo systemctl status anyi-app
```

## 容器部署

### Dockerfile

创建优化的 Dockerfile：

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o anyi-app ./cmd/main.go

# 最终阶段
FROM alpine:3.18

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 anyi && \
    adduser -D -s /bin/sh -u 1001 -G anyi anyi

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/anyi-app .
COPY --from=builder /app/config.yaml .

# 更改所有权
RUN chown -R anyi:anyi /app

# 切换到非 root 用户
USER anyi

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用程序
CMD ["./anyi-app"]
```

### Docker Compose

为本地开发创建 docker-compose.yml：

```yaml
version: "3.8"

services:
  anyi-app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=development
      - LOG_LEVEL=debug
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./logs:/app/logs
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=anyi
      - POSTGRES_USER=anyi
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  redis_data:
  postgres_data:
```

### 构建和运行

```bash
# 构建镜像
docker build -t anyi-app:latest .

# 使用 docker-compose 运行
docker-compose up -d

# 查看日志
docker-compose logs -f anyi-app
```

## Kubernetes 部署

### Kubernetes 清单

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: anyi-app
  labels:
    app: anyi-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: anyi-app
  template:
    metadata:
      labels:
        app: anyi-app
    spec:
      containers:
        - name: anyi-app
          image: anyi-app:latest
          ports:
            - containerPort: 8080
          env:
            - name: ENVIRONMENT
              value: "production"
            - name: LOG_LEVEL
              value: "info"
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: anyi-secrets
                  key: openai-api-key
            - name: ANTHROPIC_API_KEY
              valueFrom:
                secretKeyRef:
                  name: anyi-secrets
                  key: anthropic-api-key
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: anyi-app-service
spec:
  selector:
    app: anyi-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer

---
apiVersion: v1
kind: Secret
metadata:
  name: anyi-secrets
type: Opaque
data:
  openai-api-key: <base64-encoded-key>
  anthropic-api-key: <base64-encoded-key>
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: anyi-config
data:
  config.yaml: |
    clients:
      - name: "openai"
        type: "openai"
        config:
          apiKey: "$OPENAI_API_KEY"
          model: "gpt-4"
          temperature: 0.7
      
      - name: "anthropic"
        type: "anthropic"
        config:
          apiKey: "$ANTHROPIC_API_KEY"
          model: "claude-3-opus-20240229"

    flows:
      - name: "content_generation"
        clientName: "openai"
        steps:
          - name: "generate"
            executor:
              type: "llm"
              withconfig:
                template: "生成关于 {{.Text}} 的内容"
```

### 部署到 Kubernetes

```bash
# 应用配置
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml

# 检查部署状态
kubectl get deployments
kubectl get pods
kubectl get services

# 查看日志
kubectl logs -f deployment/anyi-app

# 扩展部署
kubectl scale deployment anyi-app --replicas=5
```

## 无服务器部署

### AWS Lambda

```go
// lambda/main.go
package main

import (
    "context"
    "encoding/json"
    "os"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

type Request struct {
    Message string `json:"message"`
}

type Response struct {
    StatusCode int               `json:"statusCode"`
    Headers    map[string]string `json:"headers"`
    Body       string            `json:"body"`
}

func init() {
    // 初始化 Anyi
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    anyi.NewClient("lambda", config)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    var req Request
    if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       `{"error": "无效的请求体"}`,
        }, nil
    }

    client, err := anyi.GetClient("lambda")
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       `{"error": "服务不可用"}`,
        }, nil
    }

    messages := []chat.Message{
        {Role: "user", Content: req.Message},
    }

    response, _, err := client.Chat(messages, nil)
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       `{"error": "AI 处理失败"}`,
        }, nil
    }

    responseBody, _ := json.Marshal(map[string]string{
        "response": response.Content,
    })

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
        Body: string(responseBody),
    }, nil
}

func main() {
    lambda.Start(handleRequest)
}
```

### SAM 模板

```yaml
# template.yaml
AWSTemplateFormatVersion: "2010-09-09"
Transform: AWS::Serverless-2016-10-31

Globals:
  Function:
    Timeout: 30
    MemorySize: 512

Resources:
  AnyiFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: lambda/
      Handler: main
      Runtime: go1.x
      Environment:
        Variables:
          OPENAI_API_KEY: !Ref OpenAIAPIKey
      Events:
        CatchAll:
          Type: Api
          Properties:
            Path: /chat
            Method: POST

Parameters:
  OpenAIAPIKey:
    Type: String
    NoEcho: true
    Description: OpenAI API 密钥

Outputs:
  AnyiAPI:
    Description: "API Gateway 端点"
    Value: !Sub "https://${ServerlessRestApi}.execute-api.${AWS::Region}.amazonaws.com/Prod/chat/"
```

### 部署到 AWS

```bash
# 构建和部署
sam build
sam deploy --guided

# 测试
curl -X POST https://your-api-gateway-url/Prod/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好，世界！"}'
```

## 监控和日志

### 应用程序监控

```go
// monitoring/metrics.go
package monitoring

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "anyi_requests_total",
            Help: "请求总数",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "anyi_request_duration_seconds",
            Help: "请求持续时间",
        },
        []string{"method", "endpoint"},
    )

    tokensUsed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "anyi_tokens_used_total",
            Help: "使用的令牌总数",
        },
        []string{"provider", "model"},
    )
)

func RecordRequest(method, endpoint, status string, duration time.Duration) {
    requestsTotal.WithLabelValues(method, endpoint, status).Inc()
    requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

func RecordTokenUsage(provider, model string, tokens int) {
    tokensUsed.WithLabelValues(provider, model).Add(float64(tokens))
}
```

### 结构化日志

```go
// logging/logger.go
package logging

import (
    "os"
    "github.com/sirupsen/logrus"
)

type Logger struct {
    *logrus.Logger
}

func NewLogger() *Logger {
    log := logrus.New()

    // 设置日志格式
    if os.Getenv("ENVIRONMENT") == "production" {
        log.SetFormatter(&logrus.JSONFormatter{})
    } else {
        log.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
        })
    }

    // 设置日志级别
    level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
    if err != nil {
        level = logrus.InfoLevel
    }
    log.SetLevel(level)

    return &Logger{log}
}

func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
    return l.Logger.WithFields(fields)
}

func (l *Logger) LogRequest(userID, endpoint, method string, duration time.Duration, statusCode int) {
    l.WithFields(map[string]interface{}{
        "user_id":     userID,
        "endpoint":    endpoint,
        "method":      method,
        "duration_ms": duration.Milliseconds(),
        "status_code": statusCode,
        "type":        "request",
    }).Info("API 请求")
}

func (l *Logger) LogAIInteraction(provider, model string, promptTokens, completionTokens int, duration time.Duration) {
    l.WithFields(map[string]interface{}{
        "provider":          provider,
        "model":             model,
        "prompt_tokens":     promptTokens,
        "completion_tokens": completionTokens,
        "total_tokens":      promptTokens + completionTokens,
        "duration_ms":       duration.Milliseconds(),
        "type":              "ai_interaction",
    }).Info("AI 交互")
}
```

### 健康检查

```go
// health/health.go
package health

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jieliu2000/anyi"
)

type HealthCheck struct {
    Status  string            `json:"status"`
    Version string            `json:"version"`
    Checks  map[string]string `json:"checks"`
}

func HealthHandler(version string) gin.HandlerFunc {
    return func(c *gin.Context) {
        checks := make(map[string]string)

        // 检查 Anyi 客户端
        if clients := anyi.ListClients(); len(clients) > 0 {
            checks["anyi_clients"] = "healthy"
        } else {
            checks["anyi_clients"] = "unhealthy"
        }

        // 检查数据库连接（如果适用）
        // checks["database"] = checkDatabase()

        // 检查外部 API
        checks["external_apis"] = checkExternalAPIs()

        status := "healthy"
        for _, check := range checks {
            if check != "healthy" {
                status = "unhealthy"
                break
            }
        }

        healthCheck := HealthCheck{
            Status:  status,
            Version: version,
            Checks:  checks,
        }

        if status == "healthy" {
            c.JSON(http.StatusOK, healthCheck)
        } else {
            c.JSON(http.StatusServiceUnavailable, healthCheck)
        }
    }
}

func checkExternalAPIs() string {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 测试 OpenAI API 连接
    client, err := anyi.GetClient("default")
    if err != nil {
        return "unhealthy"
    }

    // 发送简单的测试请求
    messages := []chat.Message{
        {Role: "user", Content: "test"},
    }

    _, _, err = client.Chat(messages, &chat.Options{MaxTokens: 1})
    if err != nil {
        return "unhealthy"
    }

    return "healthy"
}
```

## 性能优化

### 连接池配置

```go
// performance/pool.go
package performance

import (
    "net/http"
    "time"
)

func OptimizedHTTPClient() *http.Client {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}
```

### 缓存策略

```go
// cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisCache{client: rdb}
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    json, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return r.client.Set(ctx, key, json, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }

    return json.Unmarshal([]byte(val), dest)
}
```

## 安全配置

### TLS 配置

```go
// security/tls.go
package security

import (
    "crypto/tls"
    "net/http"
    "time"
)

func SecureTLSConfig() *tls.Config {
    return &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }
}

func SecureHTTPServer(handler http.Handler) *http.Server {
    return &http.Server{
        Handler:      handler,
        Addr:         ":8443",
        TLSConfig:    SecureTLSConfig(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
}
```

## 备份和恢复

### 数据备份策略

```bash
#!/bin/bash
# backup.sh

# 设置变量
BACKUP_DIR="/opt/anyi/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="anyi_backup_${DATE}.tar.gz"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份配置文件
tar -czf $BACKUP_DIR/$BACKUP_FILE \
    /opt/anyi/config.yaml \
    /opt/anyi/.env \
    /opt/anyi/logs

# 清理旧备份（保留 7 天）
find $BACKUP_DIR -name "anyi_backup_*.tar.gz" -mtime +7 -delete

echo "备份完成：$BACKUP_DIR/$BACKUP_FILE"
```

### 灾难恢复

```bash
#!/bin/bash
# restore.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "用法：$0 <备份文件>"
    exit 1
fi

# 停止服务
sudo systemctl stop anyi-app

# 备份当前配置
sudo cp -r /opt/anyi /opt/anyi.backup.$(date +%Y%m%d_%H%M%S)

# 恢复备份
sudo tar -xzf $BACKUP_FILE -C /

# 重启服务
sudo systemctl start anyi-app

echo "恢复完成"
```

## 最佳实践

### 1. 部署检查清单

- [ ] 生产构建优化
- [ ] 环境变量配置
- [ ] 安全设置验证
- [ ] 监控和日志配置
- [ ] 健康检查实现
- [ ] 备份策略制定

### 2. 性能优化

- 使用连接池
- 实施缓存策略
- 配置适当的超时
- 监控资源使用

### 3. 安全考虑

- 使用 HTTPS/TLS
- 实施访问控制
- 定期更新依赖
- 监控安全事件

### 4. 运维管理

- 自动化部署流程
- 实施滚动更新
- 配置监控告警
- 制定故障恢复计划

通过遵循这些部署指南和最佳实践，您可以确保 Anyi 应用程序在生产环境中稳定、安全、高效地运行。
