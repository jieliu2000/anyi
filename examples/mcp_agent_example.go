package examples

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/datasource"
	"github.com/jieliu2000/anyi/llm/openai"
)

// MCPAgentExample 演示如何使用MCP代理执行器
func MCPAgentExample() {
	// 初始化框架
	anyi.Init()

	// 注册数据源
	fileDs := datasource.NewFileDataSource("local_files", "./data")
	fileDs.Init()
	datasource.Register("local_files", fileDs)

	// 创建OpenAI客户端
	config := openai.DefaultConfig("")
	config.APIKey = "your-openai-api-key" // 请替换为你的API密钥
	client, err := openai.NewClient(config)
	if err != nil {
		log.Fatalf("创建OpenAI客户端失败: %v", err)
	}

	// 注册客户端
	anyi.RegisterClient("openai", client)

	// 创建MCP执行器
	mcpExecutor := &anyi.MCPExecutor{
		Endpoint:        "mcp.example.com",    // 替换为MCP服务端点
		APIKey:          "$MCP_API_KEY",       // 从环境变量获取API密钥
		AgentID:         "research_assistant", // 要使用的代理ID
		Timeout:         120,
		UseSSL:          true,
		IncludeMetadata: true,
		DataTransfer:    anyi.DataTransferDirect,
		ParamsTemplate: `{
			"query": "{{.query}}",
			"max_tokens": 1000,
			"options": {
				"format": "markdown"
			}
		}`,
		DataSources: []anyi.DataSourceConfig{
			{
				Name:          "local_files",
				Type:          "file",
				QueryTemplate: "{{.file_path}}",
			},
		},
	}

	// 初始化执行器
	if err := mcpExecutor.Init(); err != nil {
		log.Fatalf("初始化MCP执行器失败: %v", err)
	}

	// 创建步骤
	step := flow.NewStep(mcpExecutor, nil, client)
	step.Name = "mcp_research"

	// 创建工作流
	mcpFlow, err := flow.NewFlow(client, "mcp_research_flow", *step)
	if err != nil {
		log.Fatalf("创建工作流失败: %v", err)
	}

	// 注册工作流
	anyi.RegisterFlow("mcp_research_flow", mcpFlow)

	// 运行工作流
	context := flow.FlowContext{
		Variables: map[string]interface{}{
			"query":     "分析这份文档中的主要观点",
			"file_path": "documents/article.txt",
		},
	}

	result, err := mcpFlow.Run(context)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}

	fmt.Println("MCP代理结果:")
	fmt.Println(result.Text)
}

// MCPAgentWithWriteExample 演示如何使用MCP代理执行器并处理写入
func MCPAgentWithWriteExample() {
	// 初始化框架
	anyi.Init()

	// 注册数据源
	fileDs := datasource.NewFileDataSource("local_files", "./data")
	fileDs.Init()
	datasource.Register("local_files", fileDs)

	// 创建OpenAI客户端
	config := openai.DefaultConfig("")
	config.APIKey = "your-openai-api-key" // 请替换为你的API密钥
	client, err := openai.NewClient(config)
	if err != nil {
		log.Fatalf("创建OpenAI客户端失败: %v", err)
	}

	// 注册客户端
	anyi.RegisterClient("openai", client)

	// 创建MCP执行器
	mcpExecutor := &anyi.MCPExecutor{
		Endpoint:        "mcp.example.com",
		APIKey:          "$MCP_API_KEY",
		AgentID:         "summary_agent",
		Timeout:         120,
		UseSSL:          true,
		IncludeMetadata: true,
		DataTransfer:    anyi.DataTransferDirect,
		ParamsTemplate: `{
			"query": "{{.query}}",
			"max_tokens": 1000,
			"options": {
				"format": "markdown",
				"generate_files": true
			}
		}`,
		DataSources: []anyi.DataSourceConfig{
			{
				Name:          "local_files",
				Type:          "file",
				QueryTemplate: "{{.file_path}}",
			},
		},
	}

	// 初始化执行器
	if err := mcpExecutor.Init(); err != nil {
		log.Fatalf("初始化MCP执行器失败: %v", err)
	}

	// 创建步骤
	step := flow.NewStep(mcpExecutor, nil, client)
	step.Name = "mcp_summary"

	// 创建工作流
	mcpFlow, err := flow.NewFlow(client, "mcp_summary_flow", *step)
	if err != nil {
		log.Fatalf("创建工作流失败: %v", err)
	}

	// 注册工作流
	anyi.RegisterFlow("mcp_summary_flow", mcpFlow)

	// 特别的是，这里我们将流程变量中添加需要写入的数据
	context := flow.FlowContext{
		Variables: map[string]interface{}{
			"query":     "总结这篇文章并生成一个摘要文件",
			"file_path": "documents/article.txt",
		},
	}

	// MCP代理可以返回write_data指令，例如:
	// {
	//   "result": "文章总结完成...",
	//   "write_data": [
	//     {
	//       "data_source": "local_files",
	//       "path": "documents/summary.md",
	//       "data": "# 文章摘要\n\n这是对文章的总结..."
	//     }
	//   ]
	// }

	result, err := mcpFlow.Run(context)
	if err != nil {
		log.Fatalf("运行工作流失败: %v", err)
	}

	fmt.Println("MCP代理结果:")
	fmt.Println(result.Text)
}
