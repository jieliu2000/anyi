# Anyi使用向导和示例

## 安装

```bash
go get github.com/jieliu2000/anyi
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
	"github.com/jieliu2000/anyi/message"
)

func main() {

	// 运行前记得在环境变量中把 OPENAI_API_KEY 设置为你的 OpenAI API Key
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("openai", config)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []message.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _ := client.Chat(messages)

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
