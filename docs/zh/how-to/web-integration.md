# Web 集成

本指南展示如何将 Anyi 与流行的 Go Web 框架集成，构建 AI 驱动的 Web 应用程序和 API。

## 目录

- [与 Gin 集成](#与-gin-集成)
- [与 Echo 集成](#与-echo-集成)
- [与 Fiber 集成](#与-fiber-集成)
- [中间件和身份验证](#中间件和身份验证)
- [WebSocket 支持](#websocket-支持)
- [API 设计模式](#api-设计模式)
- [Web 上下文中的错误处理](#web-上下文中的错误处理)

## 与 Gin 集成

Gin 是最受欢迎的 Go Web 框架之一。以下是如何将 Anyi 与 Gin 集成：

### 基本设置

```go
package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Model   string `json:"model,omitempty"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
	Tokens   int    `json:"tokens,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func setupAnyi() error {
	// 配置 OpenAI 客户端
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = "gpt-3.5-turbo"

	_, err := anyi.NewClient("default", config)
	return err
}

func chatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 获取客户端
	client, err := anyi.GetClient("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "service_unavailable",
			Code:    500,
			Message: "AI 服务当前不可用",
		})
		return
	}

	// 准备消息
	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	// 发送到 AI
	response, info, err := client.Chat(messages, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ai_error",
			Code:    500,
			Message: "获取 AI 响应失败",
		})
		return
	}

	// 返回响应
	chatResp := ChatResponse{
		Response: response.Content,
		Model:    info.Model,
	}

	if info != nil {
		chatResp.Tokens = info.TotalTokens
	}

	c.JSON(http.StatusOK, chatResp)
}

func main() {
	// 初始化 Anyi
	if err := setupAnyi(); err != nil {
		panic("设置 Anyi 失败：" + err.Error())
	}

	// 设置 Gin
	r := gin.Default()

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 路由
	api := r.Group("/api/v1")
	{
		api.POST("/chat", chatHandler)
	}

	r.Run(":8080")
}
```

### 与工作流的高级 Gin 集成

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jieliu2000/anyi"
)

type WorkflowRequest struct {
	FlowName string                 `json:"flow_name" binding:"required"`
	Input    string                 `json:"input" binding:"required"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

type WorkflowResponse struct {
	Result    string                 `json:"result"`
	FlowName  string                 `json:"flow_name"`
	Duration  int64                  `json:"duration_ms"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

func workflowHandler(c *gin.Context) {
	var req WorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	// 获取流程
	flow, err := anyi.GetFlow(req.FlowName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "flow_not_found",
			"message": "指定的工作流不存在",
		})
		return
	}

	// 创建流程上下文
	flowContext := anyi.NewFlowContext(req.Input)
	if req.Context != nil {
		flowContext.Memory = req.Context
	}

	// 执行工作流
	result, err := flow.Run(flowContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "workflow_failed",
			"message": err.Error(),
		})
		return
	}

	duration := time.Since(start)

	// 返回结果
	response := WorkflowResponse{
		Result:   result.Text,
		FlowName: req.FlowName,
		Duration: duration.Milliseconds(),
	}

	if result.Memory != nil {
		response.Context = result.Memory.(map[string]interface{})
	}

	c.JSON(http.StatusOK, response)
}

func setupWorkflowRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/workflow/execute", workflowHandler)
		api.GET("/workflow/list", listWorkflowsHandler)
		api.GET("/workflow/:name/status", workflowStatusHandler)
	}
}

func listWorkflowsHandler(c *gin.Context) {
	// 这通常来自注册表或配置
	workflows := []string{"content_generation", "data_analysis", "translation"}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

func workflowStatusHandler(c *gin.Context) {
	flowName := c.Param("name")

	// 检查工作流是否存在
	_, err := anyi.GetFlow(flowName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "工作流未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":   flowName,
		"status": "active",
	})
}
```

## 与 Echo 集成

Echo 是另一个高性能的 Go Web 框架：

```go
package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

type EchoServer struct {
	echo *echo.Echo
}

func NewEchoServer() *EchoServer {
	e := echo.New()

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	return &EchoServer{echo: e}
}

func (s *EchoServer) setupRoutes() {
	// API 路由
	api := s.echo.Group("/api/v1")
	api.POST("/chat", s.chatHandler)
	api.POST("/analyze", s.analyzeHandler)
	api.GET("/models", s.modelsHandler)
}

func (s *EchoServer) chatHandler(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无效的请求格式",
		})
	}

	client, err := anyi.GetClient("default")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "服务不可用",
		})
	}

	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	response, info, err := client.Chat(messages, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "AI 处理失败",
		})
	}

	return c.JSON(http.StatusOK, ChatResponse{
		Response: response.Content,
		Model:    info.Model,
		Tokens:   info.TotalTokens,
	})
}

func (s *EchoServer) analyzeHandler(c echo.Context) error {
	type AnalyzeRequest struct {
		Text string `json:"text" validate:"required"`
		Type string `json:"type" validate:"required,oneof=sentiment summary keywords"`
	}

	var req AnalyzeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无效的请求格式",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "验证失败",
		})
	}

	// 根据类型选择不同的提示
	var prompt string
	switch req.Type {
	case "sentiment":
		prompt = "分析以下文本的情感倾向（积极/消极/中性）：\n\n" + req.Text
	case "summary":
		prompt = "总结以下文本的主要内容：\n\n" + req.Text
	case "keywords":
		prompt = "提取以下文本的关键词：\n\n" + req.Text
	}

	client, _ := anyi.GetClient("default")
	messages := []chat.Message{
		{Role: "user", Content: prompt},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "分析失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"type":   req.Type,
		"result": response.Content,
	})
}

func (s *EchoServer) modelsHandler(c echo.Context) error {
	models := []map[string]string{
		{"name": "gpt-3.5-turbo", "provider": "openai", "status": "active"},
		{"name": "gpt-4", "provider": "openai", "status": "active"},
		{"name": "claude-3-opus", "provider": "anthropic", "status": "active"},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"models": models,
		"count":  len(models),
	})
}

func main() {
	// 初始化 Anyi
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	anyi.NewClient("default", config)

	// 创建和配置服务器
	server := NewEchoServer()
	server.setupRoutes()

	// 启动服务器
	server.echo.Logger.Fatal(server.echo.Start(":8080"))
}
```

## 与 Fiber 集成

Fiber 是受 Express.js 启发的快速 HTTP Web 框架：

```go
package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func setupAnyi() error {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	_, err := anyi.NewClient("fiber", config)
	return err
}

func chatHandler(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "无效的请求体",
		})
	}

	if req.Message == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "消息不能为空",
		})
	}

	client, err := anyi.GetClient("fiber")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "服务不可用",
		})
	}

	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	response, info, err := client.Chat(messages, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "AI 处理失败",
		})
	}

	return c.JSON(ChatResponse{
		Response: response.Content,
		Model:    info.Model,
		Tokens:   info.TotalTokens,
	})
}

func streamChatHandler(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "无效的请求体",
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// 模拟流式响应
	client, _ := anyi.GetClient("fiber")
	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		c.Write([]byte("data: " + `{"error": "处理失败"}` + "\n\n"))
		return nil
	}

	// 分块发送响应
	words := strings.Fields(response.Content)
	for i, word := range words {
		data := map[string]interface{}{
			"content": word + " ",
			"index":   i,
			"done":    i == len(words)-1,
		}
		jsonData, _ := json.Marshal(data)
		c.Write([]byte("data: " + string(jsonData) + "\n\n"))

		// 模拟延迟
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func main() {
	// 初始化 Anyi
	if err := setupAnyi(); err != nil {
		panic("设置 Anyi 失败：" + err.Error())
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 中间件
	app.Use(logger.New())
	app.Use(cors.New())

	// 路由
	api := app.Group("/api/v1")
	api.Post("/chat", chatHandler)
	api.Post("/chat/stream", streamChatHandler)

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "健康",
			"time":   time.Now(),
		})
	})

	app.Listen(":8080")
}
```

## 中间件和身份验证

### JWT 认证中间件

```go
package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	secretKey []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: []byte(secret),
	}
}

func (am *AuthMiddleware) GenerateToken(userID, username, role string, permissions []string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.secretKey)
}

func (am *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return am.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := am.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(403, gin.H{"error": "权限信息不存在"})
			c.Abort()
			return
		}

		userPermissions := permissions.([]string)
		hasPermission := false
		for _, perm := range userPermissions {
			if perm == permission || perm == "admin" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(403, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

### 速率限制中间件

```go
package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// 清理过期的请求
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// 检查是否超过限制
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// 添加新请求
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用用户 ID 或 IP 作为键
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			key = userID.(string)
		}

		if !rl.Allow(key) {
			c.JSON(429, gin.H{
				"error": "请求过于频繁",
				"message": "请稍后重试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

## WebSocket 支持

### WebSocket 聊天实现

```go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该更严格
	},
}

type WSMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type ChatMessage struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
	Tokens   int    `json:"tokens"`
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败：%v", err)
		return
	}
	defer conn.Close()

	// 发送欢迎消息
	welcomeMsg := WSMessage{
		Type:    "welcome",
		Content: "欢迎使用 AI 聊天服务",
	}
	conn.WriteJSON(welcomeMsg)

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取消息失败：%v", err)
			break
		}

		switch msg.Type {
		case "chat":
			handleChatMessage(conn, msg.Content)
		case "ping":
			pongMsg := WSMessage{Type: "pong", Content: "pong"}
			conn.WriteJSON(pongMsg)
		default:
			errorMsg := WSMessage{
				Type:    "error",
				Content: "未知的消息类型",
			}
			conn.WriteJSON(errorMsg)
		}
	}
}

func handleChatMessage(conn *websocket.Conn, content interface{}) {
	// 解析聊天消息
	contentBytes, _ := json.Marshal(content)
	var chatMsg ChatMessage
	if err := json.Unmarshal(contentBytes, &chatMsg); err != nil {
		errorMsg := WSMessage{
			Type:    "error",
			Content: "无效的消息格式",
		}
		conn.WriteJSON(errorMsg)
		return
	}

	// 发送处理中状态
	processingMsg := WSMessage{
		Type:    "processing",
		Content: "正在处理您的消息...",
	}
	conn.WriteJSON(processingMsg)

	// 获取 AI 响应
	client, err := anyi.GetClient("default")
	if err != nil {
		errorMsg := WSMessage{
			Type:    "error",
			Content: "AI 服务不可用",
		}
		conn.WriteJSON(errorMsg)
		return
	}

	messages := []chat.Message{
		{Role: "user", Content: chatMsg.Message},
	}

	response, info, err := client.Chat(messages, nil)
	if err != nil {
		errorMsg := WSMessage{
			Type:    "error",
			Content: "AI 处理失败",
		}
		conn.WriteJSON(errorMsg)
		return
	}

	// 发送响应
	chatResp := ChatResponse{
		Response: response.Content,
		Model:    info.Model,
		Tokens:   info.TotalTokens,
	}

	responseMsg := WSMessage{
		Type:    "chat_response",
		Content: chatResp,
	}
	conn.WriteJSON(responseMsg)
}
```

### WebSocket 客户端示例

```html
<!DOCTYPE html>
<html>
  <head>
    <title>AI 聊天</title>
    <style>
      body {
        font-family: Arial, sans-serif;
        margin: 20px;
      }
      #messages {
        border: 1px solid #ccc;
        height: 400px;
        overflow-y: scroll;
        padding: 10px;
        margin-bottom: 10px;
      }
      #messageInput {
        width: 70%;
        padding: 5px;
      }
      #sendButton {
        padding: 5px 10px;
      }
      .message {
        margin: 5px 0;
      }
      .user {
        color: blue;
      }
      .ai {
        color: green;
      }
      .error {
        color: red;
      }
      .system {
        color: gray;
        font-style: italic;
      }
    </style>
  </head>
  <body>
    <h1>AI 聊天</h1>
    <div id="messages"></div>
    <input type="text" id="messageInput" placeholder="输入您的消息..." />
    <button id="sendButton">发送</button>

    <script>
      const ws = new WebSocket("ws://localhost:8080/ws");
      const messages = document.getElementById("messages");
      const messageInput = document.getElementById("messageInput");
      const sendButton = document.getElementById("sendButton");

      ws.onopen = function (event) {
        addMessage("已连接到聊天服务", "system");
      };

      ws.onmessage = function (event) {
        const data = JSON.parse(event.data);

        switch (data.type) {
          case "welcome":
            addMessage(data.content, "system");
            break;
          case "chat_response":
            addMessage("AI: " + data.content.response, "ai");
            addMessage(
              `模型: ${data.content.model}, 令牌: ${data.content.tokens}`,
              "system"
            );
            break;
          case "processing":
            addMessage(data.content, "system");
            break;
          case "error":
            addMessage("错误: " + data.content, "error");
            break;
        }
      };

      ws.onclose = function (event) {
        addMessage("连接已断开", "system");
      };

      ws.onerror = function (error) {
        addMessage("连接错误: " + error, "error");
      };

      function sendMessage() {
        const message = messageInput.value.trim();
        if (message) {
          addMessage("您: " + message, "user");

          const wsMessage = {
            type: "chat",
            content: {
              message: message,
              user_id: "user123",
            },
          };

          ws.send(JSON.stringify(wsMessage));
          messageInput.value = "";
        }
      }

      function addMessage(text, className) {
        const messageDiv = document.createElement("div");
        messageDiv.className = "message " + className;
        messageDiv.textContent = text;
        messages.appendChild(messageDiv);
        messages.scrollTop = messages.scrollHeight;
      }

      sendButton.addEventListener("click", sendMessage);
      messageInput.addEventListener("keypress", function (e) {
        if (e.key === "Enter") {
          sendMessage();
        }
      });
    </script>
  </body>
</html>
```

## API 设计模式

### RESTful API 设计

```go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jieliu2000/anyi"
)

// 聊天会话管理
type ChatSession struct {
	ID       string              `json:"id"`
	UserID   string              `json:"user_id"`
	Title    string              `json:"title"`
	Messages []chat.Message      `json:"messages"`
	Metadata map[string]interface{} `json:"metadata"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type ChatAPI struct {
	sessions map[string]*ChatSession
	mutex    sync.RWMutex
}

func NewChatAPI() *ChatAPI {
	return &ChatAPI{
		sessions: make(map[string]*ChatSession),
	}
}

// POST /api/v1/chat/sessions - 创建新会话
func (api *ChatAPI) CreateSession(c *gin.Context) {
	var req struct {
		Title    string                 `json:"title"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	userID, _ := c.Get("user_id")
	sessionID := generateSessionID()

	session := &ChatSession{
		ID:        sessionID,
		UserID:    userID.(string),
		Title:     req.Title,
		Messages:  []chat.Message{},
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	api.mutex.Lock()
	api.sessions[sessionID] = session
	api.mutex.Unlock()

	c.JSON(http.StatusCreated, session)
}

// GET /api/v1/chat/sessions - 获取会话列表
func (api *ChatAPI) GetSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var userSessions []*ChatSession
	api.mutex.RLock()
	for _, session := range api.sessions {
		if session.UserID == userID.(string) {
			userSessions = append(userSessions, session)
		}
	}
	api.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"sessions": userSessions,
		"count":    len(userSessions),
	})
}

// GET /api/v1/chat/sessions/:id - 获取特定会话
func (api *ChatAPI) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := c.Get("user_id")

	api.mutex.RLock()
	session, exists := api.sessions[sessionID]
	api.mutex.RUnlock()

	if !exists || session.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话未找到"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// POST /api/v1/chat/sessions/:id/messages - 发送消息
func (api *ChatAPI) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req struct {
		Message string `json:"message" binding:"required"`
		Model   string `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	api.mutex.Lock()
	session, exists := api.sessions[sessionID]
	if !exists || session.UserID != userID.(string) {
		api.mutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "会话未找到"})
		return
	}

	// 添加用户消息
	userMessage := chat.Message{
		Role:    "user",
		Content: req.Message,
	}
	session.Messages = append(session.Messages, userMessage)
	session.UpdatedAt = time.Now()
	api.mutex.Unlock()

	// 获取 AI 响应
	client, err := anyi.GetClient("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务不可用"})
		return
	}

	response, info, err := client.Chat(session.Messages, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 处理失败"})
		return
	}

	// 添加 AI 响应
	api.mutex.Lock()
	session.Messages = append(session.Messages, *response)
	session.UpdatedAt = time.Now()
	api.mutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": response.Content,
		"model":   info.Model,
		"tokens":  info.TotalTokens,
	})
}

// DELETE /api/v1/chat/sessions/:id - 删除会话
func (api *ChatAPI) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	userID, _ := c.Get("user_id")

	api.mutex.Lock()
	session, exists := api.sessions[sessionID]
	if exists && session.UserID == userID.(string) {
		delete(api.sessions, sessionID)
	}
	api.mutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话未找到"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "会话已删除"})
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
```

### 分页和过滤

```go
type PaginationParams struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	SortBy   string `form:"sort_by,default=created_at"`
	SortDir  string `form:"sort_dir,default=desc"`
	Search   string `form:"search"`
}

func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	if p.SortDir != "asc" && p.SortDir != "desc" {
		p.SortDir = "desc"
	}
	return nil
}

func (api *ChatAPI) GetSessionsPaginated(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数"})
		return
	}

	params.Validate()
	userID, _ := c.Get("user_id")

	// 过滤和搜索
	var filteredSessions []*ChatSession
	api.mutex.RLock()
	for _, session := range api.sessions {
		if session.UserID != userID.(string) {
			continue
		}

		if params.Search != "" {
			if !strings.Contains(strings.ToLower(session.Title), strings.ToLower(params.Search)) {
				continue
			}
		}

		filteredSessions = append(filteredSessions, session)
	}
	api.mutex.RUnlock()

	// 排序
	sort.Slice(filteredSessions, func(i, j int) bool {
		switch params.SortBy {
		case "title":
			if params.SortDir == "asc" {
				return filteredSessions[i].Title < filteredSessions[j].Title
			}
			return filteredSessions[i].Title > filteredSessions[j].Title
		default: // created_at
			if params.SortDir == "asc" {
				return filteredSessions[i].CreatedAt.Before(filteredSessions[j].CreatedAt)
			}
			return filteredSessions[i].CreatedAt.After(filteredSessions[j].CreatedAt)
		}
	})

	// 分页
	total := len(filteredSessions)
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedSessions := filteredSessions[start:end]

	c.JSON(http.StatusOK, gin.H{
		"sessions": paginatedSessions,
		"pagination": gin.H{
			"page":       params.Page,
			"page_size":  params.PageSize,
			"total":      total,
			"total_pages": (total + params.PageSize - 1) / params.PageSize,
		},
	})
}
```

## Web 上下文中的错误处理

### 统一错误处理

```go
package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

var (
	ErrInvalidRequest   = APIError{Code: 400, Message: "无效的请求"}
	ErrUnauthorized     = APIError{Code: 401, Message: "未授权"}
	ErrForbidden        = APIError{Code: 403, Message: "禁止访问"}
	ErrNotFound         = APIError{Code: 404, Message: "资源未找到"}
	ErrRateLimited      = APIError{Code: 429, Message: "请求过于频繁"}
	ErrInternalServer   = APIError{Code: 500, Message: "内部服务器错误"}
	ErrServiceUnavailable = APIError{Code: 503, Message: "服务不可用"}
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			switch e := err.Err.(type) {
			case APIError:
				c.JSON(e.Code, gin.H{
					"error": gin.H{
						"code":    e.Code,
						"message": e.Message,
						"details": e.Details,
					},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    500,
						"message": "内部服务器错误",
						"details": e.Error(),
					},
				})
			}
		}
	}
}

func HandleError(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}

func NewAPIError(code int, message, details string) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
```

### 使用示例

```go
func protectedHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		HandleError(c, ErrUnauthorized)
		return
	}

	// 业务逻辑
	result, err := processUserRequest(userID.(string))
	if err != nil {
		if err == sql.ErrNoRows {
			HandleError(c, ErrNotFound)
			return
		}
		HandleError(c, NewAPIError(500, "处理失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}
```

通过这些集成示例和模式，您可以轻松地将 Anyi 集成到您的 Go Web 应用程序中，构建强大的 AI 驱动的 Web 服务和 API。
