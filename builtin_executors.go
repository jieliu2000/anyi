package anyi

import (
	"errors"
	"strings"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

type DecoratedStepExecutor struct {
	ExecutorImpl flow.StepExecutor                                                              `json:"-" yaml:"-" mapstructure:"-"`
	PreRun       func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	PostRun      func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	With         *ExecutorConfig                                                                `json:"with" yaml:"with" mapstructure:"with"`
}

// Init initializes the DecratedStepExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If any of the checks fail, an error is returned.
// If all checks pass, it calls the Init method of the executor.
func (executor *DecoratedStepExecutor) Init() error {
	if executor.With != nil && executor.ExecutorImpl == nil {
		impl, err := NewExecutorFromConfig(executor.With)
		if err != nil {
			return err
		}
		executor.ExecutorImpl = impl
	}
	if executor.ExecutorImpl == nil {
		return errors.New("no executor provided")
	}

	if executor.PreRun == nil && executor.PostRun == nil {
		return errors.New("no pre or post run function provided")
	}
	return executor.ExecutorImpl.Init()
}

// The Run function executes the given step within the provided flow context.
// Parameters:
// - flowContext flow.FlowContext: The flow context in which the step will be executed.
// - step *flow.Step: The step to be executed.
// Return values:
// - *flow.FlowContext: The updated flow context after executing the step.
// - error: If an error occurs during execution, the corresponding error message is returned.
func (executor *DecoratedStepExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	context := &flowContext
	if executor.ExecutorImpl == nil {
		return context, errors.New("no executor provided")
	}
	if executor.PreRun != nil {
		var err error
		context, err := executor.PreRun(*context, step)
		if err != nil {
			return context, err
		}
	}
	context, err := executor.ExecutorImpl.Run(*context, step)
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
	OutputJSON        bool   `json:"outputJSON" yaml:"outputJSON" mapstructure:"outputJSON"`
	Trim              string `json:"trim" yaml:"trim" mapstructure:"trim"`
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

func (executor *LLMStepExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	if step == nil {
		return nil, errors.New("no step provided")
	}

	if step.GetClient() == nil {
		step.ClientImpl = flowContext.Flow.ClientImpl
	}
	if step.ClientImpl == nil {
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
		input, err = executor.TemplateFormatter.Format(flowContext)
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

	var options *chat.ChatOptions

	messages = append(messages, chat.NewUserMessage(input))

	if executor.OutputJSON {
		options = &chat.ChatOptions{
			Format: "json",
		}
	}

	output, _, err := step.ClientImpl.Chat(messages, options)
	if err != nil {
		return nil, err
	}

	flowContext.Text = output.Content
	if executor.Trim != "" {
		flowContext.Text = strings.Trim(flowContext.Text, executor.Trim)
	}
	return &flowContext, nil
}

func NewLLMStepWithTemplateFile(templateFilePath string, systemMessage string, client llm.Client) (*flow.Step, error) {

	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatterFromFile(templateFilePath)
	if err != nil {
		return nil, err
	}
	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := flow.NewStep(executor, nil, client)

	return step, nil
}

func NewLLMStepWithTemplate(tmplate string, systemMessage string, client llm.Client) (*flow.Step, error) {
	// Create a new formatter with the template
	formatter, err := chat.NewPromptTemplateFormatter(tmplate)
	if err != nil {
		return nil, err
	}

	executor := &LLMStepExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := flow.NewStep(executor, nil, client)
	return step, nil
}
