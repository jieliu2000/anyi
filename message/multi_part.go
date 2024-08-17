package message

import (
	"encoding/base64"
	"net/http"
	"os"
)

type MultiPart struct {
	Text        string `json:"text"`
	ImageUrl    string `json:"imageUrl"`
	ImageDetail string `json:"imageDetail"`
}

func (i *Message) AddImagePartFromFile(imageFilePath string, detail string) error {
	imagePart, err := NewImagePartFromFile(imageFilePath, detail)
	if err != nil {
		return err
	}

	i.MultiParts = append(i.MultiParts, *imagePart)
	return nil
}

func (i *Message) AddText(text string) error {
	textPart := NewTextPart(text)
	i.MultiParts = append(i.MultiParts, *textPart)
	return nil
}

func NewTextPart(text string) *MultiPart {
	textPart := MultiPart{
		Text: text,
	}
	return &textPart
}

func NewImagePartFromFile(imageFilePath string, detail string) (*MultiPart, error) {
	base64URL, err := toBase64(imageFilePath)
	if err != nil {
		return nil, err
	}
	imagePart := MultiPart{
		ImageDetail: detail,
		ImageUrl:    base64URL,
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
