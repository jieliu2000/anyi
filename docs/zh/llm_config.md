# 在Anyi中不同大模型的配置方法

## Openai

以下部分文档中openai模块为`"github.com/jieliu2000/anyi/llm/openai"`

```go
import "github.com/jieliu2000/anyi/llm/openai"
```

### 使用默认配置创建客户端

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("openai", config)
```

### 指定模型名称创建客户端

默认配置使用的是`gpt-3.5-turbo`模型。如果你希望使用其他模型，可以使用`openai.NewConfigWithModel()`函数创建配置：

```go
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)
client, err := anyi.NewClient("gpt4o", config)
```

模型名称可以参照以下任一文档：、
* Openai文档
* [go-openai项目中的completion.go](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
* 也可以参照Anyi中的[openai.go](../../llm/openai/openai.go)中的常数定义

### 使用不同的baseURL

如果你希望修改openAI调用的baseURL，可以使用`openai.NewConfig()`函数创建配置：

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```

其中`gpt-4o-mini`是模型名称，`https://api.openai.com/v1`是baseURL。这个参数在使用openai API兼容的其他大模型时尤为有用。

## Azure OpenAI

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