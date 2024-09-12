package flow

import (
	"errors"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

type StepExecutor interface {
	Init() error
	Run(flowContext FlowContext, Step *Step) (*FlowContext, error)
}

type DecratedStepExecutor struct {
	WithExecutor StepExecutor `json:"withExecutor" yaml:"withExecutor" mapstructure:"withExecutor"`
	PreRun       func(flowContext FlowContext, step *Step) (*FlowContext, error)
	PostRun      func(flowContext FlowContext, step *Step) (*FlowContext, error)
}

// Init initializes the DecratedStepExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If any of the checks fail, an error is returned.
// If all checks pass, it calls the Init method of the executor.
func (executor *DecratedStepExecutor) Init() error {
	if executor.WithExecutor == nil {
		return errors.New("no executor provided")
	}

	if executor.PreRun == nil && executor.PostRun == nil {
		return errors.New("no pre or post run function provided")
	}
	return executor.WithExecutor.Init()
}

// The Run function executes the given step within the provided flow context.
// Parameters:
// - flowContext FlowContext: The flow context in which the step will be executed.
// - step *Step: The step to be executed.
// Return values:
// - *FlowContext: The updated flow context after executing the step.
// - error: If an error occurs during execution, the corresponding error message is returned.
func (executor *DecratedStepExecutor) Run(flowContext FlowContext, step *Step) (*FlowContext, error) {
	context := &flowContext
	if executor.WithExecutor == nil {
		return context, errors.New("no executor provided")
	}
	if executor.PreRun != nil {
		var err error
		context, err := executor.PreRun(*context, step)
		if err != nil {
			return context, err
		}
	}
	context, err := executor.WithExecutor.Run(*context, step)
	if executor.PostRun != nil {

		context, err = executor.PostRun(*context, step)
	}
	return context, err
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
		return nil
	}
	if executor.TemplateFormatter == nil && executor.TemplateFile != "" {
		formatter, err := chat.NewPromptTemplateFormatterFromFile(executor.TemplateFile)
		if err != nil {
			return err
		}
		executor.TemplateFormatter = formatter
		return nil
	}
	return errors.New("no required parameters. You need to set either template or templateFile")
}

func (executor *LLMStepExecutor) Run(flowContext FlowContext, step *Step) (*FlowContext, error) {
	if step == nil {
		return nil, errors.New("no step provided")
	}

	if step.clientImpl == nil {
		step.clientImpl = flowContext.flow.clientImpl
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
		if flowContext.Memory == nil {
			return nil, errors.New("no non-text data provided for template execution")
		}
		input, err = executor.TemplateFormatter.Format(flowContext.Memory)
		if err != nil {
			return nil, err
		}
	} else {
		input = flowContext.Text
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

	flowContext.Text = output.Content
	return &flowContext, nil
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
