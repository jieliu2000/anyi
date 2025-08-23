package agent

import (
	"fmt"

	"github.com/jieliu2000/anyi/registry"
)

// Example demonstrates how to use the Agent framework.
func Example() {
	// Create a registry
	reg := registry.NewRegistry()

	// Register a simple LLM client (this would normally be a real client)
	// For demonstration purposes, we'll assume it's already registered

	// Create an agent configuration
	agents := []Agent{
		{
			Name:        "assistant",
			Description: "A helpful AI assistant that can perform various tasks",
			Flows:       []string{"chat", "analyze", "summarize"},
			ClientName:  "openai-gpt4",
			Memory:      NewSimpleMemory(),
			Config:      make(map[string]any),
		},
	}

	// Register the agent
	for _, agent := range agents {
		err := reg.RegisterAgent(agent.Name, agent)
		if err != nil {
			fmt.Printf("Failed to register agent %s: %v\n", agent.Name, err)
			return
		}
	}

	// Get the agent
	agentInterface, err := reg.GetAgent("assistant")
	if err != nil {
		fmt.Printf("Failed to get agent: %v\n", err)
		return
	}

	// Type assert to Agent interface
	agent, ok := agentInterface.(registry.Agent)
	if !ok {
		fmt.Println("Invalid agent type")
		return
	}

	fmt.Printf("Agent '%s' loaded successfully\n", agent.GetName())
	fmt.Printf("Available flows: %v\n", agent.GetFlows())
	fmt.Printf("Client: %s\n", agent.GetClientName())

	// Example of how execution would work (requires proper setup of flows and clients):
	// result, err := agent.Execute("Summarize the latest news about AI", reg)
	// if err != nil {
	//     fmt.Printf("Execution failed: %v\n", err)
	//     return
	// }
	// fmt.Printf("Result: %s\n", result.FinalOutput)
}

// ExampleConfig demonstrates how to load agents from YAML configuration.
func ExampleConfig() {
	yamlConfig := `
agents:
  - name: research_agent
    description: An agent specialized in research and analysis
    flows:
      - web_search
      - document_analysis
      - synthesis
    clientName: anthropic-claude
    config:
      max_sources: 10
      depth: 3
  
  - name: writing_agent
    description: An agent specialized in content creation
    flows:
      - content_planning
      - draft_writing
      - editing
    clientName: openai-gpt4
    config:
      style: academic
      length: long
`

	agents, err := LoadAgentsFromYAML([]byte(yamlConfig))
	if err != nil {
		fmt.Printf("Failed to load agents: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d agents:\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("- %s: %s\n", agent.Name, agent.Description)
		fmt.Printf("  Flows: %v\n", agent.Flows)
		fmt.Printf("  Client: %s\n", agent.ClientName)
	}
}
