// message package contains the Message related structs and their related functions.
package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"text/template"
)

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func (m *Message) ToJSON() string {

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

func NewMessage(role, content string) Message {
	return Message{Content: content, Role: role}
}

func NewSystemMessage(content string) Message {
	return Message{Content: content, Role: "system"}
}

func NewUserMessage(content string) Message {
	return Message{Content: content, Role: "user"}
}

func NewAssistantMessage(content string) Message {
	return Message{Content: content, Role: "assistant"}
}

type MessageFormatter interface {
	Format(Data any) (string, error)
}

// MessageTemplateFormatter is a struct that implements the MessageFormatter interface. It uses Golang's text/template package to format a message based on a template and parameters.
// @see https://pkg.go.dev/text/template about how to use the Golang text template.
type MessageTemplateFormatter struct {
	TemplateName   string `json:"template_name,omitempty" yaml:"template_name,omitempty" mapstructure:"template_name,omitempty"`
	TemplateString string `json:"template_string,omitempty" yaml:"template_string,omitempty" mapstructure:"template_string"`
	File           string `json:"file,omitempty" yaml:"file,omitempty" mapstructure:"file"`
	theTemplate    *template.Template
}

func (t *MessageTemplateFormatter) Init() error {

	var tmpl *template.Template
	var err error

	if t.TemplateString != "" {
		tmpl, err = template.New("template").Parse(t.TemplateString)
		if err != nil {
			return err
		}

	}
	if t.File != "" {
		tmpl, err = template.New("template").ParseFiles(t.File)
		if err != nil {
			return err
		}
	}

	t.theTemplate = tmpl
	return nil
}

func NewMessageTemplateFormatter(templateContent string) (*MessageTemplateFormatter, error) {
	tmpl := &MessageTemplateFormatter{TemplateString: templateContent}
	if err := tmpl.Init(); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func NewMessageTemplateFormatterFromFile(templateFilePath string) (*MessageTemplateFormatter, error) {

	tmpl := &MessageTemplateFormatter{File: templateFilePath}
	if err := tmpl.Init(); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (t *MessageTemplateFormatter) Format(data any) (string, error) {

	if t.theTemplate == nil {
		return "", errors.New("template is not set")
	}

	var buffer bytes.Buffer

	if err := t.theTemplate.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
