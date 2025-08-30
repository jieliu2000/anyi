package executors

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
)

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

// DeepSeekStyleResponseFilter is an executor that processes model responses containing
// DeepSeek-style special formatting (e.g., triple braces). It extracts content
// between the opening and closing braces and removes any additional formatting.
type DeepSeekStyleResponseFilter struct {
	re *regexp.Regexp
}

// Init initializes the DeepSeekStyleResponseFilter by compiling a regular expression
// to match content between triple braces.
//
// Returns:
//   - An error if the regular expression compilation fails
func (executor *DeepSeekStyleResponseFilter) Init() error {
	re, err := regexp.Compile(`\{\{\{([\s\S]*?)\}\}\}`)
	if err != nil {
		return err
	}
	executor.re = re
	return nil
}

// Run processes the text in the flow context to handle "DeepSeek-style" responses
// with triple-braced content. It extracts the content between the braces and
// updates the flow context with the cleaned text.
//
// Parameters:
//   - flowContext: The flow context containing the text to process
//   - step: The workflow step (unused in this executor)
//
// Returns:
//   - Updated flow context with the processed text
//   - Any error encountered during processing
func (executor *DeepSeekStyleResponseFilter) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	matches := executor.re.FindStringSubmatch(flowContext.Text)
	if len(matches) > 1 {
		flowContext.Text = matches[1]
	}
	return &flowContext, nil
}