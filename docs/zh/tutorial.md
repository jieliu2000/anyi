# Anyi使用向导和示例

## 安装

```bash
go get -u  github.com/jieliu2000/anyi  
```

## 大模型访问

Anyi支持对以下大模型API的访问：

- OpenAI
- Azure OpenAI
- 阿里云模型服务灵积(Dashscope)
- Ollama

Anyi使用了一套统一的大模型访问接口，允许你使用几乎同一套代码以不同配置的形式访问不同的大模型。下面是一个使用OpenAI的简单示例：

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
在以上代码中，我们通过`openai.DefaultConfig()`函数创建了OpenAI的配置，然后使用`anyi.NewClient()`函数创建了一个Anyi客户端。

### 客户端创建方式

Anyi创建客户端可以通过以下两种方式：
- 直接使用`anyi.NewClient()`函数，传入模型名称和配置。
- 使用`llm.NewClient()`函数，其中llm是Anyi的子模块`github.com/jieliu2000/anyi/llm`

两者的区别在于`llm.NewClient()`创建的Client没有客户端名称，所以你需要自己在代码中保存创建的实例，而`anyi.NewClient()`创建的Client带有客户端名称。在创建之后可以通过`anyi.GetClient()`函数获取创建的客户端实例。

下面代码是使用`anyi.NewClient()`创建客户端的代码样例：

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
_, err := anyi.NewClient("openai", config)

//通过客户端名称获取客户端实例
client, err = anyi.GetClient("openai")
```

### 客户端配置

在Anyi中，你可以通过不同的配置创建不同的大模型客户端。以下为一些代码样例。另外你可以通过[Anyi LLM的pkg.go.dev文档](https://pkg.go.dev/github.com/jieliu2000/anyi/llm)获取到更加完整的代码样例。

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
* Openai文档
* [go-openai项目中的completion.go](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
* 也可以参照Anyi中的[openai.go](../../llm/openai/openai.go)中的常数定义

##### 使用不同的baseURL

如果你希望修改openAI调用的baseURL，可以使用`openai.NewConfig()`函数创建配置：

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```
其中`gpt-4o-mini`是模型名称，`https://api.openai.com/v1`是baseURL。这个参数在使用openai API兼容的其他大模型时尤为有用。


#### Azure OpenAI

Azure OpenAI的配置方法与OpenAI不同，你需要设置以下参数才能使用Azure OpenAI：
* Azure OpenAI的API Key
* Azure OpenAI的模型部署ID (Model Deployment ID)
* Azure OpenAI的Endpoint

以下代码演示了在Anyi中如何使用Azure OpenAI。

```go
config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
client, err := llm.NewClient(config)
```

其中azureopenai模块为`"github.com/jieliu2000/anyi/llm/azureopenai"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/azureopenai"
```

由于Azure OpenAI中的所有配置项都是必须的且没有默认选项，因此在Anyi中你仅有以上一种方法创建Azure OpenAI的配置。

#### Dashscope

Dashscope灵积是阿里云的大模型API服务，你可以通过 https://dashscope.aliyun.com/ 访问。目前版本中Anyi对Dashscope的访问是通过openai兼容API的方式，具体Dashscope文档可以参见 https://help.aliyun.com/zh/dashscope/developer-reference/compatibility-of-openai-with-dashscope 。

##### 默认配置访问Dashscope

```go
//dashscope.DefaultConfig 第一个参数为Dashscope的API Key。如果使用以下代码请在环境变量中把 DASHSCOPE_API_KEY 设置为你的 Dashscope API Key
config := dashscope.DefaultConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo")
client, err := anyi.NewClient("openai", config)
```

其中`qwen-turbo`是模型名称。dashscope模块为`"github.com/jieliu2000/anyi/llm/dashscope"`。可以通过以下代码导入：

```go
import "github.com/jieliu2000/anyi/llm/dashscope"
```

这里的默认配置是指使用默认的baseURL，也就是`"https://dashscope.aliyuncs.com/compatible-mode/v1"`

##### 指定baseURL访问Dashscope

```go
//NewConfig 第一个参数为Dashscope的API Key。如果使用以下代码请在环境变量中把 DASHSCOPE_API_KEY 设置为你的 Dashscope API Key。第二个参数为模型名称，第三个参数为baseURL。
config := dashscope.NewConfig(os.Getenv("DASHSCOPE_API_KEY"), "qwen-turbo", "https://your-url.com")
client, err := anyi.NewClient("openai", config)
```

#### Ollama

Ollama是一个可以在本地计算机运行的LLM服务，你可以通过 https://ollama.ai/ 访问。Ollama支持多种开源模型，例如Llama 3.1，Mistral，qwen2，Phi-3等等。如果要使用Anyi访问ollama，你需要首先安装ollama并使用ollama拉取你希望使用的模型。如果你还不太熟悉ollama的使用，你可以先阅读[ollama文档](https://github.com/ollama/ollama/tree/main/docs)。

以下代码演示了在Anyi中如何使用Ollama

```go
// 确保你已经安装并运行了ollama，并且在ollama中已经拉取了mistral模型
config := ollama.DefaultConfig("mistral")
client, err := llm.NewClient(config)
```
其中`mistral`是模型名称。参照[ollama文档](https://github.com/ollama/ollama/tree/main/docs)和[ollma模型库列表](https://ollama.com/library)获取更多模型名称。

ollama模块为`"github.com/jieliu2000/anyi/llm/ollama"`。可以通过以下代码导入：
```go
import "github.com/jieliu2000/anyi/llm/ollama"
```

以上代码在创建配置时使用了`ollama.DefaultConfig()`函数，这个函数会根据传入的模型名称创建一个默认的Anyi ollama配置。所谓默认配置是指使用默认的baseURL，也就是`"http://localhost:11434"`。

如果你希望访问的ollama API不在localhost，或者你改了ollama服务的端口，你可以使用`ollama.NewConfig()`函数创建一个自定义的配置：

```go
config := ollama.NewConfig("mistral", "http://your-ollama-server:11434")
client, err := llm.NewClient(config)
```

### 大模型聊天调用

在Anyi中，大模型聊天调用的入口是`client.Chat()`函数。这个函数使用一个`[]chat.Message`类型的参数，表示大模型接收的聊天消息。其中`chat`模块为`"github.com/jieliu2000/anyi/llm/chat"`。可以通过以下代码导入：

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
其中`Content`是消息内容，`Role`是消息角色，`ContentParts`是针对多模态大模型的图片调用设置的属性。在后面我们将讲到如何使用ContentParts属性传递图片给大模型。如果你只是需要利用大模型进行文本消息的聊天，那么`Content`属性就足够你使用了，你可以完全忽略`ContentParts`属性。

正如我们在前面已经演示过的一样，你可以通过直接给`chat.Message`中的属性直接赋值的方式创建消息，也可以使用`chat.NewMessage()`函数创建消息。

以下代码为直接创建一个user消息并调用大模型聊天的例子：

```go

messages := []chat.Message{
	{Role: "user", Content: "5+1=?"},
}
message, _, err := client.Chat(messages, nil)
```

以下代码为使用`chat.NewMessage()`函数创建一个user消息并调用大模型聊天的例子：

```go
messages := []chat.Message{
	 chat.NewMessage("user", "Hello, world!"),
}
message, _, err := client.Chat(messages, nil)
```

#### 大模型调用返回值

你可能注意到了`client.Chat()`函数的返回值也是一个`chat.Message`类型的值。这是我们力求让Anyi代码简化的一个设计。当然通常返回值的`Role`属性通常为`"assistant"`，`Content`属性为大模型回复的内容。

### 多模态大模型调用

在Anyi中，多模态大模型的调用入口仍然是`client.Chat()`函数。和单模态大模型不同，在多模态大模型调用中，你需要使用`chat.Message`结构体中的`ContentParts`属性，用于传递图片给大模型。

#### 最简单的多模态大模型调用

在`chat`模块中，我们提供了两个简易函数用于创建多模态大模型调用所需的`chat.Message`结构体。分别是

* `chat.NewImageMessageFromUrl()` 用于创建从网络图片URL传递给大模型的`chat.Message`结构体。
* `chat.NewImageMessageFromFile()` 用于创建从本地图片文件传递给大模型的`chat.Message`结构体。

以下代码演示了如何使用`chat.NewImageMessageFromUrl()`函数创建多模态大模型调用所需的`chat.Message`结构体（使用Dashscope）：

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
		chat.NewImageMessageFromUrl("user", "这是什么？", "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"),
	}
	message, _, err := client.Chat(messages, nil)

	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
		panic(err)
	}

	log.Printf("Response: %s", message.Content)
}
```
在以上代码中，我们使用`chat.NewImageMessageFromUrl()`函数创建了一个从网络图片URL传递给大模型的`chat.Message`结构体。`chat.NewImageMessageFromUrl()`函数的第一个参数为消息角色，第二个参数为文本消息内容，第三个参数为图片URL。

需要说明的是，在以上代码中，最后被创建的`chat.Message`结构体中的`ContentParts`属性为一个长度为2的数组，数组中的第一个元素是一个文本类型的`ContentPart`结构体，其文本内容为`"这是什么？"`；第二个元素是一个图片类型的`ContentPart`结构体，其图片URL为`https://dashscope.oss-cn-beijing.aliyuncs.com/`。而`chat.Message`结构体的`Content`属性为空字符串。也就是说`chat.NewImageMessageFromUrl()`参数中的文本消息是以`ContentParts`属性体现，而不是`Content`属性体现的。这也是Anyi中多模态大模型调用和单模态大模型调用的显著不同之处。


#### 通过`chat.ContentParts`属性传递图片给大模型

在多模态大模型中，`ContentParts`属性是一个数组，数组中的每个元素都是一个`ContentPart`结构体。`ContentPart`结构体定义如下：

```go
type ContentPart struct {
	Text        string `json:"text"`
	ImageUrl    string `json:"imageUrl"`
	ImageDetail string `json:"imageDetail"`
}
```
`ContentPart`结构体中包含三个属性：`Text`、`ImageUrl`和`ImageDetail`。`Text`属性用于传递文本消息给大模型，`ImageUrl`属性用于传递图片URL给大模型，`ImageDetail`属性用于传递图片的细节程度。

需要说明的是，**`Text`和`ImageUrl`属性是互斥的**。也就是说，如果你同时给`ContentPart`结构体设置`Text`和`ImageUrl`属性，那么`ImageUrl`属性将被忽略。如果你希望传递图片，那么就将`Text`属性设置为空字符串。

ImageUrl可以是一个网络图片URL，也可以是一个图片的base64编码的URL。Anyi提供了`chat.NewImagePartFromFile()`函数用于将一个本地图片转化为`ContentPart`结构体，也提供了`chat.NewImagePartFromUrl()`函数用于将一个网络图片URL转化为`ContentPart`结构体。


