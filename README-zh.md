# Anyi(安易) - 开源的自主式 AI 智能体框架

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

| [English](README.md) | [中文](README-zh.md) |

## 介绍

Anyi(安易)是一个开源的[Go 语言](https://go.dev/)自主式 AI 智能体框架，旨在帮助你构建可以与实际工作场景无缝结合的 AI 智能体。我们还提供了便捷的大语言模型访问接口。

Anyi 需要 Go 语言[1.20](https://go.dev/doc/devel/release#go1.20)或更高版本。

## 特性

Anyi 作为一个 Go 语言的编程框架，提供以下特性：

- **多种大语言模型支持**：通过统一接口访问不同的大语言模型，目前支持的模型接口包括：

  - 阿里云灵积模型服务(Dashscope)
  - 智谱 AI 大模型服务(bigmodel.cn)
  - DeepSeek AI 大模型
  - Silicon Cloud 云服务（https://siliconflow.cn/）
  - Ollama 本地大模型部署（支持多种开源模型如Qwen、Llama等）
  - OpenAI API 服务
  - Azure OpenAI 服务

- **多模态能力**：支持图文多模态大语言模型调用，可以发送图片进行分析
- **多客户端管理**：支持同时连接多个不同来源的大语言模型，通过客户端名称进行区分和管理
- **模板化提示词系统**：基于 Go 语言模板引擎([text/template](https://pkg.go.dev/text/template))的提示词生成系统
- **工作流编排**：将多个对话任务串联成工作流，实现复杂业务逻辑
- **输出质量控制**：支持对每个工作流步骤进行校验，不符合预期则自动重试，超过重试次数则返回错误
- **灵活的模型调度**：工作流中不同步骤可以使用不同的大语言模型，实现性能与成本的平衡
- **多工作流管理**：支持定义多个工作流，并根据名称灵活调用
- **配置驱动开发**：支持通过代码动态配置或通过静态配置文件（YAML、JSON、TOML格式）定义工作流

## 文档和示例

详细的使用指南请参考[Anyi 使用教程和示例](/docs/zh/tutorial.md)。以下是基础入门内容。

## 快速开始

### 安装

```bash
go get -u github.com/jieliu2000/anyi
```

### 大语言模型调用示例

以下是使用 Anyi 调用智谱AI模型的示例：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/zhipuai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// 确保已设置智谱AI的API密钥
	config := zhipuai.DefaultConfig("glm-4")
	config.APIKey = os.Getenv("ZHIPUAI_API_KEY")
	
	client, err := anyi.NewClient("glm4", config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "请介绍一下中国的传统节日"},
	}
	message, _, _ := client.Chat(messages, nil)

	log.Printf("回复: %s\n", message.Content)
}
```

在上面的示例中，我们首先通过`zhipuai.DefaultConfig`创建智谱AI的配置，然后将该配置传递给`anyi.NewClient`创建客户端，最后通过`client.Chat`发送对话请求。

### 工作流示例

Anyi 允许你定义工作流(Flow)，然后通过工作流名称调用不同的工作流。每个工作流可以包含多个步骤(Step)，每个步骤可以定义自己的执行器（Executor）和校验器（Validator）。

工作流可以通过以下三种方式创建：

- **直接创建实例**：在代码中直接创建`flow.Flow`、`flow.Step`、`flow.StepExecutor`、`flow.StepValidator`等实例
- **动态配置**：创建一个`anyi.AnyiConfig`实例，然后通过`anyi.Config`方法初始化 Anyi
- **静态配置文件**：通过配置文件创建，支持 YAML、JSON、TOML 等[viper](https://github.com/spf13/viper)支持的格式

Anyi 支持**混合配置**，可以在代码中混合使用上述三种方式创建工作流中的对象和工作流本身。

在 Anyi 的工作流中，信息通过`flow.FlowContext`对象传递：

```go
type FlowContext struct {
	Text      string           // 传递文本信息
	Memory    ShortTermMemory  // 传递结构化数据，类型为 any
	Flow      *Flow            // 保存对工作流的引用
	ImageURLs []string         // 多模态模型的图片输入
}
```

以下是使用阿里云灵积模型服务的工作流示例：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func main() {
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "dashscope",
				Type: "dashscope",
				Config: map[string]interface{}{
					"model":  "qwen-max",
					"apiKey": os.Getenv("DASHSCOPE_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "creative_writer",
				Steps: []anyi.StepConfig{
					{
						Name: "write_scifi_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "写一篇关于{{.Text}}的科幻小说，字数控制在800字左右",
							},
						},
					},
					{
						Name: "translate_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `把下面用'''括起来的文本翻译成英文，保持原文的风格和意境，除了翻译结果外不要有其他输出。需要翻译的文本:'''{{.Text}}'''`,
							},
						},
					},
				},
			},
		},
	}

	anyi.Config(&config)
	flow, err := anyi.GetFlow("creative_writer")
	if err != nil {
		panic(err)
	}
	context, err := flow.RunWithInput("未来的北京")
	if err != nil {
		panic(err)
	}
	log.Printf("%s", context.Text)
}
```

在这个示例中：
1. 我们定义了一个使用阿里云灵积的Qwen-Max模型的客户端
2. 创建了名为"creative_writer"的工作流，包含两个步骤：
   - "write_scifi_novel"：生成一篇关于未来北京的科幻小说
   - "translate_novel"：将小说翻译成英文
3. 步骤之间通过FlowContext的Text属性传递信息
4. 最终输出翻译后的英文小说

## 配置文件支持

Anyi 支持通过配置文件来创建和管理LLM客户端和工作流，这种方式特别适合构建可部署的AI应用。

### 配置文件的优势

1. **配置与代码分离**：无需修改代码即可调整AI应用的行为
2. **环境适配**：可以为不同环境（开发、测试、生产）准备不同配置
3. **集中管理**：在一个配置文件中定义所有客户端和工作流
4. **版本控制**：将配置纳入版本管理，便于跟踪变更

### 支持的格式

Anyi 支持多种配置文件格式：
- **YAML**（推荐使用，结构清晰易读）
- **JSON**（标准格式，适合程序生成）
- **TOML**（结构化好，适合复杂配置）

配置文件可以通过`ConfigFromFile`方法加载，也可以使用`ConfigFromString`方法从字符串加载。

### 使用本地Ollama的配置文件示例

以下是一个完整的使用本地Ollama部署模型的配置文件（YAML格式）：

```yaml
# anyi-local-config.yaml
clients:
  - name: "local-qwen"
    type: "ollama"
    config:
      model: "qwen2:7b"
      ollamaApiURL: "http://localhost:11434/api"
  
  - name: "local-llama3"
    type: "ollama"
    config:
      model: "llama3"
      ollamaApiURL: "http://localhost:11434/api"

flows:
  - name: "问答流程"
    clientName: "local-qwen"  # 默认使用Qwen模型
    steps:
      - name: "回答问题"
        executor:
          type: "llm"
          withconfig:
            template: "请详细回答以下问题: {{.Text}}。回答需要全面、准确，并列出观点的理由。"
            systemMessage: "你是一个专业的知识问答助手，擅长提供详实、权威的解答。"
        maxRetryTimes: 2
      
  - name: "内容创作"
    clientName: "local-llama3"  # 使用Llama3模型
    steps:
      - name: "生成故事"
        executor:
          type: "llm"
          withconfig:
            template: "请以{{.Text}}为主题创作一个短篇故事，故事应该有完整的起承转合结构，字数在800字左右。"
            systemMessage: "你是一位富有创造力的作家，善于创作引人入胜的故事。"
      
      - name: "生成摘要"
        executor:
          type: "llm"
          withconfig:
            template: "请用三句话高度概括以下故事的主要内容：\n\n{{.Text}}"
        clientName: "local-qwen"  # 这个步骤特别指定使用Qwen模型
        validator:
          type: "string"
          withconfig:
            matchRegex: "^.{50,300}$"  # 确保摘要长度适中
```

### 加载和使用配置文件

```go
package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

func main() {
	// 加载配置文件
	err := anyi.ConfigFromFile("./anyi-local-config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 获取并运行问答流程
	qaFlow, err := anyi.GetFlow("问答流程")
	if err != nil {
		log.Fatalf("获取流程失败: %v", err)
	}
	
	result, err := qaFlow.RunWithInput("中国古代四大发明的历史意义是什么？")
	if err != nil {
		log.Fatalf("流程执行失败: %v", err)
	}

	fmt.Println("问答结果:", result.Text)
}
```

### 配置文件最佳实践

1. **环境变量使用**：使用`$VARIABLE_NAME`引用环境变量，保护API密钥等敏感信息
2. **配置模块化**：按功能将配置拆分为多个文件，便于维护
3. **添加验证器**：为关键步骤添加验证器，确保输出质量
4. **合理设置重试**：为不稳定的步骤设置适当的`maxRetryTimes`值

## 内置组件

Anyi 提供了多种内置组件，帮助你快速构建AI应用：

### 内置执行器

- **LLMExecutor**：大语言模型执行器，支持模板提示词和直接提示词
- **ConditionalFlowExecutor**：条件流执行器，根据条件选择不同的子流程执行
- **RunCommandExecutor**：系统命令执行器，可执行系统命令并获取结果
- **SetContextExecutor**：上下文设置执行器，可修改工作流上下文的属性值

### 内置校验器

- **StringValidator**：字符串校验器，支持相等比较和正则表达式匹配
- **JsonValidator**：JSON校验器，验证输出是否为有效的JSON格式

## 许可证

Anyi 遵循 [Apache License 2.0](LICENSE) 开源许可。
