package openai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
	impl "github.com/sashabaranov/go-openai"
)

type ParameterDetail struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

func convertToFuncDesc(function tools.FunctionConfig) impl.FunctionDefinition {

	parameters := map[string]any{
		"type": "object",
	}
	properties := make(map[string]any)

	for _, param := range function.Params {
		if param.Name == "type" {
			param.Name = "_type_"
		}
		properties[param.Name] = ParameterDetail{
			Type:        param.Type,
			Description: param.Description,
			Enum:        param.Enum,
		}
	}
	parameters["properties"] = properties

	return impl.FunctionDefinition{
		Name:        function.Name,
		Description: function.Description,
		Parameters:  parameters,
	}
}

func ExecuteChatWithFunctions(client *impl.Client, model string, messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	info := chat.ResponseInfo{}

	if client == nil {
		return nil, info, errors.New("client not initialized")
	}

	messagesInput := ConvertToOpenAIChatMessages(messages)
	toolsImpl := []impl.Tool{}

	for _, f := range functions {

		funcDefinition := convertToFuncDesc(f)
		toolImpl := impl.Tool{
			Type:     impl.ToolTypeFunction,
			Function: &funcDefinition,
		}

		toolsImpl = append(toolsImpl, toolImpl)
	}

	request := impl.ChatCompletionRequest{
		Model:    model,
		Messages: messagesInput,
		Tools:    toolsImpl,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		request,
	)

	if err != nil {
		return nil, info, err
	}
	choice := resp.Choices[0]

	result := chat.Message{
		Content: choice.Message.Content,
		Role:    choice.Message.Role,
	}

	toolsCalls := []chat.ToolCall{}

	if choice.Message.ToolCalls != nil {
		for _, call := range choice.Message.ToolCalls {

			params := make(map[string]any)
			json.Unmarshal([]byte(call.Function.Arguments), &params)
			funcCall := chat.FunctionCall{
				Name:      call.Function.Name,
				Arguments: params,
			}
			toolsCalls = append(toolsCalls, chat.ToolCall{
				Function: funcCall,
			})
		}
	}
	result.ToolCalls = toolsCalls

	info.PromptTokens = resp.Usage.PromptTokens
	info.CompletionTokens = resp.Usage.CompletionTokens

	return &result, info, nil
}

func ExecuteChat(client *impl.Client, model string, messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	info := chat.ResponseInfo{}

	if client == nil {
		return nil, info, errors.New("client not initialized")
	}

	messagesInput := ConvertToOpenAIChatMessages(messages)
	request := impl.ChatCompletionRequest{
		Model:    model,
		Messages: messagesInput,
	}

	if options != nil {
		if strings.ToLower(options.Format) == "json" {
			request.ResponseFormat = &impl.ChatCompletionResponseFormat{
				Type: "json_object",
			}
		}
	}
	log.Debugf("Sending request %v", request)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		request,
	)
	log.Debugf("Response: %v", resp)

	if err != nil {
		log.Errorf("Error: %v", err)
		return nil, info, err
	}
	result := chat.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}
	info.PromptTokens = resp.Usage.PromptTokens
	info.CompletionTokens = resp.Usage.CompletionTokens

	return &result, info, nil
}

func ConvertToOpenAIChatMessages(messages []chat.Message) []impl.ChatCompletionMessage {
	result := []impl.ChatCompletionMessage{}
	for _, msg := range messages {
		openaiMessage := convertToOpenAIChatMessage(msg)
		result = append(result, openaiMessage)
	}
	return result
}

func convertToOpenAIChatMessage(msg chat.Message) impl.ChatCompletionMessage {
	result := impl.ChatCompletionMessage{
		Role: msg.Role,
	}
	if msg.Content != "" {
		result.Content = msg.Content
		return result
	}
	if len(msg.ContentParts) > 0 {
		messageParts := []impl.ChatMessagePart{}
		for _, img := range msg.ContentParts {

			if img.Text != "" {
				textPart := impl.ChatMessagePart{
					Type: impl.ChatMessagePartTypeText,
					Text: img.Text,
				}

				messageParts = append(messageParts, textPart)
			} else if img.ImageUrl != "" {
				imagePart := impl.ChatMessagePart{
					Type: impl.ChatMessagePartTypeImageURL,
					ImageURL: &impl.ChatMessageImageURL{
						URL: img.ImageUrl,
					},
				}
				messageParts = append(messageParts, imagePart)
			}
		}
		result.MultiContent = messageParts
	}
	return result
}
