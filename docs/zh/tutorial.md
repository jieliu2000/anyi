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

#### OpenAI

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
// 确保设置ZHIPU_API_KEY环境变量为你的智谱API密钥
config := zhipu.DefaultConfig(os.Getenv("ZHIPU_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

zhipu 包路径为：

```go
import "github.com/jieliu2000/anyi/llm/zhipu"
```

#### SiliconCloud AI 平台 (siliconflow.cn)

##### 默认配置访问 SiliconCloud

```go
// 确保设置SILICONCLOUD_API_KEY环境变量为你的SiliconCloud API密钥
config := siliconcloud.DefaultConfig(os.Getenv("SILICONCLOUD_API_KEY"), "glm-4-flash")
client, err := llm.NewClient(config)
```

siliconcloud 包路径为：

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

正如我们在前面已经演示过的一样，你可以通过直接给`chat.Message`中的属性直接赋值的方式创建消息，也可以使用`chat.NewMessage()`函数创建消息。

以下代码为直接创建一个 user 消息并调用大模型聊天的例子：

```go

messages := []chat.Message{
	{Role: "user", Content: "5+1=?"},
}
message, responseInfo, err := client.Chat(messages, nil)
```

以下代码为使用`chat.NewMessage()`函数创建一个 user 消息并调用大模型聊天的例子：

```go
messages := []chat.Message{
	 chat.NewMessage("user", "Hello, world!"),
}
message, responseInfo, err := client.Chat(messages, nil)
```

#### 大模型调用返回值

`client.Chat()`函数有三个返回值：

    - 第一个是`*chat.Message`类型的指针，表示大模型返回的聊天消息。由于这是一个指针，因此如果调用大模型过程中发生了错误，那么程序会将这个返回值设置为nil。
    - 第二个是`chat.ResponseInfo`类型的值，表示大模型的响应信息，例如Token数量等。
    - 第三个是`error`类型的值，表示调用大模型时发生的错误。

你可能注意到了`client.Chat()`函数的返回值也是一个`chat.Message`类型的指针。这是我们力求让 Anyi 代码简化的一个设计。当然通常返回的 Message 中`Role`属性通常为`"assistant"`，`Content`属性为大模型回复的内容。

`chat.ResponseInfo`目前定义如下：

```go
type ResponseInfo struct {
	PromptTokens     int
	CompletionTokens int
}
```

正如你所看到的，`chat.ResponseInfo`结构体中的属性都很简单，分别表示大模型返回的各种 Token 数量。未来我们会在这个结构体中添加更多的属性，以支持更多大模型的调用信息。

Anyi 在支持多种大模型接口，而每种大模型返回的信息也不尽相同。因此并不一定每个大模型调用都能返回`chat.ResponseInfo`结构体中的所有属性。这时你需要根据具体的大模型接口来处理返回值。

### 多模态大模型调用

在 Anyi 中，多模态大模型的调用入口仍然是`client.Chat()`函数。和单模态大模型不同，在多模态大模型调用中，你需要使用`chat.Message`结构体中的`ContentParts`属性，用于传递图片给大模型。

#### 最简单的多模态大模型调用

在`github.com/jieliu2000/anyi/llm/chat`包中，我们提供了两个简易函数用于创建多模态大模型调用所需的`chat.Message`结构体。分别是

- `chat.NewImageMessageFromUrl()` 用于创建从网络图片 URL 传递给大模型的`chat.Message`结构体。
- `chat.NewImageMessageFromFile()` 用于创建从本地图片文件传递给大模型的`chat.Message`结构体。

这两个函数在你只需要给视觉大模型传递一个提示词字符串和一张图片时非常有用。而如果你需要传递多张图片，你就需要手动创建`chat.Message`结构体，然后手动设置`ContentParts`属性。

以下代码演示了如何使用`chat.NewImageMessageFromUrl()`函数创建多模态大模型调用所需的`chat.Message`结构体（使用 Dashscope）：

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/dashscope"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Make sure you set DASHSCOPE_API_KEY environment variable to your Dashscope API key.
	config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-vl-plus")
	client, err := anyi.NewClient("dashscope", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		chat.NewImageMessageFromUrl("user", "What's this?", "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"),
	}

	message, responseInfo, err := client.Chat(messages, nil)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
	log.Printf("Prompt tokens: %v", responseInfo.PromptTokens)
}
```

在以上代码中，我们使用`chat.NewImageMessageFromUrl()`函数创建了一个从网络图片 URL 传递给大模型的`chat.Message`结构体。`chat.NewImageMessageFromUrl()`函数的第一个参数为消息角色，第二个参数为文本消息内容，第三个参数为图片 URL。

需要说明的是，在以上代码中，最后被创建的`chat.Message`结构体中的`ContentParts`属性为一个长度为**2**的数组（而不是只有一个 ContentPart)，数组中第一个元素是一个文本类型的`ContentPart`结构体，其文本内容为`"这是什么？"`；第二个元素是一个图片类型的`ContentPart`结构体，其图片 URL 为`https://dashscope.oss-cn-beijing.aliyuncs.com/`。

而`chat.Message`结构体的`Content`属性为空字符串。也就是说`chat.NewImageMessageFromUrl()`参数中的文本消息是以`ContentParts`属性体现，而不是以`chat.Message`的`Content`属性体现的。这也是 Anyi 中多模态大模型调用和单模态大模型调用的显著不同之处。

`chat.NewImageMessageFromUrl()`用于创建从网络图片 URL 传递给大模型的`chat.Message`结构体。

`chat.NewImageMessageFromFile()`函数和`chat.NewImageMessageFromUrl()`函数类似，只是它从本地图片文件创建`chat.Message`结构体。以下为使用`chat.NewImageMessageFromFile()`函数的示例代码：

```go
messages := []chat.Message{
	chat.NewImageMessageFromFile("user", "What number is in the image?", "../internal/test/number_six.png"),
	}
```

可以看到`chat.NewImageMessageFromFile()`函数的第一个参数为消息角色，第二个参数为文本消息内容，第三个参数为图片文件路径。

被`chat.NewImageMessageFromFile()`函数创建的`chat.Message`结构体中的`ContentParts`属性为一个长度为**2**的数组（而不是只有一个 ContentPart)，数组中第一个元素是一个文本类型的`ContentPart`结构体，其文本内容为`"这是什么？"`；第二个元素是一个图片类型的`ContentPart`结构体，其`ImageUrl`属性为所传图片文件的 base64 编码 URL。

`chat.NewImageMessageFromFile()`函数会试着读取文件，如果读取失败，那么会返回一个仅包含`Role`属性的`chat.Message`值，其`Content`属性和`ContentParts`属性都为空。

#### 通过`chat.ContentParts`属性读取图片给大模型

在多模态大模型调用中，`chat.Message`结构体中的`ContentParts`属性`ContentParts`属性是一个数组，数组中的每个元素都是一个`ContentPart`结构体。`ContentPart`结构体定义如下：

```go
type ContentPart struct {
	Text        string `json:"text"`
	ImageUrl    string `json:"imageUrl"`
	ImageDetail string `json:"imageDetail"`
}
```

`ContentPart`结构体中包含三个属性：`Text`、`ImageUrl`和`ImageDetail`。`Text`属性用于传递文本消息给大模型，`ImageUrl`属性用于传递图片 URL 给大模型，`ImageDetail`属性用于传递图片的细节程度。

需要说明的是，**`Text`和`ImageUrl`属性是互斥的**。也就是说，如果你同时给`ContentPart`结构体设置`Text`和`ImageUrl`属性，那么`ImageUrl`属性将被忽略。如果你希望传递图片，那么就将`Text`属性设置为空字符串。

ImageUrl 可以是一个网络图片 URL，也可以是一个图片的 base64 编码的 URL。Anyi 提供了`chat.NewImagePartFromFile()`函数用于将一个本地图片转化为`ContentPart`结构体，也提供了`chat.NewImagePartFromUrl()`函数用于将一个网络图片 URL 转化为`ContentPart`结构体。

ImageDetail 属性用于传递图片的细节程度。例如`"low"`、`"medium"`，`"high"`和`"auto"`,用于表示图片的细节程度。如果你不清楚你应该使用哪个值，那么你可以直接设置这个参数为空字符串。

以下代码演示了如何使用`chat.NewImagePartFromUrl()`函数将一个网络图片 URL 转化为`ContentPart`结构体：

```go
imageUrl := "https://example.com/image.jpg"
contentPart, err := chat.NewImagePartFromUrl(imageUrl, "")
```

`chat.NewImagePartFromUrl()`函数不会校验图片。但是在使用`client.Chat()`函数调用大模型时，Anyi 会根据不同大模型的情况进行不同的动作：

- 大多数情况下，如果大模型 API 支持通过 URL 传递图片信息，anyi 不会检查图片的 URL 是否有效，而是直接把图片 URL 传给大模型 API。在这种情况下，你需要确保你提供的图片 URL 是有效的。
- 对于例如 ollama 之类的 API，它们不支持通过 URL 方式传递图片，在这种情况下 Anyi 会根据 URL 读取图片，并将图片转化为大模型 API 要求的格式（比如 base64 编码）传递出去。很明显如果 URL 指向了一个不可访问或者无效的图片，`client.Chat()`函数在真正与大模型交互之前就会返回错误。
