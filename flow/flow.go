package flow

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/llm"
)

const (
	DefaultMaxRetryTimes = 3
)

type Flow struct {
	Name string
	Description string `yaml:"description,omitempty" json:"description,omitempty" toml:"description,omitempty"`
	Steps []Step
	ClientImpl llm.Client
	Variables map[string]any
}
type StepExecutor interface {
	Init() error
	Run(flowContext FlowContext, Step *Step) (*FlowContext, error)
}

// StepValidator is the interface for validators of step output.
// In a flow if a step validator is set, the output of the step will be checked against the validator's Validate method.
type StepValidator interface {
	Init() error
	Validate(stepOutput string, Step *Step) bool
}

func NewFlow(client llm.Client, name string, steps ...Step) (*Flow, error) {

	if name == "" {
		return nil, errors.New("flow name cannot be empty")
	}

	flow := &Flow{Steps: steps, Name: name, ClientImpl: client}

	return flow, nil
}

type Step struct {
	ClientImpl         llm.Client
	validateClientImpl llm.Client

	Executor StepExecutor

	Validator     StepValidator
	runTimes      int
	MaxRetryTimes int
	Name          string

	// Controls whether variables can be modified during step execution
	// When true, variables cannot be modified
	VarsImmutable bool `json:"varsImmutable,omitempty" yaml:"varsImmutable,omitempty" mapstructure:"varsImmutable,omitempty"`
	
	// Controls whether text can be modified during step execution
	// When true, text cannot be modified
	TextImmutable bool `json:"textImmutable,omitempty" yaml:"textImmutable,omitempty" mapstructure:"textImmutable,omitempty"`
	
	// Controls whether memory can be modified during step execution
	// When true, memory cannot be modified
	MemoryImmutable bool `json:"memoryImmutable,omitempty" yaml:"memoryImmutable,omitempty" mapstructure:"memoryImmutable,omitempty"`
}

// GetClient function returns the client of the Step.
// If the clientImpl of the Step is not nil, it returns the clientImpl.
// Otherwise, it returns nil.
func (step *Step) GetClient() llm.Client {
	if step.ClientImpl != nil {
		return step.ClientImpl
	}
	return nil
}

type ShortTermMemory any

// FlowContext is the flowContext for a flow. It will be passed to each flow step.
// The Text field provides the input and output string for a step. For example, before the step runs, the "Text" field is the input string (which might be formatted by the template formatter). After the step runs, the "Text" field is the output string.
// The Data field could be any data you want to pass between steps.
type FlowContext struct {
	Text      string
	Memory    ShortTermMemory
	Variables map[string]any
	Flow      *Flow
	ImageURLs []string
	Think     string // Stores thinking content extracted from <think> tags in model output
}

func (fc *FlowContext) UnmarshalJsonText(entity any) error {
	return json.Unmarshal([]byte(fc.Text), entity)
}

// GetVariable gets the value of a variable. Returns nil if the variable doesn't exist
func (fc *FlowContext) GetVariable(name string) any {
	if fc.Variables == nil {
		return nil
	}
	return fc.Variables[name]
}

// SetVariable sets the value of a variable
func (fc *FlowContext) SetVariable(name string, value any) {
	if fc.Variables == nil {
		fc.Variables = make(map[string]any)
	}
	fc.Variables[name] = value
}

// GetVariableString gets the string value of a variable. Returns the defaultValue if the variable
// doesn't exist or is not a string
func (fc *FlowContext) GetVariableString(name string, defaultValue string) string {
	value := fc.GetVariable(name)
	if value == nil {
		return defaultValue
	}
	strValue, ok := value.(string)
	if !ok {
		return defaultValue
	}
	return strValue
}

// GetVariableInt gets the integer value of a variable. Returns the defaultValue if the variable
// doesn't exist or is not an integer
func (fc *FlowContext) GetVariableInt(name string, defaultValue int) int {
	value := fc.GetVariable(name)
	if value == nil {
		return defaultValue
	}
	intValue, ok := value.(int)
	if !ok {
		return defaultValue
	}
	return intValue
}

// GetVariableBool gets the boolean value of a variable. Returns the defaultValue if the variable
// doesn't exist or is not a boolean
func (fc *FlowContext) GetVariableBool(name string, defaultValue bool) bool {
	value := fc.GetVariable(name)
	if value == nil {
		return defaultValue
	}
	boolValue, ok := value.(bool)
	if !ok {
		return defaultValue
	}
	return boolValue
}

// WithVariable creates a copy of FlowContext with a new variable set
// The original FlowContext is not modified
func (fc *FlowContext) WithVariable(name string, value any) *FlowContext {
	newContext := &FlowContext{
		Text:      fc.Text,
		Memory:    fc.Memory,
		Flow:      fc.Flow,
		ImageURLs: fc.ImageURLs,
		Think:     fc.Think,
		Variables: make(map[string]any),
	}

	// Copy all existing variables
	if fc.Variables != nil {
		for k, v := range fc.Variables {
			newContext.Variables[k] = v
		}
	}

	// Set the new variable
	newContext.Variables[name] = value

	return newContext
}

func NewFlowContext(flowContext string, memory any) *FlowContext {
	return &FlowContext{Text: flowContext, Memory: memory, Variables: make(map[string]any)}
}

func (flow *Flow) NewFlowContext(flowContext string, memory any) *FlowContext {
	return &FlowContext{Text: flowContext, Memory: memory, Flow: flow, Variables: make(map[string]any)}
}

func NewStep(executor StepExecutor, validator StepValidator, client llm.Client) *Step {
	return &Step{
		Executor:      executor, 
		Validator:     validator, 
		ClientImpl:    client, 
		MaxRetryTimes: DefaultMaxRetryTimes,
		VarsImmutable: false,
		TextImmutable: false,
		MemoryImmutable:  false,
	}
}

// Create a new flow step with executor and validator.
// Parameters:
//
//   - stepConfig: the configuration of the step. This parameter is used to pass parameters to the executor and validator. This can be a flexible type object which will be used by the executor and validator. We provide a default StepConfig in Anyi, see [LLMFlowStepConfig] for more details.
//   - executor: the executor of the step. See [StepExecutor] for more details.
//   - validator: the validator of the step. Set this parameter to nil if you don't want to validate the step output. See [StepValidator] for more details.
//   - client: the default client of the step. If set to nil, the default client of the flow will be used.
//   - validateClient: the client used to validate the step output. If set to nil, the default client of the step will be used. If the step doesn't have a default client, the default client of the flow will be used.
func NewStepWithValidator(stepConfig any, executor StepExecutor, validator StepValidator, client llm.Client, validateClient llm.Client) *Step {
	return &Step{
		Executor:           executor,
		Validator:          validator,
		runTimes:           0,
		MaxRetryTimes:      DefaultMaxRetryTimes,
		ClientImpl:         client,
		validateClientImpl: validateClient,
		VarsImmutable:      false,
		TextImmutable:      false,
		MemoryImmutable:       false,
	}
}

func tryStep(step *Step, flowContext FlowContext) (*FlowContext, error) {
	var err error

	log.Debug("Running step ", step, ".")
	// Store original values if immutability is enabled
	originalVars := make(map[string]any)
	originalText := flowContext.Text
	var originalMemory any
	
	// Copy original variables if immutable
	if step.VarsImmutable && flowContext.Variables != nil {
		for k, v := range flowContext.Variables {
			originalVars[k] = v
		}
	}
	
	// Store original memory if immutable
	if step.MemoryImmutable {
		originalMemory = flowContext.Memory
	}
	
	// Run the step and get the updated flowContext
	result, err := step.Executor.Run(flowContext, step)
	step.runTimes++
	if err != nil {
		return result, err
	}
	
	// Apply immutability constraints if needed
	if result != nil {
		// Restore original variables if immutable
		if step.VarsImmutable && flowContext.Variables != nil {
			result.Variables = originalVars
		}
		
		// Restore original text if immutable
		if step.TextImmutable {
			result.Text = originalText
		}
		
		// Restore original memory if immutable
		if step.MemoryImmutable {
			result.Memory = originalMemory
		}
	}

	// Sync variables from flowContext to flow
	if result != nil && result.Variables != nil && result.Flow != nil {
		if result.Flow.Variables == nil {
			result.Flow.Variables = make(map[string]any)
		}
		for k, v := range result.Variables {
			result.Flow.Variables[k] = v
		}
	}

	if step.runTimes > step.MaxRetryTimes+1 {
		log.Error("Step retry times exceeded, returning error.")
		return result, errors.New("step retry times exceeded")
	}
	if step.Validator != nil {
		// Validate the step output
		if step.Validator.Validate(result.Text, step) {
			// If the step output is valid, update flowContext and continue to the next step
			return result, nil
		} else {
			// Otherwise, try again
			return tryStep(step, *result)
		}
	}
	// If no validator is set, simply return the updated context.
	return result, nil
}

func (flow *Flow) RunWithInput(input string) (*FlowContext, error) {
	// Create a new flowContext with the input
	flowContext := FlowContext{
		Text:      input,
		Variables: make(map[string]any),
	}

	return flow.Run(flowContext)
}

// RunWithMemory runs the flow with the provided memory object.
// This is a convenience method for creating a flow context with memory and running the flow.
//
// Parameters:
//   - memory: The memory object to use in the flow context
//
// Returns:
//   - The updated flow context after flow execution
//   - Any error encountered during flow execution
func (flow *Flow) RunWithMemory(memory ShortTermMemory) (*FlowContext, error) {
	// Create a new flowContext with the memory
	flowContext := FlowContext{
		Memory:    memory,
		Variables: make(map[string]any),
	}

	return flow.Run(flowContext)
}

// RunWithVariables runs the flow with the provided variables.
// This is a convenience method for creating a flow context with variables and running the flow.
//
// Parameters:
//   - variables: Map of variable names to values to use in the flow context
//
// Returns:
//   - The updated flow context after flow execution
//   - Any error encountered during flow execution
//
// GetVariables returns all variables in the flow
func (flow *Flow) GetVariables() map[string]any {
	if flow.Variables == nil {
		return make(map[string]any)
	}
	return flow.Variables
}

// GetVariable gets the value of a variable and returns whether it exists
func (flow *Flow) GetVariable(key string) (any, bool) {
	if flow.Variables == nil {
		return nil, false
	}
	value, exists := flow.Variables[key]
	return value, exists
}

// SetVariable sets the value of a variable
func (flow *Flow) SetVariable(key string, value any) {
	if flow.Variables == nil {
		flow.Variables = make(map[string]any)
	}
	flow.Variables[key] = value
}

func (flow *Flow) RunWithVariables(variables map[string]any) (*FlowContext, error) {

	if variables == nil {
		variables = make(map[string]any)
	}
	// Create a new flowContext with the variables
	flowContext := FlowContext{
		Variables: variables,
	}

	return flow.Run(flowContext)
}

// RunWithInputAndVariables runs the flow with the provided input text and variables.
// This is a convenience method for creating a flow context with text and variables and running the flow.
//
// Parameters:
//   - input: The input text to use in the flow context
//   - variables: Map of variable names to values to use in the flow context
//
// Returns:
//   - The updated flow context after flow execution
//   - Any error encountered during flow execution
func (flow *Flow) RunWithInputAndVariables(input string, variables map[string]any) (*FlowContext, error) {
	// Create a new flowContext with the input and variables
	flowContext := FlowContext{
		Text:      input,
		Variables: make(map[string]any),
	}
	if variables != nil {
		flowContext.Variables = variables
	}

	return flow.Run(flowContext)
}

func (flow *Flow) Run(initialFlowContext FlowContext) (*FlowContext, error) {
	flowContext := &initialFlowContext
	flowContext.Flow = flow

	// Ensure Variables is initialized
	if flowContext.Variables == nil {
		flowContext.Variables = make(map[string]any)
	}

	// Merge flow variables into context (flowContext variables take precedence)
	if flow.Variables != nil {
		for k, v := range flow.Variables {
			if _, exists := flowContext.Variables[k]; !exists {
				flowContext.Variables[k] = v
			}
		}
	}

	// Compile regular expression to extract <think> tag content
	thinkRegex, err := regexp.Compile(`(?s)<think>.*?</think>`)
	if err != nil {
		return nil, err
	}

	log.Debug("Starting run flow ", flow.Name, " with initial context.")
	// For each step in the flow
	for _, step := range flow.Steps {
		// Run the step and get the updated flowContext
		result, err := tryStep(&step, *flowContext)

		log.Debug("Step running finished. Error:", err, ".")
		if err != nil {
			return nil, err
		}

		// Sync variables from flowContext to flow after step execution
		if flowContext.Variables != nil {
			if flow.Variables == nil {
				flow.Variables = make(map[string]any)
			}
			for k, v := range flowContext.Variables {
				flow.Variables[k] = v
			}
		}

		// Check if the result contains <think> tags
		if result != nil && result.Text != "" {
			thinkMatch := thinkRegex.FindStringSubmatch(result.Text)
			if len(thinkMatch) > 0 {
				// Extract <think> tag content to the Think property
				result.Think = thinkMatch[0]
				// Remove <think> tag content, keep the cleaned text
				result.Text = strings.TrimSpace(thinkRegex.ReplaceAllString(result.Text, ""))
			}
		}

		// Update the flowContext
		flowContext = result
	}

	// Return the flowContext content
	return flowContext, nil
}
