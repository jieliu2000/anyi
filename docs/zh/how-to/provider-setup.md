# 提供商设置指南

本指南将详细介绍如何设置和配置各个 LLM 提供商，包括获取 API 密钥、配置客户端和故障排除。

## OpenAI 设置

### 1. 获取 API 密钥

1. 访问 [OpenAI 平台](https://platform.openai.com/)
2. 注册账户并登录
3. 进入 **API Keys** 页面
4. 点击 **Create new secret key**
5. 复制生成的密钥（以 `sk-` 开头）

### 2. 配置 OpenAI 客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/openai"
)

func setupOpenAI() {
    config := openai.Config{
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   2000,
        BaseURL:     "https://api.openai.com/v1", // 默认值
        Timeout:     30 * time.Second,
    }

    client, err := anyi.NewClient("openai", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 环境变量设置

```bash
# .env
OPENAI_API_KEY=sk-your-openai-api-key-here
OPENAI_MODEL=gpt-4
OPENAI_TEMPERATURE=0.7
OPENAI_MAX_TOKENS=2000
```

### 4. 可用模型

| 模型        | 描述           | 上下文长度 |
| ----------- | -------------- | ---------- |
| gpt-4o      | 最新多模态模型 | 128K       |
| gpt-4-turbo | 高性能模型     | 128K       |
| gpt-4       | 标准 GPT-4     | 8K         |
| gpt-4o-mini | 快速且经济     | 128K       |

## Anthropic Claude 设置

### 1. 获取 API 密钥

1. 访问 [Anthropic Console](https://console.anthropic.com/)
2. 注册账户并登录
3. 进入 **API Keys** 页面
4. 点击 **Create Key**
5. 复制生成的密钥（以 `sk-ant-` 开头）

### 2. 配置 Claude 客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/anthropic"
)

func setupClaude() {
    config := anthropic.Config{
        APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
        Model:       "claude-3-5-sonnet-20241022",
        MaxTokens:   1000,
        Temperature: 0.7,
        BaseURL:     "https://api.anthropic.com", // 默认值
    }

    client, err := anyi.NewClient("claude", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 环境变量设置

```bash
# .env
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key-here
ANTHROPIC_MODEL=claude-3-5-sonnet-20241022
```

### 4. 可用模型

| 模型                       | 描述             | 上下文长度 |
| -------------------------- | ---------------- | ---------- |
| claude-3-5-sonnet-20241022 | 最新 Sonnet 模型 | 200K       |
| claude-3-opus-20240229     | 最强性能模型     | 200K       |
| claude-3-haiku-20240307    | 快速模型         | 200K       |

## Azure OpenAI 设置

### 1. 创建 Azure OpenAI 资源

1. 登录 [Azure 门户](https://portal.azure.com/)
2. 创建新的 **Azure OpenAI** 资源
3. 等待资源部署完成
4. 进入资源页面，获取 **Endpoint** 和 **Key**

### 2. 部署模型

1. 在 Azure OpenAI Studio 中
2. 进入 **Deployments** 页面
3. 点击 **Create new deployment**
4. 选择模型（如 gpt-4）
5. 设置部署名称

### 3. 配置 Azure OpenAI 客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/azureopenai"
)

func setupAzureOpenAI() {
    config := azureopenai.Config{
        APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
        Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
        Model:      "gpt-4", // 您的部署名称
        APIVersion: "2024-02-15-preview",
    }

    client, err := anyi.NewClient("azure-openai", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. 环境变量设置

```bash
# .env
AZURE_OPENAI_API_KEY=your-azure-openai-key-here
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_MODEL=gpt-4
AZURE_OPENAI_API_VERSION=2024-02-15-preview
```

## Ollama 本地设置

### 1. 安装 Ollama

**macOS:**

```bash
brew install ollama
```

**Linux:**

```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
下载并安装 [Ollama for Windows](https://ollama.ai/download/windows)

### 2. 启动服务并下载模型

```bash
# 启动 Ollama 服务
ollama serve

# 在另一个终端中下载模型
ollama pull llama3
ollama pull codellama
ollama pull mistral
ollama pull llava  # 支持视觉的模型
```

### 3. 配置 Ollama 客户端

```go
package main

import (
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/ollama"
)

func setupOllama() {
    config := ollama.Config{
        Model:       "llama3",
        BaseURL:     "http://localhost:11434", // 默认值
        Temperature: 0.7,
    }

    client, err := anyi.NewClient("ollama", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. 可用模型

| 模型      | 大小  | 描述            |
| --------- | ----- | --------------- |
| llama3    | 4.7GB | Meta 的最新模型 |
| codellama | 3.8GB | 代码生成专用    |
| mistral   | 4.1GB | 高效的开源模型  |
| llava     | 4.5GB | 支持视觉的模型  |

## 智谱 AI 设置

### 1. 获取 API 密钥

1. 访问 [智谱 AI 开放平台](https://open.bigmodel.cn/)
2. 注册账户并完成实名认证
3. 进入 **API 密钥** 页面
4. 创建新的 API 密钥

### 2. 配置智谱客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/zhipu"
)

func setupZhipu() {
    config := zhipu.Config{
        APIKey:      os.Getenv("ZHIPU_API_KEY"),
        Model:       "glm-4-flash-250414",
        Temperature: 0.7,
        MaxTokens:   1000,
    }

    client, err := anyi.NewClient("zhipu", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 环境变量设置

```bash
# .env
ZHIPU_API_KEY=your-zhipu-api-key-here
ZHIPU_MODEL=glm-4-flash-250414
```

### 4. 可用模型

| 模型        | 描述         |
| ----------- | ------------ |
| glm-4       | 最新一代模型 |
| glm-3-turbo | 快速响应模型 |
| chatglm3-6b | 开源模型     |

## 通义千问设置

### 1. 获取 API 密钥

1. 访问 [阿里云控制台](https://dashscope.console.aliyun.com/)
2. 开通 DashScope 服务
3. 创建 API 密钥

### 2. 配置通义千问客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/dashscope"
)

func setupQwen() {
    config := dashscope.Config{
        APIKey:      os.Getenv("DASHSCOPE_API_KEY"),
        Model:       "qwen-max",
        Temperature: 0.7,
    }

    client, err := anyi.NewClient("qwen", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 环境变量设置

```bash
# .env
DASHSCOPE_API_KEY=your-dashscope-api-key-here
DASHSCOPE_MODEL=qwen-max
```

### 4. 可用模型

| 模型       | 描述           |
| ---------- | -------------- |
| qwen-max   | 最强性能模型   |
| qwen-plus  | 平衡性能和成本 |
| qwen-turbo | 快速响应模型   |

## DeepSeek 设置

### 1. 获取 API 密钥

1. 访问 [DeepSeek 平台](https://platform.deepseek.com/)
2. 注册账户并登录
3. 创建 API 密钥

### 2. 配置 DeepSeek 客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/deepseek"
)

func setupDeepSeek() {
    config := deepseek.Config{
        APIKey:      os.Getenv("DEEPSEEK_API_KEY"),
        Model:       "deepseek-chat",
        Temperature: 0.7,
    }

    client, err := anyi.NewClient("deepseek", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 环境变量设置

```bash
# .env
DEEPSEEK_API_KEY=your-deepseek-api-key-here
DEEPSEEK_MODEL=deepseek-chat
```

## SiliconCloud 设置

### 1. 获取 API 密钥

1. 访问 [SiliconCloud 平台](https://siliconflow.cn/)
2. 注册账户并登录
3. 创建 API 密钥

### 2. 配置 SiliconCloud 客户端

```go
package main

import (
    "os"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/siliconcloud"
)

func setupSiliconCloud() {
    config := siliconcloud.Config{
        APIKey:      os.Getenv("SILICONCLOUD_API_KEY"),
        Model:       "deepseek-ai/deepseek-chat",
        Temperature: 0.7,
    }

    client, err := anyi.NewClient("siliconcloud", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 多提供商配置

### 统一配置文件

```yaml
# providers.yaml
clients:
  # 云端提供商
  - name: "openai-gpt4"
    type: "openai"
    config:
      apiKey: "$OPENAI_API_KEY"
      model: "gpt-4"
      temperature: 0.7

  - name: "claude-sonnet"
    type: "anthropic"
    config:
      apiKey: "$ANTHROPIC_API_KEY"
      model: "claude-3-5-sonnet-20241022"

  - name: "azure-gpt4"
    type: "azureopenai"
    config:
      apiKey: "$AZURE_OPENAI_API_KEY"
      endpoint: "$AZURE_OPENAI_ENDPOINT"
      model: "gpt-4"

  # 中文提供商
  - name: "zhipu-glm4"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY"
      model: "glm-4-flash-250414"

  - name: "qwen-max"
    type: "dashscope"
    config:
      apiKey: "$DASHSCOPE_API_KEY"
      model: "qwen-max"

  - name: "deepseek-chat"
    type: "deepseek"
    config:
      apiKey: "$DEEPSEEK_API_KEY"
      model: "deepseek-chat"

  # 本地提供商
  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"
```

### 加载多提供商配置

```go
package main

import (
    "log"
    "github.com/jieliu2000/anyi"
)

func main() {
    // 加载所有提供商配置
    err := anyi.ConfigFromFile("providers.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 测试所有客户端
    testClients()
}

func testClients() {
    clientNames := []string{
        "openai-gpt4",
        "claude-sonnet",
        "zhipu-glm4",
        "qwen-max",
        "local-llama",
    }

    for _, name := range clientNames {
        client, err := anyi.GetClient(name)
        if err != nil {
            log.Printf("❌ %s: %v", name, err)
            continue
        }

        // 简单测试
        messages := []chat.Message{
            {Role: "user", Content: "你好"},
        }

        _, _, err = client.Chat(messages, nil)
        if err != nil {
            log.Printf("❌ %s: %v", name, err)
        } else {
            log.Printf("✅ %s: 连接成功", name)
        }
    }
}
```

## 故障排除

### 常见问题

#### 1. API 密钥错误

**错误信息：**

```
401 Unauthorized
```

**解决方案：**

- 检查 API 密钥是否正确
- 确认密钥没有过期
- 验证密钥格式（OpenAI: `sk-`, Anthropic: `sk-ant-`）

#### 2. 网络连接问题

**错误信息：**

```
dial tcp: lookup api.openai.com: no such host
```

**解决方案：**

- 检查网络连接
- 使用代理（如果在受限网络环境）
- 尝试使用本地模型（Ollama）

#### 3. 模型不存在

**错误信息：**

```
model 'gpt-5' not found
```

**解决方案：**

- 使用正确的模型名称
- 查看提供商文档了解可用模型
- 对于 Azure OpenAI，使用部署名称而非模型名称

#### 4. 速率限制

**错误信息：**

```
429 Too Many Requests
```

**解决方案：**

- 实施重试机制
- 添加请求间隔
- 升级到更高的配额计划

### 调试技巧

#### 1. 启用详细日志

```go
config := openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Debug:  true, // 启用调试模式
}
```

#### 2. 测试连接

```go
func testConnection(client anyi.Client) error {
    messages := []chat.Message{
        {Role: "user", Content: "test"},
    }

    _, _, err := client.Chat(messages, nil)
    return err
}
```

#### 3. 检查配置

```go
func validateConfig() {
    requiredEnvs := []string{
        "OPENAI_API_KEY",
        "ANTHROPIC_API_KEY",
        "ZHIPU_API_KEY",
    }

    for _, env := range requiredEnvs {
        if os.Getenv(env) == "" {
            log.Printf("警告: %s 环境变量未设置", env)
        }
    }
}
```

## 最佳实践

### 1. 安全性

- 使用环境变量存储 API 密钥
- 不要在代码中硬编码密钥
- 定期轮换 API 密钥
- 限制 API 密钥的权限

### 2. 性能优化

- 为不同任务选择合适的模型
- 使用连接池复用连接
- 实施请求缓存
- 监控 API 使用情况

### 3. 成本控制

- 设置使用限额
- 监控 token 使用量
- 选择性价比高的模型
- 优化提示词长度

### 4. 可靠性

- 实施重试机制
- 使用多个提供商作为备份
- 监控服务可用性
- 实施降级策略

## 下一步

现在您已经成功设置了各个提供商，可以：

1. 学习 [错误处理](error-handling.md) 来构建更健壮的应用
2. 了解 [性能优化](performance.md) 来提升应用性能
3. 探索 [Web 集成](web-integration.md) 来构建 Web 应用
4. 查看 [安全最佳实践](../advanced/security.md) 来保护您的应用
