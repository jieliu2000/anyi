package message_test

import (
	"fmt"

	"github.com/jieliu2000/anyi/message"
)

func Example_promptTemplateWithStruct() {

	// See https://pkg.go.dev/text/template about how to write templates
	template := "Write a guide on how to install and run {{.Application}} on {{.OS}}"
	formatter, err := message.NewPromptTemplateFormatter(template)

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
	formatter, err := message.NewPromptTemplateFormatter(template)

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
	formatter, err := message.NewPromptTemplateFormatter(template)

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
	formatter, err := message.NewPromptTemplateFormatter(template)

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
	formatter, err := message.NewPromptTemplateFormatter(template)

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
	formatter, err := message.NewPromptTemplateFormatter(template)

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
