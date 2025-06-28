# Deployment Guide

This guide covers various deployment strategies and best practices for deploying Anyi applications in production environments.

## Overview

Anyi applications can be deployed in various ways:

- **Standalone Applications**: Single binary deployments
- **Containerized Deployments**: Docker and Kubernetes
- **Serverless Functions**: AWS Lambda, Google Cloud Functions
- **Web Services**: HTTP API servers
- **Microservices**: Distributed service architectures

## Standalone Application Deployment

### Building for Production

Create optimized builds for production:

```bash
# Build with optimizations
go build -ldflags="-s -w" -o anyi-app ./cmd/main.go

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-linux ./cmd/main.go
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-windows.exe ./cmd/main.go
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o anyi-app-macos ./cmd/main.go
```

### Configuration Management

Use environment-specific configuration:

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

// Load Anyi configuration
func LoadAnyiConfig(configFile string) error {
    return anyi.ConfigFromFile(configFile)
}
```

### Systemd Service (Linux)

Create a systemd service file:

```ini
# /etc/systemd/system/anyi-app.service
[Unit]
Description=Anyi AI Application
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

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/anyi/logs

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Copy binary and configuration
sudo mkdir -p /opt/anyi
sudo cp anyi-app /opt/anyi/
sudo cp config.yaml /opt/anyi/
sudo cp .env /opt/anyi/

# Create user
sudo useradd -r -s /bin/false anyi
sudo chown -R anyi:anyi /opt/anyi

# Install and start service
sudo systemctl daemon-reload
sudo systemctl enable anyi-app
sudo systemctl start anyi-app

# Check status
sudo systemctl status anyi-app
```

## Container Deployment

### Dockerfile

Create an optimized Dockerfile:

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o anyi-app ./cmd/main.go

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 anyi && \
    adduser -D -s /bin/sh -u 1001 -G anyi anyi

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/anyi-app .
COPY --from=builder /app/config.yaml .

# Change ownership
RUN chown -R anyi:anyi /app

# Switch to non-root user
USER anyi

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./anyi-app"]
```

### Docker Compose

Create a docker-compose.yml for local development:

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

### Building and Running

```bash
# Build the image
docker build -t anyi-app:latest .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f anyi-app

# Scale the application
docker-compose up -d --scale anyi-app=3
```

## Kubernetes Deployment

### Kubernetes Manifests

Create Kubernetes deployment files:

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
              memory: "128Mi"
              cpu: "100m"
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
          volumeMounts:
            - name: config
              mountPath: /app/config.yaml
              subPath: config.yaml
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: anyi-config

---
# service.yaml
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
  type: ClusterIP

---
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

    flows:
      - name: "text_processor"
        clientName: "openai"
        steps:
          - name: "analyze"
            executor:
              type: "llm"
              withconfig:
                template: "Analyze: {{.Text}}"

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: anyi-secrets
type: Opaque
data:
  openai-api-key: <base64-encoded-api-key>
  anthropic-api-key: <base64-encoded-api-key>

---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: anyi-app-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - api.example.com
      secretName: anyi-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: anyi-app-service
                port:
                  number: 80
```

### Helm Chart

Create a Helm chart for easier deployment:

```yaml
# Chart.yaml
apiVersion: v2
name: anyi-app
description: Anyi AI Application Helm Chart
version: 0.1.0
appVersion: "1.0.0"

---
# values.yaml
replicaCount: 3

image:
  repository: anyi-app
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: anyi-tls
      hosts:
        - api.example.com

resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

config:
  environment: production
  logLevel: info

secrets:
  openaiApiKey: ""
  anthropicApiKey: ""
```

Deploy with Helm:

```bash
# Install the chart
helm install anyi-app ./helm/anyi-app \
  --set secrets.openaiApiKey="your-openai-key" \
  --set secrets.anthropicApiKey="your-anthropic-key"

# Upgrade the deployment
helm upgrade anyi-app ./helm/anyi-app \
  --set image.tag="v1.1.0"

# Uninstall
helm uninstall anyi-app
```

## Serverless Deployment

### AWS Lambda

Create a Lambda-compatible handler:

```go
// lambda/main.go
package main

import (
    "context"
    "encoding/json"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/jieliu2000/anyi"
)

type Request struct {
    Text string `json:"text"`
}

type Response struct {
    Result string `json:"result"`
    Error  string `json:"error,omitempty"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Parse request
    var req Request
    if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body:       `{"error": "Invalid request format"}`,
        }, nil
    }

    // Initialize Anyi (consider caching this)
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       `{"error": "Configuration failed"}`,
        }, nil
    }

    // Get flow
    flow, err := anyi.GetFlow("text_processor")
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       `{"error": "Flow not found"}`,
        }, nil
    }

    // Execute flow
    result, err := flow.RunWithInput(req.Text)
    if err != nil {
        response := Response{Error: err.Error()}
        body, _ := json.Marshal(response)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       string(body),
        }, nil
    }

    // Return response
    response := Response{Result: result.Text}
    body, _ := json.Marshal(response)

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
        Body: string(body),
    }, nil
}

func main() {
    lambda.Start(handler)
}
```

Build for Lambda:

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bootstrap lambda/main.go

# Create deployment package
zip lambda-deployment.zip bootstrap config.yaml
```

### Terraform for AWS

Deploy with Terraform:

```hcl
# main.tf
provider "aws" {
  region = var.aws_region
}

# Lambda function
resource "aws_lambda_function" "anyi_app" {
  filename         = "lambda-deployment.zip"
  function_name    = "anyi-app"
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  timeout         = 30
  memory_size     = 512

  environment {
    variables = {
      OPENAI_API_KEY    = var.openai_api_key
      ANTHROPIC_API_KEY = var.anthropic_api_key
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "anyi-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# API Gateway
resource "aws_api_gateway_rest_api" "anyi_api" {
  name        = "anyi-api"
  description = "Anyi AI API"
}

resource "aws_api_gateway_resource" "anyi_resource" {
  rest_api_id = aws_api_gateway_rest_api.anyi_api.id
  parent_id   = aws_api_gateway_rest_api.anyi_api.root_resource_id
  path_part   = "process"
}

resource "aws_api_gateway_method" "anyi_method" {
  rest_api_id   = aws_api_gateway_rest_api.anyi_api.id
  resource_id   = aws_api_gateway_resource.anyi_resource.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "anyi_integration" {
  rest_api_id = aws_api_gateway_rest_api.anyi_api.id
  resource_id = aws_api_gateway_resource.anyi_resource.id
  http_method = aws_api_gateway_method.anyi_method.http_method

  integration_http_method = "POST"
  type                   = "AWS_PROXY"
  uri                    = aws_lambda_function.anyi_app.invoke_arn
}

# Variables
variable "aws_region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "openai_api_key" {
  description = "OpenAI API key"
  sensitive   = true
}

variable "anthropic_api_key" {
  description = "Anthropic API key"
  sensitive   = true
}

# Outputs
output "api_url" {
  value = "${aws_api_gateway_deployment.anyi_deployment.invoke_url}/process"
}
```

Deploy with Terraform:

```bash
# Initialize Terraform
terraform init

# Plan deployment
terraform plan -var="openai_api_key=your-key" -var="anthropic_api_key=your-key"

# Apply deployment
terraform apply -var="openai_api_key=your-key" -var="anthropic_api_key=your-key"
```

## Monitoring and Observability

### Health Checks

Implement health check endpoints:

```go
// health/health.go
package health

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/jieliu2000/anyi"
)

type HealthStatus struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]string `json:"checks"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]string),
    }

    // Check Anyi clients
    clients := anyi.ListClients()
    if len(clients) == 0 {
        status.Status = "unhealthy"
        status.Checks["anyi_clients"] = "no clients configured"
    } else {
        status.Checks["anyi_clients"] = "ok"
    }

    // Check database connection (if applicable)
    // if err := checkDatabase(); err != nil {
    //     status.Status = "unhealthy"
    //     status.Checks["database"] = err.Error()
    // } else {
    //     status.Checks["database"] = "ok"
    // }

    w.Header().Set("Content-Type", "application/json")
    if status.Status == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(status)
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    // Check if application is ready to serve requests
    if len(anyi.ListClients()) == 0 {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("not ready"))
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ready"))
}
```

### Metrics Collection

Implement Prometheus metrics:

```go
// metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "anyi_requests_total",
            Help: "Total number of requests processed",
        },
        []string{"method", "endpoint", "status"},
    )

    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "anyi_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method", "endpoint"},
    )

    LLMRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "anyi_llm_requests_total",
            Help: "Total number of LLM requests",
        },
        []string{"provider", "model", "status"},
    )

    TokensUsed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "anyi_tokens_used_total",
            Help: "Total number of tokens used",
        },
        []string{"provider", "model", "type"},
    )
)

// Middleware for HTTP metrics
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap ResponseWriter to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", wrapped.statusCode)

        RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
        RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### Logging

Implement structured logging:

```go
// logging/logger.go
package logging

import (
    "os"
    "github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
    Logger = logrus.New()

    // Set log level from environment
    level := os.Getenv("LOG_LEVEL")
    switch level {
    case "debug":
        Logger.SetLevel(logrus.DebugLevel)
    case "info":
        Logger.SetLevel(logrus.InfoLevel)
    case "warn":
        Logger.SetLevel(logrus.WarnLevel)
    case "error":
        Logger.SetLevel(logrus.ErrorLevel)
    default:
        Logger.SetLevel(logrus.InfoLevel)
    }

    // Use JSON formatter for production
    if os.Getenv("ENVIRONMENT") == "production" {
        Logger.SetFormatter(&logrus.JSONFormatter{})
    } else {
        Logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
        })
    }
}

// Middleware for request logging
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(wrapped, r)

        Logger.WithFields(logrus.Fields{
            "method":     r.Method,
            "path":       r.URL.Path,
            "status":     wrapped.statusCode,
            "duration":   time.Since(start).Milliseconds(),
            "user_agent": r.UserAgent(),
            "remote_ip":  r.RemoteAddr,
        }).Info("Request processed")
    })
}
```

## Deployment Checklist

### Pre-deployment

- [ ] Code reviewed and tested
- [ ] Security scan completed
- [ ] Dependencies updated
- [ ] Configuration validated
- [ ] Secrets properly managed
- [ ] Documentation updated

### Deployment

- [ ] Blue-green or rolling deployment strategy
- [ ] Health checks configured
- [ ] Monitoring and alerting set up
- [ ] Backup procedures in place
- [ ] Rollback plan prepared

### Post-deployment

- [ ] Application health verified
- [ ] Metrics and logs monitored
- [ ] Performance benchmarks met
- [ ] User acceptance testing completed
- [ ] Documentation updated
- [ ] Team notified of deployment

This deployment guide provides comprehensive strategies for deploying Anyi applications across different environments and platforms. Choose the deployment method that best fits your infrastructure and requirements.
