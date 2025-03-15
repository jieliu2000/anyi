package anyi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/shello"
)

// SetContextExecutor is an executor that sets values in the flow context.
// It can modify the Text field and Memory object in the flow context.
type SetContextExecutor struct {
	Text   string               `json:"text" yaml:"text" mapstructure:"text"`
	Memory flow.ShortTermMemory `json:"memory" yaml:"memory" mapstructure:"memory"`

	Force bool `json:"force" yaml:"force" mapstructure:"force"`
}

// Init initializes the SetContextExecutor.
// This implementation has no initialization requirements.
func (executor *SetContextExecutor) Init() error {
	return nil
}

// Run sets the text and memory of the flow context.
// If the Force flag is set to true, it will override the existing text and memory.
// Otherwise, it will only set the text and memory if they are not empty.
//
// Parameters:
//   - flowContext: The current flow context to modify
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with modified text and memory
//   - Any error encountered during execution
func (executor *SetContextExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {

	if executor.Text != "" || executor.Force {
		flowContext.Text = executor.Text
	}

	if executor.Memory != nil || executor.Force {
		flowContext.Memory = executor.Memory
	}
	return &flowContext, nil
}

// DecoratedExecutor is an executor that wraps another executor with pre-run and post-run functions.
// This allows for adding behavior before and after the execution of the wrapped executor.
type DecoratedExecutor struct {
	ExecutorImpl flow.StepExecutor                                                              `json:"-" yaml:"-" mapstructure:"-"`
	PreRun       func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	PostRun      func(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) `json:"-" yaml:"-" mapstructure:"-"`
	With         *ExecutorConfig                                                                `json:"with" yaml:"with" mapstructure:"with"`
}

// Init initializes the DecoratedExecutor.
// It checks if an executor is provided and if pre or post run functions are set.
// If a configuration is provided but no executor, it creates one from the configuration.
//
// Returns:
//   - An error if no executor is provided or if neither pre nor post run functions are set
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

// Run executes the step within the provided flow context.
// It applies the pre-run function (if set), then the wrapped executor, then the post-run function (if set).
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The step to be executed
//
// Returns:
//   - Updated flow context after execution
//   - Any error encountered during execution
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

// ConditionalFlowExecutor is an executor that routes flow execution based on conditions.
// It uses the text in the flow context to determine which flow to execute next.
type ConditionalFlowExecutor struct {
	Switch map[string]string `json:"switch" yaml:"switch" mapstructure:"switch"`
	Trim   string            `json:"trim" yaml:"trim" mapstructure:"trim"`
}

// Init initializes the ConditionalFlowExecutor.
// It checks the provided switches and retrieves the corresponding flows.
//
// Returns:
//   - An error if no switches are provided or if any referenced flow cannot be found
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

// Run executes the flow based on the condition in the flow context.
// The text in the flow context is used as a key to find the next flow to execute.
//
// Parameters:
//   - flowContext: The current flow context containing the condition text
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context after the selected flow executes
//   - An error if no matching flow is found or if the flow execution fails
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

// RunCommandExecutor is an executor that runs system commands.
// It executes the command specified in the flow context's Text field.
type RunCommandExecutor struct {
	Silent          bool   `json:"silent" yaml:"silent" mapstructure:"silent"`
	OutputToContext bool   `json:"outputToContext" yaml:"outputToContext" mapstructure:"outputToContext"`
	Path            string `json:"path" yaml:"path" mapstructure:"path"`
}

// Init initializes the RunCommandExecutor.
// This implementation has no initialization requirements.
func (executor *RunCommandExecutor) Init() error {
	return nil
}

// Run executes the command specified in the flow context's Text field.
//
// Parameters:
//   - flowContext: The flow context containing the command to execute in its Text field
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context (with command output in Text field if OutputToContext is true)
//   - Any error encountered during command execution
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

// LLMExecutor is an executor that sends prompts to large language models.
// It supports template-based prompts, system messages, and JSON output formatting.
type LLMExecutor struct {
	Template          string `json:"template" yaml:"template" mapstructure:"template"`
	TemplateFile      string `json:"templateFile" yaml:"templateFile" mapstructure:"templateFile"`
	TemplateFormatter *chat.PromptyTemplateFormatter
	SystemMessage     string `json:"systemMessage" yaml:"systemMessage" mapstructure:"systemMessage"`
	OutputJSON        bool   `json:"outputJSON" yaml:"outputJSON" mapstructure:"outputJSON"`
	Trim              string `json:"trim" yaml:"trim" mapstructure:"trim"`
}

// Init initializes the LLMExecutor by creating template formatters.
// It creates a formatter based on either the Template string or TemplateFile.
//
// Returns:
//   - An error if neither Template nor TemplateFile is provided, or if formatter creation fails
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

// Run sends a prompt to a language model and processes the response.
// It formats the prompt using the template formatter, adds system messages if provided,
// handles image URLs if present, and sends the messages to the LLM client.
//
// Parameters:
//   - flowContext: The flow context containing data for prompt generation
//   - step: The workflow step containing the client to use
//
// Returns:
//   - Updated flow context with the model's response in the Text field
//   - Any error encountered during execution
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

// NewLLMStepWithTemplateFile creates a new workflow step with an LLM executor
// that uses a template from a file.
//
// Parameters:
//   - templateFilePath: Path to the file containing the prompt template
//   - systemMessage: Optional system message to include in the conversation
//   - client: LLM client to use for this step
//
// Returns:
//   - A new workflow step configured with the template and client
//   - Any error encountered during creation
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

// NewLLMStepWithTemplate creates a new workflow step with an LLM executor
// that uses an inline template string.
//
// Parameters:
//   - tmplate: String containing the prompt template
//   - systemMessage: Optional system message to include in the conversation
//   - client: LLM client to use for this step
//
// Returns:
//   - A new workflow step configured with the template and client
//   - Any error encountered during creation
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

// DeepSeekStyleResponseFilter is an executor that processes model responses containing <think> tags.
// It can either extract thinking content for debugging/analysis or clean it up for display to end users.
// This is particularly useful for models like DeepSeek that support explicit thinking steps in their responses.
type DeepSeekStyleResponseFilter struct {
	re         *regexp.Regexp
	OutputJSON bool // When true, returns both thinking and result content in JSON format
}

// Init initializes the DeepSeekStyleResponseFilter by compiling the regular expression
// used to identify and extract <think> tag content from model responses.
// Returns an error if the regular expression fails to compile.
func (executor *DeepSeekStyleResponseFilter) Init() error {
	// Compile regular expression to match <think> tag content
	re, err := regexp.Compile(`(?s)<think>.*?</think>`)
	if err != nil {
		return err
	}
	executor.re = re
	return nil
}

// Run processes the text in the flow context to handle <think> tags based on configuration.
// Parameters:
//   - flowContext: The current flow context containing the text to process
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with processed text
//   - Any error encountered during processing
//
// When OutputJSON is true, it extracts <think> content and returns both thinking and result
// in JSON format. Otherwise, it simply removes <think> tags and returns the cleaned content.
func (executor *DeepSeekStyleResponseFilter) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// If OutputJSON is true, extract <think> tag content and return in JSON format
	if executor.OutputJSON {
		// Extract <think> tag content
		thinkMatch := executor.re.FindStringSubmatch(flowContext.Text)
		thinkContent := ""
		if len(thinkMatch) > 0 {
			thinkContent = thinkMatch[0]
		}

		// Extract and trim other content
		resultContent := strings.TrimSpace(executor.re.ReplaceAllString(flowContext.Text, ""))

		// Set return content in JSON format
		flowContext.Text = fmt.Sprintf(`{"think": "%s", "result": "%s"}`, thinkContent, resultContent)
		return &flowContext, nil
	}

	// Default behavior: remove <think> tags and trim result
	flowContext.Text = strings.TrimSpace(executor.re.ReplaceAllString(flowContext.Text, ""))
	return &flowContext, nil
}
