# Anyi - Open Source Autonomouse AI Agent Framework

| [English](README.md) | [中文](README-zh.md) |

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

## Introduction

Anyi is an open source AI Agent framework designed to help you build AI Agents that can be integrated with real-world work. We also provide APIs for accessing large language models.

Anyi requires [Go](https://go.dev/) version [1.20](https://go.dev/doc/devel/release#go1.20) or higher.

## Features

Anyi, as a Go programming framework, offers the following features:

- **Access to Large Language Models**: Allows access to large language models through a uniform interface, using different configurations to access various large models. Currently supported large model interfaces include:

  - OpenAI
  - Azure OpenAI
  - Dashscope (https://dashscope.aliyun.com/)
  - Ollama
  - Zhipu AI online service (https://bigmodel.cn/)
  - Silicon Cloud AI service (https://siliconflow.cn/)

- **Multimodal Model Support**: In addition to supporting regular text-based conversations, Anyi also supports sending images to multimodal large models for processing.
- **Multiple LLM Clients Support** Supports simultaneous access to multiple large language models from different sources. Different large model clients can be distinguished by their client names.
- **Prompt Generation Based on Go Templates**: Supports generating prompts based on Go language ([text/template](https://pkg.go.dev/text/template)).
- **Workflow Support**: Allows chaining multiple conversation tasks into workflows.
- **Step Validation in Workflows**: Repeats steps whose outputs do not meet expectations until the output is as expected. If the number of retries exceeds a predefined limit, an error is returned.
- **Using Different Large Model Clients in Workflow Steps**: Different workflow steps can use different large model clients.
- **Defining Multiple Workflows**: Allows defining multiple workflows and accessing them by workflow name.
- **Configuration-Based Workflow Definition**: Allows dynamic configuration of workflows through program code or static configuration files.

## Documentation and Examples

For detailed usage guides, please refer to [Anyi Usage Guide and Examples](/docs/zh/tutorial.md). Below are some quick start instructions.

## Quick start

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

### Accessing LLMs with Anyi

Here is a simple example of using Anyi to access OpenAI:

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
	// For more documentation and examples, see github.com/jieliu2000/anyi/llm package documentation.
	// Make sure you set OPENAI_API_KEY environment variable to your OpenAI API key.
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

In the above example, an OpenAI Anyi configuration is created via `openai.DefaultConfig`. Then this config instance is passed to `anyi.NewClient` to create an OpenAI client, and finally a chat request is sent via `client.Chat`.

### Define and execute workflows

Anyi allows you to define workflows, which can then be accessed via their workflow name. Each workflow contains one or multiple steps, and each step can define its own executor and validator.

Workflows can be created in the following three ways:

- **Directly creating instances** of `flow.Flow`, `flow.Step`, `flow.StepExecutor`, `flow.StepValidator` within the program.
- **Dynamic configuration**: Creating an instance of `anyi.AnyiConfig` via `anyi.AnyiConfig`, and then initializing Anyi with the `anyi.Config` method. Anyi will create objects such as `Client`, `Flow`, etc., according to the configuration.
- **Static configuration**: Creating through a configuration file, where the format of the configuration file can be toml, yaml, json, or other formats supported by [viper](https://github.com/spf13/viper). Anyi initializes static configuration using the `anyi.ConfigFromFile` method.

Anyi allows for **hybrid configuration**, meaning you can mix and use the three methods mentioned above within your program to create various objects such as clients, steps, executors and workflows themselves.

In Anyi's workflows, objects such as Flow, Step, StepExecutor, StepValidator all exchange information through the `flow.FlowContext` object. The declaration of the `flow.FlowContext` struct is as follows:

```go
type FlowContext struct {
	Text   string
	Memory ShortTermMemory
	Flow   *Flow
}
```

The Text attribute is used to pass text information, while the Memory attribute is used to pass other structured information. In the current version, `ShortTermMemory` is actually of type `any`, thus allowing you to set it to any instance of any type. The Flow attribute is used to keep a reference to the Flow.

Below is an example of defining a workflow using dynamic configuration with Anyi:

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
				Name: "openai",
				Type: "openai",
				Config: map[string]interface{}{
					"model":  openai.GPT4o,
					"apiKey": os.Getenv("OPENAI_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "smart_writer",
				Steps: []anyi.StepConfig{
					{
						Name: "write_scifi_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "Write a science fiction story about {{.Text}}",
							},
						},
					},
					{
						Name: "translate_novel",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `Translate the text enclosed in ''' into French. Do not include any additional output except for the translation. Text to translate: '''{{.Text}}'''. Translation result:`,
							},
						},
					},
				},
			},
		},
	}

	anyi.Config(&config)
	flow, err := anyi.GetFlow("smart_writer")
	if err != nil {
		panic(err)
	}
	context, err := flow.RunWithInput("the moon")
	if err != nil {
		panic(err)
	}
	log.Printf("%s", context.Text)
}
```

In the example above, an `AnyiConfig` configuration is first created that includes `Clients` and `Flows` properties. As the names suggest, `Clients` is used to define an array of Anyi client configurations, while `Flows` is used to define an array of workflow configurations.

In `Clients` configuration, there is only one openai `Client` configuration included. Since there is only one `Client` named `openai` in the program, Anyi registers this client as the **default Client**. Anyi allows registering multiple Clients, and in Flows and Steps, you can specify which `Client` should execute the task. If none is specified, Anyi will use the _default Client_.

In `Flows` configuration, a Flow named `smart_writer` is defined. This workflow contains two steps (Steps):

- The first step "write_scifi_novel" uses an Executor of the llm type. llm is a built-in type of Executor in Anyi that can call LLM models using **direct prompts** or **prompts based on templates**. In the example above, the `template` parameter specifies the prompt template for calling the LLM model. This template uses Go language text templates ([text/template](https://pkg.go.dev/text/template)).
  The template uses {{.Text}} as a parameter, where `.Text` is a property of `flow.FlowContext`. In Anyi's llm executor, Anyi sets the `.Text` property of `flow.FlowContext` based on the user's initial input. If the Executor outputs text content, Anyi sets the `.Text` property of `flow.FlowContext` as the output.

- The second step "translate_novel" also uses an Executor of the llm type but with a different prompt template.

After configuring Anyi with `anyi.Config(&config)`, you can get the workflow named `smart_writer` created by Anyi through `anyi.GetFlow("smart_writer")`. Then run the workflow with `flow.RunWithInput("the moon")`.

Before running the Flow, the parameter "the moon" passed to RunWithInput is set to the `.Text` property of `flow.FlowContext`, which is then passed to the Executor of the first Step ("write_scifi_novel").

The Executor of the first step "write_scifi_novel" generates a prompt based on the prompt template and user input, then calls the LLM model for computation. The output of this step, which is the generated content of the story, is set to the `.Text` property of returned `flow.FlowContext` and then passed to the next step "translate_novel" for translation.

Similarly, the second step also uses Go language templates, where {{.Text}} in the template is replaced with the `.Text` property of `flow.FlowContext`, which is the output content from "write_scifi_novel". Afterward, Anyi calls the LLM model for translation and sets the translation result back to the `.Text` property of `flow.FlowContext`. Finally, Anyi returns a reference to `flow.FlowContext` as the result of the Flow execution.

## License

Anyi is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
