# MCP 集成

本文档描述了如何在 Anyi 框架中使用 MCP (Multi-agent Collaboration Protocol) 执行器，通过该执行器，您可以在 Anyi 工作流中无缝集成 MCP 代理。

## 概述

MCP 执行器允许您在 Anyi 工作流中调用 MCP 服务器上的代理，同时支持：

- 多个数据源的双向交互
- 参数模板化配置
- 灵活的数据传输方式
- 完全兼容 Anyi 的工作流系统

## 配置项

下面是 MCP 执行器的配置项说明：

| 配置项             | 类型   | 必填 | 描述                                                                 |
| ------------------ | ------ | ---- | -------------------------------------------------------------------- |
| `endpoint`         | 字符串 | 是   | MCP 服务器地址，不包含协议前缀                                       |
| `api_key`          | 字符串 | 是   | MCP 服务的 API 密钥，可使用环境变量 (如 `$MCP_API_KEY`)              |
| `agent_id`         | 字符串 | 是   | 要调用的 MCP 代理 ID                                                 |
| `timeout`          | 整数   | 否   | 请求超时时间（秒），默认 60                                          |
| `use_ssl`          | 布尔值 | 否   | 是否使用 HTTPS 连接，默认 true                                       |
| `params_template`  | 字符串 | 否   | 代理参数的 JSON 模板                                                 |
| `data_sources`     | 数组   | 否   | 数据源配置数组                                                       |
| `include_metadata` | 布尔值 | 否   | 是否在请求中包含数据源元数据，默认 false                             |
| `data_transfer`    | 字符串 | 否   | 数据传输方式，可选值：direct、reference、object_storage，默认 direct |

### 数据源配置

每个数据源配置包含以下字段：

| 配置项           | 类型   | 必填 | 描述                        |
| ---------------- | ------ | ---- | --------------------------- |
| `name`           | 字符串 | 是   | 数据源名称                  |
| `type`           | 字符串 | 是   | 数据源类型，目前支持 "file" |
| `query_template` | 字符串 | 是   | 查询模板，支持变量替换      |
| `config`         | 对象   | 否   | 数据源特定配置              |

对于文件数据源，config 可包含：

```yaml
config:
  base_dir: "./data" # 基础目录
```

## 使用示例

### 基本配置

以下是 MCP 执行器的基本配置示例：

```yaml
Flows:
  - name: "document_analysis"
    clientName: "openai"
    steps:
      - name: "analyze_with_mcp"
        clientName: "openai"
        executor:
          type: "mcp"
          withconfig:
            endpoint: "mcp.example.com"
            api_key: "${MCP_API_KEY}"
            agent_id: "research_assistant"
            timeout: 120
            data_sources:
              - name: "local_files"
                type: "file"
                query_template: "{{.document_path}}"
                config:
                  base_dir: "./documents"
```

### 参数模板

使用参数模板可以更灵活地控制发送给 MCP 代理的参数：

```yaml
params_template: |
  {
    "query": "{{.query}}",
    "max_tokens": 1000,
    "options": {
      "format": "markdown",
      "depth": "detailed"
    }
  }
```

### 写入数据

MCP 代理可以请求写入数据到已注册的可写数据源。代理响应格式如下：

```json
{
  "result": "分析完成...",
  "write_data": [
    {
      "data_source": "local_files",
      "path": "output/report.md",
      "data": "# 分析报告\n\n这是分析结果..."
    }
  ]
}
```

## 编程方式使用

以下是通过代码创建和使用 MCP 执行器的示例：

```go
// 创建 MCP 执行器
mcpExecutor := &anyi.MCPExecutor{
    Endpoint:       "mcp.example.com",
    APIKey:         "$MCP_API_KEY",
    AgentID:        "research_assistant",
    Timeout:        120,
    UseSSL:         true,
    IncludeMetadata: true,
    ParamsTemplate: `{
        "query": "{{.query}}",
        "max_tokens": 1000
    }`,
    DataSources: []anyi.DataSourceConfig{
        {
            Name:          "local_files",
            Type:          "file",
            QueryTemplate: "{{.file_path}}",
            Config: map[string]interface{}{
                "base_dir": "./data",
            },
        },
    },
}

// 初始化执行器
mcpExecutor.Init()

// 创建步骤和工作流
step := flow.NewStep(mcpExecutor, nil, client)
step.Name = "mcp_research"
mcpFlow, _ := flow.NewFlow(client, "mcp_research_flow", *step)

// 运行工作流
context := flow.FlowContext{
    Variables: map[string]interface{}{
        "query":     "分析这篇文章的主要观点",
        "file_path": "documents/article.txt",
    },
}
result, _ := mcpFlow.Run(context)
```

## 最佳实践

1. **环境变量**：使用环境变量存储敏感信息，如 API 密钥
2. **数据安全**：确保数据源配置正确设置了安全限制，特别是文件数据源的基础目录
3. **错误处理**：工作流步骤应包含适当的验证器，以处理 MCP 代理可能返回的错误
4. **并发控制**：对于重负载场景，考虑设置适当的超时和重试策略

## 故障排除

常见问题：

1. **连接超时** - 检查 MCP 服务器地址和网络连接
2. **授权失败** - 确认 API 密钥正确且未过期
3. **数据源错误** - 验证数据源名称和查询模板
4. **写入权限问题** - 确保数据源具有写入权限

## 扩展

MCP 执行器的设计允许未来扩展更多功能：

1. 支持更多数据源类型（数据库、API 等）
2. 增强数据传输安全性
3. 支持异步调用和长时间运行的任务
4. 集成 MCP 代理的流式响应
