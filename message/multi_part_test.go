package message

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
	imageFilePath := "../internal/test/number_six.png"

	part, err := NewImagePartFromFile(imageFilePath, detail)

	assert.NoError(t, err)
	assert.Equal(t, detail, part.ImageDetail)
	assert.Empty(t, part.Text)
	assert.NotEmpty(t, part.ImageUrl)
	assert.True(t, strings.HasPrefix(part.ImageUrl, "data:image/png"))
}
