# Configuring Different Large Language Models in Anyi

## OpenAI

The following documentation refers to the `openai` module as `"github.com/jieliu2000/anyi/llm/openai"`.

```go
import "github.com/jieliu2000/anyi/llm/openai"
```

### Creating a Client with Default Configuration

```go
config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
client, err := anyi.NewClient("openai", config)
```

### Creating a Client with a Specified Model Name

The default configuration uses the `gpt-3.5-turbo` model. If you want to use another model, you can create the configuration using the `openai.NewConfigWithModel()` function:

```go
config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), openai.GPT4o)
client, err := anyi.NewClient("gpt4o", config)
```

The model name can be referenced from any of the following documents:
* OpenAI documentation
* [go-openai project's completion.go](https://github.com/sashabaranov/go-openai/blob/master/completion.go)
* Or refer to the constant definitions in Anyi’s [openai.go](../../llm/openai/openai.go)

### Using a Different baseURL

If you want to modify the baseURL used for calling OpenAI, you can create the configuration using the `openai.NewConfig()` function:

```go
config := openai.NewConfig(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", "https://api.openai.com/v1")
client, err := anyi.NewClient("gpt4o", config)
```

Here, `gpt-4o-mini` is the model name, and `https://api.openai.com/v1` is the baseURL. This parameter is particularly useful when using other large models compatible with the OpenAI API.

## Azure OpenAI

The configuration method for Azure OpenAI is different from OpenAI. You need to set the following parameters to use Azure OpenAI:
* Azure OpenAI API Key
* Azure OpenAI Model Deployment ID
* Azure OpenAI Endpoint

The following code demonstrates how to use Azure OpenAI in Anyi.

```go
config := azureopenai.NewConfig(os.Getenv("AZ_OPENAI_API_KEY"), os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"), os.Getenv("AZ_OPENAI_ENDPOINT"))
client, err := llm.NewClient(config)
```

The `azureopenai` module is `"github.com/jieliu2000/anyi/llm/azureopenai"`. You can import it with the following code:

```go
import "github.com/jieliu2000/anyi/llm/azureopenai"
```