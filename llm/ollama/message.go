package ollama

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/jieliu2000/anyi/chat"
)

type OllamaMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

func ConvertToOllamaMessages(messages []chat.Message) (ollamaMessages []OllamaMessage, err error) {
	for _, message := range messages {
		ollamaMessage, err := ConvertToOllamaMessage(message)
		if err != nil {
			return nil, err
		}
		ollamaMessages = append(ollamaMessages, *ollamaMessage)
	}
	return ollamaMessages, nil
}

func ConvertToOllamaMessage(message chat.Message) (ollamaMessage *OllamaMessage, err error) {

	ollamaMessage = &OllamaMessage{
		Role:    message.Role,
		Content: message.Content,
	}
	if len(message.ContentParts) > 0 {
		for _, part := range message.ContentParts {
			if ollamaMessage.Content == "" && part.Text != "" {
				ollamaMessage.Content = part.Text
			}
			if part.ImageUrl != "" {
				base64String := ""
				if strings.HasPrefix(part.ImageUrl, "data:image") {
					parts := strings.Split(part.ImageUrl, ",")
					if len(parts) != 2 {
						return nil, errors.New("invalid image url")
					}
					base64String = parts[1]
				} else {
					base64String, err = base64Encode(part.ImageUrl)
					if err != nil {
						return nil, err
					}
				}
				ollamaMessage.Images = append(ollamaMessage.Images, base64String)
			}
		}
	}
	return ollamaMessage, nil
}

func base64Encode(url string) (string, error) {
	// Create a new buffer base on file size
	var b bytes.Buffer
	// Create a new writer using our buffer
	w := base64.NewEncoder(base64.StdEncoding, &b)
	// Create a new http get request using our url
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// Use our writer to write the body of our http response to our buffer
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return "", err
	}
	// Flush our writer
	defer w.Close()
	// Close our response
	resp.Body.Close()
	// Return the base64 string of our buffer
	return b.String(), nil
}
