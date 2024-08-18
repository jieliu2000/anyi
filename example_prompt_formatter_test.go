package anyi_test

import (
	"fmt"

	"github.com/jieliu2000/anyi"
)

func Example_promptTemplateFormatter() {

	// See documentation of github.com/jieliu2000/anyi/chat for more information and examples on how to use the prompt formatter
	// See https://pkg.go.dev/text/template about how to write templates
	template := `Write a guide on how to install and run {{.Application}} on {{.OS}}`
	formatter, err := anyi.NewPromptTemplateFormatter("template1", template)

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
