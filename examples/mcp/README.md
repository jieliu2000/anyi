# ZDNet Tech News Scraper with Playwright MCP

This example demonstrates how to use the [Microsoft Playwright MCP server](https://github.com/microsoft/playwright-mcp) with the Anyi framework to scrape the latest tech news from ZDNet and generate AI-powered summaries.

## Overview

The example consists of:

- **Example Test Format**: Uses Go's example test pattern for documentation and testing
- **Browser automation** using Playwright MCP to navigate and extract content from ZDNet
- **AI summarization** using Ollama with Qwen3 model to analyze and summarize news articles
- **Structured output** displaying both individual articles and an overall summary

## Features

- **Web Scraping**: Automated navigation to ZDNet using Playwright
- **AI Analysis**: Intelligent extraction and summarization of tech news
- **Structured Output**: Clean command-line display of results
- **Real-time**: Fetches today's latest tech news
- **Configurable**: Easy to modify for other news sources

## Prerequisites

### 1. Install Node.js and npm

The Playwright MCP server requires Node.js:

```bash
# Install Node.js (if not already installed)
# Visit https://nodejs.org/ or use your package manager
```

### 2. Install Playwright MCP Server

```bash
# Install the Playwright MCP server globally
npm install -g @playwright/mcp

# Verify installation
npx @playwright/mcp --help
```

### 3. Install Ollama and Qwen3 Model

```bash
# Install Ollama (visit https://ollama.ai/)
# Then pull the Qwen3 model
ollama pull qwen3
```

### 4. Install Go Dependencies

```bash
# From the anyi project root
go mod tidy
```

## Usage

### Running the Example

1. Navigate to the MCP example directory:

```bash
cd examples/mcp
```

2. Run the ZDNet news scraper:

```bash
go test -run ExampleZDNetNewsScraper -v
```

3. Run other example functions:

```bash
# Test NewsItem structure usage
go test -run ExampleNewsItem -v

# Test NewsCollection structure usage
go test -run ExampleNewsCollection -v

# Run all examples
go test -v
```

### Expected Output

The program will:

1. Start Playwright MCP server via npx
2. Navigate to ZDNet website
3. Extract current tech news articles
4. Generate an AI summary using Qwen3
5. Display results in a formatted output

Example output:

```
================================================================================
   ZDNET TECH NEWS SUMMARY - 2024-01-15
================================================================================

AI GENERATED SUMMARY:
----------------------------------------
Today's tech news from ZDNet highlights several significant developments in the technology sector.
The focus appears to be on artificial intelligence advancements, cybersecurity updates, and enterprise
technology solutions. Key companies mentioned include major cloud providers and software manufacturers
who are driving innovation in their respective domains.

These developments indicate a continued emphasis on AI integration across various industries,
enhanced security measures in response to evolving threats, and the ongoing digital transformation
efforts by enterprises worldwide.

LATEST ARTICLES (5 found):
----------------------------------------

[1] AI-powered development tools gain momentum in enterprise environments
    Summary: Latest survey shows increased adoption of AI coding assistants
    Link: https://www.zdnet.com/article/ai-development-tools-enterprise/
    Date: 2024-01-15

[2] Cybersecurity threats evolve with new attack vectors
    Summary: Security researchers identify emerging threats targeting cloud infrastructure
    Link: https://www.zdnet.com/article/cybersecurity-cloud-threats/
    Date: 2024-01-15

...
================================================================================
Scraped from ZDNet on 2024-01-15 14:30:22
================================================================================
```

## Configuration

### Modifying the Target Website

To scrape a different news source, update the `config.yml` file:

```yaml
flows:
  - name: zdnetNewsExtraction
    steps:
      - name: navigateToSite
        executor:
          type: mcp
          withConfig:
            server:
              name: "playwright"
              # ... server config
            toolArgs:
              url: "https://your-news-site.com/" # Change this URL
```

### Customizing AI Model

To use a different Ollama model, update the client configuration:

```yaml
clients:
  - name: ollama
    type: ollama
    config:
      model: llama2 # Change to your preferred model
```

### Browser Settings

Modify browser behavior in the MCP server configuration:

```yaml
server:
  env:
    PLAYWRIGHT_HEADLESS: "false" # Set to false to see browser window
    PLAYWRIGHT_BROWSER: "firefox" # Change browser (chromium, firefox, webkit)
```

## Troubleshooting

### Common Issues

1. **"Command npx not found"**

   - Install Node.js and npm
   - Verify with: `node --version` and `npm --version`

2. **"@playwright/mcp not found"**

   - Install the package: `npm install -g @playwright/mcp`
   - Or run without global install: `npx @playwright/mcp`

3. **"Model qwen3 not found"**

   - Pull the model: `ollama pull qwen3`
   - Check available models: `ollama list`

4. **Browser fails to start**

   - Install browser dependencies: `npx playwright install`
   - Check system requirements for Playwright

5. **Connection timeout**
   - Increase timeout in config: `timeout: 60s`
   - Check internet connection
   - Verify the target website is accessible

### Debug Mode

Enable debug logging by modifying the Go code:

```go
log.SetLevel(log.DebugLevel)
```

Or set environment variable:

```bash
export LOG_LEVEL=debug
go test -run ExampleZDNetNewsScraper -v
```

## Architecture

### Workflow Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Go Program    │────│  Anyi Framework  │────│  Playwright MCP │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌──────────────────┐             │
         └──────────────│   Ollama/Qwen3   │─────────────┘
                        └──────────────────┘
```

### Component Responsibilities

- **Go Program**: Main application logic, data structures, output formatting
- **Anyi Framework**: Workflow orchestration, configuration management
- **Playwright MCP**: Browser automation, web scraping, page interaction
- **Ollama/Qwen3**: AI-powered content analysis and summarization

## Go Example Test Format

This example uses Go's example test format (`Example*` functions) which provides several benefits:

- **Documentation**: Example functions serve as executable documentation
- **Testing**: Can be run as tests to verify functionality
- **Testable Examples**: Output can be verified using `// Output:` comments
- **Godoc Integration**: Examples appear in generated documentation

### Example Function Types

1. **ExampleZDNetNewsScraper**: Complete workflow demonstration
2. **ExampleNewsItem**: Shows how to use the NewsItem struct
3. **ExampleNewsCollection**: Demonstrates NewsCollection usage

### Running Examples

```bash
# Run a specific example
go test -run ExampleZDNetNewsScraper -v

# Run all examples with verbose output
go test -v

# Run examples and check output (if // Output: comments are present)
go test
```

## Extending the Example

### Adding More News Sources

1. Create new workflow configurations in `config.yml`
2. Implement source-specific extraction logic
3. Modify the Go code to handle multiple sources

### Enhanced Content Analysis

1. Add sentiment analysis to news articles
2. Implement topic categorization
3. Generate trend analysis across multiple days

### Data Persistence

1. Save results to JSON files
2. Store in database for historical analysis
3. Generate reports over time

## API Reference

### Key Functions

- `ExampleZDNetNewsScraper()`: Main example function demonstrating the complete workflow
- `extractZDNetNews()`: Main scraping function using Playwright MCP
- `summarizeNews()`: AI-powered summarization using Ollama
- `displayResults()`: Formatted console output
- `saveResultsToFile()`: Optional data persistence
- `ExampleNewsItem()`: Demonstrates NewsItem structure usage
- `ExampleNewsCollection()`: Demonstrates NewsCollection structure usage

### Data Structures

```go
type NewsItem struct {
    Title   string `json:"title"`
    Summary string `json:"summary"`
    Link    string `json:"link"`
    Date    string `json:"date"`
}

type NewsCollection struct {
    Articles []NewsItem `json:"articles"`
    Source   string     `json:"source"`
    Date     string     `json:"date"`
}
```

## License

This example is part of the Anyi project. Please refer to the main project license.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## References

- [Microsoft Playwright MCP](https://github.com/microsoft/playwright-mcp)
- [Ollama Documentation](https://ollama.ai/)
- [Anyi Framework](https://github.com/jieliu2000/anyi)
- [Playwright Documentation](https://playwright.dev/)
