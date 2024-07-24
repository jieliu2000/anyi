// message package contains the Message and Prompt related structs and their related functions.
package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"text/template"

	"github.com/jieliu2000/anyi/llm/tools"
)

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func (m *Message) ToJSON() string {

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// Creates a new message with the given role and content.
func NewMessage(role, content string) Message {
	return Message{Content: content, Role: role}
}

// Creates a new system message with the given content.
func NewSystemMessage(content string) Message {
	return Message{Content: content, Role: "system"}
}

// Creates a new user message with the given content.
func NewUserMessage(content string) Message {
	return Message{Content: content, Role: "user"}
}

// Creates a new assistant message with the given content.
func NewAssistantMessage(content string) Message {
	return Message{Content: content, Role: "assistant"}
}

type PromptFormatter interface {
	Format(Data any) (string, error)
}

// PromptyTemplateFormatter is a struct that implements the [PromptFormatter] interface. It uses [Golang's text/template package] to format a message based on a template and parameters.
// See https://pkg.go.dev/text/template about how to use the Golang text template.
//
// [Golang's text/template package]: https://pkg.go.dev/text/template
type PromptyTemplateFormatter struct {
	TemplateName   string `json:"template_name,omitempty" yaml:"template_name,omitempty" mapstructure:"template_name,omitempty"`
	TemplateString string `json:"template_string,omitempty" yaml:"template_string,omitempty" mapstructure:"template_string"`
	File           string `json:"file,omitempty" yaml:"file,omitempty" mapstructure:"file"`
	theTemplate    *template.Template
}

func (t *PromptyTemplateFormatter) SetTemplate(template *template.Template) {
	t.theTemplate = template
}

// Initializes the PromptyTemplateFormatter.
// This method will init the TemplateFormatter based based on the values of the TemplateString and File fields. It works in the following order:
//  1. If TemplateString is not empty, it will parse the string as a template and set the Template field to that parsed template.
//  2. If TemplateString is empty and File is not empty, it will parse the file as a template and set the Template field to that parsed template.
//  3. If both TemplateString and File are empty, it will return an error.
func (t *PromptyTemplateFormatter) Init() error {

	var tmpl *template.Template
	var err error

	if t.TemplateString != "" {
		tmpl, err = template.New("template").Parse(t.TemplateString)
		if err != nil {
			return err
		}

	} else if t.File != "" {
		fileBase := filepath.Base(t.File)
		tmpl, err = template.New(fileBase).ParseFiles(t.File)
		if err != nil {
			return err
		}
	}

	t.theTemplate = tmpl
	return nil
}

// Creates a new PromptyTemplateFormatter with the given template string. See [Golang's text/template package documentation] about how to use the Golang text template.
//
// [Golang's text/template package documentation]: https://pkg.go.dev/text/template
func NewPromptTemplateFormatter(templateContent string) (*PromptyTemplateFormatter, error) {
	tmpl := &PromptyTemplateFormatter{TemplateString: templateContent}
	if err := tmpl.Init(); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Creates a new PromptyTemplateFormatter with the given template file path. See [Golang's text/template package documentation] about how to use the Golang text template.
// Parameter templateFilePath is the path to the template file.
//
// [Golang's text/template package documentation]: https://pkg.go.dev/text/template
func NewPromptTemplateFormatterFromFile(templateFilePath string) (*PromptyTemplateFormatter, error) {

	tmpl := &PromptyTemplateFormatter{File: templateFilePath}
	if err := tmpl.Init(); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Formats the given data using the template. See [Golang's text/template package documentation] about how to use the Golang text template.
//
// [Golang's text/template package documentation]: https://pkg.go.dev/text/template
func (t *PromptyTemplateFormatter) Format(data any) (string, error) {

	if t.theTemplate == nil {
		return "", errors.New("template is not set")
	}

	var buffer bytes.Buffer

	if err := t.theTemplate.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func AddFunctionDirectivesToPrompt(objective string, functions []tools.FunctionConfig) (string, error) {

	templateString := `Your task is to generate a task list in JSON array format to archieve this target: '''{{.Objective}}'''
	You can use the following functions in generating the task list:
	{{range .Functions}}* {{.Name}}: {{.Description}}. {{if .Params}} Parameters:{{range .Params}}	- {{.Name}}(type: {{.Type}}): {{.Description}}, {{end}} {{end}}
	{{end}}
	Each task json should be in this format:
	{"function": "$function_name", "params": [{"param_name": "$param_name", "param_value": "$param_value"}]}

	The output should be a JSON array of JSONs in above format.

	For example, if you only need one task with a function "add" with two params "a" and "b" which have values 1 and 2 , you should finally out
	put an array like this:
	[{"function": "add", "params": [{"param_name": "a", "param_value": 1}, {"param_name": "b", "param_value": 2}]}]
	DO NOT add any extra text execept the task list JSON array.`
	type TemplateData struct {
		Objective string
		Functions []tools.FunctionConfig
	}

	data := TemplateData{
		Objective: objective,
		Functions: functions,
	}

	tmpl, err := NewPromptTemplateFormatter(templateString)
	if err != nil {
		return "", err
	}

	return tmpl.Format(data)
}
