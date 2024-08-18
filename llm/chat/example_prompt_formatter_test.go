package chat

import (
	"fmt"
)

func Example_promptTemplateWithStruct() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Write a guide on how to install and run {{.Application}} on {{.OS}}"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	type App struct {
		Application string
		OS          string
	}

	app := App{
		Application: "VS Code",
		OS:          "Ubuntu",
	}

	result, _ := formatter.Format(app)

	fmt.Print(result)
	// Output: Write a guide on how to install and run VS Code on Ubuntu
}

func Example_promptTemplateWithMap() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Write a guide on how to install and run {{.Application}} on {{.OS}}"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	app := map[string]interface{}{
		"Application": "VS Code",
		"OS":          "Ubuntu",
	}

	result, _ := formatter.Format(app)

	fmt.Print(result)
	// Output: Write a guide on how to install and run VS Code on Ubuntu
}

func Example_promptTemplateWithStructPointer() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Write a guide on how to install and run {{.Application}} on {{.OS}}"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	app := map[string]interface{}{
		"Application": "VS Code",
		"OS":          "Ubuntu",
	}

	result, _ := formatter.Format(&app)

	fmt.Print(result)
	// Output: Write a guide on how to install and run VS Code on Ubuntu
}

func Example_promptTemplateWithString() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Hello, {{.}}!"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	result, _ := formatter.Format("world")

	fmt.Print(result)
	// Output: Hello, world!
}

func Example_promptTemplateWithArray() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Hello, {{index . 0}}!"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	input := []string{
		"world",
	}
	result, _ := formatter.Format(input)

	fmt.Print(result)
	// Output: Hello, world!
}

func Example_promptTemplateWithStructArray() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Hello, {{(index . 0).Name}}!"
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	type Person struct {
		Name string
	}

	input := []Person{
		{Name: "John Doe"},
	}
	result, _ := formatter.Format(input)

	fmt.Print(result)
	// Output: Hello, John Doe!
}

func Example_promptTemplateWithArrayField() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := `{{/* This template is a prompt template in babyagi. See https://github.com/yoheinakajima/babyagi/blob/main/babyagi.py for details */}}
You are tasked with prioritizing the following tasks: 
{{range $index, $task := .Tasks}}* {{$task}}.
{{end}}
Consider the ultimate objective of your team: {{.Objective}}.
Tasks should be sorted from highest to lowest priority, where higher-priority tasks are those that act as pre-requisites or are more essential for meeting the objective.
Do not remove any tasks. Return the ranked tasks as a numbered list in the format:

#. First task
#. Second task

The entries must be consecutively numbered, starting with 1. The number of each entry must be followed by a period.
Do not include any headers before your ranked list or follow your list with any other output.`
	formatter, err := NewPromptTemplateFormatter(template)

	if err != nil {
		panic(err)
	}

	type AgentTasks struct {
		Tasks     []string
		Objective string
	}

	input := AgentTasks{
		Tasks:     []string{"task1", "task2"},
		Objective: "objective",
	}

	result, _ := formatter.Format(input)

	fmt.Print(len(result))
	// Output: 625
}
