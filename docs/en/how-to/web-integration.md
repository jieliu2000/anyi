# Web Integration

This guide shows how to integrate Anyi with popular Go web frameworks to build AI-powered web applications and APIs.

## Table of Contents

- [Integration with Gin](#integration-with-gin)
- [Integration with Echo](#integration-with-echo)
- [Integration with Fiber](#integration-with-fiber)
- [Middleware and Authentication](#middleware-and-authentication)
- [WebSocket Support](#websocket-support)
- [API Design Patterns](#api-design-patterns)
- [Error Handling in Web Context](#error-handling-in-web-context)

## Integration with Gin

Gin is one of the most popular Go web frameworks. Here's how to integrate Anyi with Gin:

### Basic Setup

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
	// Configure OpenAI client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	    config.Model = "gpt-4o-mini"

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

	// Get client
	client, err := anyi.GetClient("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "service_unavailable",
			Code:    500,
			Message: "AI service is currently unavailable",
		})
		return
	}

	// Prepare messages
	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	// Send to AI
	response, info, err := client.Chat(messages, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ai_error",
			Code:    500,
			Message: "Failed to get AI response",
		})
		return
	}

	// Return response
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
	// Initialize Anyi
	if err := setupAnyi(); err != nil {
		panic("Failed to setup Anyi: " + err.Error())
	}

	// Setup Gin
	r := gin.Default()

	// Add CORS middleware
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

	// Routes
	api := r.Group("/api/v1")
	{
		api.POST("/chat", chatHandler)
	}

	r.Run(":8080")
}
```

### Advanced Gin Integration with Workflows

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

	// Get flow
	flow, err := anyi.GetFlow(req.FlowName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "flow_not_found",
			"message": "The specified workflow does not exist",
		})
		return
	}

	// Create flow context
	flowContext := anyi.NewFlowContext(req.Input)
	if req.Context != nil {
		flowContext.Memory = req.Context
	}

	// Execute workflow
	result, err := flow.Run(flowContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "workflow_failed",
			"message": err.Error(),
		})
		return
	}

	duration := time.Since(start)

	// Return result
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
	// This would typically come from a registry or configuration
	workflows := []string{"content_generation", "data_analysis", "translation"}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

func workflowStatusHandler(c *gin.Context) {
	flowName := c.Param("name")

	// Check if workflow exists
	_, err := anyi.GetFlow(flowName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":   flowName,
		"status": "available",
	})
}
```

## Integration with Echo

Echo is another popular Go web framework with excellent performance:

### Basic Echo Setup

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

type APIHandler struct {
	anyiClient *anyi.Client
}

func NewAPIHandler() (*APIHandler, error) {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("echo-client", config)
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		anyiClient: client,
	}, nil
}

func (h *APIHandler) Chat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Message is required")
	}

	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	response, info, err := h.anyiClient.Chat(messages, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "AI service error")
	}

	return c.JSON(http.StatusOK, ChatResponse{
		Response: response.Content,
		Model:    info.Model,
		Tokens:   info.TotalTokens,
	})
}

func (h *APIHandler) StreamChat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Set headers for Server-Sent Events
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	// For demonstration - in real implementation, you'd use streaming API
	response, _, err := h.anyiClient.Chat(messages, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "AI service error")
	}

	// Simulate streaming by sending words one by one
	words := strings.Fields(response.Content)
	for i, word := range words {
		data := fmt.Sprintf("data: {\"token\": \"%s\", \"done\": %t}\n\n", word, i == len(words)-1)
		if _, err := c.Response().Write([]byte(data)); err != nil {
			return err
		}
		c.Response().Flush()
		time.Sleep(100 * time.Millisecond) // Simulate streaming delay
	}

	return nil
}

func main() {
	handler, err := NewAPIHandler()
	if err != nil {
		panic("Failed to create API handler: " + err.Error())
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	api := e.Group("/api/v1")
	api.POST("/chat", handler.Chat)
	api.POST("/chat/stream", handler.StreamChat)

	e.Logger.Fatal(e.Start(":8080"))
}
```

## Integration with Fiber

Fiber is a fast HTTP web framework inspired by Express.js:

### Basic Fiber Setup

```go
package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func setupAnyiWithFiber() error {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	_, err := anyi.NewClient("fiber-client", config)
	return err
}

func chatHandler(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	if req.Message == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Message is required",
		})
	}

	client, err := anyi.GetClient("fiber-client")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Service unavailable",
		})
	}

	messages := []chat.Message{
		{Role: "user", Content: req.Message},
	}

	start := time.Now()
	response, info, err := client.Chat(messages, nil)
	duration := time.Since(start)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "AI service error",
		})
	}

	return c.JSON(fiber.Map{
		"response":    response.Content,
		"model":       info.Model,
		"tokens":      info.TotalTokens,
		"duration_ms": duration.Milliseconds(),
	})
}

func batchChatHandler(c *fiber.Ctx) error {
	var req struct {
		Messages []string `json:"messages"`
		Model    string   `json:"model,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	if len(req.Messages) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "At least one message is required",
		})
	}

	client, err := anyi.GetClient("fiber-client")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Service unavailable",
		})
	}

	// Process messages concurrently
	type result struct {
		Index    int    `json:"index"`
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}

	results := make([]result, len(req.Messages))
	done := make(chan struct{})

	for i, msg := range req.Messages {
		go func(index int, message string) {
			messages := []chat.Message{
				{Role: "user", Content: message},
			}

			response, _, err := client.Chat(messages, nil)
			if err != nil {
				results[index] = result{
					Index: index,
					Error: err.Error(),
				}
			} else {
				results[index] = result{
					Index:    index,
					Response: response.Content,
				}
			}

			if index == len(req.Messages)-1 {
				done <- struct{}{}
			}
		}(i, msg)
	}

	// Wait for all requests to complete
	<-done

	return c.JSON(fiber.Map{
		"results": results,
		"count":   len(results),
	})
}

func main() {
	if err := setupAnyiWithFiber(); err != nil {
		panic("Failed to setup Anyi: " + err.Error())
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	api := app.Group("/api/v1")
	api.Post("/chat", chatHandler)
	api.Post("/chat/batch", batchChatHandler)

	app.Listen(":8080")
}
```

## Middleware and Authentication

### JWT Authentication Middleware

```go
package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func JWTAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
			})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// Rate limiting middleware
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	type client struct {
		requests []time.Time
	}

	clients := make(map[string]*client)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		if clients[clientIP] == nil {
			clients[clientIP] = &client{
				requests: []time.Time{now},
			}
			c.Next()
			return
		}

		// Remove requests older than 1 minute
		validRequests := []time.Time{}
		for _, reqTime := range clients[clientIP].requests {
			if now.Sub(reqTime) < time.Minute {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		clients[clientIP].requests = append(validRequests, now)
		c.Next()
	}
}

// Usage example
func setupSecureRoutes(r *gin.Engine) {
	// Public routes
	public := r.Group("/api/v1/public")
	{
		public.POST("/login", loginHandler)
		public.GET("/health", healthHandler)
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(JWTAuthMiddleware(os.Getenv("JWT_SECRET")))
	protected.Use(RateLimitMiddleware(10)) // 10 requests per minute
	{
		protected.POST("/chat", authenticatedChatHandler)
		protected.POST("/workflow/execute", authenticatedWorkflowHandler)
	}
}

func authenticatedChatHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")

	// Log user activity
	log.Printf("Chat request from user %s (%s)", username, userID)

	// Your existing chat handler logic here
	chatHandler(c)
}
```

## WebSocket Support

### Real-time Chat with WebSockets

```go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/chat"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type WSMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

type ChatSession struct {
	conn     *websocket.Conn
	client   *anyi.Client
	messages []chat.Message
}

func NewChatSession(conn *websocket.Conn, client *anyi.Client) *ChatSession {
	return &ChatSession{
		conn:     conn,
		client:   client,
		messages: []chat.Message{},
	}
}

func (cs *ChatSession) handleMessages() {
	defer cs.conn.Close()

	for {
		var msg WSMessage
		err := cs.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch msg.Type {
		case "chat":
			cs.handleChatMessage(msg.Content)
		case "clear":
			cs.messages = []chat.Message{}
			cs.sendMessage(WSMessage{
				Type:    "system",
				Content: "Chat history cleared",
			})
		case "ping":
			cs.sendMessage(WSMessage{
				Type:    "pong",
				Content: "pong",
			})
		}
	}
}

func (cs *ChatSession) handleChatMessage(content string) {
	// Add user message to history
	cs.messages = append(cs.messages, chat.Message{
		Role:    "user",
		Content: content,
	})

	// Send to AI
	response, _, err := cs.client.Chat(cs.messages, nil)
	if err != nil {
		cs.sendMessage(WSMessage{
			Type:  "error",
			Error: "Failed to get AI response",
		})
		return
	}

	// Add AI response to history
	cs.messages = append(cs.messages, *response)

	// Send response back to client
	cs.sendMessage(WSMessage{
		Type:    "response",
		Content: response.Content,
	})
}

func (cs *ChatSession) sendMessage(msg WSMessage) {
	if err := cs.conn.WriteJSON(msg); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client, err := anyi.GetClient("default")
	if err != nil {
		log.Printf("Failed to get Anyi client: %v", err)
		conn.Close()
		return
	}

	session := NewChatSession(conn, client)
	session.handleMessages()
}

func setupWebSocketRoutes(r *gin.Engine) {
	r.GET("/ws/chat", websocketHandler)

	// Serve static files for WebSocket client
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", gin.H{
			"title": "AI Chat",
		})
	})
}
```

### WebSocket Client Example (HTML/JavaScript)

```html
<!-- templates/chat.html -->
<!DOCTYPE html>
<html>
  <head>
    <title>{{.title}}</title>
    <style>
      #messages {
        height: 400px;
        overflow-y: scroll;
        border: 1px solid #ccc;
        padding: 10px;
        margin-bottom: 10px;
      }
      .message {
        margin-bottom: 10px;
        padding: 5px;
        border-radius: 5px;
      }
      .user {
        background-color: #e3f2fd;
      }
      .ai {
        background-color: #f3e5f5;
      }
      .system {
        background-color: #e8f5e8;
      }
      .error {
        background-color: #ffebee;
        color: #c62828;
      }
    </style>
  </head>
  <body>
    <h1>AI Chat</h1>
    <div id="messages"></div>
    <input
      type="text"
      id="messageInput"
      placeholder="Type your message..."
      style="width: 70%;"
    />
    <button onclick="sendMessage()">Send</button>
    <button onclick="clearChat()">Clear</button>

    <script>
      const ws = new WebSocket("ws://localhost:8080/ws/chat");
      const messages = document.getElementById("messages");
      const messageInput = document.getElementById("messageInput");

      ws.onmessage = function (event) {
        const msg = JSON.parse(event.data);
        addMessage(msg.content, msg.type);
      };

      ws.onopen = function (event) {
        addMessage("Connected to AI chat", "system");
      };

      ws.onclose = function (event) {
        addMessage("Disconnected from AI chat", "system");
      };

      function addMessage(content, type) {
        const messageDiv = document.createElement("div");
        messageDiv.className = "message " + type;
        messageDiv.textContent = content;
        messages.appendChild(messageDiv);
        messages.scrollTop = messages.scrollHeight;
      }

      function sendMessage() {
        const content = messageInput.value.trim();
        if (content) {
          addMessage(content, "user");
          ws.send(
            JSON.stringify({
              type: "chat",
              content: content,
            })
          );
          messageInput.value = "";
        }
      }

      function clearChat() {
        ws.send(
          JSON.stringify({
            type: "clear",
          })
        );
        messages.innerHTML = "";
      }

      messageInput.addEventListener("keypress", function (e) {
        if (e.key === "Enter") {
          sendMessage();
        }
      });
    </script>
  </body>
</html>
```

## API Design Patterns

### RESTful API Design

```go
package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jieliu2000/anyi"
)

// RESTful endpoints for conversations
func setupRESTfulAPI(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// Conversations
		conversations := api.Group("/conversations")
		{
			conversations.GET("", listConversations)
			conversations.POST("", createConversation)
			conversations.GET("/:id", getConversation)
			conversations.DELETE("/:id", deleteConversation)
			conversations.POST("/:id/messages", addMessage)
			conversations.GET("/:id/messages", getMessages)
		}

		// Models
		models := api.Group("/models")
		{
			models.GET("", listAvailableModels)
			models.GET("/:name/status", getModelStatus)
		}

		// Workflows
		workflows := api.Group("/workflows")
		{
			workflows.GET("", listWorkflows)
			workflows.POST("/:name/execute", executeWorkflow)
			workflows.GET("/:name", getWorkflowInfo)
		}
	}
}

type Conversation struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Messages []chat.Message `json:"messages"`
	Created  time.Time      `json:"created"`
	Updated  time.Time      `json:"updated"`
}

// In-memory storage (use database in production)
var conversations = make(map[string]*Conversation)
var nextID = 1

func listConversations(c *gin.Context) {
	var result []*Conversation
	for _, conv := range conversations {
		result = append(result, conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": result,
		"count":         len(result),
	})
}

func createConversation(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := strconv.Itoa(nextID)
	nextID++

	conv := &Conversation{
		ID:       id,
		Title:    req.Title,
		Messages: []chat.Message{},
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	conversations[id] = conv

	c.JSON(http.StatusCreated, conv)
}

func addMessage(c *gin.Context) {
	convID := c.Param("id")
	conv, exists := conversations[convID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add user message
	userMessage := chat.Message{
		Role:    "user",
		Content: req.Content,
	}
	conv.Messages = append(conv.Messages, userMessage)

	// Get AI response
	client, err := anyi.GetClient("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service unavailable"})
		return
	}

	response, _, err := client.Chat(conv.Messages, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error"})
		return
	}

	// Add AI response
	conv.Messages = append(conv.Messages, *response)
	conv.Updated = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"message": response,
		"conversation": conv,
	})
}
```

## Error Handling in Web Context

### Centralized Error Handling

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Details string `json:"details,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			var apiErr APIError

			switch err.Type {
			case gin.ErrorTypeBind:
				apiErr = APIError{
					Code:    http.StatusBadRequest,
					Message: "Invalid request format",
					Type:    "validation_error",
					Details: err.Error(),
				}
			case gin.ErrorTypePublic:
				apiErr = APIError{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Type:    "server_error",
				}
			default:
				apiErr = APIError{
					Code:    http.StatusInternalServerError,
					Message: "An unexpected error occurred",
					Type:    "unknown_error",
				}
			}

			c.JSON(apiErr.Code, apiErr)
		}
	}
}

// Custom error types
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type BusinessLogicError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleBusinessError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *BusinessLogicError:
		c.JSON(http.StatusBadRequest, APIError{
			Code:    http.StatusBadRequest,
			Message: e.Message,
			Type:    "business_error",
			Details: e.Code,
		})
	default:
		c.JSON(http.StatusInternalServerError, APIError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Type:    "server_error",
		})
	}
}
```

This comprehensive guide covers the major aspects of integrating Anyi with popular Go web frameworks, providing you with the foundation to build robust AI-powered web applications and APIs.
