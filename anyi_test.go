package anyi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
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
		RegisterNewDefaultClient("", client)
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
	t.Run("Concurrent access", func(t *testing.T) {
		// Setup
		GlobalRegistry.Clients = make(map[string]llm.Client)
		client1 := &test.MockClient{}
		client2 := &test.MockClient{}
		RegisterClient("client1", client1)
		RegisterClient("client2", client2)
		SetDefaultClient("client1")

		// Run concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client, err := GetDefaultClient()
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}()
		}
		wg.Wait()
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

	t.Run("ConcurrentRegistration", func(t *testing.T) {
		// Reset registry for clean test
		GlobalRegistry.Clients = make(map[string]llm.Client)

		var wg sync.WaitGroup
		clients := make([]llm.Client, 100)
		for i := range clients {
			clients[i] = &test.MockClient{}
		}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := fmt.Sprintf("client%d", idx)
				err := RegisterClient(name, clients[idx])
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()

		assert.Equal(t, 100, len(GlobalRegistry.Clients))
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

	t.Run("ConcurrentRegistration", func(t *testing.T) {
		// Reset registry for clean test
		GlobalRegistry.Flows = make(map[string]*flow.Flow)

		var wg sync.WaitGroup
		flows := make([]*flow.Flow, 100)
		for i := range flows {
			flows[i] = &flow.Flow{Name: fmt.Sprintf("flow%d", i)}
		}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := fmt.Sprintf("flow%d", idx)
				err := RegisterFlow(name, flows[idx])
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()

		assert.Equal(t, 100, len(GlobalRegistry.Flows))
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

func TestNewFlowContextWithVariables(t *testing.T) {
	// Test with empty variables
	t.Run("With nil variables", func(t *testing.T) {
		text := "test text"
		memory := 42
		flowContext := NewFlowContextWithVariables(text, memory, nil)

		assert.IsType(t, &flow.FlowContext{}, flowContext)
		assert.Equal(t, text, flowContext.Text)
		assert.Equal(t, memory, flowContext.Memory)
		assert.NotNil(t, flowContext.Variables)
		assert.Equal(t, 0, len(flowContext.Variables))
	})

	// Test with variables
	t.Run("With provided variables", func(t *testing.T) {
		text := "test text"
		memory := "test memory"
		variables := map[string]any{
			"strVar":  "string value",
			"intVar":  42,
			"boolVar": true,
		}

		flowContext := NewFlowContextWithVariables(text, memory, variables)

		assert.IsType(t, &flow.FlowContext{}, flowContext)
		assert.Equal(t, text, flowContext.Text)
		assert.Equal(t, memory, flowContext.Memory)
		assert.NotNil(t, flowContext.Variables)
		assert.Equal(t, 3, len(flowContext.Variables))
		assert.Equal(t, "string value", flowContext.Variables["strVar"])
		assert.Equal(t, 42, flowContext.Variables["intVar"])
		assert.Equal(t, true, flowContext.Variables["boolVar"])
	})

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
	assert.NotNil(t, GlobalRegistry.Executors["setVariables"])
	assert.NotNil(t, GlobalRegistry.Executors["setVariable"]) // backward compatibility

	assert.NotNil(t, GlobalRegistry.Validators["json"])
	assert.NotNil(t, GlobalRegistry.Validators["string"])
}

func TestGetExecutor(t *testing.T) {
	t.Run("PointerTypeExecutor", func(t *testing.T) {
		// 注册指针类型的executor（需要完整参数）
		exec := &LLMExecutor{
			Template:      "test template", // 必须设置模板
			SystemMessage: "test system",
			OutputJSON:    true,
		}
		name := "pointer_llm_exec"
		RegisterExecutor(name, exec)

		// 第一次获取
		got1, err := GetExecutor(name)
		assert.NoError(t, err)
		assert.IsType(t, &LLMExecutor{}, got1)
		assert.Equal(t, "test template", got1.(*LLMExecutor).Template)

		// 验证返回的是新实例
		got2, err := GetExecutor(name)
		assert.NoError(t, err)
		assert.NotSame(t, got1, got2, "应该返回新的实例指针")
	})

	t.Run("ValueTypeExecutor", func(t *testing.T) {
		// 注册值类型的executor（需要完整参数）
		name := "value_llm_exec"
		RegisterExecutor(name, &LLMExecutor{
			TemplateFile:  "test.tmpl", // 使用模板文件
			SystemMessage: "value system",
		})

		// 获取executor
		got, err := GetExecutor(name)
		assert.NoError(t, err)
		assert.IsType(t, &LLMExecutor{}, got)
		assert.Equal(t, "value system", got.(*LLMExecutor).SystemMessage)
	})

	t.Run("NonExistingExecutor", func(t *testing.T) {
		_, err := GetExecutor("not_exist")
		assert.Error(t, err)
		assert.EqualError(t, err, "no executor found with the given name: not_exist")
	})
}

func TestGetValidator(t *testing.T) {
	t.Run("PointerTypeValidator", func(t *testing.T) {
		// 注册指针类型的validator
		val := &StringValidator{
			EqualTo: "test",
		}
		name := "pointer_string_val"
		RegisterValidator(name, val)

		// 第一次获取
		got1, err := GetValidator(name)
		assert.NoError(t, err)
		assert.IsType(t, &StringValidator{}, got1)
		assert.Equal(t, "test", got1.(*StringValidator).EqualTo)

		// 验证返回的是同一个实例（假设validator是单例模式）
		got2, err := GetValidator(name)
		assert.NoError(t, err)
		assert.NotSame(t, got1, got2, "应该返回新的实例指针")
	})

	t.Run("ValueTypeValidator", func(t *testing.T) {
		// 注册值类型的validator
		name := "value_string_val"
		RegisterValidator(name, &StringValidator{
			EqualTo: "test",
		})

		// 获取validator
		got, err := GetValidator(name)
		assert.NoError(t, err)
		assert.IsType(t, &StringValidator{}, got)
		assert.Equal(t, "test", got.(*StringValidator).EqualTo)
	})

	t.Run("NonExistingValidator", func(t *testing.T) {
		_, err := GetValidator("not_exist_val")
		assert.Error(t, err)
		assert.EqualError(t, err, "no validator found with the given name: not_exist_val")
	})
}

// TestRegisterFormatter tests the RegisterFormatter function
func TestRegisterFormatter(t *testing.T) {
	// Save the original registry and restore it after tests
	origRegistry := GlobalRegistry
	defer func() { GlobalRegistry = origRegistry }()

	t.Run("Success case", func(t *testing.T) {
		// Setup a fresh registry
		GlobalRegistry = &anyiRegistry{
			Formatters: make(map[string]chat.PromptFormatter),
		}

		// Create a formatter to register
		templateString := "Hello, {{.Name}}!"
		formatter, err := chat.NewPromptTemplateFormatter(templateString)
		assert.NoError(t, err)

		// Execute the register function
		err = RegisterFormatter("test-formatter", formatter)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, formatter, GlobalRegistry.Formatters["test-formatter"])
	})

	t.Run("Empty name", func(t *testing.T) {
		// Setup
		GlobalRegistry = &anyiRegistry{
			Formatters: make(map[string]chat.PromptFormatter),
		}

		// Create a formatter
		templateString := "Hello, {{.Name}}!"
		formatter, err := chat.NewPromptTemplateFormatter(templateString)
		assert.NoError(t, err)

		// Execute with empty name
		err = RegisterFormatter("", formatter)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, "name cannot be empty", err.Error())
	})

	t.Run("Overwriting existing formatter", func(t *testing.T) {
		// Setup
		GlobalRegistry = &anyiRegistry{
			Formatters: make(map[string]chat.PromptFormatter),
		}

		// Create and register a formatter
		formatter1, _ := chat.NewPromptTemplateFormatter("Template 1")
		err := RegisterFormatter("formatter", formatter1)
		assert.NoError(t, err)

		// Create a second formatter
		formatter2, _ := chat.NewPromptTemplateFormatter("Template 2")

		// Execute - register with the same name
		err = RegisterFormatter("formatter", formatter2)

		// Verify - should overwrite without error
		assert.NoError(t, err)
		assert.Equal(t, formatter2, GlobalRegistry.Formatters["formatter"])
		assert.NotEqual(t, formatter1, GlobalRegistry.Formatters["formatter"])
	})
}

// TestNewClientFromConfigFile tests the NewClientFromConfigFile function
func TestNewClientFromConfigFile(t *testing.T) {
	// Save original registry to restore after tests
	origRegistry := GlobalRegistry
	defer func() { GlobalRegistry = origRegistry }()

	// Create a test config file
	configContent := `
type: "openai"
config:
  apiKey: "test-key"
  model: "test-model"
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	t.Run("Success case with name", func(t *testing.T) {
		// Setup a fresh registry
		GlobalRegistry = &anyiRegistry{
			Clients: make(map[string]llm.Client),
		}

		// Execute
		client, err := NewClientFromConfigFile("test-client", tmpFile.Name())

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// Check that the client was registered
		registeredClient, err := GetClient("test-client")
		assert.NoError(t, err)
		assert.Equal(t, client, registeredClient)
	})

	t.Run("Success case without name", func(t *testing.T) {
		// Setup a fresh registry
		GlobalRegistry = &anyiRegistry{
			Clients: make(map[string]llm.Client),
		}

		// Execute with empty name
		client, err := NewClientFromConfigFile("", tmpFile.Name())

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// Check that the client wasn't registered
		assert.Equal(t, 0, len(GlobalRegistry.Clients))
	})

	t.Run("Invalid config file", func(t *testing.T) {
		// Create an invalid config file
		invalidContent := `
type: "invalid"
config:
  invalid: true
`
		invalidFile, err := os.CreateTemp("", "invalid-config-*.yaml")
		assert.NoError(t, err)
		defer os.Remove(invalidFile.Name())

		_, err = invalidFile.WriteString(invalidContent)
		assert.NoError(t, err)
		err = invalidFile.Close()
		assert.NoError(t, err)

		// Execute
		client, err := NewClientFromConfigFile("test-client", invalidFile.Name())

		// Verify
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		// Execute with a non-existent file
		client, err := NewClientFromConfigFile("test-client", "non-existent-file.yaml")

		// Verify
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
