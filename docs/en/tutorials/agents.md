# Building Autonomous Agents with Anyi

Anyi's Agent framework enables you to create intelligent autonomous agents that can plan and execute complex tasks by leveraging available workflows. This tutorial will guide you through the process of building and using autonomous agents.

## Understanding Agents

An Agent in Anyi is an intelligent entity that can:

1. **Plan** - Analyze tasks and create execution plans using available workflows
2. **Execute** - Run workflows in sequence to accomplish objectives
3. **Adapt** - Adjust plans based on execution results and feedback

Agents are particularly powerful for complex, multi-step tasks that require intelligent decision-making.

## Creating an Agent

To create an agent, use the [NewAgent](../../reference/api.md#NewAgent) function:

```go
agent, err := anyi.NewAgent(
    "researcher",                                    // Agent name (for registry)
    "Research Assistant",                           // Agent role
    "Expert at researching topics and writing reports", // Agent backstory
    []string{"research_flow", "analyze_flow"},      // Available flows
    client,                                         // LLM client for planning
)
```

### Parameters Explained

- **name**: Optional name for registering the agent in the global registry
- **role**: The role or job function of the agent
- **backstory**: Background information that helps the agent understand its purpose
- **availableFlows**: List of workflow names the agent can use
- **client**: LLM client used for intelligent planning (can be nil for simple planning)

## Agent Configuration

Agents can be configured with specific parameters to control their behavior:

```go
// Default configuration
config := agent.DefaultConfig()

// Custom configuration
config := agent.Config{
    MaxIterations: 15,    // Maximum execution iterations
    MaxRetries:    5,     // Maximum retry attempts
    Timeout:       60 * time.Minute, // Execution timeout
}

// Apply configuration
myAgent.Config = config
```

## Executing Tasks

Once created, agents can execute tasks using the [Execute](../../reference/api.md#Agent.Execute) method:

```go
result, context, err := agent.Execute(
    "Research the impact of AI on healthcare and write a comprehensive report",
    agent.AgentContext{
        Variables: map[string]interface{}{
            "depth":   "detailed",
            "sources": 10,
            "format":  "markdown",
        },
    },
)
```

The execution process involves:

1. **Planning**: The agent analyzes the task and creates an execution plan
2. **Execution**: Workflows are executed in sequence according to the plan
3. **Monitoring**: Results are monitored to determine completion
4. **Adaptation**: Plans may be adjusted based on execution results

## Planning Strategies

Anyi agents support multiple planning strategies:

### AI-Based Planning

When an LLM client is provided, agents use AI to create intelligent plans:

```go
// Agent with LLM client for AI planning
agent := agent.NewAgentWithClient(
    "AI Researcher",
    "Intelligent research assistant",
    []string{"web_search", "analyze", "summarize"},
    registry.Global,
    openaiClient, // LLM client for planning
)
```

### Simple Planning

When no LLM client is available, agents use a simple sequential approach:

```go
// Agent without LLM client uses simple planning
agent := agent.NewAgent(
    "Basic Researcher",
    "Research assistant",
    []string{"research", "analyze", "summarize"},
    registry.Global, // Flow getter
)
```

## Working with Context

Agents use [AgentContext](../../reference/api.md#AgentContext) to maintain state and pass variables:

```go
context := agent.AgentContext{
    Variables: map[string]interface{}{
        "topic": "AI in Healthcare",
        "tone":  "professional",
        "style": "academic",
    },
    Memory: "Previous research results",
    History: []string{
        "Previous task results",
    },
}
```

Context is passed by value, ensuring thread safety and preventing unintended modifications.

## Example: Research Agent

Here's a complete example of a research agent:

```go
package main

import (
    "log"
    "os"
    
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent"
    "github.com/jieliu2000/anyi/llm/openai"
    "github.com/jieliu2000/anyi/registry"
)

func main() {
    // Create LLM client
    config := openai.DefaultConfig("gpt-4")
    config.APIKey = os.Getenv("OPENAI_API_KEY")
    client, err := anyi.NewClient("researcher", config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Create research agent
    agent, err := anyi.NewAgent(
        "ai_researcher",                            // Registry name
        "AI Research Assistant",                    // Role
        "Expert at researching topics and synthesizing information", // Backstory
        []string{"web_research", "analyze", "write_report"}, // Available flows
        client,                                     // Planning client
    )
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Execute research task
    result, context, err := agent.Execute(
        "Analyze the latest developments in quantum computing and their potential impact on cybersecurity",
        agent.AgentContext{
            Variables: map[string]interface{}{
                "depth":        "comprehensive",
                "perspective":  "technical",
                "citations":    true,
                "word_limit":   2000,
            },
        },
    )
    if err != nil {
        log.Fatalf("Agent execution failed: %v", err)
    }

    log.Printf("Research completed. Result length: %d characters", len(result))
    log.Printf("Execution history: %v", context.History)
}
```

## Best Practices

### 1. Design Specific Agents

Create agents with specific roles and capabilities:

```go
// Good: Specific role and capabilities
writerAgent, _ := anyi.NewAgent(
    "technical_writer",
    "Technical Documentation Writer",
    "Specialist in creating clear technical documentation",
    []string{"outline", "draft", "review", "finalize"},
    client,
)

// Avoid: Generic, unfocused agent
genericAgent, _ := anyi.NewAgent(
    "helper",
    "General Assistant",
    "Does everything",
    []string{"all_flows"},
    client,
)
```

### 2. Provide Clear Instructions

Give agents clear, specific tasks:

```go
// Good: Clear, specific task
result, _, _ := agent.Execute(
    "Create a step-by-step guide for setting up a Kubernetes cluster",
    context,
)

// Avoid: Vague, ambiguous task
result, _, _ := agent.Execute(
    "Help with Kubernetes",
    context,
)
```

### 3. Use Appropriate Flows

Match flows to agent capabilities:

```go
// Research agent with research-focused flows
researchAgent, _ := anyi.NewAgent(
    "market_researcher",
    "Market Research Analyst",
    "Expert in market analysis and competitive intelligence",
    []string{
        "competitor_analysis",
        "market_trends",
        "customer_survey",
        "data_analysis",
        "report_generation",
    },
    client,
)
```

## Error Handling

Agents can encounter various errors during execution:

```go
result, context, err := agent.Execute(task, initialContext)
if err != nil {
    switch {
    case errors.Is(err, agent.ErrPlanningFailed):
        log.Printf("Planning failed: %v", err)
    case errors.Is(err, agent.ErrExecutionFailed):
        log.Printf("Execution failed: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
    
    // Access partial results from context if needed
    log.Printf("Partial results: %v", context.History)
}
```

## Advanced Features

### Custom Flow Getters

For advanced use cases, you can implement custom flow getters:

```go
type CustomFlowGetter struct {
    // Custom flow management logic
}

func (c *CustomFlowGetter) GetFlow(name string) (interface{}, error) {
    // Custom flow retrieval logic
    return flow, nil
}

agent := agent.NewAgentWithClient(
    "Custom Agent",
    "Agent with custom flow management",
    []string{"custom_flow1", "custom_flow2"},
    &CustomFlowGetter{},
    client,
)
```

### Monitoring and Logging

Implement custom monitoring for agent activities:

```go
// Before execution
log.Printf("Starting agent execution for task: %s", task)

// During execution (in custom executors)
log.Printf("Agent executing flow: %s", flowName)

// After execution
log.Printf("Agent execution completed. Result length: %d", len(result))
```

## Next Steps

- Learn about [workflows](workflows.md) to create the flows your agents will use
- Explore [configuration management](configuration.md) for complex agent setups
- Review the [API reference](../../reference/api.md#agent) for detailed agent documentation