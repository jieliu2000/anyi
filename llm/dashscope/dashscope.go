// Use the openai compatible interface to access the DashScope service.
// @see https://help.aliyun.com/zh/dashscope/developer-reference/compatibility-of-openai-with-dashscope/?spm=a2c4g.11186623.0.0.17504ad0abpnzJ for details
package dashscope

import (
	"context"
	"errors"

	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/message"

	impl "github.com/sashabaranov/go-openai"
)

const (
	DefaultBaseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

type DashScopeModelConfig struct {
	APIKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
	Model   string `json:"model"`
}

type DashScopeClient struct {
	Config     *DashScopeModelConfig
	clientImpl *impl.Client
}

func (c *DashScopeClient) Init() error {

	return nil
}

// Creats a default DashScope model config.
func DefaultConfig(apiKey string, model string) *DashScopeModelConfig {
	return &DashScopeModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: DefaultBaseUrl,
	}
}

func NewConfig(apiKey string, model string, baseUrl string) *DashScopeModelConfig {
	if len(baseUrl) == 0 {
		baseUrl = DefaultBaseUrl
	}
	return &DashScopeModelConfig{
		APIKey:  apiKey,
		Model:   model,
		BaseUrl: baseUrl,
	}
}

func NewClient(config *DashScopeModelConfig) (*DashScopeClient, error) {

	configImpl := impl.DefaultConfig(config.APIKey)
	configImpl.BaseURL = config.BaseUrl

	configImpl.AzureModelMapperFunc = func(modelId string) string { return modelId }

	client := &DashScopeClient{
		Config:     config,
		clientImpl: impl.NewClientWithConfig(configImpl),
	}

	return client, nil
}

func (c *DashScopeClient) Chat(messages []message.Message) (*message.Message, error) {

	client := c.clientImpl
	if client == nil {
		return nil, errors.New("client not initialized")
	}

	messagesInput := openai.ConvertToOpenAIChatMessages(messages)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		impl.ChatCompletionRequest{
			Model:    c.Config.Model,
			Messages: messagesInput,
		},
	)

	if err != nil {
		return nil, err
	}
	result := message.Message{
		Content: resp.Choices[0].Message.Content,
		Role:    resp.Choices[0].Message.Role,
	}
	return &result, nil
}
