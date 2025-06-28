# 安全最佳实践

本文档提供使用 Anyi 框架构建和部署 AI 应用程序的全面安全指南。

## 概述

AI 应用程序的安全涉及多个层面：

- **API 安全**：保护 LLM API 密钥和端点
- **输入验证**：防止提示注入和恶意输入
- **输出清理**：确保 AI 输出使用安全
- **数据隐私**：保护敏感信息
- **基础设施安全**：保护部署环境
- **访问控制**：管理用户权限和身份验证

## API 密钥管理

### 环境变量

永远不要在源代码中硬编码 API 密钥。使用环境变量：

```go
// ❌ 错误：硬编码 API 密钥
config := openai.DefaultConfig("sk-1234567890abcdef...")

// ✅ 正确：环境变量
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
```

### 配置文件

在配置文件中使用环境变量替换：

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY" # 运行时替换
      model: "gpt-4"
```

### 密钥管理

对于生产环境，使用专用的密钥管理：

```go
// 使用 AWS Secrets Manager 的示例
func getAPIKey(secretName string) (string, error) {
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess)

    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }

    result, err := svc.GetSecretValue(input)
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}

func main() {
    apiKey, err := getAPIKey("openai-api-key")
    if err != nil {
        log.Fatal(err)
    }

    config := openai.DefaultConfig(apiKey)
    client, err := anyi.NewClient("openai", config)
    // ...
}
```

### 密钥轮换

实现定期 API 密钥轮换：

```go
type RotatingClient struct {
    currentClient anyi.Client
    keyRotator    *KeyRotator
    mutex         sync.RWMutex
}

func (rc *RotatingClient) Chat(messages []chat.Message, options *chat.Options) (*chat.Message, chat.ResponseInfo, error) {
    rc.mutex.RLock()
    client := rc.currentClient
    rc.mutex.RUnlock()

    response, info, err := client.Chat(messages, options)

    // 检查错误是否表示无效的 API 密钥
    if isInvalidKeyError(err) {
        rc.rotateKey()
        // 使用新密钥重试
        rc.mutex.RLock()
        client = rc.currentClient
        rc.mutex.RUnlock()
        return client.Chat(messages, options)
    }

    return response, info, err
}

func (rc *RotatingClient) rotateKey() {
    rc.mutex.Lock()
    defer rc.mutex.Unlock()

    newKey, err := rc.keyRotator.GetNewKey()
    if err != nil {
        log.Printf("密钥轮换失败：%v", err)
        return
    }

    newConfig := openai.DefaultConfig(newKey)
    newClient, err := anyi.NewClient("openai-rotated", newConfig)
    if err != nil {
        log.Printf("创建新客户端失败：%v", err)
        return
    }

    rc.currentClient = newClient
}
```

## 输入验证和清理

### 提示注入防护

防护提示注入攻击：

````go
type SecurePromptBuilder struct {
    maxLength     int
    blockedPhrases []string
    allowedChars   *regexp.Regexp
}

func NewSecurePromptBuilder() *SecurePromptBuilder {
    return &SecurePromptBuilder{
        maxLength: 4000,
        blockedPhrases: []string{
            "忽略之前的指令",
            "无视上面的内容",
            "忘记所有内容",
            "新指令：",
            "系统：",
            "助手：",
            "ignore previous instructions",
            "disregard the above",
            "forget everything",
            "new instructions:",
            "system:",
            "assistant:",
        },
        allowedChars: regexp.MustCompile(`^[a-zA-Z0-9\s\.,!?;:()\-'"，。！？；：（）【】""'']+$`),
    }
}

func (spb *SecurePromptBuilder) ValidateInput(input string) error {
    // 检查长度
    if len(input) > spb.maxLength {
        return fmt.Errorf("输入过长：%d 字符（最大：%d）", len(input), spb.maxLength)
    }

    // 检查被阻止的短语
    inputLower := strings.ToLower(input)
    for _, phrase := range spb.blockedPhrases {
        if strings.Contains(inputLower, phrase) {
            return fmt.Errorf("输入包含被阻止的短语：%s", phrase)
        }
    }

    // 检查字符白名单
    if !spb.allowedChars.MatchString(input) {
        return fmt.Errorf("输入包含无效字符")
    }

    return nil
}

func (spb *SecurePromptBuilder) SanitizeInput(input string) string {
    // 移除潜在的注入标记
    input = strings.ReplaceAll(input, "```", "")
    input = strings.ReplaceAll(input, "---", "")

    // 转义特殊字符
    input = html.EscapeString(input)

    // 限制长度
    if len(input) > spb.maxLength {
        input = input[:spb.maxLength]
    }

    return input
}

// 在执行器中使用
type SecureLLMExecutor struct {
    *anyi.LLMExecutor
    promptBuilder *SecurePromptBuilder
}

func (e *SecureLLMExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 验证和清理输入
    if err := e.promptBuilder.ValidateInput(context.Text); err != nil {
        return nil, fmt.Errorf("输入验证失败：%v", err)
    }

    sanitizedText := e.promptBuilder.SanitizeInput(context.Text)

    // 创建清理后的上下文
    sanitizedContext := &anyi.FlowContext{
        Text:   sanitizedText,
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return e.LLMExecutor.Execute(sanitizedContext)
}
````

### 内容过滤

实现不当输入的内容过滤：

```go
type ContentFilter struct {
    profanityList []string
    toxicityAPI   string
}

func (cf *ContentFilter) FilterContent(text string) (string, error) {
    // 检查脏话
    textLower := strings.ToLower(text)
    for _, word := range cf.profanityList {
        if strings.Contains(textLower, word) {
            return "", fmt.Errorf("内容包含不当语言")
        }
    }

    // 使用外部 API 检查毒性
    if cf.toxicityAPI != "" {
        toxic, err := cf.checkToxicity(text)
        if err != nil {
            log.Printf("毒性检查失败：%v", err)
            // 如果毒性检查失败，继续处理
        } else if toxic {
            return "", fmt.Errorf("内容被标记为有毒")
        }
    }

    return text, nil
}

func (cf *ContentFilter) checkToxicity(text string) (bool, error) {
    // 外部毒性 API 的实现
    // 这是一个占位符 - 使用实际服务如 Perspective API
    return false, nil
}
```

## 输出验证和清理

### 响应验证

在使用 AI 响应之前验证它们：

```go
type ResponseValidator struct {
    maxLength      int
    requiredFields []string
    forbiddenTerms []string
}

func (rv *ResponseValidator) ValidateResponse(response string) error {
    // 检查长度
    if len(response) > rv.maxLength {
        return fmt.Errorf("响应过长")
    }

    // 检查禁止词汇
    responseLower := strings.ToLower(response)
    for _, term := range rv.forbiddenTerms {
        if strings.Contains(responseLower, term) {
            return fmt.Errorf("响应包含禁止词汇：%s", term)
        }
    }

    // 其他验证逻辑...
    return nil
}

func (rv *ResponseValidator) SanitizeResponse(response string) string {
    // 移除潜在的脚本标签
    response = regexp.MustCompile(`<script[^>]*>.*?</script>`).ReplaceAllString(response, "")

    // 移除其他潜在危险的 HTML
    response = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(response, "")

    // 转义剩余的 HTML 实体
    response = html.EscapeString(response)

    return response
}
```

### 代码执行防护

永远不要直接执行 AI 生成的代码：

```go
// ❌ 危险：直接执行
func dangerousExecuteCode(code string) error {
    return exec.Command("sh", "-c", code).Run()
}

// ✅ 安全：沙盒执行和验证
type SafeCodeExecutor struct {
    allowedCommands map[string]bool
    timeout         time.Duration
    workingDir      string
}

func (sce *SafeCodeExecutor) ExecuteCode(code string) (string, error) {
    // 解析和验证代码
    if err := sce.validateCode(code); err != nil {
        return "", err
    }

    // 创建临时文件
    tmpFile, err := os.CreateTemp(sce.workingDir, "safe_code_*.py")
    if err != nil {
        return "", err
    }
    defer os.Remove(tmpFile.Name())

    // 将代码写入文件
    if _, err := tmpFile.WriteString(code); err != nil {
        return "", err
    }
    tmpFile.Close()

    // 在沙盒中执行
    ctx, cancel := context.WithTimeout(context.Background(), sce.timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, "python", tmpFile.Name())
    cmd.Dir = sce.workingDir

    output, err := cmd.Output()
    return string(output), err
}

func (sce *SafeCodeExecutor) validateCode(code string) error {
    // 检查危险操作
    dangerous := []string{
        "import os",
        "import subprocess",
        "exec(",
        "eval(",
        "open(",
        "__import__",
    }

    for _, danger := range dangerous {
        if strings.Contains(code, danger) {
            return fmt.Errorf("代码包含危险操作：%s", danger)
        }
    }

    return nil
}
```

## 数据隐私和保护

### 敏感数据检测

检测和处理敏感信息：

```go
type SensitiveDataDetector struct {
    patterns map[string]*regexp.Regexp
}

func NewSensitiveDataDetector() *SensitiveDataDetector {
    return &SensitiveDataDetector{
        patterns: map[string]*regexp.Regexp{
            "id_card":     regexp.MustCompile(`\b\d{15}|\d{18}\b`), // 身份证号
            "phone":       regexp.MustCompile(`\b1[3-9]\d{9}\b`),   // 手机号
            "credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
            "email":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
            "bank_card":   regexp.MustCompile(`\b\d{16,19}\b`), // 银行卡号
        },
    }
}

func (sdd *SensitiveDataDetector) DetectSensitiveData(text string) map[string][]string {
    detected := make(map[string][]string)

    for dataType, pattern := range sdd.patterns {
        matches := pattern.FindAllString(text, -1)
        if len(matches) > 0 {
            detected[dataType] = matches
        }
    }

    return detected
}

func (sdd *SensitiveDataDetector) RedactSensitiveData(text string) string {
    for _, pattern := range sdd.patterns {
        text = pattern.ReplaceAllString(text, "[已脱敏]")
    }
    return text
}

// 在工作流中使用
type PrivacyAwareLLMExecutor struct {
    *anyi.LLMExecutor
    detector *SensitiveDataDetector
}

func (e *PrivacyAwareLLMExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // 检查敏感数据
    sensitiveData := e.detector.DetectSensitiveData(context.Text)
    if len(sensitiveData) > 0 {
        log.Printf("检测到敏感数据：%v", sensitiveData)

        // 在发送到 LLM 之前脱敏敏感数据
        redactedText := e.detector.RedactSensitiveData(context.Text)
        redactedContext := &anyi.FlowContext{
            Text:   redactedText,
            Memory: context.Memory,
            Think:  context.Think,
            Images: context.Images,
        }

        return e.LLMExecutor.Execute(redactedContext)
    }

    return e.LLMExecutor.Execute(context)
}
```

### 数据加密

加密内存和存储中的敏感数据：

```go
type EncryptedFlowContext struct {
    encryptedText   []byte
    encryptedMemory []byte
    key             []byte
}

func NewEncryptedFlowContext(context *anyi.FlowContext, key []byte) (*EncryptedFlowContext, error) {
    // 加密文本
    encryptedText, err := encrypt([]byte(context.Text), key)
    if err != nil {
        return nil, err
    }

    // 加密内存
    memoryJSON, _ := json.Marshal(context.Memory)
    encryptedMemory, err := encrypt(memoryJSON, key)
    if err != nil {
        return nil, err
    }

    return &EncryptedFlowContext{
        encryptedText:   encryptedText,
        encryptedMemory: encryptedMemory,
        key:             key,
    }, nil
}

func (efc *EncryptedFlowContext) Decrypt() (*anyi.FlowContext, error) {
    // 解密文本
    textBytes, err := decrypt(efc.encryptedText, efc.key)
    if err != nil {
        return nil, err
    }

    // 解密内存
    memoryBytes, err := decrypt(efc.encryptedMemory, efc.key)
    if err != nil {
        return nil, err
    }

    var memory interface{}
    json.Unmarshal(memoryBytes, &memory)

    return &anyi.FlowContext{
        Text:   string(textBytes),
        Memory: memory,
    }, nil
}

func encrypt(data, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func decrypt(data, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, errors.New("密文过短")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}
```

## 访问控制和身份验证

### JWT 身份验证

实现基于 JWT 的身份验证：

```go
type AuthMiddleware struct {
    secretKey []byte
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
    return &AuthMiddleware{
        secretKey: []byte(secretKey),
    }
}

func (am *AuthMiddleware) ValidateToken(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("意外的签名方法：%v", token.Header["alg"])
        }
        return am.secretKey, nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, fmt.Errorf("无效令牌")
    }

    return token, nil
}

func (am *AuthMiddleware) AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "缺少授权头"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := am.ValidateToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "无效令牌"})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"])
            c.Set("user_role", claims["role"])
        }

        c.Next()
    }
}
```

### 基于角色的访问控制 (RBAC)

```go
type Permission string

const (
    PermissionReadChats   Permission = "read:chats"
    PermissionWriteChats  Permission = "write:chats"
    PermissionManageFlows Permission = "manage:flows"
    PermissionAdminAccess Permission = "admin:access"
)

type Role struct {
    Name        string
    Permissions []Permission
}

var (
    RoleUser = Role{
        Name:        "user",
        Permissions: []Permission{PermissionReadChats, PermissionWriteChats},
    }

    RoleAdmin = Role{
        Name: "admin",
        Permissions: []Permission{
            PermissionReadChats,
            PermissionWriteChats,
            PermissionManageFlows,
            PermissionAdminAccess,
        },
    }
)

func RequirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.JSON(403, gin.H{"error": "未找到用户角色"})
            c.Abort()
            return
        }

        roleStr := userRole.(string)
        var role Role

        switch roleStr {
        case "user":
            role = RoleUser
        case "admin":
            role = RoleAdmin
        default:
            c.JSON(403, gin.H{"error": "未知角色"})
            c.Abort()
            return
        }

        // 检查权限
        hasPermission := false
        for _, perm := range role.Permissions {
            if perm == permission {
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

## 速率限制

### 基于令牌桶的速率限制

```go
type RateLimiter struct {
    clients map[string]*TokenBucket
    mutex   sync.RWMutex
}

type TokenBucket struct {
    capacity    int
    tokens      int
    refillRate  int
    lastRefill  time.Time
    mutex       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        clients: make(map[string]*TokenBucket),
    }
}

func (rl *RateLimiter) Allow(clientID string, capacity, refillRate int) bool {
    rl.mutex.Lock()
    bucket, exists := rl.clients[clientID]
    if !exists {
        bucket = &TokenBucket{
            capacity:   capacity,
            tokens:     capacity,
            refillRate: refillRate,
            lastRefill: time.Now(),
        }
        rl.clients[clientID] = bucket
    }
    rl.mutex.Unlock()

    return bucket.consume()
}

func (tb *TokenBucket) consume() bool {
    tb.mutex.Lock()
    defer tb.mutex.Unlock()

    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)

    // 重新填充令牌
    tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
    tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
    tb.lastRefill = now

    if tb.tokens > 0 {
        tb.tokens--
        return true
    }

    return false
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientID := c.ClientIP()
        if userID, exists := c.Get("user_id"); exists {
            clientID = userID.(string)
        }

        if !limiter.Allow(clientID, 100, 10) { // 每秒 10 个请求，最大 100 个
            c.JSON(429, gin.H{
                "error": "速率限制",
                "message": "请求过于频繁，请稍后重试",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

## 审计和日志

### 安全事件日志

```go
type SecurityLogger struct {
    logger *log.Logger
    file   *os.File
}

func NewSecurityLogger(filename string) (*SecurityLogger, error) {
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
    if err != nil {
        return nil, err
    }

    logger := log.New(file, "", log.LstdFlags|log.Lmicroseconds)

    return &SecurityLogger{
        logger: logger,
        file:   file,
    }, nil
}

func (sl *SecurityLogger) LogSecurityEvent(event string, userID string, details map[string]interface{}) {
    eventData := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "event":     event,
        "user_id":   userID,
        "details":   details,
    }

    jsonData, _ := json.Marshal(eventData)
    sl.logger.Println(string(jsonData))
}

func (sl *SecurityLogger) Close() error {
    return sl.file.Close()
}

// 使用示例
func secureEndpointHandler(secLogger *SecurityLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, _ := c.Get("user_id")

        // 记录访问
        secLogger.LogSecurityEvent("api_access", userID.(string), map[string]interface{}{
            "endpoint": c.Request.URL.Path,
            "method":   c.Request.Method,
            "ip":       c.ClientIP(),
        })

        // 处理请求...
    }
}
```

## 网络安全

### HTTPS 和 TLS 配置

```go
func setupTLSServer() *http.Server {
    // TLS 配置
    tlsConfig := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }

    server := &http.Server{
        Addr:         ":8443",
        TLSConfig:    tlsConfig,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return server
}

func main() {
    r := gin.Default()

    // 添加安全头
    r.Use(func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    })

    server := setupTLSServer()
    server.Handler = r

    log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}
```

## 最佳实践总结

### 1. 输入安全

- 始终验证和清理用户输入
- 实现提示注入防护
- 使用内容过滤检测不当内容
- 设置输入长度限制

### 2. 输出安全

- 验证 AI 响应内容
- 清理输出中的潜在危险内容
- 永远不要直接执行 AI 生成的代码
- 实现输出长度和格式限制

### 3. 数据保护

- 检测和脱敏敏感数据
- 加密存储和传输中的敏感信息
- 实现数据保留策略
- 定期清理临时数据

### 4. 访问控制

- 实现强身份验证
- 使用基于角色的访问控制
- 实施速率限制
- 记录和监控访问模式

### 5. 基础设施安全

- 使用 HTTPS/TLS 进行所有通信
- 实施网络分段
- 定期更新依赖项
- 监控安全漏洞

### 6. 监控和审计

- 记录所有安全相关事件
- 实施实时监控和警报
- 定期进行安全审计
- 制定事件响应计划

通过遵循这些安全最佳实践，您可以构建安全、可靠的 AI 应用程序，保护用户数据和系统完整性。
