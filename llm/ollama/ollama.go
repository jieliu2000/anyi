package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
)

const (
	DefaultOllamaUrl = "http://localhost:11434/api"
)

type OllamaModelConfig struct {
	//The url of the ollama server. Note that don't add "/chat" to the end of this url. In [Chat] function it will be added automatically.
	OllamaApiURL string `json:"ollama_api_url"`

	//The model name used by ollama. See [Ollama's documentation] for more information on the available models.
	//
	//[Ollama's documentation]: https://github.com/ollama/ollama/blob/main/README.md#quickstart
	Model string `json:"model"`
}

type OllamaClient struct {
	Config     *OllamaModelConfig
	clientImpl *http.Client
}

// Creats a default Ollama model config.
func DefaultConfig(model string) *OllamaModelConfig {
	return &OllamaModelConfig{
		Model:        model,
		OllamaApiURL: DefaultOllamaUrl,
	}
}

// NewConfig creates and returns a pointer to OllamaModelConfig.
// Parameters:
// - model string: The model name used by ollama. See [Ollama's documentation] for more information on the available models.
// - ollamaApiURL string: The Ollama API URL. Leaving this blank will use the default Ollama API URL. See [DefaultOllamaUrl]. Note that don't add "/chat" to the end of this url. In [Chat] function it will be added automatically.
//
// [Ollama's documentation]: https://github.com/ollama/ollama/blob/main/README.md
func NewConfig(model string, ollamaApiURL string) *OllamaModelConfig {
	if len(ollamaApiURL) == 0 {
		ollamaApiURL = DefaultOllamaUrl
	}
	return &OllamaModelConfig{
		Model:        model,
		OllamaApiURL: ollamaApiURL,
	}
}

// NewClient creates a new OllamaClient instance based on the provided OllamaModelConfig.
// If the config is nil, it will return an error. If the model in the config is empty, it will return an error as well because ollama chat cannot be called without a model.
// the OllamaApiURL in the config can be left blank. The default Ollama API URL will be used in that case.
func NewClient(config *OllamaModelConfig) (*OllamaClient, error) {

	// Check if the config is nil to prevent panic or unexpected behavior
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	if config.OllamaApiURL == "" {
		config.OllamaApiURL = DefaultOllamaUrl
	}

	if config.Model == "" {
		return nil, errors.New("model cannot be empty")
	}

	// Create a new OllamaClient using the provided config and the configured client implementation
	client := &OllamaClient{
		Config: config,
	}

	client.clientImpl = &http.Client{}

	// Return the newly created OllamaClient and nil error
	return client, nil
}

type OllamaRequest struct {
	Model    string                   `json:"model"`
	Messages []OllamaMessage          `json:"messages"`
	Stream   bool                     `json:"stream"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
	Format   string                   `json:"format,omitempty"`
}

type OllamaParameterDetail struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type OllamaParameters struct {
	Type       string                           `json:"type"`
	Properties map[string]OllamaParameterDetail `json:"properties"`
}

type OllamaFunction struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Parameters  OllamaParameters `json:"parameters,omitempty"`
}

type OllamaResponse struct {
	Message         chat.Message `json:"message"`
	CreatedAt       time.Time    `json:"created_at"`
	Done            bool         `json:"done"`
	TotalDuration   int          `json:"total_duration"`
	LoadDuration    int          `json:"load_duration"`
	PromptEvalCount int          `json:"prompt_eval_count"`
	EvalCount       int          `json:"eval_count"`
}

func convertToOllamaFunction(function tools.FunctionConfig) OllamaFunction {

	properties := make(map[string]OllamaParameterDetail)

	for _, param := range function.Params {
		properties[param.Name] = OllamaParameterDetail{
			Type:        param.Type,
			Description: param.Description,
			Enum:        param.Enum,
		}
	}

	return OllamaFunction{
		Name:        function.Name,
		Description: function.Description,
		Parameters: OllamaParameters{
			Type:       "object",
			Properties: properties,
		},
	}
}

func ConvertToOllamaTools(functions []tools.FunctionConfig) ([]map[string]interface{}, error) {
	if len(functions) == 0 {
		return nil, errors.New("no functions provided")
	}

	var ollamaFunctions []map[string]interface{}
	for _, function := range functions {
		ollamaFuncDesciption := make(map[string]interface{})
		ollamaFuncDesciption["type"] = "function"

		ollamaFuncDesciption["function"] = convertToOllamaFunction(function)

		ollamaFunctions = append(ollamaFunctions, ollamaFuncDesciption)
	}

	return ollamaFunctions, nil
}

func (c *OllamaClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {

	response := chat.ResponseInfo{}
	httpClient := c.clientImpl

	if httpClient == nil {
		return nil, response, errors.New("http client cannot be nil, maybe you didn't initiatialize the client. Considering using NewClient function")
	}

	ollamaMessages, err := ConvertToOllamaMessages(messages)
	if err != nil {
		return nil, response, err
	}

	tools, err := ConvertToOllamaTools(functions)

	if err != nil {
		return nil, response, err
	}

	request := &OllamaRequest{}
	chat.SetChatOptions(options, &request)
	request.Model = c.Config.Model
	request.Messages = ollamaMessages
	request.Tools = tools

	return c.callOllamaAPI(request, response, httpClient)
}

func (c *OllamaClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {

	response := chat.ResponseInfo{}
	httpClient := c.clientImpl

	if httpClient == nil {
		return nil, response, errors.New("http client cannot be nil, maybe you didn't initiatialize the client. Considering using NewClient function")
	}

	ollamaMessages, err := ConvertToOllamaMessages(messages)
	if err != nil {
		return nil, response, err
	}

	request := &OllamaRequest{}
	chat.SetChatOptions(options, &request)
	request.Model = c.Config.Model
	request.Messages = ollamaMessages

	return c.callOllamaAPI(request, response, httpClient)
}

func (c *OllamaClient) callOllamaAPI(request *OllamaRequest, response chat.ResponseInfo, httpClient *http.Client) (*chat.Message, chat.ResponseInfo, error) {
	requestJson, err := json.Marshal(*request)

	if err != nil {
		return nil, response, err
	}

	res, err := httpClient.Post(c.Config.OllamaApiURL+"/chat", "application/json", bytes.NewBuffer(requestJson))
	log.Print(string(requestJson))

	if err != nil {
		return nil, response, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, response, fmt.Errorf("error response status from ollama chat api: %d", res.StatusCode)
	}

	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, response, err
	}

	ollamaResponse := OllamaResponse{}
	err = json.Unmarshal(responseBody, &ollamaResponse)
	if err != nil {
		return nil, response, err
	}

	response.PromptTokens = ollamaResponse.PromptEvalCount
	response.CompletionTokens = ollamaResponse.EvalCount

	return &ollamaResponse.Message, response, nil
}
