package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPrompt(t *testing.T) {

	msg := NewMessage("user", "hello")
	assert.Equal(t, "user", msg.Role, "role should be user")
	assert.Equal(t, "hello", msg.Content, "content should be hello")

	msg = NewSystemMessage("you are an assisstant")
	assert.Equal(t, "system", msg.Role, "role should be system")
	assert.Equal(t, "you are an assisstant", msg.Content, "content should be 'you are an assisstant'")

	msg = NewUserMessage("6+1=")
	assert.Equal(t, "user", msg.Role, "role should be user")
	assert.Equal(t, "6+1=", msg.Content, "content should be '6+1='")

	msg = NewAssistantMessage("7")
	assert.Equal(t, "assistant", msg.Role, "role should be assistant")
	assert.Equal(t, "7", msg.Content, "content should be '7'")
}

func TestMessage_ToJSON(t *testing.T) {
	m := NewMessage("user", "Hello, world!")
	expected := `{"content":"Hello, world!","role":"user"}`
	assert.Equal(t, expected, m.ToJSON())
}

func TestNewSystemMessage(t *testing.T) {
	m := NewSystemMessage("Hello, world!")

	assert.Equal(t, "system", m.Role)
	assert.Equal(t, "Hello, world!", m.Content)
}

func TestNewUserMessage(t *testing.T) {
	m := NewUserMessage("Hello, world!")
	assert.Equal(t, "user", m.Role)
	assert.Equal(t, "Hello, world!", m.Content)
}

func TestNewAssistantMessage(t *testing.T) {
	m := NewAssistantMessage("Hello, world!")
	assert.Equal(t, "assistant", m.Role)
	assert.Equal(t, "Hello, world!", m.Content)
}

func TestNewPromptTemplateFormatter(t *testing.T) {

	t.Run("Success with map", func(t *testing.T) {
		f, err := NewPromptTemplateFormatter("Hello, {{.Name}}!")

		assert.Nil(t, err)

		input := map[string]interface{}{
			"Name": "world",
		}

		expected := "Hello, world!"

		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)
	})
	t.Run("Success with struct", func(t *testing.T) {
		type User struct {
			Name string
		}
		f, err := NewPromptTemplateFormatter("Hello, {{.Name}}!")
		assert.Nil(t, err)
		input := User{
			Name: "world",
		}
		expected := "Hello, world!"
		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)
	})
	t.Run("Success with struct pointer", func(t *testing.T) {

		type User struct {
			Name string
		}
		f, err := NewPromptTemplateFormatter("Hello, {{.Name}}!")
		assert.NoError(t, err)
		input := &User{
			Name: "world",
		}
		expected := "Hello, world!"
		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("Success with plain text", func(t *testing.T) {
		f, err := NewPromptTemplateFormatter("Hello, {{.}}!")
		assert.Nil(t, err)
		input := "world"
		expected := "Hello, world!"
		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("Success with array", func(t *testing.T) {

		f, err := NewPromptTemplateFormatter("Hello, {{index . 0}}!")
		assert.Nil(t, err)
		input := []string{
			"world",
		}
		expected := "Hello, world!"
		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("Success with struct array", func(t *testing.T) {

		type User struct {
			Name string
		}
		input := []User{
			{
				Name: "world",
			},
		}

		f, err := NewPromptTemplateFormatter(`Hello, {{(index . 0).Name}}!`)
		assert.Nil(t, err)

		expected := "Hello, world!"
		output, err := f.Format(input)
		assert.Equal(t, expected, output)
		assert.NoError(t, err)

	})

	t.Run("Error with invalid template", func(t *testing.T) {
		f, err := NewPromptTemplateFormatter("Hello, {{.Name")
		assert.Nil(t, f)
		assert.Error(t, err)
	})

	t.Run("Error with invalid input", func(t *testing.T) {
		f, err := NewPromptTemplateFormatter("Hello, {{.Name}}!")
		assert.NotNil(t, f)
		assert.NoError(t, err)

		input := "world"
		_, err = f.Format(input)

		assert.Error(t, err)

	})

}
