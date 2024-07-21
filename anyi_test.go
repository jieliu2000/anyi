package anyi

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/stretchr/testify/assert"
)

func TestNewClientWithName(t *testing.T) {
	openaiConfig := openai.DefaultConfig("test")
	client, err := NewClient(openaiConfig, "openai")

	assert.NoError(t, err)
	assert.NotNil(t, client)

	client1, err := GetClient("openai")
	assert.NoError(t, err)
	assert.Equal(t, client1, client)

	client, err = NewClient(nil, "openai")
	assert.Error(t, err)
	assert.Nil(t, client)

	client, err = NewClient(openaiConfig, "")
	assert.NoError(t, err)
	assert.NotNil(t, client)

}

func TestAddClient(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		client := &test.MockClient{}
		name := "test_client"
		err := AddClient(client, name)
		assert.Nil(t, err)
		assert.Equal(t, client, Anyi.Clients[name])

		client1, err := GetClient(name)
		assert.NoError(t, err)
		assert.Equal(t, client1, client)
	})

	t.Run("EmptyName", func(t *testing.T) {
		client := &test.MockClient{}
		name := ""
		err := AddClient(client, name)
		assert.Equal(t, err, errors.New("name cannot be empty"))
	})

	t.Run("NilClient", func(t *testing.T) {
		client := llm.Client(nil)
		name := "nil_client"
		err := AddClient(client, name)
		assert.Equal(t, err, errors.New("client cannot be empty"))
	})

	t.Run("NilParams", func(t *testing.T) {
		err := AddClient(nil, "")
		assert.Equal(t, err, errors.New("client cannot be empty"))
	})
}

func TestGetClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := &test.MockClient{}
		name := "get_client"
		err := AddClient(client, name)
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
		AddClient(client, name)
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
		formatter, err := NewPromptTemplateFormatter(template)
		assert.NoError(t, err)
		assert.NotNil(t, formatter)
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		template := "Hello, {{.name" // Incomplete placeholder
		formatter, err := NewPromptTemplateFormatter(template)
		assert.Error(t, err)
		assert.Nil(t, formatter)
	})

}
