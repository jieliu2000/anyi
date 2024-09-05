package flow

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

type StepExecutor interface {
	Init() error

	Run(memory ShortTermMemory, Step *Step) (*ShortTermMemory, error)
}

type LLMStepExecutor struct {
	Template          string `json:"template" yaml:"template" mapstructure:"template"`
	TemplateFile      string `json:"templateFile" yaml:"templateFile" mapstructure:"templateFile"`
	TemplateFormatter *chat.PromptyTemplateFormatter
	SystemMessage     string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`
}

func (executor *LLMStepExecutor) Init() error {
	if executor.TemplateFormatter == nil && executor.Template != "" {
		formatter, err := chat.NewPromptTemplateFormatter(executor.Template)
		if err != nil {
			return err
		}
		executor.TemplateFormatter = formatter
	}
	if executor.TemplateFormatter == nil && executor.TemplateFile != "" {
		formatter, err := chat.NewPromptTemplateFormatterFromFile(executor.TemplateFile)
		if err != nil {
			return err
		}
		executor.TemplateFormatter = formatter
	}
	return nil
}

func (executor *LLMStepExecutor) Run(memory ShortTermMemory, step *Step) (*ShortTermMemory, error) {
	if step == nil {
		return nil, errors.New("no step provided")
	}

	if step.clientImpl == nil {
		step.clientImpl = memory.flow.clientImpl
	}
	if step.clientImpl == nil {
		return nil, errors.New("no client set for flow step")
	}

	if executor.TemplateFormatter == nil && executor.Template != "" {
		var err error
		executor.TemplateFormatter, err = chat.NewPromptTemplateFormatter(executor.Template)
		if err != nil {
			return nil, err
		}
	}

	if executor.TemplateFormatter == nil && executor.TemplateFile != "" {
		var err error
		executor.TemplateFormatter, err = chat.NewPromptTemplateFormatterFromFile(executor.TemplateFile)
		if err != nil {
			return nil, err
		}
	}

	var input string
	if executor.TemplateFormatter != nil {
		var err error
		input, err = executor.TemplateFormatter.Format(memory)
		if err != nil {
			return nil, err
		}
	} else {
		input = memory.Text
	}

	messages := make([]chat.Message, 0, 2)
	if executor.SystemMessage != "" {
		messages = append(messages, chat.NewSystemMessage(executor.SystemMessage))
	}
	messages = append(messages, chat.NewUserMessage(input))

	output, _, err := step.clientImpl.Chat(messages, nil)
	if err != nil {
		return nil, err
	}

	memory.Text = output.Content
	return &memory, nil
}

func NewLLMStepWithTemplateFile(templateFilePath string, systemMessage string, client llm.Client) (*Step, error) {

	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatterFromFile(templateFilePath)
	if err != nil {
		return nil, err
	}
	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := NewStep(executor, nil, client)

	return step, nil
}

func NewLLMStepWithTemplate(tmplate string, systemMessage string, client llm.Client) (*Step, error) {
	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatter(tmplate)
	if err != nil {
		return nil, err
	}

	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := NewStep(executor, nil, client)
	return step, nil
}
