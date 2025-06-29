#!/bin/bash

# ZDNet Tech News Scraper Runner
# This script sets up and runs the ZDNet news scraper example

set -e  # Exit on any error

echo "Starting ZDNet Tech News Scraper with Playwright MCP..."

# Check prerequisites
echo "Checking prerequisites..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "ERROR: Node.js is not installed. Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "ERROR: npm is not installed. Please install npm."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed. Please install Go from https://golang.org/"
    exit 1
fi

# Check if Ollama is installed and running
if ! command -v ollama &> /dev/null; then
    echo "ERROR: Ollama is not installed. Please install Ollama from https://ollama.ai/"
    exit 1
fi

echo "All prerequisites are installed"

# Check if Playwright MCP is available
echo "Checking Playwright MCP availability..."
if ! npx @playwright/mcp --help &> /dev/null; then
    echo "Installing Playwright MCP..."
    npm install -g @playwright/mcp
else
    echo "Playwright MCP is available"
fi

# Check if qwen3 model is available
echo "Checking Ollama qwen3 model..."
if ! ollama list | grep -q "qwen3"; then
    echo "Pulling qwen3 model (this may take a few minutes)..."
    ollama pull qwen3
else
    echo "qwen3 model is available"
fi

# Ensure Ollama is running
echo "Starting Ollama service..."
ollama serve &> /dev/null &
OLLAMA_PID=$!
sleep 3  # Give Ollama time to start

# Install Playwright browsers if needed
echo "Installing Playwright browsers..."
npx playwright install chromium

echo "Running ZDNet news scraper..."
echo "=================================================="

# Run the Go example test
go test -run ExampleZDNetNewsScraper -v

echo "=================================================="
echo "ZDNet news scraper completed successfully!"

# Cleanup: Stop Ollama if we started it
if [ ! -z "$OLLAMA_PID" ]; then
    kill $OLLAMA_PID &> /dev/null || true
fi

echo "Done! Check the output above for your tech news summary." 