# 配置管理教程

本教程将介绍如何在 Anyi 中管理配置，包括配置文件、环境变量、动态配置和最佳实践。

## 配置概述

Anyi 支持多种配置方式：

1. **代码配置**：直接在代码中定义配置
2. **配置文件**：使用 YAML、JSON 或 TOML 文件
3. **环境变量**：通过环境变量动态配置
4. **混合配置**：结合多种方式

### 配置优先级

```
环境变量 > 配置文件 > 代码默认值
```

## 配置文件格式

### YAML 配置（推荐）

```yaml
# config.yaml
clients:
  - name: "zhipu-glm4"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY"
      model: "glm-4-flash-250414"
      temperature: 0.7
      maxTokens: 2000

  - name: "claude-sonnet"
    type: "anthropic"
    config:
      apiKey: "$ANTHROPIC_API_KEY"
      model: "claude-3-5-sonnet-20241022"
      maxTokens: 1000

  - name: "local-llama"
    type: "ollama"
    config:
      model: "llama3"
      baseURL: "http://localhost:11434"
      temperature: 0.5

flows:
  - name: "content_analyzer"
    clientName: "zhipu-glm4"
    steps:
      - name: "analyze_sentiment"
        executor:
          type: "llm"
          withconfig:
            template: "分析以下文本的情感倾向：{{.Text}}"
            systemMessage: "你是一个专业的情感分析师。"
        validator:
          type: "string"
          withconfig:
            minLength: 10
        maxRetryTimes: 3

      - name: "extract_keywords"
        executor:
          type: "llm"
          withconfig:
            template: "从以下文本中提取关键词：{{.Text}}"
            systemMessage: "你是一个关键词提取专家。"
```

### JSON 配置

```json
{
  "clients": [
    {
      "name": "zhipu-glm4",
      "type": "zhipu",
      "config": {
        "apiKey": "$ZHIPU_API_KEY",
        "model": "glm-4-flash-250414",
        "temperature": 0.7
      }
    }
  ],
  "flows": [
    {
      "name": "simple_chat",
      "clientName": "zhipu-glm4",
      "steps": [
        {
          "name": "chat",
          "executor": {
            "type": "llm",
            "withconfig": {
              "template": "回答用户问题：{{.Text}}",
              "systemMessage": "你是一个有用的助手。"
            }
          }
        }
      ]
    }
  ]
}
```

## 环境变量管理

### 基本环境变量

创建 `.env` 文件：

```bash
# .env
# 中文 LLM 配置（推荐优先使用）
ZHIPU_API_KEY=your-zhipu-key-here
DASHSCOPE_API_KEY=your-dashscope-key-here
DEEPSEEK_API_KEY=your-deepseek-key-here

# Anthropic 配置
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key-here

# Azure OpenAI 配置
AZURE_OPENAI_API_KEY=your-azure-key-here
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/

# OpenAI 配置（如需使用国外服务）
OPENAI_API_KEY=sk-your-openai-key-here
OPENAI_MODEL=gpt-4
OPENAI_TEMPERATURE=0.7

# 应用配置
APP_ENV=development
APP_LOG_LEVEL=info
APP_PORT=8080
```

### 加载环境变量

```go
package main

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
    "github.com/jieliu2000/anyi"
)

func init() {
    // 加载 .env 文件
    if err := godotenv.Load(); err != nil {
        log.Println("未找到 .env 文件，使用系统环境变量")
    }
}

func main() {
    // 读取环境变量
    zhipuKey := os.Getenv("ZHIPU_API_KEY")
    if zhipuKey == "" {
        log.Fatal("ZHIPU_API_KEY 环境变量未设置")
    }

    temperature, err := strconv.ParseFloat(os.Getenv("ZHIPU_TEMPERATURE"), 64)
    if err != nil {
        temperature = 0.7 // 默认值
    }

    // 使用环境变量配置
    config := anyi.AnyiConfig{
        Clients: []anyi.ClientConfig{
            {
                Name: "zhipu",
                Type: "zhipu",
                Config: map[string]interface{}{
                    "apiKey":     zhipuKey,
                    "model":      getEnvOrDefault("ZHIPU_MODEL", "glm-4-flash-250414"),
                    "temperature": temperature,
                },
            },
        },
    }

    err = anyi.Config(&config)
    if err != nil {
        log.Fatal(err)
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## 多环境配置

### 环境特定配置

```bash
# 目录结构
configs/
├── base.yaml
├── development.yaml
├── staging.yaml
└── production.yaml
```

```yaml
# configs/base.yaml
clients:
  - name: "openai"
    type: "openai"
    config:
      model: "gpt-4o-mini"
      temperature: 0.7

flows:
  - name: "chat"
    clientName: "openai"
    steps:
      - name: "respond"
        executor:
          type: "llm"
          withconfig:
            template: "回答：{{.Text}}"
```

```yaml
# configs/development.yaml
clients:
  - name: "zhipu"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY_DEV"
      temperature: 0.9 # 开发环境使用更高创造性

server:
  port: 8080
  debug: true
```

```yaml
# configs/production.yaml
clients:
  - name: "zhipu"
    type: "zhipu"
    config:
      apiKey: "$ZHIPU_API_KEY_PROD"
      model: "glm-4-flash-250414" # 生产环境使用更好的模型
      temperature: 0.3 # 生产环境更保守

server:
  port: 80
  debug: false
```

### 配置合并

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jieliu2000/anyi"
    "gopkg.in/yaml.v2"
)

func loadEnvironmentConfig(env string) error {
    // 加载基础配置
    baseConfig, err := loadConfigFile("configs/base.yaml")
    if err != nil {
        return err
    }

    // 加载环境特定配置
    envConfigFile := fmt.Sprintf("configs/%s.yaml", env)
    envConfig, err := loadConfigFile(envConfigFile)
    if err != nil {
        return err
    }

    // 合并配置
    mergedConfig := mergeConfigs(baseConfig, envConfig)

    // 应用配置
    return anyi.Config(mergedConfig)
}

func loadConfigFile(filename string) (*anyi.AnyiConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config anyi.AnyiConfig
    err = yaml.Unmarshal(data, &config)
    return &config, err
}

func mergeConfigs(base, env *anyi.AnyiConfig) *anyi.AnyiConfig {
    // 简化的配置合并逻辑
    merged := *base

    // 合并客户端配置
    for _, envClient := range env.Clients {
        found := false
        for i, baseClient := range merged.Clients {
            if baseClient.Name == envClient.Name {
                // 合并配置项
                for k, v := range envClient.Config {
                    merged.Clients[i].Config[k] = v
                }
                found = true
                break
            }
        }
        if !found {
            merged.Clients = append(merged.Clients, envClient)
        }
    }

    return &merged
}

func main() {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }

    err := loadEnvironmentConfig(env)
    if err != nil {
        log.Fatalf("加载 %s 环境配置失败: %v", env, err)
    }

    fmt.Printf("已加载 %s 环境配置\n", env)
}
```

## 配置验证

### 基本验证

```go
package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/jieliu2000/anyi"
)

func validateConfig(config *anyi.AnyiConfig) error {
    // 验证客户端配置
    if len(config.Clients) == 0 {
        return fmt.Errorf("至少需要一个客户端配置")
    }

    for _, client := range config.Clients {
        if client.Name == "" {
            return fmt.Errorf("客户端名称不能为空")
        }

        if err := validateClientConfig(client); err != nil {
            return fmt.Errorf("客户端 %s 配置错误: %v", client.Name, err)
        }
    }

    // 验证工作流配置
    for _, flow := range config.Flows {
        if flow.Name == "" {
            return fmt.Errorf("工作流名称不能为空")
        }

        if len(flow.Steps) == 0 {
            return fmt.Errorf("工作流 %s 至少需要一个步骤", flow.Name)
        }
    }

    return nil
}

func validateClientConfig(client anyi.ClientConfig) error {
    switch client.Type {
    case "openai":
        if apiKey, ok := client.Config["apiKey"].(string); ok {
            if !strings.HasPrefix(apiKey, "sk-") {
                return fmt.Errorf("无效的 OpenAI API 密钥格式")
            }
        } else {
            return fmt.Errorf("缺少 OpenAI API 密钥")
        }
    case "anthropic":
        if apiKey, ok := client.Config["apiKey"].(string); ok {
            if !strings.HasPrefix(apiKey, "sk-ant-") {
                return fmt.Errorf("无效的 Anthropic API 密钥格式")
            }
        } else {
            return fmt.Errorf("缺少 Anthropic API 密钥")
        }
    }
    return nil
}

func main() {
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 获取当前配置并验证
    config := anyi.GetCurrentConfig()
    if err := validateConfig(config); err != nil {
        log.Fatalf("配置验证失败: %v", err)
    }

    fmt.Println("配置验证通过")
}
```

## 动态配置

### 配置热重载

```go
package main

import (
    "log"
    "time"
    "path/filepath"

    "github.com/fsnotify/fsnotify"
    "github.com/jieliu2000/anyi"
)

type ConfigManager struct {
    configFile string
    watcher    *fsnotify.Watcher
}

func NewConfigManager(configFile string) *ConfigManager {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    cm := &ConfigManager{
        configFile: configFile,
        watcher:    watcher,
    }

    // 监听配置文件变化
    go cm.watchConfig()

    // 添加文件到监听列表
    err = watcher.Add(filepath.Dir(configFile))
    if err != nil {
        log.Fatal(err)
    }

    return cm
}

func (cm *ConfigManager) watchConfig() {
    for {
        select {
        case event, ok := <-cm.watcher.Events:
            if !ok {
                return
            }

            if event.Op&fsnotify.Write == fsnotify.Write {
                if filepath.Base(event.Name) == filepath.Base(cm.configFile) {
                    log.Println("配置文件已更改，重新加载...")
                    time.Sleep(100 * time.Millisecond) // 等待文件写入完成
                    cm.reloadConfig()
                }
            }

        case err, ok := <-cm.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("监听配置文件错误: %v", err)
        }
    }
}

func (cm *ConfigManager) reloadConfig() {
    err := anyi.ConfigFromFile(cm.configFile)
    if err != nil {
        log.Printf("重新加载配置失败: %v", err)
    } else {
        log.Println("配置重新加载成功")
    }
}

func (cm *ConfigManager) Close() {
    cm.watcher.Close()
}

func main() {
    configManager := NewConfigManager("config.yaml")
    defer configManager.Close()

    // 初始加载配置
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 应用运行...
    select {} // 保持程序运行
}
```

## 最佳实践

### 1. 配置组织

```
project/
├── configs/
│   ├── base.yaml           # 基础配置
│   ├── development.yaml    # 开发环境
│   ├── staging.yaml        # 测试环境
│   └── production.yaml     # 生产环境
├── .env.example           # 环境变量模板
├── .env                   # 本地环境变量
└── config/
    ├── loader.go          # 配置加载器
    ├── validator.go       # 配置验证器
    └── manager.go         # 配置管理器
```

### 2. 安全考虑

- **敏感信息**：使用环境变量存储 API 密钥
- **配置文件**：不要将包含密钥的配置文件提交到版本控制
- **访问控制**：限制配置文件的读取权限
- **加密存储**：对敏感配置进行加密存储

### 3. 配置文档化

```yaml
# config.yaml
# Anyi 应用配置文件
# 版本: 1.0

# LLM 客户端配置
clients:
  # 智谱AI 客户端配置
  - name: "zhipu-glm4" # 客户端名称
    type: "zhipu" # 客户端类型
    config:
      apiKey: "$ZHIPU_API_KEY" # API 密钥（环境变量）
      model: "glm-4-flash-250414" # 使用的模型
      temperature: 0.7 # 创造性参数 (0.0-2.0)
      maxTokens: 2000 # 最大 token 数

# 工作流配置
flows:
  # 内容分析工作流
  - name: "content_analyzer" # 工作流名称
    clientName: "zhipu-glm4" # 使用的客户端
    description: "分析文本内容" # 工作流描述
    steps:
      - name: "analyze" # 步骤名称
        executor:
          type: "llm"
          withconfig:
            template: "分析：{{.Text}}"
```

### 4. 配置测试

```go
package config_test

import (
    "testing"
    "os"
    "github.com/jieliu2000/anyi"
)

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name        string
        configFile  string
        expectError bool
    }{
        {
            name:        "有效配置",
            configFile:  "testdata/valid_config.yaml",
            expectError: false,
        },
        {
            name:        "缺少客户端配置",
            configFile:  "testdata/missing_clients.yaml",
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := anyi.ConfigFromFile(tt.configFile)

            if tt.expectError && err == nil {
                t.Errorf("期望错误但没有收到错误")
            }

            if !tt.expectError && err != nil {
                t.Errorf("不期望错误但收到错误: %v", err)
            }
        })
    }
}
```

## 下一步

现在您已经掌握了配置管理，可以：

1. 学习 [多模态应用](multimodal.md) 来处理图像和文本
2. 查看 [错误处理指南](../how-to/error-handling.md) 来构建更健壮的应用
3. 了解 [性能优化](../how-to/performance.md) 来提升应用性能
4. 探索 [安全最佳实践](../advanced/security.md) 来保护您的应用

通过合理的配置管理，您可以构建更灵活、可维护、安全的 AI 应用程序！
