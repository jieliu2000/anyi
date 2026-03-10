package builtin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/utils"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
)

// ResponseFilter is an interface for filtering LLM responses.
type ResponseFilter interface {
	Filter(response string) (string, error)
}

// LLMExecutor is an executor that runs LLM (Large Language Model) steps.
type LLMExecutor struct {
	// Template for the LLM step. Can be a string or a file path.
	// If it's a file path, it should start with "file://" prefix.
	Template string `json:"template" yaml:"template" mapstructure:"template"`

	// SystemMessage is the system message for the LLM step.
	SystemMessage string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`

	// TemplateFormatter is the template formatter for generating prompts.
	TemplateFormatter *chat.PromptyTemplateFormatter `json:"-" yaml:"-" mapstructure:"-"`

	// Config is the configuration for the LLM client.
	Config *llm.ClientConfig `json:"config" yaml:"config" mapstructure:"config"`

	// Client is the LLM client to use for the step.
	Client llm.Client `json:"-" yaml:"-" mapstructure:"-"`

	// ResponseFilter is the response filter to use for the step.
	ResponseFilter ResponseFilter `json:"-" yaml:"-" mapstructure:"-"`
}

// Init initializes the LLMExecutor.
// It initializes the LLM client and response filter.
func (executor *LLMExecutor) Init() error {
	var err error
	if executor.Config != nil {
		modelConfig, err := llm.NewModelConfigFromClientConfig(executor.Config)
		if err != nil {
			return err
		}
		executor.Client, err = llm.NewClient(modelConfig)
		if err != nil {
			return err
		}
	} else {
		executor.Client, err = llm.NewClient(openai.DefaultConfig(""))
		if err != nil {
			return err
		}
	}

	if executor.ResponseFilter != nil {
		if initableResponseFilter, ok := executor.ResponseFilter.(flow.Initable); ok {
			err = initableResponseFilter.Init()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Run executes the LLM step.
// It creates an LLM step with the provided template and runs it.
func (executor *LLMExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	var prompt string
	var err error

	if executor.TemplateFormatter != nil {
		prompt, err = executor.TemplateFormatter.Format(flowContext)
		if err != nil {
			return &flowContext, err
		}
	} else {
		if strings.HasPrefix(executor.Template, "file://") {
			filePath := strings.TrimPrefix(executor.Template, "file://")
			templateContent, err := utils.GetFileContent(filePath)
			if err != nil {
				return &flowContext, err
			}
			prompt = templateContent
		} else {
			prompt = executor.Template
		}
	}

	if prompt == "" {
		return &flowContext, errors.New("prompt is empty")
	}

	messages := []chat.Message{
		{Role: "user", Content: prompt},
	}

	if executor.SystemMessage != "" {
		messages = []chat.Message{
			{Role: "system", Content: executor.SystemMessage},
			{Role: "user", Content: prompt},
		}
	}

	response, _, err := executor.Client.Chat(messages, nil)
	if err != nil {
		return &flowContext, err
	}

	result := response.Content
	if executor.ResponseFilter != nil {
		filteredResult, err := executor.ResponseFilter.Filter(result)
		if err != nil {
			return &flowContext, err
		}
		result = filteredResult
	}

	flowContext.Text = result
	return &flowContext, nil
}

// NewLLMStepWithTemplateFile creates a new LLM step with a template file.
// The filePath should be a path to a template file.
func NewLLMStepWithTemplateFile(filePath string, client llm.Client, systemMessage string) (*flow.Step, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	templateContent, err := utils.GetFileContent(filePath)
	if err != nil {
		return nil, err
	}

	executor := &LLMExecutor{
		Template:      templateContent,
		SystemMessage: systemMessage,
		Client:        client,
	}

	return flow.NewStep(executor, nil, client), nil
}

// NewLLMStepWithTemplate creates a new LLM step with a template string.
// The template should be a template string.
func NewLLMStepWithTemplate(template string, client llm.Client, systemMessage string) (*flow.Step, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	executor := &LLMExecutor{
		Template:      template,
		SystemMessage: systemMessage,
		Client:        client,
	}

	return flow.NewStep(executor, nil, client), nil
}

// DeepSeekStyleResponseFilter is a response filter that extracts code blocks from the response.
// It is designed to work with DeepSeek-style responses.
type DeepSeekStyleResponseFilter struct {
	// CodeBlockIndex is the index of the code block to extract.
	// If it is -1, all code blocks will be extracted and joined with newlines.
	CodeBlockIndex int `json:"codeBlockIndex" yaml:"codeBlockIndex" mapstructure:"codeBlockIndex"`
}

// Init initializes the DeepSeekStyleResponseFilter.
// This implementation has no initialization requirements.
func (filter *DeepSeekStyleResponseFilter) Init() error {
	return nil
}

// Filter filters the response by extracting code blocks.
// If CodeBlockIndex is -1, all code blocks will be extracted and joined with newlines.
// Otherwise, only the code block at the specified index will be extracted.
func (filter *DeepSeekStyleResponseFilter) Filter(response string) (string, error) {
	codeBlocks := utils.ExtractAllCodeBlocks(response)
	if len(codeBlocks) == 0 {
		return "", fmt.Errorf("no code blocks found in response: %s", response)
	}

	if filter.CodeBlockIndex == -1 {
		return strings.Join(codeBlocks, "\n"), nil
	}

	if filter.CodeBlockIndex >= len(codeBlocks) {
		return "", fmt.Errorf("code block index %d is out of range", filter.CodeBlockIndex)
	}

	return codeBlocks[filter.CodeBlockIndex], nil
}