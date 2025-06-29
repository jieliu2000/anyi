@echo off
REM ZDNet Tech News Scraper Runner for Windows
REM This script sets up and runs the ZDNet news scraper example

echo Starting ZDNet Tech News Scraper with Playwright MCP...

REM Check prerequisites
echo Checking prerequisites...

REM Check if Node.js is installed
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js is not installed. Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

REM Check if npm is installed
npm --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: npm is not installed. Please install npm.
    pause
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed. Please install Go from https://golang.org/
    pause
    exit /b 1
)

REM Check if Ollama is installed
ollama --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Ollama is not installed. Please install Ollama from https://ollama.ai/
    pause
    exit /b 1
)

echo All prerequisites are installed

REM Check if Playwright MCP is available
echo Checking Playwright MCP availability...
npx @playwright/mcp --help >nul 2>&1
if %errorlevel% neq 0 (
    echo Installing Playwright MCP...
    npm install -g @playwright/mcp
) else (
    echo Playwright MCP is available
)

REM Check if qwen3 model is available
echo Checking Ollama qwen3 model...
ollama list | findstr "qwen3" >nul 2>&1
if %errorlevel% neq 0 (
    echo Pulling qwen3 model (this may take a few minutes)...
    ollama pull qwen3
) else (
    echo qwen3 model is available
)

REM Start Ollama service in background
echo Starting Ollama service...
start /min ollama serve

REM Wait a moment for Ollama to start
timeout /t 3 /nobreak >nul

REM Install Playwright browsers if needed
echo Installing Playwright browsers...
npx playwright install chromium

echo Running ZDNet news scraper...
echo ==================================================

REM Run the Go example test
go test -run ExampleZDNetNewsScraper -v

echo ==================================================
echo ZDNet news scraper completed successfully!

echo Done! Check the output above for your tech news summary.
pause 