# Anyi 使用向导和示例

| [English](../en/tutorial.md) | [中文](../zh/tutorial.md) |

## 目录

- [安装](#安装)
- [大模型访问](#大模型访问)
  - [客户端创建方式](#客户端创建方式)
  - [客户端配置](#客户端配置)
    - [OpenAI](#openai)
    - [Azure OpenAI](#azure-openai)
    - [Dashscope](#dashscope)
    - [Ollama](#ollama)
    - [智谱 AI 开放平台 API](#智谱ai开放平台api-bigmodelcn)
    - [SiliconCloud AI 平台](#siliconcloud-ai平台-siliconflowcn)
- [大模型聊天调用](#大模型聊天调用)
  - [Message 结构体](#chatmessage结构体)
  - [调用返回值](#大模型调用返回值)
- [多模态大模型调用](#多模态大模型调用)
  - [简单调用](#最简单的多模态大模型调用)
  - [ContentParts 属性](#通过chatcontentparts属性读取图片给大模型)

## 安装

```bash
go get -u  github.com/jieliu2000/anyi
```

## 大模型访问

Anyi 支持对以下大模型 API 的访问：

- OpenAI
- Azure OpenAI
- 阿里云模型服务灵积(Dashscope)
- Ollama

Anyi 使用了一套统一的大模型访问接口，允许你使用几乎同一套代码以不同配置的形式访问不同的大模型。下面是一个使用 OpenAI 的简单示例：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {

	// 运行前记得在环境变量中把 OPENAI_API_KEY 设置为你的 OpenAI API Key
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _, _:= client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}

```

在以上代码中，我们通过`openai.DefaultConfig()`函数创建了 OpenAI 的配置，然后使用`anyi.NewClient()`函数创建了一个 Anyi 客户端。

### 客户端创建方式

Anyi 创建客户端可以通过以下两种方式：

- 直接使用`anyi.NewClient()`函数，传入模型名称和配置。
- 使用`llm.NewClient()`函数，其中 llm 是 Anyi 的子包`github.com/jieliu2000/anyi/llm`

两者的区别在于`llm.NewClient()`创建的 Client 没有客户端名称，所以你需要自己在代码中保存创建的实例，而`anyi.NewClient()`创建的 Client 带有客户端名称。在创建之后可以通过`anyi.GetClient()`函数获取创建的客户端实例。

下面代码是使用`anyi.NewClient()`创建客户端的代码样例：

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
_, err := anyi.NewClient("openai", config)

//通过客户端名称获取客户端实例
client, err = anyi.GetClient("openai")
```

### 客户端配置

在 Anyi 中，你可以通过不同的配置创建不同的大模型客户端。以下为一些代码样例。另外你可以通过[Anyi LLM 的 pkg.go.dev 文档](https://pkg.go.dev/github.com/jieliu2000/anyi/llm)获取到更加完整的代码样例。

#### Openai

##### 使用默认配置创建客户端

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("openai", config)
```

##### 指定模型名称创建客户端

默认配置使用的是`gpt-3.5-turbo`模型。如果你希望使用其他模型，可以使用`openai.NewConfigWithModel()`函数创建配置：

```go
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)
client, err := anyi.NewClient("gpt4o", config)
```

模型名称可以参照以下任一文档：、

- Openai 文档
- [go-openai 项目中的 completion.go](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
- 也可以参照 Anyi 中的[openai.go](../../llm/openai/openai.go)中的常数定义

##### 使用不同的 baseURL

如果你希望修改 openAI 调用的 baseURL，可以使用`openai.NewConfig()`函数创建配置：

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```

其中`gpt-4o-mini`是模型名称，`https://api.openai.com/v1`是 baseURL。这个参数在使用 openai API 兼容的其他大模型时尤为有用。

#### Azure OpenAI

Azure OpenAI 的配置方法与 OpenAI 不同，你需要设置以下参数才能使用 Azure OpenAI：

- Azure OpenAI 的 API Key
- Azure OpenAI 的模型部署 ID (Model Deployment ID)
- Azure OpenAI 的 Endpoint

以下代码演示了在 Anyi 中如何使用 Azure OpenAI。

```go
config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
client, err := llm.NewClient(config)
```

其中 azureopenai 包为`"github.com/jieliu2000/anyi/llm/azureopenai"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/azureopenai"
```

由于 Azure OpenAI 中的所有配置项都是必须的且没有默认选项，因此在 Anyi 中你仅有以上一种方法创建 Azure OpenAI 的配置。

#### Dashscope

Dashscope 灵积是阿里云的大模型 API 服务，你可以通过 https://dashscope.aliyun.com/ 访问。目前版本中 Anyi 对 Dashscope 的访问是通过 openai 兼容 API 的方式，具体 Dashscope 文档可以参见 https://help.aliyun.com/zh/dashscope/developer-reference/compatibility-of-openai-with-dashscope 。

##### 默认配置访问 Dashscope

```go
//dashscope.DefaultConfig 第一个参数为Dashscope的API Key。如果使用以下代码请在环境变量中把 DASHSCOPE_API_KEY 设置为你的 Dashscope API Key
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
client, err := anyi.NewClient("openai", config)
```

其中`qwen-turbo`是模型名称。dashscope 包为`"github.com/jieliu2000/anyi/llm/dashscope"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/dashscope"
```

这里的默认配置是指使用默认的 baseURL，也就是`"https://dashscope.aliyuncs.com/compatible-mode/v1"`

##### 指定 baseURL 访问 Dashscope

```go
//NewConfig 第一个参数为Dashscope的API Key。如果使用以下代码请在环境变量中把 DASHSCOPE_API_KEY 设置为你的 Dashscope API Key。第二个参数为模型名称，第三个参数为baseURL。
config := dashscope.NewConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo", "https://your-url.com")
client, err := anyi.NewClient("openai", config)
```

#### Ollama

Ollama 是一个可以在本地计算机运行的 LLM 服务，你可以通过 https://ollama.ai/ 访问。Ollama 支持多种开源模型，例如 Llama 3.1，Mistral，qwen2，Phi-3 等等。如果要使用 Anyi 访问 ollama，你需要首先安装 ollama 并使用 ollama 拉取你希望使用的模型。如果你还不太熟悉 ollama 的使用，你可以先阅读[ollama 文档](https://github.com/ollama/ollama/tree/main/docs)。

以下代码演示了在 Anyi 中如何使用 Ollama

```go
// 确保你已经安装并运行了ollama，并且在ollama中已经拉取了mistral模型
config := ollama.DefaultConfig("mistral")
client, err := llm.NewClient(config)
```

其中`mistral`是模型名称。参照[ollama 文档](https://github.com/ollama/ollama/tree/main/docs)和[ollma 模型库列表](https://ollama.com/library)获取更多模型名称。

ollama 包为`"github.com/jieliu2000/anyi/llm/ollama"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/ollama"
```

以上代码在创建配置时使用了`ollama.DefaultConfig()`函数，这个函数会根据传入的模型名称创建一个默认的 Anyi ollama 配置。所谓默认配置是指使用默认的 baseURL，也就是`"http://localhost:11434"`。

如果你希望访问的 ollama API 不在 localhost，或者你改了 ollama 服务的端口，你可以使用`ollama.NewConfig()`函数创建一个自定义的配置：

```go
config := ollama.NewConfig("mistral", "http://your-ollama-server:11434")
client, err := llm.NewClient(config)
```

#### 智谱 AI 开放平台 API (bigmodel.cn)

##### 默认配置访问智谱 AI 开放平台 API

```go
// Make sure you set ZHIPU_API_KEY environment variable to your Zhipu API key.
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

其中`glm-4-flash`是模型名称。zhipu 包为`"github.com/jieliu2000/anyi/llm/zhipu"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/zhipu"
```

#### SiliconCloud AI 平台 (siliconflow.cn)

##### 默认配置访问 SiliconCloud AI 平台

```go
// Make sure you set SILICONCLOUD_API_KEY environment variable to your Siliconcloud API key.
config := siliconcloud.DefaultConfig(os.Getenv("SILICONCLOUD_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

其中`glm-4-flash`是模型名称。siliconcloud 包为`"github.com/jieliu2000/anyi/llm/siliconcloud"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/siliconcloud"
```

### 大模型聊天调用

在 Anyi 中，普通大模型聊天调用的入口是`client.Chat()`函数。这个函数使用一个`[]chat.Message`类型的参数，表示大模型接收的聊天消息。

上面所提到的`chat`包为`"github.com/jieliu2000/anyi/llm/chat"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/chat"
```

#### `chat.Message`结构体

`chat.Message`结构体定义如下：

```go
type Message struct {
	Content    string      `json:"content,omitempty"`
	Role       string      `json:"role"`
	ContentParts []ContentPart `json:"contentParts,omitempty"`
}
```

其中`Content`是消息内容，`Role`是消息角色，`ContentParts`是针对多模态大模型的图片调用设置的属性。在后面我们将讲到如何使用 ContentParts 属性传递图片给大模型。如果你只是需要利用大模型进行文本消息的聊天，那么`Content`属性就足够你使用了，你可以完全忽略`ContentParts`属性。

_\* 关于如何使用`ContentParts`属性，可以参照[多模态大模型调用](#多模态大模型调用)_

正如我们在前面已经演示过的一样，你可以通过直接给`chat.Message`中的属性直接赋值的方式创建消息，也可以使用`
