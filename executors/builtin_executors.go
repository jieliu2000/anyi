package executors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
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

// SetVariablesExecutor is an executor that sets multiple variables in the flow context at once
type SetVariablesExecutor struct {
	// Variables to set, map of variable names to their corresponding values
	// Example: { "var1": "value1", "var2": 123, "var3": true }
	Variables map[string]any `json:"variables" yaml:"variables" mapstructure:"variables"`
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
// In both cases, the extracted thinking content is stored in the FlowContext.Think field.
func (executor *DeepSeekStyleResponseFilter) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// Extract <think> tag content
	thinkMatch := executor.re.FindStringSubmatch(flowContext.Text)
	thinkContent := ""
	if len(thinkMatch) > 0 {
		thinkContent = thinkMatch[0]
		// Store the thinking content in the FlowContext.Think field
		flowContext.Think = thinkContent
	}

	// Extract and trim other content (remove <think> tags)
	resultContent := strings.TrimSpace(executor.re.ReplaceAllString(flowContext.Text, ""))

	// If OutputJSON is true, return both thinking and result in JSON format
	if executor.OutputJSON {
		// Set return content in JSON format
		flowContext.Text = fmt.Sprintf(`{"think": "%s", "result": "%s"}`, thinkContent, resultContent)
		return &flowContext, nil
	}

	// Default behavior: set the clean content as Text
	flowContext.Text = resultContent
	return &flowContext, nil
}

// DelayExecutor is an executor that delays execution for a specified number of microseconds
type DelayExecutor struct {

	// Milliseconds is the number of milliseconds to delay (alternative to Microseconds)
	Milliseconds int `json:"milliseconds" yaml:"milliseconds" mapstructure:"milliseconds"`
}

// Init initializes the DelayExecutor
func (executor *DelayExecutor) Init() error {
	// Validate that either Milliseconds is set
	if executor.Milliseconds <= 0 {
		return errors.New("milliseconds must be set")
	}
	return nil
}

// Run delays execution for the specified time
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The current workflow step
//
// Returns:
//   - The unchanged flow context
//   - Any error encountered during execution
func (executor *DelayExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// Calculate the delay duration
	var delay = time.Millisecond * time.Duration(executor.Milliseconds)

	log.Debugf("Delaying execution for %v", delay)

	// Perform the delay
	time.Sleep(delay)

	log.Debugf("Delay completed")

	return &flowContext, nil
}

// Run executes the variable setting operation for multiple variables at once
//
// Parameters:
//   - flowContext: The current flow context
//   - step: The current workflow step
//
// Returns:
//   - Updated flow context with the new variables set
//   - Any error encountered during execution
//
// Example usage in configuration:
//
//	{
//	  "type": "setVariables",
//	  "variables": {
//	    "username": "john_doe",
//	    "age": 30,
//	    "isActive": true,
//	    "preferences": { "theme": "dark", "notifications": false }
//	  }
//	}
func (executor *SetVariablesExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// Check if variables are immutable for this step
	if step != nil && step.VarsImmutable {
		// If variables are immutable, return the context unchanged
		log.Debug("Variables are immutable for this step, skipping variable modification")
		return &flowContext, nil
	}

	// Ensure Variables is initialized in flowContext
	if flowContext.Variables == nil {
		flowContext.Variables = make(map[string]any)
	}

	// Set multiple variables simultaneously. Each key in the map is a variable name,
	// and its corresponding value will be assigned to that variable.
	// For example, with Variables = {"name": "John", "age": 30, "active": true},
	// this will create/update three different variables with their respective values.
	if executor.Variables != nil {
		for name, value := range executor.Variables {
			if name == "" {
				continue // Skip empty variable names
			}

			// Set the variable (always overwrite existing values)
			flowContext.Variables[name] = value
		}
	}

	return &flowContext, nil
}

// Init initializes SetVariablesExecutor
func (executor *SetVariablesExecutor) Init() error {
	return nil
}
