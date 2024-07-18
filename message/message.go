package message

import (
	"bytes"
	"errors"
	"text/template"
)

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type MessageFormatter interface {
	Format(Data any) (string, error)
}

// MessageTemplateFormatter is a struct that implements the MessageFormatter interface. It uses Golang's text/template package to format a message based on a template and parameters.
// @see https://pkg.go.dev/text/template about how to use the Golang text template.
type MessageTemplateFormatter struct {
	File        string
	theTemplate *template.Template
}

func (t *MessageTemplateFormatter) Init() error {
	tmpl, err := template.New(t.File).ParseFiles(t.File)
	if err != nil {
		return err
	}
	t.theTemplate = tmpl
	return nil
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

func NewMessageTemplateFormatter(file string) (*MessageTemplateFormatter, error) {
	t := &MessageTemplateFormatter{File: file}
	err := t.Init()

	return t, err
}
