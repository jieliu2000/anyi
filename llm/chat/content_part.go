package chat

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
)

type ContentPart struct {
	Text        string `json:"text"`
	ImageUrl    string `json:"imageUrl"`
	ImageDetail string `json:"imageDetail"`
}

// The AddImagePartFromFile function adds an image part to the message by reading from a file.
// Parameters:
// - imageFilePath string: The path to the image file.
// - detail string: detail parameter of the image. Set this to an empty string if you don't know what value you should set
// return value:
// - error: If an error occurs, the corresponding error message is returned.
func (i *Message) AddImagePartFromFile(imageFilePath string, detail string) error {
	imagePart, err := NewImagePartFromFile(imageFilePath, detail)
	if err != nil {
		return err
	}

	i.ContentParts = append(i.ContentParts, *imagePart)
	return nil
}

func (i *Message) AddImagePartFromUrl(imageUrl string, detail string) error {
	imagePart, _ := NewImagePartFromUrl(imageUrl, detail)

	i.ContentParts = append(i.ContentParts, *imagePart)
	return nil
}

// The AddTextPart function adds a text part to the chat.
// Parameters:
// - text string: The text content to be added.
// return value:
// - error: If an error occurs, the corresponding error message is returned. Otherwise, nil is returned.
func (i *Message) AddTextPart(text string) error {
	textPart := NewTextPart(text)
	i.ContentParts = append(i.ContentParts, *textPart)
	return nil
}

// NewTextPart creates a new ContentPart struct with the given text.
// Parameters:
// - text string: The text to be set in the ContentPart struct.
// Return value:
// - *ContentPart: A pointer to the newly created ContentPart struct with the specified text.
func NewTextPart(text string) *ContentPart {
	textPart := ContentPart{
		Text: text,
	}
	return &textPart
}

// NewImagePartFromFile creates a new ContentPart object from an image file. This function will read the file, convert it to a base64 encoded string, and pass the resulting string to the ImageUrl property of the ContentPart object.
// If the file path is not valid or the file is not an image, the function will return an error.
// Parameters:
// - imageFilePath string: The path to the image file.
// - detail string: detail parameter of the image. Set this to an empty string if you don't know what value you should set
// Return value:
// - *ContentPart: A pointer to the newly created ContentPart object.
// - error: If an error occurs, the corresponding error message is returned.
func NewImagePartFromFile(imageFilePath string, detail string) (*ContentPart, error) {
	base64URL, err := toBase64(imageFilePath)
	if err != nil {
		return nil, err
	}
	imagePart := ContentPart{
		ImageDetail: detail,
		ImageUrl:    base64URL,
	}
	return &imagePart, nil
}

// NewImagePartFromUrl creates a new ContentPart object from an image URL and detail string.
// Please note that this function doesn't validate the URL. If you pass an invalid URL, the function will not return an error. The AI engine will return an error obviously in this case in the AI invocation.
// Parameters:
// - imageUrl string: The URL of the image.
// - detail string: detail parameter of the image. Set this to an empty string if you don't know what value you should set
// Return value:
// - *ContentPart: A pointer to the newly created ContentPart object.
// - error: If an error occurs, the corresponding error message is returned.
func NewImagePartFromUrl(imageUrl string, detail string) (*ContentPart, error) {
	if imageUrl == "" {
		return nil, errors.New("image url is empty")
	}
	imagePart := ContentPart{
		ImageDetail: detail,
		ImageUrl:    imageUrl,
	}
	return &imagePart, nil
}

func toBase64(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}

	// Append the base64 encoded output
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	return base64Encoding, nil
}
