package anyi

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/stretchr/testify/assert"
)

func TestNewClientWithName(t *testing.T) {
	openaiConfig := openai.DefaultConfig("test")
	client, err := NewClient("openai", openaiConfig)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	client1, err := GetClient("openai")
	assert.NoError(t, err)
	assert.Equal(t, client1, client)

	client, err = NewClient("openai", nil)
	assert.Error(t, err)
	assert.Nil(t, client)

	client, err = NewClient("", openaiConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)

}

func TestGetDefaultClient(t *testing.T) {
	t.Run("No default client", func(t *testing.T) {
		GlobalRegistry.Clients = make(map[string]llm.Client)
		_, err := GetDefaultClient()
		assert.Error(t, err)
	})
	t.Run("Set default client via RegisterDefaultClient", func(t *testing.T) {
		client := &test.MockClient{}
		RegisterDefaultClient("", client)
		got, err := GetDefaultClient()
		assert.NoError(t, err)
		assert.Equal(t, client, got)
	})
	t.Run("Set default client", func(t *testing.T) {
		client := &test.MockClient{}
		GlobalRegistry.Clients["default"] = client
		got, err := GetDefaultClient()
		assert.NoError(t, err)
		assert.Equal(t, client, got)
	})
	t.Run("Only one client", func(t *testing.T) {
		// Arrange
		GlobalRegistry.Clients = make(map[string]llm.Client)
		GlobalRegistry.Clients["test"] = &test.MockClient{}

		// Act
		client, err := GetDefaultClient()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestRegisterClient(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		client := &test.MockClient{}
		name := "test_client"
		err := RegisterClient(name, client)
		assert.Nil(t, err)
		assert.Equal(t, client, GlobalRegistry.Clients[name])

		client1, err := GetClient(name)
		assert.NoError(t, err)
		assert.Equal(t, client1, client)
	})

	t.Run("EmptyName", func(t *testing.T) {
		client := &test.MockClient{}
		name := ""
		err := RegisterClient(name, client)
		assert.Equal(t, err, errors.New("name cannot be empty"))
	})

	t.Run("NilClient", func(t *testing.T) {
		client := llm.Client(nil)
		name := "nil_client"
		err := RegisterClient(name, client)
		assert.Equal(t, err, errors.New("client cannot be empty"))
	})

	t.Run("NilParams", func(t *testing.T) {
		err := RegisterClient("", nil)
		assert.Equal(t, err, errors.New("client cannot be empty"))
	})
}

func TestRegisterFlow(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		flow := &flow.Flow{}
		name := "test_flow"
		err := RegisterFlow(name, flow)
		assert.Nil(t, err)
		assert.Equal(t, flow, GlobalRegistry.Flows[name])

		client1, err := GetFlow(name)
		assert.NoError(t, err)
		assert.Equal(t, client1, flow)
	})

	t.Run("EmptyName", func(t *testing.T) {
		flow := &flow.Flow{}
		name := ""
		err := RegisterFlow(name, flow)
		assert.Equal(t, err, errors.New("name cannot be empty"))
	})

	t.Run("NilParams", func(t *testing.T) {
		err := RegisterFlow("", nil)
		assert.Equal(t, err, errors.New("name cannot be empty"))
	})
}

func TestGetClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := &test.MockClient{}
		name := "get_client"
		err := RegisterClient(name, client)
		assert.Nil(t, err)
		client1, err := GetClient(name)
		assert.NoError(t, err)
		assert.Equal(t, client1, client)
	})
	t.Run("EmptyName", func(t *testing.T) {
		_, err := GetClient("")
		assert.Error(t, err)
	})
	t.Run("NotExist", func(t *testing.T) {
		client := &test.MockClient{}
		name := "get_client"
		RegisterClient(name, client)
		result, err := GetClient("not_exist")
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
func TestNewMessage(t *testing.T) {

	role := "user"
	content := "Hello, world!"
	msg := NewMessage(role, content)

	jsonString := msg.ToJSON()

	target := make(map[string]string)

	json.Unmarshal([]byte(jsonString), &target)

	assert.Equal(t, "user", target["role"])
	assert.Equal(t, "Hello, world!", target["content"])
}

func TestNewPromptTemplateFormatter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		template := "Hello, {{.Name}}!"
		formatter, err := NewPromptTemplateFormatter("template1", template)
		assert.NoError(t, err)
		assert.NotNil(t, formatter)

		formatter, ok := (GetFormatter("template1")).(*chat.PromptyTemplateFormatter)
		assert.True(t, ok)
		assert.Equal(t, template, formatter.TemplateString)
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		template := "Hello, {{.name" // Incomplete placeholder
		formatter, err := NewPromptTemplateFormatter("name1", template)
		assert.Error(t, err)

		assert.Nil(t, GetFormatter("name1"))
		assert.Nil(t, formatter)
	})

}

func TestNewLLMStepExecutorWithFormatter(t *testing.T) {
	name := "test_executor"
	formatter := &chat.PromptyTemplateFormatter{
		TemplateString: "Hello, {{.Name}}!",
	}
	systemMessage := "Reactions please!"
	client := &test.MockClient{}

	stepExecutor := NewLLMStepExecutorWithFormatter(name, formatter, systemMessage, client)
	assert.NotNil(t, stepExecutor)
	assert.Equal(t, formatter, stepExecutor.TemplateFormatter)
	assert.Equal(t, systemMessage, stepExecutor.SystemMessage)

	retrievedExecutor := GlobalRegistry.Executors[name]
	assert.Equal(t, stepExecutor, retrievedExecutor)
}

func TestGetFlow(t *testing.T) {
	t.Run("with an existing flow", func(t *testing.T) {
		flowName := "test_flow"
		GlobalRegistry.Flows[flowName] = &flow.Flow{
			Name: flowName,
		}
		f, err := GetFlow(flowName)
		assert.Nil(t, err)
		assert.Equal(t, flowName, f.Name)
	})
	t.Run("with a non-existing flow", func(t *testing.T) {
		flowName := "non_existing_flow"
		f, err := GetFlow(flowName)
		assert.Nil(t, f)
		assert.EqualError(t, err, "no flow found with the given name: "+flowName)
	})
	t.Run("with an empty name", func(t *testing.T) {
		f, err := GetFlow("")
		assert.Nil(t, f)
		assert.EqualError(t, err, "name cannot be empty")
	})
}

func TestNewFlowContext(t *testing.T) {
	input := "test"
	flowContext := NewFlowContextWithText(input)

	assert.IsType(t, &flow.FlowContext{}, flowContext)
	assert.Equal(t, input, flowContext.Text)
	assert.Nil(t, flowContext.Memory)

	flowContext = NewFlowContextWithMemory(5)
	assert.Equal(t, 5, flowContext.Memory)
	assert.Equal(t, "", flowContext.Text)

}

func TestNewFlow(t *testing.T) {
	t.Run("creates a new flow with the given name and steps", func(t *testing.T) {
		name := "test_flow"
		client := test.MockClient{}
		steps := []flow.Step{
			{},
			{},
		}

		flow, err := NewFlow(name, &client, steps...)
		assert.NoError(t, err)
		assert.Equal(t, name, flow.Name)
		assert.Equal(t, len(steps), len(flow.Steps))

	})

	t.Run("returns an error if the name is empty", func(t *testing.T) {
		client := test.MockClient{}
		steps := []flow.Step{}

		flow, err := NewFlow("", &client, steps...)
		assert.Error(t, err)
		assert.Nil(t, flow)
	})

	t.Run("returns an error if the flow cannot be created", func(t *testing.T) {
		name := "invalid_flow"
		client := test.MockClient{}

		flow, err := NewFlow(name, &client)
		assert.Error(t, err)
		assert.Nil(t, flow)
	})
}

func TestInit(t *testing.T) {
	// Execute
	Init()
	// Verify
	assert.NotNil(t, GlobalRegistry.Executors["llm"])
	assert.NotNil(t, GlobalRegistry.Executors["condition"])
	assert.NotNil(t, GlobalRegistry.Executors["exec"])
	assert.NotNil(t, GlobalRegistry.Executors["setContext"])

	assert.NotNil(t, GlobalRegistry.Validators["json"])
	assert.NotNil(t, GlobalRegistry.Validators["string"])

}
