// message package contains the Message and Prompt related structs and their related functions.
package chat

import (
	"encoding/json"
)

type Message struct {
	Content      string        `json:"content,omitempty"`
	Role         string        `json:"role"`
	ContentParts []ContentPart `json:"contentParts,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (m *Message) ToJSON() string {

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// Creates a new message with the given role and content.
func NewMessage(role, content string) Message {
	return Message{Content: content, Role: role}
}

func NewEmptyMessage(role string) Message {
	return Message{Role: role}
}

// Creates a new image message with the given role, content and image.
func NewImageMessageFromUrl(role, content string, imageUrl string) Message {
	if imageUrl == "" {
		return Message{Role: role}
	}
	textPart := NewTextPart(content)
	imageContent, err := NewImagePartFromUrl(imageUrl, "")

	if err != nil {
		// In this case because the image content is invalid, we will return an empty message with the given role. Obviously it will cause an error when the message is sent.
		return Message{Role: role}
	}

	return Message{Role: role, ContentParts: []ContentPart{*textPart, *imageContent}}
}

// Creates a new image message with the given role, content and image.
func NewImageMessageFromFile(role, content string, filePath string) Message {
	if filePath == "" {
		return Message{Role: role}
	}
	textPart := NewTextPart(content)
	imageContent, err := NewImagePartFromFile(filePath, "")

	if err != nil {
		// In this case because the image content is invalid, we will return an empty message with the given role. Obviously it will cause an error when the message is sent.
		return Message{Role: role}
	}

	return Message{Role: role, ContentParts: []ContentPart{*textPart, *imageContent}}
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
