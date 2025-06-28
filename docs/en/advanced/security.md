# Security Best Practices

This document provides comprehensive security guidelines for building and deploying AI applications with the Anyi framework.

## Overview

Security in AI applications involves multiple layers:

- **API Security**: Protecting LLM API keys and endpoints
- **Input Validation**: Preventing prompt injection and malicious inputs
- **Output Sanitization**: Ensuring AI outputs are safe to use
- **Data Privacy**: Protecting sensitive information
- **Infrastructure Security**: Securing deployment environments
- **Access Control**: Managing user permissions and authentication

## API Key Management

### Environment Variables

Never hardcode API keys in your source code. Use environment variables:

```go
// ❌ Bad: Hardcoded API key
config := openai.DefaultConfig("sk-1234567890abcdef...")

// ✅ Good: Environment variable
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
```

### Configuration Files

Use environment variable substitution in configuration files:

```yaml
# config.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY" # Substituted at runtime
      model: "gpt-4"
```

### Secrets Management

For production environments, use dedicated secrets management:

```go
// Example with AWS Secrets Manager
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

### Key Rotation

Implement regular API key rotation:

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

    // Check if error indicates invalid API key
    if isInvalidKeyError(err) {
        rc.rotateKey()
        // Retry with new key
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
        log.Printf("Failed to rotate key: %v", err)
        return
    }

    newConfig := openai.DefaultConfig(newKey)
    newClient, err := anyi.NewClient("openai-rotated", newConfig)
    if err != nil {
        log.Printf("Failed to create new client: %v", err)
        return
    }

    rc.currentClient = newClient
}
```

## Input Validation and Sanitization

### Prompt Injection Prevention

Protect against prompt injection attacks:

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
            "ignore previous instructions",
            "disregard the above",
            "forget everything",
            "new instructions:",
            "system:",
            "assistant:",
        },
        allowedChars: regexp.MustCompile(`^[a-zA-Z0-9\s\.,!?;:()\-'"]+$`),
    }
}

func (spb *SecurePromptBuilder) ValidateInput(input string) error {
    // Check length
    if len(input) > spb.maxLength {
        return fmt.Errorf("input too long: %d characters (max: %d)", len(input), spb.maxLength)
    }

    // Check for blocked phrases
    inputLower := strings.ToLower(input)
    for _, phrase := range spb.blockedPhrases {
        if strings.Contains(inputLower, phrase) {
            return fmt.Errorf("input contains blocked phrase: %s", phrase)
        }
    }

    // Check character allowlist
    if !spb.allowedChars.MatchString(input) {
        return fmt.Errorf("input contains invalid characters")
    }

    return nil
}

func (spb *SecurePromptBuilder) SanitizeInput(input string) string {
    // Remove potential injection markers
    input = strings.ReplaceAll(input, "```", "")
    input = strings.ReplaceAll(input, "---", "")

    // Escape special characters
    input = html.EscapeString(input)

    // Limit length
    if len(input) > spb.maxLength {
        input = input[:spb.maxLength]
    }

    return input
}

// Usage in executor
type SecureLLMExecutor struct {
    *anyi.LLMExecutor
    promptBuilder *SecurePromptBuilder
}

func (e *SecureLLMExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Validate and sanitize input
    if err := e.promptBuilder.ValidateInput(context.Text); err != nil {
        return nil, fmt.Errorf("input validation failed: %v", err)
    }

    sanitizedText := e.promptBuilder.SanitizeInput(context.Text)

    // Create sanitized context
    sanitizedContext := &anyi.FlowContext{
        Text:   sanitizedText,
        Memory: context.Memory,
        Think:  context.Think,
        Images: context.Images,
    }

    return e.LLMExecutor.Execute(sanitizedContext)
}
````

### Content Filtering

Implement content filtering for inappropriate inputs:

```go
type ContentFilter struct {
    profanityList []string
    toxicityAPI   string
}

func (cf *ContentFilter) FilterContent(text string) (string, error) {
    // Check for profanity
    textLower := strings.ToLower(text)
    for _, word := range cf.profanityList {
        if strings.Contains(textLower, word) {
            return "", fmt.Errorf("content contains inappropriate language")
        }
    }

    // Check toxicity using external API
    if cf.toxicityAPI != "" {
        toxic, err := cf.checkToxicity(text)
        if err != nil {
            log.Printf("Toxicity check failed: %v", err)
            // Continue processing if toxicity check fails
        } else if toxic {
            return "", fmt.Errorf("content flagged as toxic")
        }
    }

    return text, nil
}

func (cf *ContentFilter) checkToxicity(text string) (bool, error) {
    // Implementation for external toxicity API
    // This is a placeholder - use actual service like Perspective API
    return false, nil
}
```

## Output Validation and Sanitization

### Response Validation

Validate AI responses before using them:

```go
type ResponseValidator struct {
    maxLength      int
    requiredFields []string
    forbiddenTerms []string
}

func (rv *ResponseValidator) ValidateResponse(response string) error {
    // Check length
    if len(response) > rv.maxLength {
        return fmt.Errorf("response too long")
    }

    // Check for forbidden terms
    responseLower := strings.ToLower(response)
    for _, term := range rv.forbiddenTerms {
        if strings.Contains(responseLower, term) {
            return fmt.Errorf("response contains forbidden term: %s", term)
        }
    }

    // Additional validation logic...
    return nil
}

func (rv *ResponseValidator) SanitizeResponse(response string) string {
    // Remove potential script tags
    response = regexp.MustCompile(`<script[^>]*>.*?</script>`).ReplaceAllString(response, "")

    // Remove other potentially dangerous HTML
    response = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(response, "")

    // Escape remaining HTML entities
    response = html.EscapeString(response)

    return response
}
```

### Code Execution Prevention

Never execute AI-generated code directly:

```go
// ❌ Dangerous: Direct execution
func dangerousExecuteCode(code string) error {
    return exec.Command("sh", "-c", code).Run()
}

// ✅ Safe: Sandboxed execution with validation
type SafeCodeExecutor struct {
    allowedCommands map[string]bool
    timeout         time.Duration
    workingDir      string
}

func (sce *SafeCodeExecutor) ExecuteCode(code string) (string, error) {
    // Parse and validate code
    if err := sce.validateCode(code); err != nil {
        return "", err
    }

    // Create temporary file
    tmpFile, err := os.CreateTemp(sce.workingDir, "safe_code_*.py")
    if err != nil {
        return "", err
    }
    defer os.Remove(tmpFile.Name())

    // Write code to file
    if _, err := tmpFile.WriteString(code); err != nil {
        return "", err
    }
    tmpFile.Close()

    // Execute in sandbox
    ctx, cancel := context.WithTimeout(context.Background(), sce.timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, "python", tmpFile.Name())
    cmd.Dir = sce.workingDir

    output, err := cmd.Output()
    return string(output), err
}

func (sce *SafeCodeExecutor) validateCode(code string) error {
    // Check for dangerous operations
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
            return fmt.Errorf("code contains dangerous operation: %s", danger)
        }
    }

    return nil
}
```

## Data Privacy and Protection

### Sensitive Data Detection

Detect and handle sensitive information:

```go
type SensitiveDataDetector struct {
    patterns map[string]*regexp.Regexp
}

func NewSensitiveDataDetector() *SensitiveDataDetector {
    return &SensitiveDataDetector{
        patterns: map[string]*regexp.Regexp{
            "ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
            "credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
            "email":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
            "phone":       regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`),
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
        text = pattern.ReplaceAllString(text, "[REDACTED]")
    }
    return text
}

// Usage in workflow
type PrivacyAwareLLMExecutor struct {
    *anyi.LLMExecutor
    detector *SensitiveDataDetector
}

func (e *PrivacyAwareLLMExecutor) Execute(context *anyi.FlowContext) (*anyi.FlowContext, error) {
    // Check for sensitive data
    sensitiveData := e.detector.DetectSensitiveData(context.Text)
    if len(sensitiveData) > 0 {
        log.Printf("Sensitive data detected: %v", sensitiveData)

        // Redact sensitive data before sending to LLM
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

### Data Encryption

Encrypt sensitive data in memory and storage:

```go
type EncryptedFlowContext struct {
    encryptedText   []byte
    encryptedMemory []byte
    key             []byte
}

func NewEncryptedFlowContext(context *anyi.FlowContext, key []byte) (*EncryptedFlowContext, error) {
    // Encrypt text
    encryptedText, err := encrypt([]byte(context.Text), key)
    if err != nil {
        return nil, err
    }

    // Encrypt memory
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
    // Decrypt text
    textBytes, err := decrypt(efc.encryptedText, efc.key)
    if err != nil {
        return nil, err
    }

    // Decrypt memory
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
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
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
        return nil, errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

## Access Control and Authentication

### Rate Limiting

Implement rate limiting to prevent abuse:

```go
type RateLimiter struct {
    requests map[string][]time.Time
    limit    int
    window   time.Duration
    mutex    sync.RWMutex
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (rl *RateLimiter) Allow(userID string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()

    // Clean old requests
    requests := rl.requests[userID]
    validRequests := make([]time.Time, 0)

    for _, req := range requests {
        if now.Sub(req) < rl.window {
            validRequests = append(validRequests, req)
        }
    }

    // Check if limit exceeded
    if len(validRequests) >= rl.limit {
        return false
    }

    // Add current request
    validRequests = append(validRequests, now)
    rl.requests[userID] = validRequests

    return true
}

// Usage in HTTP handler
func chatHandler(w http.ResponseWriter, r *http.Request, rateLimiter *RateLimiter) {
    userID := getUserID(r) // Extract user ID from request

    if !rateLimiter.Allow(userID) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }

    // Process request...
}
```

### Authentication Middleware

Implement authentication for API endpoints:

```go
type AuthMiddleware struct {
    jwtSecret []byte
}

func (am *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // Remove "Bearer " prefix
        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = tokenString[7:]
        }

        // Validate JWT token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return am.jwtSecret, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Extract user information from token
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        // Add user context to request
        ctx := context.WithValue(r.Context(), "userID", claims["userID"])
        r = r.WithContext(ctx)

        next(w, r)
    }
}
```

## Logging and Monitoring

### Security Event Logging

Log security-relevant events:

```go
type SecurityLogger struct {
    logger *log.Logger
}

func NewSecurityLogger() *SecurityLogger {
    return &SecurityLogger{
        logger: log.New(os.Stdout, "[SECURITY] ", log.LstdFlags),
    }
}

func (sl *SecurityLogger) LogAuthFailure(userID, reason string) {
    sl.logger.Printf("AUTH_FAILURE user=%s reason=%s", userID, reason)
}

func (sl *SecurityLogger) LogSensitiveDataDetected(userID string, dataTypes []string) {
    sl.logger.Printf("SENSITIVE_DATA_DETECTED user=%s types=%v", userID, dataTypes)
}

func (sl *SecurityLogger) LogRateLimitExceeded(userID string) {
    sl.logger.Printf("RATE_LIMIT_EXCEEDED user=%s", userID)
}

func (sl *SecurityLogger) LogPromptInjectionAttempt(userID, input string) {
    sl.logger.Printf("PROMPT_INJECTION_ATTEMPT user=%s input_hash=%x", userID, sha256.Sum256([]byte(input)))
}
```

### Audit Trail

Maintain an audit trail of all operations:

```go
type AuditTrail struct {
    db *sql.DB
}

type AuditEvent struct {
    Timestamp time.Time
    UserID    string
    Action    string
    Resource  string
    Success   bool
    Details   map[string]interface{}
}

func (at *AuditTrail) LogEvent(event AuditEvent) error {
    detailsJSON, _ := json.Marshal(event.Details)

    _, err := at.db.Exec(`
        INSERT INTO audit_log (timestamp, user_id, action, resource, success, details)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, event.Timestamp, event.UserID, event.Action, event.Resource, event.Success, detailsJSON)

    return err
}

// Usage in workflow
func auditedExecute(executor anyi.Executor, context *anyi.FlowContext, userID string, audit *AuditTrail) (*anyi.FlowContext, error) {
    start := time.Now()

    result, err := executor.Execute(context)

    // Log audit event
    event := AuditEvent{
        Timestamp: start,
        UserID:    userID,
        Action:    "EXECUTE_WORKFLOW",
        Resource:  "anyi_executor",
        Success:   err == nil,
        Details: map[string]interface{}{
            "execution_time": time.Since(start).Milliseconds(),
            "input_length":   len(context.Text),
        },
    }

    if err != nil {
        event.Details["error"] = err.Error()
    } else {
        event.Details["output_length"] = len(result.Text)
    }

    audit.LogEvent(event)

    return result, err
}
```

## Deployment Security

### Container Security

Use secure container practices:

```dockerfile
# Use minimal base image
FROM alpine:3.18

# Create non-root user
RUN addgroup -g 1001 anyi && \
    adduser -D -s /bin/sh -u 1001 -G anyi anyi

# Install only necessary packages
RUN apk add --no-cache ca-certificates

# Copy binary
COPY anyi /usr/local/bin/anyi
RUN chmod +x /usr/local/bin/anyi

# Switch to non-root user
USER anyi

# Set security headers
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/anyi"]
```

### Network Security

Configure secure networking:

```go
func createSecureServer() *http.Server {
    // TLS configuration
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

    // HTTP server configuration
    server := &http.Server{
        Addr:         ":8080",
        TLSConfig:    tlsConfig,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return server
}
```

## Security Checklist

### Development Phase

- [ ] API keys stored in environment variables
- [ ] Input validation implemented
- [ ] Output sanitization in place
- [ ] Sensitive data detection configured
- [ ] Rate limiting implemented
- [ ] Authentication/authorization configured
- [ ] Security logging enabled
- [ ] Error handling doesn't leak sensitive information

### Deployment Phase

- [ ] TLS/SSL configured
- [ ] Secrets management system in use
- [ ] Container security hardened
- [ ] Network security configured
- [ ] Monitoring and alerting set up
- [ ] Backup and recovery procedures tested
- [ ] Incident response plan documented

### Operational Phase

- [ ] Regular security audits performed
- [ ] API keys rotated regularly
- [ ] Dependencies updated regularly
- [ ] Security logs monitored
- [ ] Penetration testing conducted
- [ ] Staff security training completed

By following these security best practices, you can build robust and secure AI applications with Anyi that protect both your organization and your users' data.
