# Building Workflows

This guide teaches you how to create complex AI workflows using Anyi's powerful workflow system. You'll learn to chain multiple steps together, handle data flow between steps, and implement robust error handling.

## Table of Contents

- [Understanding Workflows](#understanding-workflows)
- [Core Workflow Concepts](#core-workflow-concepts)
- [Creating Your First Workflow](#creating-your-first-workflow)
- [Data Flow and Context](#data-flow-and-context)
- [Advanced Workflow Patterns](#advanced-workflow-patterns)
- [Error Handling and Retries](#error-handling-and-retries)
- [Best Practices](#best-practices)

## Understanding Workflows

A workflow in Anyi is a sequence of steps that process data through various AI models and operations. Each step can:

- Send prompts to different LLM providers
- Validate outputs against specific criteria
- Transform data between steps
- Execute conditional logic
- Perform external operations

Workflows enable you to build sophisticated AI applications that go beyond simple question-and-answer interactions.

## Core Workflow Concepts

### Flow

A Flow is the main container that orchestrates the execution of multiple steps in sequence.

### Step

A Step represents a single operation in the workflow, consisting of:

- **Executor**: Performs the actual work (LLM call, data transformation, etc.)
- **Validator**: Ensures the output meets quality criteria
- **Retry Logic**: Handles failures gracefully

### FlowContext

The FlowContext carries data between steps and contains:

- **Text**: Primary text content being processed
- **Memory**: Structured data shared between steps
- **Think**: Extracted thinking process from models
- **ImageURLs**: List of image URLs for multimodal processing

## Creating Your First Workflow

Let's start with a simple two-step workflow that generates and improves content:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
)

func main() {
	// Create client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.Model = "gpt-4"
	client, err := anyi.NewClient("gpt4", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Step 1: Generate initial content
	step1, err := anyi.NewLLMStepWithTemplate(
		"Write a brief introduction about {{.Text}}. Keep it under 200 words.",
		"You are a professional technical writer.",
		client,
	)
	if err != nil {
		log.Fatalf("Failed to create step1: %v", err)
	}
	step1.Name = "generate_intro"

	// Step 2: Improve the content
	step2, err := anyi.NewLLMStepWithTemplate(
		"Improve the following introduction by making it more engaging and adding specific examples:\n\n{{.Text}}",
		"You are an expert editor who specializes in making technical content accessible.",
		client,
	)
	if err != nil {
		log.Fatalf("Failed to create step2: %v", err)
	}
	step2.Name = "improve_content"

	// Create the flow
	flow, err := anyi.NewFlow("content_workflow", client, *step1, *step2)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	// Execute the workflow
	result, err := flow.RunWithInput("machine learning")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	log.Printf("Final result: %s", result.Text)
}
```

## Data Flow and Context

### Using Memory for Complex Data

For workflows that need to maintain structured data across steps:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
)

// Define structured data for the workflow
type ResearchProject struct {
	Topic       string   `json:"topic"`
	Questions   []string `json:"questions"`
	Findings    []string `json:"findings"`
	Conclusion  string   `json:"conclusion"`
	Completed   bool     `json:"completed"`
}

func main() {
	// Create client
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	client, err := anyi.NewClient("researcher", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Initialize research project data
	project := ResearchProject{
		Topic:     "Quantum Computing Applications",
		Questions: []string{
			"What are the current practical applications?",
			"What are the main technical challenges?",
			"What is the timeline for widespread adoption?",
		},
		Findings:  []string{},
		Completed: false,
	}

	// Create workflow with memory
	context := anyi.NewFlowContextWithMemory(project)

	// Step 1: Research each question
	researchStep := &flow.Step{
		Name: "research_questions",
		Executor: &anyi.LLMExecutor{
			Template: `Research the topic: {{.Memory.Topic}}

Focus on answering these questions:
{{range .Memory.Questions}}
- {{.}}
{{end}}

Provide detailed findings for each question.`,
			SystemMessage: "You are a research scientist with expertise in emerging technologies.",
		},
		VarsImmutable: true,  // Preserve variables during execution
	}

	// Step 2: Synthesize conclusions
	synthesisStep := &flow.Step{
		Name: "synthesize_findings",
		Executor: &anyi.LLMExecutor{
			Template: `Based on the research findings below, provide a comprehensive conclusion about {{.Memory.Topic}}:

{{.Text}}

Summarize the key insights and implications.`,
			SystemMessage: "You are an expert analyst who synthesizes complex information.",
		},
		MemoryImmutable: true,  // Preserve memory data during execution
	}

	// Create and run flow
	researchFlow, err := anyi.NewFlow("research_flow", client, *researchStep, *synthesisStep)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	result, err := researchFlow.Run(context)
	if err != nil {
		log.Fatalf("Research workflow failed: %v", err)
	}

	log.Printf("Research conclusion: %s", result.Text)
}
```

### Working with Thinking Process

Some models (like DeepSeek) support explicit thinking tags. Here's how to capture and use them:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/deepseek"
)

func main() {
	// Create DeepSeek client
	config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")
	client, err := anyi.NewClient("deepseek", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Step 1: Problem analysis with thinking
	analysisStep := &flow.Step{
		Name: "analyze_problem",
		Executor: &anyi.LLMExecutor{
			Template: `<think>
Let me break down this problem step by step:
1. Understand what's being asked
2. Identify key components
3. Consider different approaches
4. Choose the best solution
</think>

Analyze the following problem and provide a solution: {{.Text}}`,
			SystemMessage: "You are a problem-solving expert. Use <think> tags to show your reasoning process.",
		},
	}

	// Step 2: Use the thinking process for refinement
	refinementStep := &flow.Step{
		Name: "refine_solution",
		Executor: &anyi.LLMExecutor{
			Template: `Previous analysis: {{.Text}}

Thinking process from previous step:
{{.Think}}

Based on this thinking process, provide an improved and more detailed solution.`,
			SystemMessage: "You are a solution architect who builds on previous analysis.",
		},
	}

	// Create flow
	problemSolvingFlow, err := anyi.NewFlow("problem_solving", client, *analysisStep, *refinementStep)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	// Run the workflow
	result, err := problemSolvingFlow.RunWithInput("How can we optimize database query performance for a high-traffic e-commerce application?")
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	log.Printf("Final solution: %s", result.Text)
	if result.Think != "" {
		log.Printf("Final thinking process: %s", result.Think)
	}
}
```

## Advanced Workflow Patterns

### Conditional Branching

Create workflows that make decisions based on intermediate results:

```go
// Create a conditional executor
conditionalExecutor := &anyi.ConditionalFlowExecutor{
	Condition: "{{.Text | contains \"urgent\"}}",
	TrueFlow:  urgentProcessingFlow,
	FalseFlow: normalProcessingFlow,
}

conditionalStep := &flow.Step{
	Name:     "route_request",
	Executor: conditionalExecutor,
}
```

### Multi-Client Workflows

Use different models for different steps based on their strengths:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/anthropic"
)

func main() {
	// Create different clients for different tasks
	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openaiClient, _ := anyi.NewClient("openai", openaiConfig)

	anthropicConfig := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))
	anthropicClient, _ := anyi.NewClient("anthropic", anthropicConfig)

	// Step 1: Use OpenAI for creative generation
	creativeStep := &flow.Step{
		Name: "generate_ideas",
		Executor: &anyi.LLMExecutor{
			Template:      "Generate creative ideas for {{.Text}}",
			SystemMessage: "You are a creative brainstorming expert.",
			Client:        openaiClient, // Specify client for this step
		},
	}

	// Step 2: Use Anthropic for analysis and refinement
	analysisStep := &flow.Step{
		Name: "analyze_ideas",
		Executor: &anyi.LLMExecutor{
			Template:      "Analyze and refine these ideas:\n\n{{.Text}}",
			SystemMessage: "You are a strategic analyst who evaluates ideas.",
			Client:        anthropicClient, // Different client for analysis
		},
	}

	// Create flow with default client (can be overridden per step)
	multiClientFlow, err := anyi.NewFlow("multi_client_workflow", openaiClient, *creativeStep, *analysisStep)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	result, err := multiClientFlow.RunWithInput("sustainable urban transportation")
	if err != nil {
		log.Fatalf("Multi-client workflow failed: %v", err)
	}

	log.Printf("Final analysis: %s", result.Text)
}
```

### Parallel Processing

For steps that can run independently:

```go
// Note: This is a conceptual example - actual parallel execution
// would require custom implementation or future Anyi features
func runParallelSteps(input string) ([]string, error) {
	// Create multiple clients
	client1, _ := anyi.GetClient("openai")
	client2, _ := anyi.GetClient("anthropic")
	client3, _ := anyi.GetClient("deepseek")

	// Create channels for results
	results := make(chan string, 3)
	errors := make(chan error, 3)

	// Run steps in parallel
	go func() {
		result, _, err := client1.Chat([]chat.Message{
			{Role: "user", Content: "Analyze from perspective 1: " + input},
		}, nil)
		if err != nil {
			errors <- err
			return
		}
		results <- result.Content
	}()

	go func() {
		result, _, err := client2.Chat([]chat.Message{
			{Role: "user", Content: "Analyze from perspective 2: " + input},
		}, nil)
		if err != nil {
			errors <- err
			return
		}
		results <- result.Content
	}()

	go func() {
		result, _, err := client3.Chat([]chat.Message{
			{Role: "user", Content: "Analyze from perspective 3: " + input},
		}, nil)
		if err != nil {
			errors <- err
			return
		}
		results <- result.Content
	}()

	// Collect results
	var finalResults []string
	for i := 0; i < 3; i++ {
		select {
		case result := <-results:
			finalResults = append(finalResults, result)
		case err := <-errors:
			return nil, err
		}
	}

	return finalResults, nil
}
```

## Error Handling and Retries

### Step-Level Retry Configuration

```go
// Configure retry behavior for individual steps
step := &flow.Step{
	Name:           "critical_analysis",
	MaxRetryTimes:  3, // Retry up to 3 times
	Executor:       myExecutor,
	Validator:      myValidator, // Validation failure triggers retry
}
```

### Custom Error Handling

```go
package main

import (
	"log"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

func createRobustWorkflow() *flow.Flow {
	// Create steps with comprehensive error handling
	step1 := &flow.Step{
		Name:          "data_extraction",
		MaxRetryTimes: 3,
		Executor: &anyi.LLMExecutor{
			Template:      "Extract key information from: {{.Text}}",
			SystemMessage: "You are a data extraction expert.",
		},
		Validator: &anyi.JsonValidator{
			RequiredFields: []string{"summary", "key_points"},
		},
	}

	step2 := &flow.Step{
		Name:          "data_analysis",
		MaxRetryTimes: 2,
		Executor: &anyi.LLMExecutor{
			Template:      "Analyze the extracted data: {{.Text}}",
			SystemMessage: "You are a data analyst.",
		},
		Validator: &anyi.StringValidator{
			MinLength: 100,
			MaxLength: 1000,
		},
	}

	// Create flow with error handling
	client, _ := anyi.GetClient("default")
	robustFlow, _ := anyi.NewFlow("robust_workflow", client, *step1, *step2)

	return robustFlow
}

func runWithErrorHandling() {
	workflow := createRobustWorkflow()

	// Implement custom retry logic at the workflow level
	maxAttempts := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := workflow.RunWithInput("complex data to process")

		if err == nil {
			log.Printf("Workflow succeeded on attempt %d", attempt)
			log.Printf("Result: %s", result.Text)
			return
		}

		if attempt < maxAttempts {
			log.Printf("Workflow failed on attempt %d: %v", attempt, err)
			log.Printf("Retrying in %v...", backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	log.Fatalf("Workflow failed after %d attempts", maxAttempts)
}
```

## Best Practices

### 1. Design for Modularity

- Keep steps focused on single responsibilities
- Make steps reusable across different workflows
- Use clear, descriptive names for steps and flows

### 2. Handle Data Flow Carefully

- Validate data at step boundaries
- Use structured memory for complex data
- Document expected input/output formats

### 3. Implement Robust Error Handling

- Set appropriate retry counts based on step criticality
- Use validators to catch quality issues early
- Implement graceful degradation when possible

### 4. Optimize for Performance

- Choose the right model for each step's complexity
- Consider caching for repeated operations
- Use local models for non-critical steps

### 5. Monitor and Debug

- Add logging at key points in your workflow
- Track token usage and costs
- Implement health checks for long-running workflows

### 6. Security Considerations

- Validate and sanitize all inputs
- Be cautious with dynamic template content
- Implement rate limiting for user-facing workflows

### Example: Production-Ready Workflow

```go
package main

import (
	"log"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

func createProductionWorkflow() *flow.Flow {
	// Load configuration from file
	err := anyi.ConfigFromFile("./production-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get pre-configured clients
	primaryClient, _ := anyi.GetClient("primary")
	fallbackClient, _ := anyi.GetClient("fallback")

	// Step 1: Input validation and preprocessing
	preprocessStep := &flow.Step{
		Name: "preprocess",
		Executor: &anyi.SetContextExecutor{
			TextTemplate: "Preprocessed: {{.Text | trim | lower}}",
		},
		MaxRetryTimes: 1,
	}

	// Step 2: Main processing with fallback
	mainProcessingStep := &flow.Step{
		Name: "main_processing",
		Executor: &anyi.LLMExecutor{
			Template:      "Process this content: {{.Text}}",
			SystemMessage: "You are a professional content processor.",
			Client:        primaryClient,
		},
		Validator: &anyi.StringValidator{
			MinLength: 50,
			MaxLength: 2000,
		},
		MaxRetryTimes: 2,
	}

	// Step 3: Quality assurance
	qaStep := &flow.Step{
		Name: "quality_assurance",
		Executor: &anyi.LLMExecutor{
			Template:      "Review and improve if necessary: {{.Text}}",
			SystemMessage: "You are a quality assurance specialist.",
			Client:        fallbackClient,
		},
		Validator: &anyi.StringValidator{
			MinLength: 100,
		},
		MaxRetryTimes: 1,
	}

	// Create production workflow
	productionFlow, err := anyi.NewFlow("production_workflow", primaryClient,
		*preprocessStep, *mainProcessingStep, *qaStep)
	if err != nil {
		log.Fatalf("Failed to create production workflow: %v", err)
	}

	return productionFlow
}

func main() {
	workflow := createProductionWorkflow()

	// Production execution with monitoring
	start := time.Now()
	result, err := workflow.RunWithInput("User input content here")
	duration := time.Since(start)

	if err != nil {
		log.Printf("Workflow failed after %v: %v", duration, err)
		// Implement alerting/monitoring here
		return
	}

	log.Printf("Workflow completed successfully in %v", duration)
	log.Printf("Result: %s", result.Text)
	// Implement success metrics tracking here
}
```

This comprehensive guide should help you build robust, efficient workflows using the Anyi framework. Remember to start simple and gradually add complexity as your requirements evolve.
