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

在Anyi中，你可以通过不同的配置创建不同的大模型客户端。以下为一些代码样例。

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

#### 访问其他大模型

如果你需要访问其他大模型，请参照[其他大模型设置](./llm_config.md)中的说明。


