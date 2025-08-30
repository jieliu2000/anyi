package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

func main() {
	// Load configuration from file
	err := anyi.ConfigFromFile("config.yml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get an agent by name
	agent, err := anyi.GetAgent("codeReviewer")
	if err != nil {
		log.Fatalf("Failed to get agent: %v", err)
	}

	// Use the agent
	codeToReview := `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	result, _, err := agent.Execute("Review this Go code for best practices and potential issues", anyi.AgentContext{
		Variables: map[string]interface{}{
			"code": codeToReview,
		},
	})
	if err != nil {
		log.Fatalf("Failed to execute agent: %v", err)
	}

	fmt.Println("Code Review Result:")
	fmt.Println(result)
}
