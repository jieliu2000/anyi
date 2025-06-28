# 配置参考

本文档提供配置 Anyi 应用程序的全面参考。涵盖所有配置选项、文件格式和管理设置的最佳实践。

## 配置方法

Anyi 支持多种配置应用程序的方式：

1. **程序化配置**：在代码中定义配置
2. **配置文件**：使用 YAML、JSON 或 TOML 文件
3. **环境变量**：使用环境变量覆盖设置
4. **混合方法**：结合多种方法以获得灵活性

## 配置结构

### 根配置

根配置结构包含客户端和流程：

```go
type AnyiConfig struct {
    Clients []llm.ClientConfig `yaml:"clients" json:"clients" toml:"clients"`
    Flows   []FlowConfig       `yaml:"flows" json:"flows" toml:"flows"`
}
```

## 客户端配置

### ClientConfig 结构

```go
type ClientConfig struct {
    Name   string                 `yaml:"name" json:"name" toml:"name"`
    Type   string                 `yaml:"type" json:"type" toml:"type"`
    Config map[string]interface{} `yaml:"config" json:"config" toml:"config"`
}
```

**字段：**

- `Name`: 客户端的唯一标识符
- `Type`: 提供商类型（openai、anthropic、ollama 等）
- `Config`: 提供商特定的配置选项

### 提供商特定配置

#### OpenAI 配置

```yaml
clients:
  - name: "openai-gpt4"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"
      baseURL: "https://api.openai.com/v1" # 可选
      orgID: "$OPENAI_ORG_ID" # 可选
      temperature: 0.7 # 可选
      maxTokens: 2000 # 可选
```

**配置选项：**

- `apiKey`（必需）：OpenAI API 密钥
- `model`（必需）：模型名称（gpt-4、gpt-3.5-turbo 等）
- `baseURL`（可选）：自定义 API 端点
- `orgID`（可选）：组织 ID
- `temperature`（可选）：默认温度（0.0-2.0）
- `maxTokens`（可选）：默认最大令牌数

#### Anthropic 配置

```yaml
clients:
  - name: "claude"
    type: "anthropic"
    config:
      apiKey: "$ANTHROPIC_API_KEY"
      model: "claude-3-opus-20240229"
      baseURL: "https://api.anthropic.com" # 可选
      version: "2023-06-01" # 可选
      temperature: 0.5 # 可选
      maxTokens: 1000 # 可选
```

**配置选项：**

- `apiKey`（必需）：Anthropic API 密钥
- `model`（必需）：模型名称
- `baseURL`（可选）：自定义 API 端点
- `version`（可选）：API 版本
- `temperature`（可选）：默认温度
- `maxTokens`（可选）：默认最大令牌数

#### Azure OpenAI 配置

```yaml
clients:
  - name: "azure-openai"
    type: "azure"
    config:
      apiKey: "$AZURE_OPENAI_API_KEY"
      endpoint: "$AZURE_OPENAI_ENDPOINT"
      deploymentName: "gpt-4-deployment"
      apiVersion: "2023-12-01-preview" # 可选
      temperature: 0.7 # 可选
      maxTokens: 2000 # 可选
```

**配置选项：**

- `apiKey`（必需）：Azure OpenAI API 密钥
- `endpoint`（必需）：Azure OpenAI 端点 URL
- `deploymentName`（必需）：部署名称
- `apiVersion`（可选）：API 版本
- `temperature`（可选）：默认温度
- `maxTokens`（可选）：默认最大令牌数

#### Ollama 配置

```yaml
clients:
  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434" # 可选
      options: # 可选
        temperature: 0.8
        top_p: 0.9
        top_k: 40
```

**配置选项：**

- `model`（必需）：Ollama 模型名称
- `baseURL`（可选）：Ollama 服务器 URL
- `options`（可选）：模型特定选项

#### 智谱 AI 配置

```yaml
clients:
  - name: "zhipu"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY"
      model: "glm-4"
      baseURL: "https://open.bigmodel.cn/api/paas/v4" # 可选
      temperature: 0.7 # 可选
      maxTokens: 1000 # 可选
```

#### 通义千问配置

```yaml
clients:
  - name: "dashscope"
    type: "dashscope"
    config:
      apiKey: "$DASHSCOPE_API_KEY"
      model: "qwen-turbo"
      baseURL: "https://dashscope.aliyuncs.com/api/v1" # 可选
      temperature: 0.7 # 可选
      maxTokens: 1500 # 可选
```

#### DeepSeek 配置

```yaml
clients:
  - name: "deepseek"
    type: "deepseek"
    config:
      apiKey: "$DEEPSEEK_API_KEY"
      model: "deepseek-chat"
      baseURL: "https://api.deepseek.com/v1" # 可选
      temperature: 0.7 # 可选
      maxTokens: 2000 # 可选
```

#### SiliconCloud 配置

```yaml
clients:
  - name: "siliconcloud"
    type: "siliconcloud"
    config:
      apiKey: "$SILICONCLOUD_API_KEY"
      model: "meta-llama/Llama-2-7b-chat-hf"
      baseURL: "https://api.siliconflow.cn/v1" # 可选
      temperature: 0.7 # 可选
      maxTokens: 1000 # 可选
```

## 流程配置

### FlowConfig 结构

```go
type FlowConfig struct {
    Name       string       `yaml:"name" json:"name" toml:"name"`
    ClientName string       `yaml:"clientName,omitempty" json:"clientName,omitempty" toml:"clientName,omitempty"`
    Steps      []StepConfig `yaml:"steps" json:"steps" toml:"steps"`
}
```

**字段：**

- `Name`: 流程的唯一标识符
- `ClientName`: 所有步骤使用的默认客户端
- `Steps`: 步骤配置数组

### StepConfig 结构

```go
type StepConfig struct {
    Name          string           `yaml:"name" json:"name" toml:"name"`
    ClientName    string           `yaml:"clientName,omitempty" json:"clientName,omitempty" toml:"clientName,omitempty"`
    Executor      *ExecutorConfig  `yaml:"executor,omitempty" json:"executor,omitempty" toml:"executor,omitempty"`
    Validator     *ValidatorConfig `yaml:"validator,omitempty" json:"validator,omitempty" toml:"validator,omitempty"`
    MaxRetryTimes int              `yaml:"maxRetryTimes,omitempty" json:"maxRetryTimes,omitempty" toml:"maxRetryTimes,omitempty"`
}
```

**字段：**

- `Name`: 步骤标识符
- `ClientName`: 此步骤使用的客户端（覆盖流程默认值）
- `Executor`: 执行器配置
- `Validator`: 验证器配置
- `MaxRetryTimes`: 最大重试次数

### ExecutorConfig 结构

```go
type ExecutorConfig struct {
    Type       string                 `yaml:"type" json:"type" toml:"type"`
    WithConfig map[string]interface{} `yaml:"withconfig,omitempty" json:"withconfig,omitempty" toml:"withconfig,omitempty"`
}
```

**执行器类型：**

- `llm`: 用于 AI 处理的 LLM 执行器
- `setcontext`: 上下文操作
- `conditional`: 条件分支
- `command`: Shell 命令执行

#### LLM 执行器配置

```yaml
executor:
  type: "llm"
  withconfig:
    template: "分析以下文本：{{.Text}}"
    systemMessage: "你是一个专业分析师。"
    temperature: 0.7
    maxTokens: 1000
    extractThink: true
```

#### SetContext 执行器配置

```yaml
executor:
  type: "setcontext"
  withconfig:
    text: "新的文本内容"
    memory:
      key1: "value1"
      key2: "value2"
    think: "初始思考过程"
    images:
      - "https://example.com/image1.jpg"
```

#### 条件执行器配置

```yaml
executor:
  type: "conditional"
  withconfig:
    condition: '{{contains .Text "错误"}}'
    trueFlow: "错误处理器"
    falseFlow: "正常处理器"
```

#### 命令执行器配置

```yaml
executor:
  type: "command"
  withconfig:
    command: "python"
    args:
      - "script.py"
      - "{{.Text}}"
    workingDir: "/path/to/scripts"
    timeout: 30
```

## 验证器配置

### ValidatorConfig 结构

```go
type ValidatorConfig struct {
    Type       string                 `yaml:"type" json:"type" toml:"type"`
    WithConfig map[string]interface{} `yaml:"withconfig,omitempty" json:"withconfig,omitempty" toml:"withconfig,omitempty"`
}
```

**验证器类型：**

- `string`: 字符串验证
- `json`: JSON 结构验证
- `regex`: 正则表达式验证

#### 字符串验证器配置

```yaml
validator:
  type: "string"
  withconfig:
    minLength: 100
    maxLength: 2000
    contains: "必需短语"
    notContains: "禁止词汇"
    startsWith: "摘要："
    endsWith: "。"
```

#### JSON 验证器配置

```yaml
validator:
  type: "json"
  withconfig:
    requiredFields:
      - "title"
      - "content"
      - "summary"
    schema: |
      {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "content": {"type": "string"},
          "summary": {"type": "string"}
        }
      }
```

## 环境变量

### 支持的环境变量

Anyi 支持以下环境变量来覆盖配置：

```bash
# OpenAI
OPENAI_API_KEY=your-openai-api-key
OPENAI_ORG_ID=your-org-id
OPENAI_BASE_URL=https://api.openai.com/v1

# Anthropic
ANTHROPIC_API_KEY=your-anthropic-api-key
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_VERSION=2023-06-01

# Azure OpenAI
AZURE_OPENAI_API_KEY=your-azure-api-key
AZURE_OPENAI_ENDPOINT=your-azure-endpoint
AZURE_OPENAI_API_VERSION=2023-12-01-preview

# 智谱 AI
ZHIPU_API_KEY=your-zhipu-api-key

# 通义千问
DASHSCOPE_API_KEY=your-dashscope-api-key

# DeepSeek
DEEPSEEK_API_KEY=your-deepseek-api-key

# SiliconCloud
SILICONCLOUD_API_KEY=your-siliconcloud-api-key

# Ollama
OLLAMA_BASE_URL=http://localhost:11434
```

### 环境变量优先级

环境变量的优先级从高到低：

1. 运行时设置的环境变量
2. `.env` 文件中的变量
3. 配置文件中的值
4. 默认值

## 配置文件格式

### YAML 格式

```yaml
# config.yaml
clients:
  - name: "openai-gpt4"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"
      temperature: 0.7

  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"

flows:
  - name: "文本分析"
    clientName: "openai-gpt4"
    steps:
      - name: "分析步骤"
        executor:
          type: "llm"
          withconfig:
            template: "分析这个文本：{{.Text}}"
            maxTokens: 1000
        validator:
          type: "string"
          withconfig:
            minLength: 50
```

### JSON 格式

```json
{
  "clients": [
    {
      "name": "openai-gpt4",
      "type": "openai",
      "config": {
        "apiKey": "$OPENAI_API_KEY",
        "model": "gpt-4",
        "temperature": 0.7
      }
    }
  ],
  "flows": [
    {
      "name": "文本分析",
      "clientName": "openai-gpt4",
      "steps": [
        {
          "name": "分析步骤",
          "executor": {
            "type": "llm",
            "withconfig": {
              "template": "分析这个文本：{{.Text}}",
              "maxTokens": 1000
            }
          }
        }
      ]
    }
  ]
}
```

### TOML 格式

```toml
# config.toml
[[clients]]
name = "openai-gpt4"
type = "openai"

[clients.config]
apiKey = "$OPENAI_API_KEY"
model = "gpt-4"
temperature = 0.7

[[flows]]
name = "文本分析"
clientName = "openai-gpt4"

[[flows.steps]]
name = "分析步骤"

[flows.steps.executor]
type = "llm"

[flows.steps.executor.withconfig]
template = "分析这个文本：{{.Text}}"
maxTokens = 1000
```

## 配置最佳实践

### 1. 安全性

```yaml
# 使用环境变量存储敏感信息
clients:
  - name: "secure-client"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY" # 永远不要硬编码 API 密钥
      orgID: "$OPENAI_ORG_ID"
```

### 2. 环境分离

```bash
# 开发环境
export ANYI_CONFIG_FILE=config.dev.yaml

# 生产环境
export ANYI_CONFIG_FILE=config.prod.yaml
```

### 3. 配置验证

```go
func validateConfig(config *AnyiConfig) error {
    if len(config.Clients) == 0 {
        return errors.New("至少需要配置一个客户端")
    }

    for _, client := range config.Clients {
        if client.Name == "" {
            return errors.New("客户端名称不能为空")
        }
        if client.Type == "" {
            return errors.New("客户端类型不能为空")
        }
    }

    return nil
}
```

### 4. 动态配置重载

```go
func watchConfigFile(filename string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    err = watcher.Add(filename)
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Println("配置文件已修改，重新加载...")
                err := anyi.ConfigFromFile(filename)
                if err != nil {
                    log.Printf("重新加载配置失败：%v", err)
                } else {
                    log.Println("配置重新加载成功")
                }
            }
        case err := <-watcher.Errors:
            log.Printf("配置文件监控错误：%v", err)
        }
    }
}
```

## 故障排除

### 常见配置错误

1. **API 密钥错误**

   - 检查环境变量是否正确设置
   - 验证 API 密钥格式
   - 确认密钥权限

2. **模型名称错误**

   - 检查提供商支持的模型列表
   - 验证模型名称拼写
   - 确认模型可用性

3. **网络连接问题**

   - 检查防火墙设置
   - 验证代理配置
   - 测试网络连接

4. **配置文件格式错误**
   - 验证 YAML/JSON/TOML 语法
   - 检查缩进和引号
   - 使用配置验证工具

### 调试配置

```go
func debugConfig() {
    // 打印所有已注册的客户端
    clients := anyi.ListClients()
    log.Printf("已注册的客户端：%v", clients)

    // 测试客户端连接
    for _, clientName := range clients {
        client, err := anyi.GetClient(clientName)
        if err != nil {
            log.Printf("获取客户端 %s 失败：%v", clientName, err)
            continue
        }

        // 发送测试消息
        testMessage := []chat.Message{
            {Role: "user", Content: "Hello"},
        }

        _, _, err = client.Chat(testMessage, nil)
        if err != nil {
            log.Printf("客户端 %s 测试失败：%v", clientName, err)
        } else {
            log.Printf("客户端 %s 测试成功", clientName)
        }
    }
}
```
