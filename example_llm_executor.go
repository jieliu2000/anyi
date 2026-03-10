package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi/executors/builtin"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
)

func main() {
	fmt.Println("Testing LLMExecutor...")

	// Create a default OpenAI client
	client := openai.NewOpenAICompatClient(openai.DefaultConfig(""))

	// Create an LLMExecutor
	executor := &builtin.LLMExecutor{
		Client:   client,
		Template: "Please provide a short summary of the following text:\n{{.Text}}",
	}

	// Initialize the executor
	if err := executor.Init(); err != nil {
		log.Fatal("Failed to initialize executor:", err)
	}

	// Create a flow context
	context := flow.FlowContext{
		Text: "Artificial intelligence (AI) is a branch of computer science that aims to create software or machines that exhibit human-like intelligence. This can include learning from experience, understanding natural language, solving problems, and recognizing patterns. AI technologies are increasingly prevalent in our daily lives, from voice assistants like Siri and Alexa to recommendation systems on streaming platforms.",
	}

	// Create a step
	step := &flow.Step{
		Name: "LLM Step",
	}

	// Run the executor
	result, err := executor.Run(context, step)
	if err != nil {
		log.Fatal("Failed to run executor:", err)
	}

	fmt.Println("LLM Executor Result:")
	fmt.Println(result.Text)
}