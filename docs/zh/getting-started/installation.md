# 安装指南

本指南将带您完成 Anyi 框架的安装和设置过程。

## 系统要求

### Go 版本

Anyi 需要 Go 1.20 或更高版本。检查您的 Go 版本：

```bash
go version
```

如果您需要安装或升级 Go，请访问 [Go 官方网站](https://golang.org/dl/)。

### 操作系统支持

Anyi 支持所有主要操作系统：

- **Linux** (Ubuntu, CentOS, Debian 等)
- **macOS** (10.15 或更高版本)
- **Windows** (Windows 10 或更高版本)

### 网络连接

某些功能需要互联网连接：

- 访问 LLM API（OpenAI、Anthropic 等）
- 下载依赖包
- 获取模型更新（适用于本地提供商如 Ollama）

## 安装方法

### 方法 1：Go 模块（推荐）

这是最简单、最常用的安装方法：

```bash
# 初始化新的 Go 项目
mkdir my-anyi-project
cd my-anyi-project
go mod init my-anyi-project

# 添加 Anyi 依赖
go get -u github.com/jieliu2000/anyi
```

### 方法 2：克隆仓库

如果您想要最新的开发版本或贡献代码：

```bash
# 克隆仓库
git clone https://github.com/jieliu2000/anyi.git
cd anyi

# 安装依赖
go mod download
```

### 方法 3：从源码构建

构建您自己的二进制文件：

```bash
# 克隆仓库
git clone https://github.com/jieliu2000/anyi.git
cd anyi

# 构建
go build -o anyi ./cmd/...

# 可选：安装到 GOPATH
go install ./cmd/...
```

## 验证安装

创建一个简单的测试文件来验证安装：

```go
// main.go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/llm/chat"
)

func main() {
    fmt.Println("Anyi 安装验证")

    // 创建一个简单的消息来测试导入
    message := chat.Message{
        Role:    "user",
        Content: "Hello, Anyi!",
    }

    fmt.Printf("创建消息成功: %+v\n", message)
    fmt.Println("Anyi 安装成功！")
}
```

运行测试：

```bash
go run main.go
```

您应该看到：

```
Anyi 安装验证
创建消息成功: {Role:user Content:Hello, Anyi! Images:[] Function:<nil>}
Anyi 安装成功！
```

## 环境设置

### API 密钥配置

为了使用外部 LLM 提供商，您需要设置 API 密钥。创建一个 `.env` 文件：

```bash
# .env
OPENAI_API_KEY=your-openai-api-key-here
ANTHROPIC_API_KEY=your-anthropic-api-key-here
AZURE_OPENAI_API_KEY=your-azure-openai-api-key-here
AZURE_OPENAI_ENDPOINT=your-azure-openai-endpoint-here
ZHIPU_API_KEY=your-zhipu-api-key-here
DASHSCOPE_API_KEY=your-dashscope-api-key-here
DEEPSEEK_API_KEY=your-deepseek-api-key-here
SILICONCLOUD_API_KEY=your-siliconcloud-api-key-here
```

### 加载环境变量

在您的 Go 应用中加载环境变量：

```go
package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

func init() {
    // 加载 .env 文件
    if err := godotenv.Load(); err != nil {
        log.Println("未找到 .env 文件，使用系统环境变量")
    }
}

func main() {
    // 现在可以访问环境变量
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY 环境变量未设置")
    }

    // 您的应用代码...
}
```

安装 godotenv 包：

```bash
go get github.com/joho/godotenv
```

## 本地 LLM 设置（可选）

### Ollama 安装

如果您想使用本地 LLM，可以安装 Ollama：

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

启动 Ollama 并下载模型：

```bash
# 启动 Ollama 服务
ollama serve

# 在另一个终端中下载模型
ollama pull llama3
ollama pull codellama
```

## IDE 设置

### VS Code

推荐的 VS Code 扩展：

1. **Go** (Google) - Go 语言支持
2. **Go Outliner** - Go 代码大纲
3. **REST Client** - 测试 HTTP API

### GoLand

GoLand 提供开箱即用的优秀 Go 支持。确保启用：

1. Go 模块集成
2. 代码检查
3. 调试器支持

## 故障排除

### 常见问题

**问题：`go get` 失败，提示网络错误**

解决方案：

```bash
# 设置 Go 代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 重试安装
go get -u github.com/jieliu2000/anyi
```

**问题：导入错误 "module not found"**

解决方案：

```bash
# 确保您在正确的目录中
pwd

# 检查 go.mod 文件
cat go.mod

# 重新初始化模块
go mod init your-project-name
go get github.com/jieliu2000/anyi
```

**问题：API 密钥不工作**

解决方案：

1. 检查密钥格式是否正确
2. 验证密钥是否有效且未过期
3. 确保环境变量名称正确
4. 重启应用以重新加载环境变量

### 获取帮助

如果您遇到安装问题：

1. 查看 [常见问题](../reference/faq.md)
2. 搜索 [GitHub Issues](https://github.com/jieliu2000/anyi/issues)
3. 创建新的 issue 并包含：
   - 您的操作系统和版本
   - Go 版本
   - 完整的错误消息
   - 您尝试的步骤

## 下一步

安装完成后，您可以：

1. 阅读 [快速入门指南](quickstart.md) 构建您的第一个应用
2. 了解 [基本概念](concepts.md) 来理解 Anyi 的工作原理
3. 浏览 [LLM 客户端教程](../tutorials/llm-clients.md) 学习如何连接不同的 AI 提供商

恭喜！您已经成功安装了 Anyi。现在可以开始构建您的 AI 应用了！
