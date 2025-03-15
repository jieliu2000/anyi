# Anyi - Open Source Autonomous AI Agent Framework

[![Go Reference](https://pkg.go.dev/badge/github.com/jieliu2000/anyi.svg)](https://pkg.go.dev/github.com/jieliu2000/anyi)
[![Go Report Card](https://goreportcard.com/badge/github.com/jieliu2000/anyi)](https://goreportcard.com/report/github.com/jieliu2000/anyi)

| [English](README.md) | [中文](README-zh.md) |

## Introduction

Anyi is an open-source Autonomous AI Agent framework written in [Go](https://go.dev/), designed to help you build AI agents that integrate with real-world workflows. We also provide APIs for accessing large language models.

Anyi requires Go version [1.20](https://go.dev/doc/devel/release#go1.20) or higher.

## Features

As a Go programming framework, Anyi offers the following features:

- **LLM Access**: Access different large language models through the same interface with different configurations. Currently supported LLM interfaces include:

  - OpenAI
  - Azure OpenAI
  - Anthropic
  - Ollama (for local model deployment)
  - DeepSeek
  - Zhipu AI Cloud Service (bigmodel.cn)
  - Silicon Cloud Service (https://siliconflow.cn/)
  - Aliyun Model Service Lingji (Dashscope)

- **Multimodal Model Support**: In addition to text-based chat, Anyi supports sending images to multimodal LLMs
- **Multi-client Support**: Access multiple LLMs from different sources simultaneously, with different LLM clients distinguished by name
- **Go Template-based Prompt Generation**: Support for prompt generation based on Go language [text/template](https://pkg.go.dev/text/template)
- **Workflow Support**: Chain multiple conversation tasks to form a workflow
- **Workflow Step Validation**: If a step's output doesn't meet expectations, the step repeats until the output is valid. If execution exceeds a set limit, an error is returned
- **Multi-client Workflow Steps**: Different steps in a workflow can use different LLM clients
- **Multiple Workflow Definitions**: Define multiple workflows and access them by name
- **Configuration-based Workflow Definition**: Define workflows dynamically through code or via static configuration files (YAML, JSON, TOML formats)

## Documentation and Examples

For detailed usage guides, please refer to [Anyi Tutorial and Examples](/docs/tutorial.md). Below are some simple getting started guides.

## Quick Start

### Installation

```bash
go get -u github.com/jieliu2000/anyi
```

### LLM Access Example

Here's a simple example of using Anyi to access OpenAI:

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
	// Make sure you have set the OPENAI_API_KEY environment variable
	config := openai.DefaultConfig("gpt-4")
	config.APIKey = os.Getenv("OPENAI_API_KEY")
	
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "What is 5+1?"},
	}
	message, _, _ := client.Chat(messages, nil)

	log.Printf("Response: %s\n", message.Content)
}
```

In the example above, we first create an OpenAI configuration through `openai.DefaultConfig`, then pass this configuration to `anyi.NewClient` to create a client, and finally send a chat request through `client.Chat`.

### Workflow Example

Anyi allows you to define workflows (Flow) and access different workflows by name. Each workflow can contain multiple steps (Step), and each step can define its own executor and validator.

Workflows can be created in three ways:

- **Direct Instance Creation**: Create `flow.Flow`, `flow.Step`, `flow.StepExecutor`, `flow.StepValidator` instances directly in code
- **Dynamic Configuration**: Create an `anyi.AnyiConfig` instance and initialize Anyi using the `anyi.Config` method
- **Static Configuration File**: Create through configuration files, which can be in YAML, JSON, TOML or any format supported by [viper](https://github.com/spf13/viper)

Anyi allows **mixed configuration**, meaning you can combine the three methods above to create various objects and workflows.

In Anyi workflows, information is passed through the `flow.FlowContext` object:

```go
type FlowContext struct {
	Text      string           // For passing text information
	Memory    ShortTermMemory  // For passing structured information, type 'any'
	Flow      *Flow            // Stores a reference to the Flow
	ImageURLs []string         // For image inputs to multimodal models
}
```

Here's an example of defining a workflow with dynamic configuration:

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
					"model":  "gpt-4",
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
								"template": `Translate the following text enclosed in triple backticks to French. Provide only the translation with no additional output. Text to translate: '''{{.Text}}'''. Translation:`,
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
	context, err := flow.RunWithInput("the Moon")
	if err != nil {
		panic(err)
	}
	log.Printf("%s", context.Text)
}
```

In this example:
1. We define an OpenAI client (as the default client)
2. Create a "smart_writer" workflow with two steps:
   - "write_scifi_novel": Generates a science fiction story
   - "translate_novel": Translates the story to French
3. Information is passed between steps via the Text property of FlowContext
4. The final translation result is output

## Configuration File Support

Anyi supports creating and managing LLM clients and workflows through configuration files, which is ideal for building deployable AI applications.

### Advantages of Configuration Files

1. **Separation of Configuration and Code**: Adjust AI application behavior without modifying code
2. **Environment Adaptation**: Prepare different configuration files for different environments (development, testing, production)
3. **Centralized Management**: Define configurations for all clients and workflows in one file
4. **Version Control**: Configurations can be tracked in version control systems

### Supported Formats

Anyi supports multiple configuration file formats:
- **YAML** (most commonly used, good readability)
- **JSON** (good compatibility, suitable for program generation)
- **TOML** (well-structured, suitable for complex configurations)

Configuration files can be loaded using the `ConfigFromFile` method, or from a string using the `ConfigFromString` method.

### Example Configuration File Using OpenAI

Here's a complete example configuration file (YAML format) using OpenAI:

```yaml
# anyi-openai-config.yaml
clients:
  - name: "gpt-4"
    type: "openai"
    config:
      model: "gpt-4"
      apiKey: "$OPENAI_API_KEY"
  
  - name: "gpt-3.5"
    type: "openai"
    config:
      model: "gpt-3.5-turbo"
      apiKey: "$OPENAI_API_KEY"

flows:
  - name: "qa-flow"
    clientName: "gpt-4"  # Default to using GPT-4
    steps:
      - name: "answer-question"
        executor:
          type: "llm"
          withconfig:
            template: "Please answer the following question in detail: {{.Text}}. Your answer should be comprehensive, accurate, and provide reasoning."
            systemMessage: "You are a professional knowledge assistant, skilled at providing accurate, detailed answers."
        maxRetryTimes: 2
      
  - name: "creative-writing"
    clientName: "gpt-3.5"  # Use GPT-3.5 for creative tasks
    steps:
      - name: "generate-story"
        executor:
          type: "llm"
          withconfig:
            template: "Write a short story about {{.Text}} with a clear beginning, middle, and end."
            systemMessage: "You are a creative fiction writer."
      
      - name: "summarize-story"
        executor:
          type: "llm"
          withconfig:
            template: "Summarize the following story in three sentences:\n\n{{.Text}}"
        clientName: "gpt-4"  # This step specifically uses GPT-4
        validator:
          type: "string"
          withconfig:
            matchRegex: "^.{30,500}$"  # Ensure the summary is of reasonable length
```

### Loading and Using Configuration Files

```go
package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

func main() {
	// Load configuration file
	err := anyi.ConfigFromFile("./anyi-openai-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get and run the QA workflow
	qaFlow, err := anyi.GetFlow("qa-flow")
	if err != nil {
		log.Fatalf("Failed to get workflow: %v", err)
	}
	
	result, err := qaFlow.RunWithInput("What is the history of artificial intelligence?")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Println("QA Result:", result.Text)
}
```

### Configuration File Best Practices

1. **Environment Variable Substitution**: Use `$VARIABLE_NAME` in configuration files to reference environment variables, protecting sensitive information
2. **Modular Configuration Files**: Split configurations into multiple files by functionality for easier management
3. **Validator Usage**: Add validators for critical steps to ensure outputs meet expectations
4. **Step Retries**: Set reasonable `maxRetryTimes` values for unstable steps

## Built-in Components

Anyi provides various built-in components to help build AI applications:

### Built-in Executors

- **LLMExecutor**: LLM-based executor, supporting template prompts and direct prompts
- **ConditionalFlowExecutor**: Condition-based executor that selects different sub-workflows based on conditions
- **RunCommandExecutor**: System command executor for executing system commands
- **SetContextExecutor**: Context setting executor that can set properties in the workflow context

### Built-in Validators

- **StringValidator**: String validator supporting equality comparison and regex matching
- **JsonValidator**: JSON validator that verifies if output is valid JSON

## License

Anyi is licensed under the [Apache License 2.0](LICENSE).
