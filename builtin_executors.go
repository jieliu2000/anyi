package anyi

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/shello"
)

type SetContextExecutor struct {
	Text   string               `json:"text" yaml:"text" mapstructure:"text"`
	Memory flow.ShortTermMemory `json:"memory" yaml:"memory" mapstructure:"memory"`

	Force bool `json:"force" yaml:"force" mapstructure:"force"`
}

func (executor *SetContextExecutor) Init() error {
	return nil
}

// Sets the text and memory of the flow context. If the Force flag is set to true, it will override the existing text and memory. Otherwise, it will only set the text and memory if they are not empty.
func (executor *SetContextExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {

	if executor.Text != "" || executor.Force {
		flowContext.Text = executor.Text
	}

	if executor.Memory != nil || executor.Force {
		flowContext.Memory = executor.Memory
	}
	return &flowContext, nil
}

type DecoratedExecutor struct {
	ExecutorImpl flow.StepExecutor                                                              `json:"-" yaml:"-" mapstructure:"-"`
	PreRun       func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	PostRun      func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	With         *ExecutorConfig                                                                `json:"with" yaml:"with" mapstructure:"with"`
}

// Init initializes the DecratedStepExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If any of the checks fail, an error is returned.
// If all checks pass, it calls the Init method of the executor.
func (executor *DecoratedExecutor) Init() error {
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
func (executor *DecoratedExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
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

type ConditionalFlowExecutor struct {
	Switch map[string]string `json:"switch" yaml:"switch" mapstructure:"switch"`
	Trim   string            `json:"trim" yaml:"trim" mapstructure:"trim"`
}

// The Init function initializes the ConditionalFlowExecutor.
// It checks the provided switches and retrieves the corresponding flows.
// If any error occurs during the initialization process, an error is returned.
func (executor *ConditionalFlowExecutor) Init() error {
	if executor.Switch == nil || len(executor.Switch) == 0 {
		return errors.New("no switch provided")
	}
	for _, value := range executor.Switch {
		flow, err := GetFlow(value)
		if err != nil {
			return errors.Join(err, errors.New("failed to get flow "+value))
		}
		if flow == nil {
			return errors.New("flow " + value + " not found")
		}
	}
	return nil
}

// Run is the function of ConditionalFlowExecutor, which is responsible for executing the flow based on the given condition.
// Parameters:
// - flowContext flow.FlowContext: The current flow context.
// - step *flow.Step: The current step of the flow.
// Return value:
// - *flow.FlowContext: The updated flow context after execution.
// - error: If an error occurs during execution, the corresponding error message is returned.
func (executor *ConditionalFlowExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	condition := flowContext.Text
	if executor.Trim != "" {
		condition = strings.Trim(condition, executor.Trim)
	}
	flowName, ok := executor.Switch[condition]
	if !ok || flowName == "" {
		return &flowContext, fmt.Errorf("no matching flow found for condition %s", condition)
	}

	flow, err := GetFlow(flowName)
	if err != nil {
		return &flowContext, err
	}
	if flow == nil {
		return &flowContext, fmt.Errorf("flow %s not found", flowName)
	}

	return flow.Run(flowContext)
}

type RunCommandExecutor struct {
	Silent          bool   `json:"silent" yaml:"silent" mapstructure:"silent"`
	OutputToContext bool   `json:"outputToContext" yaml:"outputToContext" mapstructure:"outputToContext"`
	Path            string `json:"path" yaml:"path" mapstructure:"path"`
}

func (executor *RunCommandExecutor) Init() error {
	return nil
}

// Run executes the provided command in the Text field of the flow context.
// Parameters:
// - flowContext flow.FlowContext: The current flow context. The Text field of the flow context is used as the command to be executed.
// - step *flow.Step: The current step in the flow.
// Return values:
// - *flow.FlowContext: This executor doesn't change anything in the flow context.
// - error: Any error that occurred during command execution.
func (executor *RunCommandExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	commandText := flowContext.Text
	if commandText == "" {
		return &flowContext, errors.New("no command provided")
	}
	if !executor.Silent {
		log.Infof("Running command: %s", commandText)
	}

	outputString, _, err := shello.RunOutputWithDir(commandText, executor.Path)

	if err != nil {
		return &flowContext, err
	}
	if !executor.Silent {
		log.Infof("%s\n", outputString)
	}
	if executor.OutputToContext {
		flowContext.Text = outputString
	}
	return &flowContext, nil
}

type LLMExecutor struct {
	Template          string `json:"template" yaml:"template" mapstructure:"template"`
	TemplateFile      string `json:"templateFile" yaml:"templateFile" mapstructure:"templateFile"`
	TemplateFormatter *chat.PromptyTemplateFormatter
	SystemMessage     string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`
	OutputJSON        bool   `json:"outputJSON" yaml:"outputJSON" mapstructure:"outputJSON"`
	Trim              string `json:"trim" yaml:"trim" mapstructure:"trim"`
}

func (executor *LLMExecutor) Init() error {
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

func (executor *LLMExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
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

	if len(flowContext.ImageURLs) > 0 {
		msg := chat.Message{
			Role: "user",
		}

		if input != "" {
			msg.ContentParts = append(msg.ContentParts, chat.ContentPart{
				Text: input,
			})
		}

		for _, imgURL := range flowContext.ImageURLs {
			msg.ContentParts = append(msg.ContentParts, chat.ContentPart{
				ImageUrl: imgURL,
			})
		}

		messages = append(messages, msg)
	} else {
		messages = append(messages, chat.NewUserMessage(input))
	}

	var options *chat.ChatOptions
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
	executor := &LLMExecutor{
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

	executor := &LLMExecutor{
		TemplateFormatter: formatter,
		SystemMessage:     systemMessage,
	}
	step := flow.NewStep(executor, nil, client)
	return step, nil
}
