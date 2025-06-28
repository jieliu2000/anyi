# MCP Executor 传输方式详解

## 概述

MCP (Model Context Protocol) Executor 现在完全支持三种传输方式：HTTP、Server-Sent Events (SSE) 和 STDIO。每种传输方式都有其特定的使用场景和优势。

## 支持的传输方式

### 1. HTTP 传输 (`transport: "http"`)

**适用场景**：

- 与基于 REST API 的 MCP 服务器通信
- 需要标准 HTTP 认证的场景
- 简单的请求-响应模式

**配置示例**：

```yaml
executor: "mcp"
config:
  transport: "http"
  serverEndpoint: "http://localhost:8080/mcp"
  apiKey: "your-api-key" # 可选
  operation: "tool_call"
  toolName: "example_tool"
  toolArgs:
    param1: "value1"
  timeout: "30s"
  retryAttempts: 3
```

**特点**：

- ✅ 实现完整，生产就绪
- ✅ 支持 Bearer Token 认证
- ✅ 支持超时和重试机制
- ✅ 标准 HTTP 状态码处理

### 2. Server-Sent Events (SSE) 传输 (`transport: "sse"`)

**适用场景**：

- 需要实时数据流的场景
- 长时间运行的任务监控
- 需要服务器主动推送数据

**配置示例**：

```yaml
executor: "mcp"
config:
  transport: "sse"
  serverEndpoint: "http://localhost:8080/mcp/events"
  apiKey: "your-api-key" # 可选
  operation: "tool_call"
  toolName: "streaming_tool"
  toolArgs:
    stream: true
  timeout: "60s"
```

**特点**：

- ✅ 实现完整，生产就绪
- ✅ 支持实时事件流
- ✅ 自动重连机制
- ✅ 支持事件过滤和匹配
- ✅ 优雅的连接关闭处理

**工作原理**：

1. 建立 SSE 连接到指定端点
2. 通过单独的 HTTP POST 端点发送请求
3. 通过 SSE 流接收响应
4. 根据请求 ID 匹配响应

### 3. STDIO 传输 (`transport: "stdio"`)

**适用场景**：

- 与本地 MCP 服务器进程通信
- 需要启动外部程序作为 MCP 服务器
- 高性能的本地通信

**配置示例**：

```yaml
executor: "mcp"
config:
  transport: "stdio"
  serverCommand: "python"
  serverArgs: ["-m", "my_mcp_server"]
  operation: "tool_call"
  toolName: "local_tool"
  toolArgs:
    input: "data"
  timeout: "30s"
```

**特点**：

- ✅ 实现完整，生产就绪
- ✅ 自动进程管理（启动/停止）
- ✅ 完整的 MCP 协议初始化
- ✅ 异步 I/O 处理
- ✅ 优雅的进程终止
- ✅ 错误输出监控

**工作原理**：

1. 启动指定的命令作为子进程
2. 建立 stdin/stdout/stderr 管道
3. 发送 MCP 初始化请求
4. 通过 JSON-RPC over STDIO 进行通信
5. 监控进程状态和错误输出

## 操作类型

所有传输方式都支持以下 MCP 操作：

### 1. 工具调用 (`operation: "tool_call"`)

```yaml
operation: "tool_call"
toolName: "calculator"
toolArgs:
  expression: "2 + 2"
toolArgVars: ["userInput"] # 从流程变量获取参数
```

### 2. 资源读取 (`operation: "resource_read"`)

```yaml
operation: "resource_read"
resourceUri: "file://path/to/resource.txt"
```

### 3. 提示获取 (`operation: "prompt_get"`)

```yaml
operation: "prompt_get"
promptName: "code_review"
promptArgs:
  language: "python"
  style: "detailed"
```

### 4. 列出工具 (`operation: "list_tools"`)

```yaml
operation: "list_tools"
```

### 5. 列出资源 (`operation: "list_resources"`)

```yaml
operation: "list_resources"
```

## 配置参数详解

### 通用参数

| 参数              | 类型     | 必需 | 描述                             |
| ----------------- | -------- | ---- | -------------------------------- |
| `transport`       | string   | ✅   | 传输方式：`http`、`sse`、`stdio` |
| `operation`       | string   | ✅   | MCP 操作类型                     |
| `timeout`         | duration | ❌   | 超时时间（默认：30s）            |
| `retryAttempts`   | int      | ❌   | 重试次数（默认：3）              |
| `outputToContext` | bool     | ❌   | 是否将结果输出到流程上下文       |
| `resultVarName`   | string   | ❌   | 结果变量名（默认：mcpResult）    |

### HTTP/SSE 特定参数

| 参数             | 类型   | 必需 | 描述                     |
| ---------------- | ------ | ---- | ------------------------ |
| `serverEndpoint` | string | ✅   | 服务器端点 URL           |
| `apiKey`         | string | ❌   | API 密钥（Bearer Token） |

### STDIO 特定参数

| 参数            | 类型     | 必需 | 描述       |
| --------------- | -------- | ---- | ---------- |
| `serverCommand` | string   | ✅   | 服务器命令 |
| `serverArgs`    | []string | ❌   | 命令参数   |

### 操作特定参数

#### 工具调用

| 参数          | 类型     | 必需 | 描述                   |
| ------------- | -------- | ---- | ---------------------- |
| `toolName`    | string   | ✅   | 工具名称               |
| `toolArgs`    | map      | ❌   | 静态工具参数           |
| `toolArgVars` | []string | ❌   | 从流程变量获取的参数名 |

#### 资源读取

| 参数          | 类型   | 必需 | 描述     |
| ------------- | ------ | ---- | -------- |
| `resourceUri` | string | ✅   | 资源 URI |

#### 提示获取

| 参数         | 类型   | 必需 | 描述     |
| ------------ | ------ | ---- | -------- |
| `promptName` | string | ✅   | 提示名称 |
| `promptArgs` | map    | ❌   | 提示参数 |

## 错误处理

### 自动重试

- 所有传输方式都支持自动重试
- 可通过 `retryAttempts` 配置重试次数
- 重试间隔递增（1s、2s、3s...）

### 错误类型

1. **连接错误**：网络连接失败、进程启动失败
2. **超时错误**：请求超时
3. **协议错误**：MCP 协议错误、JSON-RPC 错误
4. **业务错误**：MCP 服务器返回的业务错误

### 错误处理最佳实践

```yaml
config:
  timeout: "60s" # 适当的超时时间
  retryAttempts: 5 # 不稳定服务增加重试次数
  outputToContext: true # 便于调试
```

## 性能考虑

### HTTP 传输

- **优点**：简单、标准、易调试
- **缺点**：每次请求建立连接的开销
- **适用**：低频率调用、简单场景

### SSE 传输

- **优点**：实时性好、支持流式数据
- **缺点**：连接管理复杂、资源占用较高
- **适用**：需要实时更新、长时间任务

### STDIO 传输

- **优点**：性能最佳、延迟最低
- **缺点**：仅限本地、进程管理复杂
- **适用**：高频率调用、本地服务

## 安全考虑

### HTTP/SSE 传输

- 使用 HTTPS 加密传输
- 妥善保管 API 密钥
- 验证服务器证书

### STDIO 传输

- 验证服务器命令路径
- 限制命令参数
- 监控子进程行为

## 故障排查

### 常见问题

1. **连接失败**

   ```
   错误：failed to connect to SSE endpoint
   解决：检查端点 URL、网络连接、服务器状态
   ```

2. **认证失败**

   ```
   错误：HTTP error 401: Unauthorized
   解决：检查 API 密钥是否正确、是否有权限
   ```

3. **进程启动失败**

   ```
   错误：failed to start MCP server process
   解决：检查命令路径、参数、权限
   ```

4. **超时错误**
   ```
   错误：timeout waiting for response
   解决：增加超时时间、检查服务器性能
   ```

### 调试技巧

1. **启用详细日志**：

   ```yaml
   config:
     outputToContext: true # 输出完整响应
   ```

2. **检查变量**：

   ```yaml
   resultVarName: "debugResult" # 使用明确的变量名
   ```

3. **分步测试**：
   - 先测试 `list_tools` 操作
   - 再测试具体的工具调用

## 示例集合

完整的配置示例请参考：`examples/mcp_executor_complete_example.yaml`

该文件包含：

- 各种传输方式的基础用法
- 复杂工作流示例
- 错误处理示例
- 动态参数示例
