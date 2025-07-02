# Multimodal Applications

This tutorial covers how to work with multimodal AI models that can process both text and images. You'll learn to send images to vision-capable models, handle different image formats, and build sophisticated multimodal workflows.

## Table of Contents

- [Understanding Multimodal Models](#understanding-multimodal-models)
- [Basic Image Processing](#basic-image-processing)
- [Image Input Methods](#image-input-methods)
- [Multimodal Workflows](#multimodal-workflows)
- [Advanced Techniques](#advanced-techniques)
- [Best Practices](#best-practices)

## Understanding Multimodal Models

Multimodal AI models can process and understand multiple types of input simultaneously, typically text and images. These models can:

- **Analyze Images**: Describe what they see in images
- **Answer Questions**: About image content
- **Compare Images**: Identify similarities and differences
- **Extract Information**: Read text from images (OCR)
- **Generate Descriptions**: Create detailed image captions
- **Understand Context**: Combine visual and textual information

## Basic Image Processing

### Simple Image Analysis

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create a vision-capable client
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create message with image
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "What do you see in this image? Describe it in detail.",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/image.jpg",
					},
				},
			},
		},
	}

	// Send request
	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	log.Printf("Image description: %s", response.Content)
}
```

### Using Anthropic Claude for Image Analysis

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/anthropic"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	// Create Claude client
	config := anthropic.DefaultConfigWithModel(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-sonnet-20240229")
	client, err := anyi.NewClient("claude", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Analyze a document image
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "Please extract and summarize the key information from this document.",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/document.png",
					},
				},
			},
		},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Document analysis failed: %v", err)
	}

	log.Printf("Document summary: %s", response.Content)
}
```

## Image Input Methods

### 1. Image URLs

The most straightforward method is using publicly accessible image URLs:

```go
imageURL := &chat.ImageURL{
	URL: "https://example.com/image.jpg",
}

contentPart := chat.ContentPart{
	Type:     "image_url",
	ImageURL: imageURL,
}
```

### 2. Base64 Encoded Images

For local images or when you need to send image data directly:

```go
package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func encodeImageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(imageData)
	return encoded, nil
}

func main() {
	// Create client
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Encode local image
	base64Image, err := encodeImageToBase64("./local-image.jpg")
	if err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	// Create data URL
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// Create message with base64 image
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "Analyze this local image and tell me what you see.",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: dataURL,
					},
				},
			},
		},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Image analysis failed: %v", err)
	}

	log.Printf("Analysis: %s", response.Content)
}
```

### 3. Multiple Images

Send multiple images in a single request:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func main() {
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Compare multiple images
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "Compare these two images and tell me the differences:",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/image1.jpg",
					},
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL: "https://example.com/image2.jpg",
					},
				},
			},
		},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("Image comparison failed: %v", err)
	}

	log.Printf("Comparison result: %s", response.Content)
}
```

## Multimodal Workflows

### Document Processing Workflow

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
)

func main() {
	// Create vision-capable client
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Step 1: Extract text from document image
	extractStep := &flow.Step{
		Name: "extract_text",
		Executor: &anyi.LLMExecutor{
			Template: `Extract all text content from this document image.
			Maintain the original structure and formatting as much as possible.

			Image URLs: {{range .ImageURLs}}{{.}}{{end}}`,
			SystemMessage: "You are an expert OCR system that accurately extracts text from images.",
		},
	}

	// Step 2: Summarize extracted content
	summarizeStep := &flow.Step{
		Name: "summarize_content",
		Executor: &anyi.LLMExecutor{
			Template: `Summarize the following extracted text content:

{{.Text}}

Provide a concise summary highlighting the key points.`,
			SystemMessage: "You are a professional summarizer.",
		},
		Validator: &anyi.StringValidator{
			MinLength: 100,
			MaxLength: 500,
		},
	}

	// Step 3: Generate action items
	actionItemsStep := &flow.Step{
		Name: "generate_actions",
		Executor: &anyi.LLMExecutor{
			Template: `Based on this document summary, generate a list of action items:

{{.Text}}

Format as a numbered list of specific, actionable tasks.`,
			SystemMessage: "You are a project manager who creates actionable task lists.",
		},
	}

	// Create workflow
	documentFlow, err := anyi.NewFlow("document_processor", client,
		*extractStep, *summarizeStep, *actionItemsStep)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	// Create context with image URLs
	context := anyi.NewFlowContext("Document processing request")
	context.ImageURLs = []string{
		"https://example.com/document.pdf",
		"https://example.com/chart.png",
	}

	// Run workflow
	result, err := documentFlow.Run(context)
	if err != nil {
		log.Fatalf("Document processing failed: %v", err)
	}

	log.Printf("Action items: %s", result.Text)
}
```

### Image Analysis and Content Generation

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/anthropic"
)

func main() {
	// Create different clients for different tasks
	visionConfig := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	visionClient, _ := anyi.NewClient("vision", visionConfig)

	writerConfig := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))
	writerClient, _ := anyi.NewClient("writer", writerConfig)

	// Step 1: Analyze image with vision model
	analyzeStep := &flow.Step{
		Name: "analyze_image",
		Executor: &anyi.LLMExecutor{
			Template: `Analyze this image in detail. Describe:
1. The main subject and objects
2. The setting and environment
3. Colors, lighting, and mood
4. Any text or signage visible
5. Overall composition and style

Image URLs: {{range .ImageURLs}}{{.}}{{end}}`,
			SystemMessage: "You are an expert image analyst with a keen eye for detail.",
			Client:        visionClient,
		},
	}

	// Step 2: Generate creative content based on analysis
	createStep := &flow.Step{
		Name: "create_content",
		Executor: &anyi.LLMExecutor{
			Template: `Based on this detailed image analysis, write a creative short story (300-500 words):

{{.Text}}

The story should be inspired by the image but not limited to describing it.`,
			SystemMessage: "You are a creative writer who crafts engaging stories.",
			Client:        writerClient,
		},
		Validator: &anyi.StringValidator{
			MinLength: 300,
			MaxLength: 500,
		},
	}

	// Step 3: Create social media caption
	captionStep := &flow.Step{
		Name: "create_caption",
		Executor: &anyi.LLMExecutor{
			Template: `Create an engaging social media caption for the original image based on this story:

{{.Text}}

Keep it under 280 characters and include relevant hashtags.`,
			SystemMessage: "You are a social media expert who creates viral content.",
			Client:        writerClient,
		},
		Validator: &anyi.StringValidator{
			MaxLength: 280,
		},
	}

	// Create workflow
	creativeFlow, err := anyi.NewFlow("creative_content", visionClient,
		*analyzeStep, *createStep, *captionStep)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	// Run with image
	context := anyi.NewFlowContext("Creative content generation")
	context.ImageURLs = []string{"https://example.com/inspiration-image.jpg"}

	result, err := creativeFlow.Run(context)
	if err != nil {
		log.Fatalf("Creative workflow failed: %v", err)
	}

	log.Printf("Social media caption: %s", result.Text)
}
```

## Advanced Techniques

### Image Quality and Detail Control

Some providers allow you to control image processing quality:

```go
// High detail analysis (OpenAI)
imageURL := &chat.ImageURL{
	URL:    "https://example.com/high-res-image.jpg",
	Detail: "high", // Options: "low", "high", "auto"
}

contentPart := chat.ContentPart{
	Type:     "image_url",
	ImageURL: imageURL,
}
```

### Batch Image Processing

Process multiple images efficiently:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

type ImageAnalysisResult struct {
	ImageURL    string
	Description string
	Error       error
}

func analyzeImage(client *anyi.Client, imageURL string) ImageAnalysisResult {
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "Describe this image in one sentence.",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{URL: imageURL},
				},
			},
		},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		return ImageAnalysisResult{ImageURL: imageURL, Error: err}
	}

	return ImageAnalysisResult{
		ImageURL:    imageURL,
		Description: response.Content,
	}
}

func main() {
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// List of images to analyze
	imageURLs := []string{
		"https://example.com/image1.jpg",
		"https://example.com/image2.jpg",
		"https://example.com/image3.jpg",
		"https://example.com/image4.jpg",
	}

	// Process images concurrently
	var wg sync.WaitGroup
	results := make(chan ImageAnalysisResult, len(imageURLs))

	for _, url := range imageURLs {
		wg.Add(1)
		go func(imageURL string) {
			defer wg.Done()
			result := analyzeImage(client, imageURL)
			results <- result
		}(url)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		if result.Error != nil {
			log.Printf("Error analyzing %s: %v", result.ImageURL, result.Error)
		} else {
			fmt.Printf("Image: %s\nDescription: %s\n\n", result.ImageURL, result.Description)
		}
	}
}
```

### Multimodal RAG (Retrieval-Augmented Generation)

Combine image analysis with document retrieval:

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/openai"
)

type ImageContext struct {
	ImageURL    string
	Context     string
	Metadata    map[string]string
}

func main() {
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("multimodal_rag", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Step 1: Analyze image and extract key features
	extractFeaturesStep := &flow.Step{
		Name: "extract_features",
		Executor: &anyi.LLMExecutor{
			Template: `Analyze this image and extract key searchable features:

Image: {{range .ImageURLs}}{{.}}{{end}}

Extract:
1. Objects and entities
2. Scene type and setting
3. Colors and visual style
4. Any text or numbers
5. Technical specifications if visible

Format as JSON with clear categories.`,
			SystemMessage: "You are an expert feature extractor for image search systems.",
		},
		Validator: &anyi.JsonValidator{},
	}

	// Step 2: Retrieve relevant context (simulated)
	retrieveContextStep := &flow.Step{
		Name: "retrieve_context",
		Executor: &anyi.SetContextExecutor{
			TextTemplate: `{{.Text}}

RETRIEVED CONTEXT:
Based on the extracted features, here is relevant information:
- Technical specifications for similar products
- Historical data and trends
- Related documentation and manuals
- User reviews and feedback

QUERY: {{.Memory.Query}}`,
		},
	}

	// Step 3: Generate comprehensive response
	generateResponseStep := &flow.Step{
		Name: "generate_response",
		Executor: &anyi.LLMExecutor{
			Template: `Based on the image analysis and retrieved context, provide a comprehensive answer to: {{.Memory.Query}}

Image Analysis:
{{.Text}}

Use both the visual information and the retrieved context to provide a detailed, accurate response.`,
			SystemMessage: "You are an expert consultant who combines visual analysis with background knowledge.",
		},
	}

	// Create RAG workflow
	ragFlow, err := anyi.NewFlow("multimodal_rag", client,
		*extractFeaturesStep, *retrieveContextStep, *generateResponseStep)
	if err != nil {
		log.Fatalf("Failed to create RAG flow: %v", err)
	}

	// Create context with query and image
	type QueryData struct {
		Query string
	}

	context := anyi.NewFlowContextWithMemory(QueryData{
		Query: "What is this product and how does it compare to similar items?",
	})
	context.ImageURLs = []string{"https://example.com/product-image.jpg"}

	// Run multimodal RAG
	result, err := ragFlow.Run(context)
	if err != nil {
		log.Fatalf("Multimodal RAG failed: %v", err)
	}

	log.Printf("RAG Response: %s", result.Text)
}
```

## Best Practices

### 1. Image Quality and Format

- **Use high-quality images**: Better quality leads to more accurate analysis
- **Supported formats**: Stick to PNG, JPEG, WebP for maximum compatibility
- **File size limits**: Respect provider limits (typically 5-20MB)
- **Resolution**: Higher resolution for detailed analysis, lower for simple tasks

### 2. Prompt Engineering for Vision

```go
// Good: Specific and structured prompts
template := `Analyze this medical X-ray image and provide:
1. Anatomical structures visible
2. Any abnormalities or concerns
3. Recommended follow-up actions
4. Confidence level in assessment

Image: {{range .ImageURLs}}{{.}}{{end}}`

// Better: Include context and constraints
template := `As a radiologist, analyze this chest X-ray image:

PATIENT CONTEXT: {{.Memory.PatientInfo}}
CLINICAL QUESTION: {{.Memory.ClinicalQuestion}}

Please provide:
1. Technical quality assessment
2. Anatomical findings
3. Pathological findings (if any)
4. Clinical correlation recommendations
5. Urgency level (routine/urgent/emergent)

Use standard medical terminology and be specific about locations.

Image: {{range .ImageURLs}}{{.}}{{end}}`
```

### 3. Error Handling and Fallbacks

```go
package main

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/jieliu2000/anyi/llm/chat"
)

func analyzeImageWithFallback(client *anyi.Client, imageURL string) (string, error) {
	// Primary attempt with high detail
	messages := []chat.Message{
		{
			Role: "user",
			ContentParts: []chat.ContentPart{
				{
					Type: "text",
					Text: "Provide a detailed analysis of this image.",
				},
				{
					Type: "image_url",
					ImageURL: &chat.ImageURL{
						URL:    imageURL,
						Detail: "high",
					},
				},
			},
		},
	}

	response, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Printf("High detail analysis failed: %v", err)

		// Fallback: Try with low detail
		messages[0].ContentParts[1].ImageURL.Detail = "low"
		response, _, err = client.Chat(messages, nil)
		if err != nil {
			return "", err
		}

		log.Println("Used low detail fallback")
	}

	return response.Content, nil
}

func main() {
	config := openai.NewConfigWithModel(os.Getenv("OPENAI_API_KEY"), "gpt-4o")
	client, err := anyi.NewClient("vision", config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	result, err := analyzeImageWithFallback(client, "https://example.com/large-image.jpg")
	if err != nil {
		log.Fatalf("Image analysis failed: %v", err)
	}

	log.Printf("Analysis: %s", result)
}
```

### 4. Cost Optimization

- **Choose appropriate detail levels**: Use "low" for simple tasks, "high" for complex analysis
- **Batch processing**: Group related images in single requests when possible
- **Cache results**: Store analysis results for frequently accessed images
- **Preprocessing**: Resize or compress images when high resolution isn't needed

### 5. Privacy and Security

```go
// For sensitive images, use base64 encoding instead of URLs
func processSensitiveImage(imagePath string) error {
	// Read and encode image locally
	base64Image, err := encodeImageToBase64(imagePath)
	if err != nil {
		return err
	}

	// Process without exposing image URL
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// Use dataURL in your multimodal requests
	// This keeps the image data within the request

	return nil
}

// For compliance, implement data retention policies
func processImageWithRetention(imageURL string) {
	// Process image
	result := analyzeImage(imageURL)

	// Log processing for audit trail
	log.Printf("Processed image: %s at %s", imageURL, time.Now())

	// Implement data deletion after retention period
	// scheduleDataDeletion(result.ID, retentionPeriod)
}
```

### 6. Testing Multimodal Applications

```go
package main

import (
	"testing"

	"github.com/jieliu2000/anyi"
)

func TestImageAnalysis(t *testing.T) {
	// Use test images with known content
	testCases := []struct {
		name        string
		imageURL    string
		expectedKey string
	}{
		{
			name:        "Cat Image",
			imageURL:    "https://example.com/test-cat.jpg",
			expectedKey: "cat",
		},
		{
			name:        "Document Image",
			imageURL:    "https://example.com/test-document.png",
			expectedKey: "document",
		},
	}

	client, err := anyi.GetClient("test_vision")
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzeImage(client, tc.imageURL)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			// Check if expected content is detected
			if !strings.Contains(strings.ToLower(result), tc.expectedKey) {
				t.Errorf("Expected to find '%s' in analysis result: %s", tc.expectedKey, result)
			}
		})
	}
}
```

By following these practices and examples, you can build robust multimodal applications that effectively combine text and image processing capabilities using the Anyi framework.
