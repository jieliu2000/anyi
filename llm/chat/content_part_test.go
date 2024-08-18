package chat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImageContentFromFile_InvalidFile(t *testing.T) {
	text := "Here is an image:"
	imageFilePath := "testdata/invalid_image.jpg"

	_, err := NewImagePartFromFile(text, imageFilePath)
	if err == nil {
		t.Errorf("Expected error for invalid file, got nil")
	}
}

func TestNewImageContentFromFile_ValidFile(t *testing.T) {
	detail := "low"
	imageFilePath := "../../internal/test/number_six.png"

	part, err := NewImagePartFromFile(imageFilePath, detail)

	assert.NoError(t, err)
	assert.Equal(t, detail, part.ImageDetail)
	assert.Empty(t, part.Text)
	assert.NotEmpty(t, part.ImageUrl)
	assert.True(t, strings.HasPrefix(part.ImageUrl, "data:image/png"))
}

func TestNewImagePartFromUrl(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		imageUrl := "https://example.com/image.jpg"
		detail := "This is an example image"

		contentPart, err := NewImagePartFromUrl(imageUrl, detail)
		assert.Nil(t, err)
		assert.NotNil(t, contentPart)
		assert.Equal(t, detail, contentPart.ImageDetail)
		assert.Equal(t, imageUrl, contentPart.ImageUrl)
	})

	t.Run("Empty image URL", func(t *testing.T) {
		imageUrl := ""
		detail := "This is an example image"

		contentPart, err := NewImagePartFromUrl(imageUrl, detail)
		assert.Nil(t, contentPart)
		assert.Error(t, err)
	})
	t.Run("Empty detail", func(t *testing.T) {
		imageUrl := "https://example.com/image.jpg"
		detail := ""

		contentPart, err := NewImagePartFromUrl(imageUrl, detail)
		assert.Nil(t, err)
		assert.NotNil(t, contentPart)
		assert.Equal(t, detail, contentPart.ImageDetail)
		assert.Equal(t, imageUrl, contentPart.ImageUrl)
	})
}

func TestAddImagePartFromUrl_Success(t *testing.T) {
	message := Message{ContentParts: []ContentPart{}}
	err := message.AddImagePartFromUrl("http://example.com/image.jpg", "A beautiful sunny day")
	assert.Nil(t, err)
	assert.Len(t, message.ContentParts, 1)
	assert.Equal(t, "A beautiful sunny day", message.ContentParts[0].ImageDetail)
	assert.Equal(t, "http://example.com/image.jpg", message.ContentParts[0].ImageUrl)
}

func TestAddImagePartFromUrl_EmptyUrl(t *testing.T) {
	message := Message{ContentParts: []ContentPart{}}
	err := message.AddImagePartFromUrl("", "A beautiful sunny day")
	assert.NotNil(t, err)
	assert.Equal(t, "image url is empty", err.Error())
	assert.Len(t, message.ContentParts, 0)
}
