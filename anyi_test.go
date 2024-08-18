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
