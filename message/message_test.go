package message

import (
	"strings"
	"testing"

	"github.com/jieliu2000/anyi/llm/tools"
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

func TestNewPromptTemplateFormatterFromFile(t *testing.T) {
	var tmplFile = "../internal/test/test_prompt1.tmpl"
	formatter, err := NewPromptTemplateFormatterFromFile(tmplFile)

	if err != nil {
		panic(err)
	}

	type AgentTasks struct {
		Tasks     []string
		Objective string
	}

	tasks := AgentTasks{
		Tasks:     []string{"task1", "task2"},
		Objective: "objective",
	}
	result, err := formatter.Format(tasks)
	if err != nil {
		panic(err)
	}
	assert.Greater(t, len(result), 0)
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

func TestAddFunctionDirectivesToPrompt(t *testing.T) {
	testCases := []struct {
		objective   string
		functions   []tools.FunctionConfig
		expectError bool
		wantResult  string
	}{
		{
			objective:   "Test objective",
			functions:   []tools.FunctionConfig{{Name: "add", Description: "Add two numbers", Params: []tools.ParameterConfig{{Name: "a", Type: "int", Description: "First number"}, {Name: "b", Type: "int", Description: "Second number"}}}},
			expectError: false,
			wantResult:  "Your task is to generate a task list in JSON array format to archieve this target: '''Test objective'''\n\tYou can use the following functions in generating the task list:\n\t* add: Add two numbers.  Parameters:\t- a(type: int): First number, \t- b(type: int): Second number,  \n\t\n\tEach task json should be in this format:\n\t{\"function\": \"$function_name\", \"params\": [{\"param_name\": \"$param_name\", \"param_value\": \"$param_value\"}]}\n\n\tThe output should be a JSON array of JSONs in above format.\n\n\tFor example, if you only need one task with a function \"add\" with two params \"a\" and \"b\" which have values 1 and 2 , you should finally out\n\tput an array like this:\n\t[{\"function\": \"add\", \"params\": [{\"param_name\": \"a\", \"param_value\": 1}, {\"param_name\": \"b\", \"param_value\": 2}]}]\n\tDO NOT add any extra text execept the task list JSON array.",
		},
	}
	for _, tc := range testCases {
		result, err := AddFunctionDirectivesToPrompt(tc.objective, tc.functions)
		if tc.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
		}
	}
}

func TestNewImageMessage(t *testing.T) {
	t.Run("With valid image", func(t *testing.T) {
		role := "user"
		content := "image.jpg"
		filePath := "../internal/test/number_six.png"

		message := NewImageMessageFromFile(role, content, filePath)

		assert.Equal(t, role, message.Role, "Expected role to be %s, but got %s", role, message.Role)

		assert.Equal(t, 2, len(message.MultiParts), "Expected image contents to have one element, but got %d", len(message.MultiParts))
		textPart := message.MultiParts[0]
		assert.Equal(t, content, textPart.Text)
		imagePart := message.MultiParts[1]
		assert.True(t, strings.HasPrefix(imagePart.ImageUrl, "data:image/png"), "Expected image url to start with data:image, but got %s", textPart.ImageUrl)
	})

	t.Run("With empty filePath", func(t *testing.T) {
		role := "user"
		content := "content"

		message := NewImageMessageFromFile("user", content, "")

		assert.Equal(t, role, message.Role, "Expected role to be %s, but got %s", role, message.Role)
		assert.Equal(t, "", message.Content)
		assert.Equal(t, 0, len(message.MultiParts), "Expected image contents to have 0 element, but got %d", len(message.MultiParts))
	})

	t.Run("With invalid path", func(t *testing.T) {
		role := "user"
		content := "image.jpg"
		filePath := "../internal/test/invalid_file.png"
		message := NewImageMessageFromFile(role, content, filePath)
		assert.Equal(t, role, message.Role, "Expected role to be %s, but got %s", role, message.Role)
		assert.Equal(t, "", message.Content)
		assert.Equal(t, 0, len(message.MultiParts), "Expected image contents to have 0 element, but got %d", len(message.MultiParts))
	})

}
