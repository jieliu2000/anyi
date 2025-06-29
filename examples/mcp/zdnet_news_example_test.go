package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

// NewsItem represents a single news article
type NewsItem struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Link    string `json:"link"`
	Date    string `json:"date"`
}

// NewsCollection represents the collection of news articles
type NewsCollection struct {
	Articles []NewsItem `json:"articles"`
	Source   string     `json:"source"`
	Date     string     `json:"date"`
}

// ExampleZDNetNewsScraper demonstrates how to use Playwright MCP with Anyi framework
// to scrape the latest tech news from ZDNet and generate AI-powered summaries.
//
// This example shows:
//   - Browser automation using Playwright MCP to navigate and extract content from ZDNet
//   - AI summarization using Ollama with Qwen3 model to analyze and summarize news articles
//   - Structured output displaying both individual articles and an overall summary
//
// Prerequisites:
//   - Node.js and npm installed
//   - Playwright MCP server: npm install -g @playwright/mcp
//   - Ollama with qwen3 model: ollama pull qwen3
//   - Configuration file: config.yml in the same directory
//
// Usage:
//
//	go test -run ExampleZDNetNewsScraper
func Example_openAI() {
	log.Printf("Starting ZDNet Tech News Scraper using Playwright MCP...")

	// Initialize Anyi with configuration
	err := anyi.ConfigFromFile("config.yml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Run the news extraction workflow
	newsData, err := extractZDNetNews()
	if err != nil {
		log.Fatalf("Failed to extract news: %v", err)
	}

	// Summarize the news using AI
	summary, err := summarizeNews(newsData)
	if err != nil {
		log.Fatalf("Failed to summarize news: %v", err)
	}

	// Display the results
	displayResults(newsData, summary)
	// Output: hello
	// Output example (actual output will vary based on current news):
	// Starting ZDNet Tech News Scraper using Playwright MCP...
	// Navigating to ZDNet to extract latest tech news...
	// Successfully extracted X news articles from ZDNet
	// Generating AI summary of the news articles...
	// AI summary generated successfully
	//
	// ================================================================================
	//    ZDNET TECH NEWS SUMMARY - 2024-01-15
	// ================================================================================
	//
	// AI GENERATED SUMMARY:
	// ----------------------------------------
	// Today's tech news highlights...
	//
	// LATEST ARTICLES (X found):
	// ----------------------------------------
	//
	// [1] Article Title Here
	//     Summary: Article summary here
	//     Link: https://www.zdnet.com/article/...
	//     Date: 2024-01-15
	// ================================================================================
}

// extractZDNetNews uses Playwright MCP to navigate to ZDNet and extract news articles
func extractZDNetNews() (*NewsCollection, error) {
	log.Printf("Navigating to ZDNet to extract latest tech news...")

	// Get the news extraction workflow
	extractFlow, err := anyi.GetFlow("zdnetNewsExtraction")
	if err != nil {
		return nil, fmt.Errorf("failed to get news extraction flow: %w", err)
	}

	// Create context with today's date
	today := time.Now().Format("2006-01-02")
	context := &flow.FlowContext{
		Variables: map[string]interface{}{
			"targetUrl": "https://www.zdnet.com/",
			"date":      today,
		},
	}

	// Run the extraction workflow
	result, err := extractFlow.Run(*context)
	if err != nil {
		return nil, fmt.Errorf("failed to run extraction workflow: %w", err)
	}

	// Parse the results
	var newsCollection NewsCollection
	err = json.Unmarshal([]byte(result.Text), &newsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to parse news data: %w", err)
	}

	log.Printf("Successfully extracted %d news articles from ZDNet", len(newsCollection.Articles))
	return &newsCollection, nil
}

// summarizeNews uses ollama/qwen3 to create a summary of the news articles
func summarizeNews(newsData *NewsCollection) (string, error) {
	log.Printf("Generating AI summary of the news articles...")

	// Get the summarization workflow
	summaryFlow, err := anyi.GetFlow("newsSummarization")
	if err != nil {
		return "", fmt.Errorf("failed to get summarization flow: %w", err)
	}

	// Prepare news data for summarization
	newsJSON, err := json.Marshal(newsData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal news data: %w", err)
	}

	// Create context with news data
	context := &flow.FlowContext{
		Text: string(newsJSON),
		Variables: map[string]interface{}{
			"source": "ZDNet",
			"date":   newsData.Date,
		},
	}

	// Run the summarization workflow
	result, err := summaryFlow.Run(*context)
	if err != nil {
		return "", fmt.Errorf("failed to run summarization workflow: %w", err)
	}

	log.Printf("AI summary generated successfully")
	return result.Text, nil
}

// displayResults outputs the news articles and summary to the command line
func displayResults(newsData *NewsCollection, summary string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("   ZDNET TECH NEWS SUMMARY - %s\n", newsData.Date)
	fmt.Println(strings.Repeat("=", 80))

	// Display AI Summary
	fmt.Println("\nAI GENERATED SUMMARY:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("%s\n", summary)

	// Display individual articles
	fmt.Printf("\nLATEST ARTICLES (%d found):\n", len(newsData.Articles))
	fmt.Println(strings.Repeat("-", 40))

	for i, article := range newsData.Articles {
		fmt.Printf("\n[%d] %s\n", i+1, article.Title)
		if article.Summary != "" {
			fmt.Printf("    Summary: %s\n", article.Summary)
		}
		if article.Link != "" {
			fmt.Printf("    Link: %s\n", article.Link)
		}
		if article.Date != "" {
			fmt.Printf("    Date: %s\n", article.Date)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Scraped from %s on %s\n", newsData.Source, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}

// saveResultsToFile demonstrates how to save results to a file
// This function shows how to structure the output data for persistence
func saveResultsToFile(newsData *NewsCollection, summary string, filename string) error {
	output := map[string]interface{}{
		"summary":     summary,
		"articles":    newsData.Articles,
		"source":      newsData.Source,
		"scraped_at":  time.Now().Format("2006-01-02 15:04:05"),
		"total_count": len(newsData.Articles),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// This could be implemented using a file write executor
	log.Printf("Results would be saved to %s", filename)
	_ = data // Use data to write to file when implementing actual file operations
	return nil
}

// ExampleNewsItem demonstrates the NewsItem structure usage
func ExampleNewsItem() {
	item := NewsItem{
		Title:   "AI-powered development tools gain momentum",
		Summary: "Latest survey shows increased adoption of AI coding assistants",
		Link:    "https://www.zdnet.com/article/ai-development-tools/",
		Date:    "2024-01-15",
	}

	// Convert to JSON for API usage
	data, _ := json.MarshalIndent(item, "", "  ")
	fmt.Printf("NewsItem JSON:\n%s\n", data)

	// Output:
	// NewsItem JSON:
	// {
	//   "title": "AI-powered development tools gain momentum",
	//   "summary": "Latest survey shows increased adoption of AI coding assistants",
	//   "link": "https://www.zdnet.com/article/ai-development-tools/",
	//   "date": "2024-01-15"
	// }
}

// ExampleNewsCollection demonstrates the NewsCollection structure usage
func ExampleNewsCollection() {
	collection := NewsCollection{
		Articles: []NewsItem{
			{
				Title:   "Cloud computing trends in 2024",
				Summary: "Analysis of emerging cloud technologies",
				Link:    "https://www.zdnet.com/article/cloud-trends/",
				Date:    "2024-01-15",
			},
			{
				Title:   "Cybersecurity best practices",
				Summary: "Essential security measures for enterprises",
				Link:    "https://www.zdnet.com/article/cybersecurity/",
				Date:    "2024-01-15",
			},
		},
		Source: "ZDNet",
		Date:   "2024-01-15",
	}

	fmt.Printf("Found %d articles from %s\n", len(collection.Articles), collection.Source)
	for i, article := range collection.Articles {
		fmt.Printf("[%d] %s\n", i+1, article.Title)
	}

	// Output:
	// Found 2 articles from ZDNet
	// [1] Cloud computing trends in 2024
	// [2] Cybersecurity best practices
}
