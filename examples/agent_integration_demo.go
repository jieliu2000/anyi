package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

// This example demonstrates the complete Agent framework integration with Anyi.
// It shows how to:
// 1. Load configuration that includes agents
// 2. Retrieve configured agents
// 3. Execute tasks with simple string objectives
func main() {
	// Set up environment variables for testing (in real usage, these would be set externally)
	// os.Setenv("OPENAI_API_KEY", "your-openai-key")
	// os.Setenv("ANTHROPIC_API_KEY", "your-anthropic-key")

	fmt.Println("=== Anyi Agent Framework Integration Demo ===\n")

	// 1. Load complete configuration including agents
	fmt.Println("📁 Loading configuration from file...")
	err := anyi.ConfigFromFile("examples/agent_config.yaml")
	if err != nil {
		log.Printf("⚠️  Config loading failed (expected if API keys not set): %v\n", err)
		fmt.Println("💡 To run this demo with real LLMs, set environment variables:")
		fmt.Println("   export OPENAI_API_KEY=your-key")
		fmt.Println("   export ANTHROPIC_API_KEY=your-key")
		fmt.Println()
		
		// Continue with demo using mock setup
		fmt.Println("🔧 Setting up demo with mock configuration...")
		setupMockDemo()
		return
	}
	fmt.Println("✅ Configuration loaded successfully!\n")

	// 2. List available agents
	fmt.Println("👥 Available agents:")
	agents := anyi.ListAgents()
	for i, agentName := range agents {
		fmt.Printf("   %d. %s\n", i+1, agentName)
	}
	fmt.Println()

	// 3. Get specific agent
	fmt.Println("🤖 Getting research assistant agent...")
	agent, err := anyi.GetAgent("research_assistant")
	if err != nil {
		log.Fatal("Failed to get agent:", err)
	}
	fmt.Printf("✅ Agent '%s' loaded successfully\n", agent.Name)
	fmt.Printf("   Description: %s\n", agent.Description)
	fmt.Printf("   Available flows: %v\n", agent.Flows)
	fmt.Printf("   Client: %s\n\n", agent.ClientName)

	// 4. Execute task with simple objective string
	fmt.Println("🎯 Executing task...")
	objective := "Research the latest developments in artificial intelligence and create a comprehensive report"
	fmt.Printf("   Objective: %s\n\n", objective)

	result, err := agent.Execute(objective)
	if err != nil {
		log.Printf("Task execution failed: %v\n", err)
		return
	}

	// 5. Display results
	fmt.Println("📊 Task Results:")
	fmt.Printf("   Status: %s\n", result.Status)
	fmt.Printf("   Duration: %v\n", result.Duration)
	fmt.Printf("   Steps executed: %d\n", len(result.StepResults))
	fmt.Printf("   Final output: %s\n\n", result.FinalOutput)

	// 6. Show execution history
	fmt.Println("📚 Execution History:")
	history := agent.GetExecutionHistory()
	fmt.Printf("   Total tasks completed: %d\n", len(history))
	for i, task := range history {
		fmt.Printf("   %d. %s (Status: %s)\n", i+1, task.Objective, task.Status)
	}
}

// setupMockDemo sets up a demonstration with mock configuration when API keys are not available
func setupMockDemo() {
	fmt.Println("🎭 This would normally:")
	fmt.Println("   1. Load agent configurations from YAML")
	fmt.Println("   2. Validate that referenced flows and clients exist")
	fmt.Println("   3. Create agent instances with registry access")
	fmt.Println("   4. Enable simple agent.Execute(objective) calls")
	fmt.Println()
	
	fmt.Println("📋 Example usage after proper setup:")
	fmt.Println(`
   // Load configuration
   err := anyi.ConfigFromFile("config.yaml")
   
   // Get agent
   agent, err := anyi.GetAgent("research_assistant")
   
   // Execute task with simple string objective
   result, err := agent.Execute("Research AI safety developments")
   
   // Get results
   fmt.Printf("Result: %s\n", result.FinalOutput)
	`)
	
	fmt.Println("🎯 Key benefits of this integration:")
	fmt.Println("   ✅ Single configuration file for all components")
	fmt.Println("   ✅ Simple agent.Execute(objective) interface")
	fmt.Println("   ✅ Automatic registry integration")
	fmt.Println("   ✅ Backward compatibility with existing flows")
	fmt.Println("   ✅ Intelligent planning using existing Flow infrastructure")
}
