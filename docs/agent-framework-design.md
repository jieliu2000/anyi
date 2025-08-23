# Anyi Agent Framework 设计方案

## 概述

本文档提出了为 Anyi 框架添加智能代理支持的详细设计方案。该方案将在保持 Anyi 现有简洁性和模块化设计的基础上，引入智能代理（Agent）概念，使 LLM 能够根据任务目标自主规划和执行工作流。与 CrewAI 不同，我们的设计更加简洁：**Agent 就是具有智能规划能力的工作流集合，而具体的执行单元仍然是 Anyi 现有的 Flow**。

## 项目集成要求

基于用户需求，Agent 框架必须无缝集成到现有的 Anyi 配置系统中：

### 核心使用场景

1. **配置加载**: 通过 `anyi.ConfigFromFile()` 一次性配置所有组件（包括 Agent）
2. **Agent 获取**: 通过 `anyi.GetAgent("agentName")` 获取已配置的 Agent
3. **任务执行**: 通过 `agent.Execute("Objective")` 执行任务

### 集成设计原则

- **统一配置**: Agent 配置与现有 Clients、Flows 配置使用相同的配置文件和加载机制
- **全局注册**: Agent 注册到 GlobalRegistry，与其他组件一致
- **简化接口**: Agent.Execute() 只需一个字符串参数，无需手动构建复杂对象
- **向后兼容**: 不影响现有 Flow 和 Client 的使用方式

## 当前架构分析

### 现有组件

- **GlobalRegistry**: 全局注册表，管理 Clients、Flows、Validators、Executors、Formatters
- **Flow**: 工作流，包含多个顺序执行的 Step（这是我们的核心执行单元）
- **Step**: 工作流步骤，包含 Executor 和 Validator
- **LLMExecutor**: 核心执行器，处理 LLM 交互
- **FlowContext**: 工作流上下文，传递数据和状态

### 架构特点

- 配置驱动开发（支持 YAML/JSON/TOML）
- 统一的 LLM 接口
- 强类型的 Go 设计
- 模块化组件管理

## Agent Framework 设计

### 1. 配置集成设计

#### 1.1 扩展 AnyiConfig 结构

```go
// AnyiConfig 扩展以支持 Agent 配置
type AnyiConfig struct {
    Clients    []llm.ClientConfig  `mapstructure:"clients"`
    Flows      []FlowConfig        `mapstructure:"flows"`
    Formatters []FormatterConfig   `mapstructure:"formatters"`
    Agents     []AgentConfig       `mapstructure:"agents"`  // 新增 Agent 配置
}

// AgentConfig 定义 Agent 的配置结构
type AgentConfig struct {
    Name        string                 `mapstructure:"name"`
    Description string                 `mapstructure:"description"`
    Flows       []string               `mapstructure:"flows"`     // 可用的 Flow 名称列表
    ClientName  string                 `mapstructure:"clientName"` // LLM 客户端名称
    Config      map[string]interface{} `mapstructure:"config"`    // Agent 特定配置
}
```

#### 1.2 配置文件示例

```yaml
# anyi_config.yaml
clients:
  - name: openai-gpt4
    type: openai
    apiKey: "your-api-key"
    default: true

flows:
  - name: web_search
    clientName: openai-gpt4
    steps:
      - name: search
        executor:
          type: llm
          withconfig:
            prompt: "Search for: {{.input}}"

  - name: document_analysis
    clientName: openai-gpt4
    steps:
      - name: analyze
        executor:
          type: llm
          withconfig:
            prompt: "Analyze this document: {{.input}}"

agents: # 新增 Agent 配置部分
  - name: research_assistant
    description: "An AI assistant specialized in research and analysis"
    flows:
      - web_search
      - document_analysis
    clientName: openai-gpt4
    config:
      max_iterations: 5
      timeout: 300
```

### 2. 核心概念重新定义

#### 2.1 Agent 定义（简化版）

```go
// Agent 是一个智能代理，能够根据目标自主规划和执行多个工作流
type Agent struct {
    Name        string            `mapstructure:"name"`
    Description string            `mapstructure:"description"`

    // 可用的工作流列表 - Agent 的核心能力就是这些 Flow
    Flows       []string          `mapstructure:"flows"`

    // Agent 的规划客户端（用于任务规划）
    ClientName  string            `mapstructure:"clientName"`

    // Agent 的工作记忆
    Memory      AgentMemory       `mapstructure:"-"`

    // Agent 配置参数
    Config      map[string]any    `mapstructure:"config"`
}
```

#### 2.2 简化的执行接口

Agent 的核心执行方法应该极其简单，只需要一个目标描述字符串：

```go
// Agent 的核心执行方法 - 极简接口
func (agent *Agent) Execute(objective string) (*TaskResult, error)

// 内部任务表示（用户不需要直接创建）
type agentTask struct {
    id          string
    objective   string
    createdAt   time.Time
}
```

#### 2.3 执行计划

```go
// ExecutionPlan 是 Agent 根据任务生成的执行计划
// 计划中的每一步都是一个 Flow 的调用
type ExecutionPlan struct {
    Objective   string            `mapstructure:"objective"`     // 任务目标
    Steps       []ExecutionStep   `mapstructure:"steps"`
    Description string            `mapstructure:"description"`
}

type ExecutionStep struct {
    FlowName    string                 `mapstructure:"flowName"`       // 要执行的 Flow 名称
    Input       string                 `mapstructure:"input"`          // Flow 的输入
    Variables   map[string]interface{} `mapstructure:"variables"`      // Flow 的变量
    Description string                 `mapstructure:"description"`    // 这一步的说明
    Order       int                    `mapstructure:"order"`          // 执行顺序
}
```

### 4. 主配置系统集成

#### 4.1 扩展配置加载函数

```go
// config.go 中的配置加载函数扩展
func Config(config *AnyiConfig) error {
    Init()

    log.Debug("Config Anyi with: ", config)

    // 现有的客户端初始化
    for _, clientConfig := range config.Clients {
        if clientConfig.Name != "" {
            _, err := NewClientFromConfig(&clientConfig)
            if err != nil {
                return err
            }
        }
    }

    // 现有的工作流初始化
    for _, flowConfig := range config.Flows {
        _, err := NewFlowFromConfig(&flowConfig)
        if err != nil {
            return err
        }
    }

    // 新增：Agent 初始化
    for _, agentConfig := range config.Agents {
        err := NewAgentFromConfig(&agentConfig)
        if err != nil {
            return err
        }
    }

    log.Debug("Config loaded successfully")
    return nil
}

// 新增 Agent 配置加载函数
func NewAgentFromConfig(agentConfig *AgentConfig) error {
    if agentConfig == nil {
        return errors.New("agent config is nil")
    }

    // 验证配置
    if err := ValidateAgentConfig(agentConfig); err != nil {
        return fmt.Errorf("invalid agent config: %w", err)
    }

    // 创建 Agent 实例
    agent := &Agent{
        Name:        agentConfig.Name,
        Description: agentConfig.Description,
        Flows:       agentConfig.Flows,
        ClientName:  agentConfig.ClientName,
        Memory:      NewSimpleMemory(),
        Config:      agentConfig.Config,
    }

    // 注册到全局注册表
    return RegisterAgent(agentConfig.Name, agent)
}
```

#### 4.2 新增全局 Agent 访问函数

```go
// anyi.go 中新增的便利函数

// GetAgent 从全局注册表获取 Agent
func GetAgent(name string) (*Agent, error) {
    agentInterface, err := GlobalRegistry.GetAgent(name)
    if err != nil {
        return nil, err
    }

    agent, ok := agentInterface.(*Agent)
    if !ok {
        return nil, fmt.Errorf("invalid agent type for %s", name)
    }

    return agent, nil
}

// RegisterAgent 注册 Agent 到全局注册表
func RegisterAgent(name string, agent *Agent) error {
    return GlobalRegistry.RegisterAgent(name, agent)
}

// ListAgents 列出所有已注册的 Agent
func ListAgents() ([]string, error) {
    return GlobalRegistry.ListAgents()
}
```

### 5. Agent 完整使用流程

#### 5.1 配置文件示例（完整版）

```yaml
# config.yaml - 完整的 Anyi + Agent 配置

# LLM 客户端配置
clients:
  - name: openai-gpt4
    type: openai
    apiKey: "${OPENAI_API_KEY}"
    model: gpt-4
    default: true

  - name: anthropic-claude
    type: anthropic
    apiKey: "${ANTHROPIC_API_KEY}"
    model: claude-3-opus

# 工作流配置
flows:
  - name: web_search
    clientName: openai-gpt4
    steps:
      - name: search_query_generation
        executor:
          type: llm
          withconfig:
            prompt: |
              Generate search queries for: {{.input}}
              Provide 3-5 specific search terms.

  - name: document_analysis
    clientName: anthropic-claude
    steps:
      - name: content_analysis
        executor:
          type: llm
          withconfig:
            prompt: |
              Analyze the following document and provide key insights:
              {{.input}}

              Focus on:
              1. Main topics
              2. Key findings
              3. Actionable recommendations

  - name: report_generation
    clientName: openai-gpt4
    steps:
      - name: synthesis
        executor:
          type: llm
          withconfig:
            prompt: |
              Based on the research findings below, create a comprehensive report:
              {{.input}}

              Include:
              - Executive Summary
              - Detailed Analysis
              - Recommendations

# Agent 配置（新增）
agents:
  - name: research_assistant
    description: "An AI research assistant that can search, analyze, and synthesize information"
    flows:
      - web_search
      - document_analysis
      - report_generation
    clientName: openai-gpt4
    config:
      max_search_results: 10
      analysis_depth: "comprehensive"
      report_format: "professional"

  - name: content_creator
    description: "An AI content creator specialized in writing and editing"
    flows:
      - document_analysis
      - report_generation
    clientName: anthropic-claude
    config:
      writing_style: "engaging"
      target_audience: "general"
```

#### 5.2 代码使用示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/jieliu2000/anyi"
)

func main() {
    // 1. 加载配置（包括 Agent）
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // 2. 获取 Agent
    agent, err := anyi.GetAgent("research_assistant")
    if err != nil {
        log.Fatal("Failed to get agent:", err)
    }

    // 3. 执行任务 - 极简接口！
    result, err := agent.Execute("Research the latest developments in AI safety and create a comprehensive report")
    if err != nil {
        log.Fatal("Failed to execute task:", err)
    }

    // 4. 获取结果
    fmt.Printf("Task completed successfully!\n")
    fmt.Printf("Final Output: %s\n", result.FinalOutput)
    fmt.Printf("Execution took: %v\n", result.Duration)

    // 5. 查看执行历史
    history := agent.GetExecutionHistory()
    fmt.Printf("Agent has completed %d tasks\n", len(history))
}
```

## 实现计划

### 阶段 1: 配置系统集成 ✅

#### 1.1 扩展 AnyiConfig 结构

- [x] 在 `config.go` 中添加 `AgentConfig` 类型
- [x] 扩展 `AnyiConfig` 包含 `Agents` 字段
- [x] 创建 `NewAgentFromConfig` 函数

#### 1.2 更新配置加载流程

- [x] 修改 `Config()` 函数支持 Agent 配置加载
- [x] 在 `ConfigFromFile()` 和 `ConfigFromString()` 中集成 Agent

### 阶段 2: 核心 Agent 框架 ✅

#### 2.1 创建 registry 包

- [x] 重构 GlobalRegistry 到 `registry/registry.go`
- [x] 创建 `registry/types.go` 定义接口避免循环引用
- [x] 实现 Agent 注册和检索方法

#### 2.2 实现 agent 包

- [x] 创建 `agent/types.go` - 核心类型定义
- [x] 创建 `agent/memory.go` - Agent 记忆系统
- [x] 创建 `agent/agent.go` - Agent 核心逻辑
- [x] 创建 `agent/planner.go` - 智能任务规划器
- [x] 创建 `agent/executor.go` - 任务执行器
- [x] 创建 `agent/config.go` - 配置加载工具

### 阶段 3: 主框架集成 🔄

#### 3.1 更新主配置文件

- [x] 在 `config.go` 中添加 `AgentConfig` 结构体
- [ ] 实现 `NewAgentFromConfig()` 函数
- [ ] 集成到 `Config()` 主函数

#### 3.2 添加便利函数

- [ ] 在 `anyi.go` 中添加 `GetAgent()` 函数
- [ ] 添加 `RegisterAgent()` 函数
- [ ] 添加 `ListAgents()` 函数

### 阶段 4: 测试和示例

#### 4.1 单元测试

- [x] Agent 基础功能测试
- [x] Memory 系统测试
- [ ] 集成测试

#### 4.2 示例和文档

- [x] 创建使用示例
- [ ] 编写详细文档
- [ ] 性能测试

## 关键技术决策

### 1. 配置集成策略

**决策**: 扩展现有的 `AnyiConfig` 结构，而不是创建独立的配置系统

**原因**:

- 保持配置的一致性和统一性
- 利用现有的配置加载机制（支持多种格式）
- 避免用户需要维护多个配置文件
- 确保所有组件（Client、Flow、Agent）在同一配置中协调工作

### 2. 接口设计原则

**决策**: Agent.Execute() 只需要一个字符串参数

**原因**:

- 符合用户的简化需求
- 隐藏内部复杂性
- 让 LLM 负责智能规划，而不是用户手动构建复杂对象
- 提供最佳的用户体验

### 3. 循环依赖解决方案

**决策**: 创建独立的 `registry` 包，使用接口隔离

**原因**:

- 彻底解决 Agent ↔ Flow 循环引用问题
- 保持代码的清晰架构
- 提供类型安全的访问方式
- 为未来扩展留下空间

### 4. Agent 执行模型

**决策**: Agent 不直接执行任务，而是规划和协调 Flow 执行

**原因**:

- 保持与现有 Anyi 架构的兼容性
- Flow 仍然是实际的执行单元
- Agent 专注于智能规划和协调
- 重用现有的 Step、Executor、Validator 生态系统

## 向后兼容性

### 现有代码保持不变

1. **Flow 和 Step**: 完全向后兼容，无需修改现有工作流
2. **Client 配置**: 现有 LLM 客户端配置保持不变
3. **配置文件**: 现有配置文件无需修改，只需添加 agents 部分（可选）
4. **API**: 所有现有的 `anyi.GetClient()`, `anyi.GetFlow()` 等函数保持不变

### 渐进式采用

用户可以：

1. 继续使用现有的 Flow 方式
2. 逐步添加 Agent 配置
3. 混合使用 Flow 和 Agent
4. 完全迁移到 Agent 模式（如果需要）

## 使用场景示例

### 场景 1: 研究助手

```go
// 配置文件中定义研究助手 Agent
agent, _ := anyi.GetAgent("research_assistant")
result, _ := agent.Execute("分析人工智能在医疗领域的最新应用，并生成详细报告")
```

### 场景 2: 内容创作

```go
// 配置文件中定义内容创作 Agent
agent, _ := anyi.GetAgent("content_creator")
result, _ := agent.Execute("为我们的产品写一篇引人入胜的营销文章")
```

### 场景 3: 数据分析

```go
// 配置文件中定义数据分析 Agent
agent, _ := anyi.GetAgent("data_analyst")
result, _ := agent.Execute("分析销售数据趋势并提供业务建议")
```

## 结论

这个设计方案实现了用户的核心需求：

1. ✅ **统一配置**: 通过 `anyi.ConfigFromFile()` 一次性配置所有组件
2. ✅ **简单获取**: 通过 `anyi.GetAgent("name")` 获取 Agent
3. ✅ **极简执行**: 通过 `agent.Execute("objective")` 执行任务

同时保持了：

- 向后兼容性
- 架构清晰性
- 类型安全
- 扩展性

该方案不仅满足了当前需求，还为未来的 Agent 生态系统发展奠定了坚实的基础。

#### 6.1 新的包结构

```
anyi/
├── registry/          # 新增：注册表包
│   ├── registry.go    # 核心注册表逻辑
│   └── types.go       # 注册表相关类型定义
├── agent/             # 新增：Agent 框架包
│   ├── agent.go       # Agent 定义和基础功能
│   ├── types.go       # 类型定义
│   ├── memory.go      # Agent 记忆系统
│   ├── planner.go     # 任务规划器
│   ├── executor.go    # 任务执行器
│   ├── config.go      # Agent 配置加载
│   └── example.go     # 使用示例
├── flow/              # 现有：工作流包
├── llm/               # 现有：LLM 包
├── config.go          # 更新：集成 Agent 配置
└── anyi.go            # 更新：添加 Agent 便利函数
```

#### 6.2 Registry 包设计

```go
// registry/registry.go
package registry

import (
    "sync"
    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/llm"
    "github.com/jieliu2000/anyi/llm/chat"
)

type Registry struct {
    mu         sync.RWMutex
    clients    map[string]llm.Client
    flows      map[string]*flow.Flow
    validators map[string]flow.StepValidator
    executors  map[string]flow.StepExecutor
    formatters map[string]chat.PromptFormatter
    agents     map[string]interface{} // 使用 interface{} 避免循环引用
    defaultClientName string
}

// Agent 接口定义（避免循环引用）
type Agent interface {
    GetName() string
    GetFlows() []string
    GetClientName() string
}

// 新增 Agent 相关方法
func (r *Registry) RegisterAgent(name string, agent interface{}) error
func (r *Registry) GetAgent(name string) (interface{}, error)
func (r *Registry) GetFlows(agent interface{}) ([]*flow.Flow, error)
func (r *Registry) ListAgents() ([]string, error)
```

    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/llm"
    "github.com/jieliu2000/anyi/llm/chat"

)

type Registry struct {
mu sync.RWMutex
clients map[string]llm.Client
flows map[string]*flow.Flow
validators map[string]flow.StepValidator
executors map[string]flow.StepExecutor
formatters map[string]chat.PromptFormatter
agents map[string]*Agent // 新增：Agent 注册
defaultClientName string
}

var GlobalRegistry = NewRegistry()

func NewRegistry() *Registry {
return &Registry{
clients: make(map[string]llm.Client),
flows: make(map[string]*flow.Flow),
validators: make(map[string]flow.StepValidator),
executors: make(map[string]flow.StepExecutor),
formatters: make(map[string]chat.PromptFormatter),
agents: make(map[string]\*Agent),
}
}

// Agent 相关方法
func (r *Registry) RegisterAgent(name string, agent *Agent) error
func (r *Registry) GetAgent(name string) (*Agent, error)
func (r *Registry) GetFlows(agent *Agent) ([]\*flow.Flow, error)

````

### 3. Agent 核心功能

#### 3.1 智能任务规划器（核心组件）

```go
// agent/planner.go
type TaskPlanner struct {
    registry *registry.Registry
    client   llm.Client
}

// PlanExecution 是核心方法：根据目标字符串和 Agent 能力生成执行计划
func (p *TaskPlanner) PlanExecution(objective string, agent *Agent) (*ExecutionPlan, error) {
    // 1. 获取 Agent 可用的工作流信息
    flows, err := p.registry.GetFlows(agent)
    if err != nil {
        return nil, err
    }

    // 2. 构建规划提示
    planningPrompt := p.buildPlanningPrompt(objective, agent, flows)

    // 3. 调用 LLM 进行智能规划
    planJSON, err := p.generateExecutionPlan(planningPrompt)
    if err != nil {
        return nil, err
    }

    // 4. 解析和验证执行计划
    return p.parseAndValidatePlan(planJSON, flows)
}func (p *TaskPlanner) buildPlanningPrompt(objective string, agent *Agent, flows []*flow.Flow) string {
    return fmt.Sprintf(`
你是一个智能任务规划器。请根据以下信息制定详细的执行计划：

**任务目标**: %s

**Agent能力描述**: %s

**可用的工作流**:
%s

请制定一个执行计划，将任务分解为一系列工作流调用步骤。每个步骤必须使用上述可用工作流之一。

输出格式为 JSON：
{
  "objective": "%s",
  "description": "执行计划的总体描述",
  "steps": [
    {
      "flowName": "工作流名称",
      "input": "该步骤的输入文本",
      "variables": {"key": "value"},
      "description": "该步骤的作用说明",
      "order": 1
    }
  ]
}
    `, objective, agent.Description, p.formatFlowsInfo(flows), objective)
}
````

#### 3.2 Agent 执行器（极简版）

```go
// agent/executor.go
type AgentExecutor struct {
    Agent    *Agent
    Planner  *TaskPlanner
    Registry *registry.Registry
}

// Execute 是 Agent 的核心执行方法 - 只需要目标字符串
func (e *AgentExecutor) Execute(objective string) (*TaskResult, error) {
    // 1. 使用智能规划器生成执行计划
    plan, err := e.Planner.PlanExecution(objective, e.Agent)
    if err != nil {
        return nil, fmt.Errorf("planning failed: %w", err)
    }

    // 2. 顺序执行计划中的每个 Flow 步骤
    context := &flow.FlowContext{
        Text:      objective, // 初始输入就是目标描述
        Variables: make(map[string]any),
    }

    executionResults := make([]StepResult, 0, len(plan.Steps))

    for _, step := range plan.Steps {
        // 获取要执行的 Flow
        targetFlow, err := e.Registry.GetFlow(step.FlowName)
        if err != nil {
            return nil, fmt.Errorf("flow %s not found: %w", step.FlowName, err)
        }

        // 准备 Flow 的输入和变量
        context.Text = step.Input
        if step.Variables != nil {
            for k, v := range step.Variables {
                context.Variables[k] = v
            }
        }

        // 执行 Flow
        result, err := targetFlow.Run(*context)
        if err != nil {
            return nil, fmt.Errorf("flow %s execution failed: %w", step.FlowName, err)
        }

        // 记录步骤结果
        executionResults = append(executionResults, StepResult{
            FlowName:    step.FlowName,
            Description: step.Description,
            Input:       step.Input,
            Output:      result.Text,
            Variables:   result.Variables,
        })

        // 将结果传递给下一步
        context = result
    }

    return &TaskResult{
        Objective:        objective,
        FinalOutput:      context.Text,
        ExecutionPlan:    plan,
        StepResults:      executionResults,
        Status:           TaskStatusCompleted,
        FinalVariables:   context.Variables,
    }, nil
}

// Agent 本身也提供 Execute 方法，自动创建执行器
func (agent *Agent) Execute(objective string) (*TaskResult, error) {
    executor := &AgentExecutor{
        Agent:    agent,
        Planner:  NewTaskPlanner(registry.GlobalRegistry, agent.getClient()),
        Registry: registry.GlobalRegistry,
    }
    return executor.Execute(objective)
}

type StepResult struct {
    FlowName    string         `mapstructure:"flowName"`
    Description string         `mapstructure:"description"`
    Input       string         `mapstructure:"input"`
    Output      string         `mapstructure:"output"`
    Variables   map[string]any `mapstructure:"variables"`
}

type TaskResult struct {
    Objective        string                 `mapstructure:"objective"`
    FinalOutput      string                 `mapstructure:"finalOutput"`
    ExecutionPlan    *ExecutionPlan         `mapstructure:"executionPlan"`
    StepResults      []StepResult           `mapstructure:"stepResults"`
    Status           TaskStatus             `mapstructure:"status"`
    FinalVariables   map[string]any         `mapstructure:"finalVariables"`
    ExecutedAt       time.Time              `mapstructure:"executedAt"`
}
```

#### 3.3 Agent 记忆系统（简化版）

```go
// agent/memory.go
type AgentMemory interface {
    Store(key string, value any) error
    Retrieve(key string) (any, error)
    Search(query string) ([]MemoryItem, error)
}

type MemoryItem struct {
    Key       string    `mapstructure:"key"`
    Value     any       `mapstructure:"value"`
    Timestamp time.Time `mapstructure:"timestamp"`
}

// 简单的内存实现
type SimpleMemory struct {
    items map[string]MemoryItem
    mutex sync.RWMutex
}

func (m *SimpleMemory) Store(key string, value any) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.items[key] = MemoryItem{
        Key:       key,
        Value:     value,
        Timestamp: time.Now(),
    }
    return nil
}
```

### 4. 配置文件支持

#### 4.1 Agent 配置（简化版）

````yaml
# agents.yaml
agents:
  - name: "research_agent"
    description: "专门用于研究和分析的智能代理"
    flows:
      - "web_search_flow"
      - "data_analysis_flow"
      - "report_generation_flow"
    clientName: "gpt4"

  - name: "content_agent"
    description: "专门用于内容创作的智能代理"
    flows:
      - "content_generation_flow"
      - "editing_flow"
      - "formatting_flow"
    clientName: "claude"
```#### 4.2 配置加载

```go
// agent/config.go
type AgentConfig struct {
    Agents []Agent `mapstructure:"agents"`
}

func LoadAgentConfig(configFile string) error {
    var config AgentConfig
    err := utils.LoadConfig(configFile, &config)
    if err != nil {
        return err
    }

    // 注册所有 Agents
    for _, agent := range config.Agents {
        err = registry.GlobalRegistry.RegisterAgent(agent.Name, &agent)
        if err != nil {
            return err
        }
    }

    return nil
}
````

### 5. 使用示例

#### 5.1 编程方式使用（极简版）

```go
package main

import (
    "log"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent"
    "github.com/jieliu2000/anyi/registry"
)

func main() {
    // 初始化 Anyi（注册基础 Flow）
    anyi.Init()

    // 注册一些工作流
    // ... 注册 web_search_flow, data_analysis_flow 等

    // 创建研究 Agent
    researchAgent := &agent.Agent{
        Name:        "research_agent",
        Description: "专门用于研究和分析的智能代理",
        Flows:       []string{"web_search_flow", "data_analysis_flow", "report_generation_flow"},
        ClientName:  "gpt4",
        Memory:      agent.NewSimpleMemory(),
    }

    // 注册 Agent
    registry.GlobalRegistry.RegisterAgent("research_agent", researchAgent)

    // 执行任务 - 只需要一个目标字符串！
    result, err := researchAgent.Execute("研究2025年AI发展趋势并生成详细报告")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("研究报告: %s", result.FinalOutput)
    log.Printf("执行了 %d 个步骤", len(result.StepResults))

    // 查看执行计划
    for i, step := range result.StepResults {
        log.Printf("步骤 %d: %s -> %s", i+1, step.FlowName, step.Description)
    }
}
```

#### 5.2 配置文件方式使用（极简版）

```go
package main

import (
    "log"
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent"
    "github.com/jieliu2000/anyi/registry"
)

func main() {
    // 加载基础配置（LLM客户端和工作流）
    err := anyi.ConfigFromFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 加载 Agent 配置
    err = agent.LoadAgentConfig("agents.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 获取 Agent
    researchAgent, err := registry.GlobalRegistry.GetAgent("research_agent")
    if err != nil {
        log.Fatal(err)
    }

    // 直接执行任务 - 极简接口
    result, err := researchAgent.Execute("分析当前AI市场竞争格局")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("分析结果: %s", result.FinalOutput)
}
```

#### 5.3 批量任务执行

````go
// 同一个 Agent 可以连续执行多个相关任务
objectives := []string{
    "收集最新的AI技术新闻",
    "分析AI技术发展趋势",
    "生成市场预测报告",
}

for _, objective := range objectives {
    result, err := researchAgent.Execute(objective)
    if err != nil {
        log.Printf("执行失败: %v", err)
        continue
    }
    log.Printf("完成任务: %s", objective)
    log.Printf("结果: %s\n", result.FinalOutput)
}
```### 6. 高级特性

#### 6.1 动态工作流推荐

```go
// Agent 可以根据任务特点推荐最适合的工作流组合
type FlowRecommender struct {
    registry *registry.Registry
    client   llm.Client
}

func (r *FlowRecommender) RecommendFlows(objective string, flows []string) ([]string, error) {
    // 使用 LLM 分析任务需求，推荐最佳工作流组合
}
````

#### 6.2 执行结果学习

```go
// 根据执行效果优化未来的规划
type ExecutionLearner struct {
    memory AgentMemory
}

func (l *ExecutionLearner) LearnFromExecution(objective string, result *TaskResult) error {
    // 分析执行效果，存储经验用于优化未来规划
}
```

## 7. 实施计划

### 阶段一：核心架构重构（1-2 周）

1. 创建 `registry` 包，迁移注册表逻辑
2. 重构 `anyi.go`，保持向后兼容
3. 创建 `agent` 包基础结构

### 阶段二：核心 Agent 功能（2-3 周）

1. 实现 Agent、AgentTask 基础类型
2. 实现智能任务规划器（TaskPlanner）
3. 实现 Agent 执行器（AgentExecutor）

### 阶段三：配置和集成（1-2 周）

1. 实现 Agent 记忆系统
2. 添加配置文件支持
3. 完善文档和示例

### 阶段四：高级功能（可选，1-2 周）

1. 添加动态工作流推荐
2. 实现执行结果学习
3. 性能优化和测试

## 8. 核心优势

### 8.1 极简的接口设计

- **一行代码执行任务**：`agent.Execute("任务目标")` 就能完成复杂任务
- **无需复杂配置**：Agent 从自身配置中读取所有必要信息
- **自动智能规划**：LLM 根据目标和可用 Flow 自动生成最佳执行计划

### 8.2 概念简化

- **没有复杂的 Task 对象**：用户只需要提供目标字符串
- **Agent 是核心抽象**：Agent 包含了执行任务所需的全部信息
- **复用现有 Flow**：Agent 的执行单元仍然是 Anyi 的 Flow，无需重新发明轮子

### 8.3 与现有架构的完美融合

- 完全向后兼容现有的 Flow 和 Step 系统
- 复用现有的 LLM 客户端和执行器
- 保持配置驱动的设计理念
- 利用现有的验证和重试机制

### 8.4 智能化程度高

- **LLM 驱动的任务规划**：根据目标描述和可用 Flow 智能生成执行计划
- **动态执行**：不是预定义的步骤，而是根据任务动态规划
- **上下文传递**：执行过程中的结果自动传递给下一个 Flow
- **自适应能力**：同一个 Agent 可以处理不同类型的相关任务

### 8.5 极简但强大

- **最简 API**：`agent.Execute(objective)` 是唯一需要的接口
- **配置驱动**：通过 YAML 配置 Agent 的能力范围
- **扩展性强**：可以轻松添加新的工作流到 Agent 的能力列表
- **易于测试**：简单的输入输出，便于单元测试

## 9. 与 CrewAI 的区别

| 特性       | CrewAI              | Anyi Agent Framework  |
| ---------- | ------------------- | --------------------- |
| 基本概念   | Agent + Crew + Task | Agent + 目标字符串    |
| 执行单元   | 自定义 Tool         | 现有的 Flow           |
| 规划方式   | 预定义任务流程      | LLM 动态规划          |
| 配置方式   | Python 代码         | YAML/JSON 配置 + 代码 |
| 输入接口   | 复杂的 Task 对象    | 简单的目标字符串      |
| 语言       | Python              | Go                    |
| 架构复杂度 | 相对复杂            | 更简洁                |

## 10. 设计总结

这个优化后的设计方案实现了以下关键改进：

### 10.1 **极简的输入接口**

- **唯一输入**：Agent 执行任务只需要一个目标字符串 `agent.Execute("研究AI趋势")`
- **内部自给自足**：Agent 从自身配置中获取所有必要信息（可用 Flow、规划器客户端等）
- **无外部依赖**：用户不需要创建复杂的 Task 对象或提供额外配置

### 10.2 **智能自主规划**

- **LLM 驱动**：根据目标字符串和 Agent 能力智能生成执行计划
- **动态适应**：同一个 Agent 可以处理多种相关任务，无需预定义流程
- **上下文流转**：执行过程中自动处理 Flow 间的数据传递

### 10.3 **完美的架构融合**

- **零破坏性**：完全复用现有的 Flow 系统，无需重写任何现有代码
- **注册表重构**：通过 registry 包避免循环引用，保持架构清洁
- **配置驱动**：保持 Anyi 的配置化理念，Agent 能力通过 YAML 定义

### 10.4 **核心设计理念**

Agent 是一个**智能的工作流协调器**，它：

- 知道自己能做什么（Flows）
- 知道如何规划（ClientName）
- 能够根据任务目标动态决定执行哪些 Flow 以及执行顺序
- 自动处理 Flow 间的数据流转和上下文管理

这种设计让 Anyi 能够支持真正智能的代理功能，同时保持其独特的简洁性和强大的配置驱动特性。用户可以通过一行代码让 Agent 完成复杂的多步骤任务，而 Agent 会自主地规划和执行最佳的工作流序列。
