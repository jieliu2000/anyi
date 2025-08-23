# Anyi Agent System Design - CrewAI风格集成

## 概述

本文档提出在Anyi框架中添加类似CrewAI的agent支持系统的设计方案。该系统将允许用户定义具有角色、背景故事和可用工作流的agent，并通过LLM进行自主任务规划和动态执行。

## 当前Anyi架构分析

### 核心组件
1. **Registry系统**: 全局注册表管理Clients、Flows、Validators、Executors、Formatters
2. **Flow系统**: 工作流执行引擎，支持多步骤顺序执行
3. **Executor系统**: 步骤执行器（LLM、Conditional、Exec、SetContext等）
4. **配置系统**: 基于Viper的灵活配置管理

### 现有AGI能力
- 示例中展示了基本的任务分解和执行能力
- 支持任务生成、优先级排序、执行
- 但缺乏正式的agent定义和管理机制

## 设计目标

1. **类似CrewAI的agent定义**: 支持角色、背景故事、目标、工具集
2. **动态任务规划**: LLM根据agent能力和目标自主规划执行步骤
3. **工作流集成**: agent可以调用Anyi的现有工作流
4. **避免循环引用**: 保持清晰的包结构
5. **极简设计**: 保持代码简洁性和可维护性

## 架构设计

### 包结构规划

```
anyi/
├── agent/                    # 新增agent包
│   ├── model/               # Agent数据模型
│   │   ├── agent.go         # Agent定义
│   │   ├── crew.go          # Crew/Team定义
│   │   └── task.go          # Task定义
│   ├── manager/             # Agent管理
│   │   ├── registry.go      # Agent注册表
│   │   └── planner.go       # 任务规划器
│   ├── executor/            # Agent执行器
│   │   ├── base.go          # 基础执行器
│   │   └── flow_executor.go # 工作流执行器
│   └── agent.go             # 主入口文件
├── flow/                    # 现有flow包
└── llm/                     # 现有llm包
```

### 核心接口定义

```go
// agent/model/agent.go
package agentmodel

type Role string

type Agent struct {
    ID          string
    Name        string
    Role        Role
    Backstory   string
    Goal        string
    Capabilities []string  // 可用的工作流名称
    Tools       []Tool     // 可用工具
    Verbose     bool
}

type Tool interface {
    Name() string
    Description() string
    Execute(input string) (string, error)
}

// agent/model/crew.go
type Crew struct {
    ID      string
    Name    string
    Agents  []*Agent
    Tasks   []*Task
    Process ProcessType // sequential, hierarchical, etc.
}

type ProcessType string

const (
    ProcessSequential ProcessType = "sequential"
    ProcessHierarchical ProcessType = "hierarchical"
)

// agent/model/task.go
type Task struct {
    ID          string
    Description string
    ExpectedOutput string
    Agent       *Agent
    Context     map[string]interface{}
    Status      TaskStatus
}

type TaskStatus string
```

### Agent注册表

```go
// agent/manager/registry.go
package agentmanager

import (
    "sync"
    "github.com/jieliu2000/anyi/agent/agentmodel"
)

type AgentRegistry struct {
    mu     sync.RWMutex
    agents map[string]*agentmodel.Agent
    crews  map[string]*agentmodel.Crew
}

var GlobalAgentRegistry = &AgentRegistry{
    agents: make(map[string]*agentmodel.Agent),
    crews:  make(map[string]*agentmodel.Crew),
}

func (r *AgentRegistry) RegisterAgent(agent *agentmodel.Agent) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.agents[agent.ID]; exists {
        return fmt.Errorf("agent with ID %s already exists", agent.ID)
    }
    
    r.agents[agent.ID] = agent
    return nil
}

func (r *AgentRegistry) GetAgent(id string) (*agentmodel.Agent, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    agent, exists := r.agents[id]
    if !exists {
        return nil, fmt.Errorf("agent with ID %s not found", id)
    }
    
    return agent, nil
}
```

### 任务规划器

```go
// agent/manager/planner.go
package agentmanager

import (
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent/agentmodel"
    "github.com/jieliu2000/anyi/llm/chat"
)

type TaskPlanner struct {
    client anyi.LLMClient
}

func NewTaskPlanner(client anyi.LLMClient) *TaskPlanner {
    return &TaskPlanner{client: client}
}

func (p *TaskPlanner) PlanTask(agent *agentmodel.Agent, objective string) ([]*agentmodel.Task, error) {
    // 构建规划提示
    prompt := p.buildPlanningPrompt(agent, objective)
    
    // 调用LLM进行规划
    messages := []chat.Message{
        chat.NewSystemMessage("You are an expert task planner for AI agents."),
        chat.NewUserMessage(prompt),
    }
    
    response, _, err := p.client.Chat(messages, nil)
    if err != nil {
        return nil, err
    }
    
    // 解析LLM响应为任务列表
    tasks, err := p.parseTaskPlan(response.Content, agent)
    if err != nil {
        return nil, err
    }
    
    return tasks, nil
}

func (p *TaskPlanner) buildPlanningPrompt(agent *agentmodel.Agent, objective string) string {
    return fmt.Sprintf(`Agent Role: %s
Agent Backstory: %s
Agent Goal: %s
Available Capabilities: %v

Objective: %s

Please break down this objective into a sequence of tasks that this agent can execute using its available capabilities. Each task should be specific and actionable.

Return the tasks in JSON format:
{
    "tasks": [
        {
            "description": "task description",
            "expected_output": "what should be produced",
            "capability": "flow_name_to_use"
        }
    ]
}`, agent.Role, agent.Backstory, agent.Goal, agent.Capabilities, objective)
}
```

### 工作流执行器

```go
// agent/executor/flow_executor.go
package agentexecutor

import (
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent/agentmodel"
    "github.com/jieliu2000/anyi/flow"
)

type FlowExecutor struct {
    registry *anyi.AnyiRegistry
}

func NewFlowExecutor() *FlowExecutor {
    return &FlowExecutor{
        registry: anyi.GlobalRegistry,
    }
}

func (e *FlowExecutor) ExecuteFlow(flowName string, context *flow.FlowContext) (*flow.FlowContext, error) {
    flow, err := e.registry.GetFlow(flowName)
    if err != nil {
        return nil, err
    }
    
    return flow.Run(*context)
}

func (e *FlowExecutor) CanExecute(capability string) bool {
    _, err := e.registry.GetFlow(capability)
    return err == nil
}
```

### 主Agent接口

```go
// agent/agent.go
package agent

import (
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent/agentmanager"
    "github.com/jieliu2000/anyi/agent/agentmodel"
    "github.com/jieliu2000/anyi/flow"
)

type AgentSystem struct {
    planner  *agentmanager.TaskPlanner
    executor *FlowExecutor
}

func NewAgentSystem(client anyi.LLMClient) *AgentSystem {
    return &AgentSystem{
        planner:  agentmanager.NewTaskPlanner(client),
        executor: NewFlowExecutor(),
    }
}

func (s *AgentSystem) ExecuteObjective(agentID, objective string) ([]*agentmodel.TaskResult, error) {
    // 获取agent
    agent, err := agentmanager.GlobalAgentRegistry.GetAgent(agentID)
    if err != nil {
        return nil, err
    }
    
    // 规划任务
    tasks, err := s.planner.PlanTask(agent, objective)
    if err != nil {
        return nil, err
    }
    
    // 执行任务
    results := make([]*agentmodel.TaskResult, 0, len(tasks))
    for _, task := range tasks {
        result, err := s.executeTask(agent, task)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    
    return results, nil
}

func (s *AgentSystem) executeTask(agent *agentmodel.Agent, task *agentmodel.Task) (*agentmodel.TaskResult, error) {
    // 创建执行上下文
    context := flow.NewFlowContextWithVariables(
        task.Description,
        task.Context,
        make(map[string]any),
    )
    
    // 执行工作流
    result, err := s.executor.ExecuteFlow(task.Capability, context)
    if err != nil {
        return nil, err
    }
    
    return &agentmodel.TaskResult{
        Task:        task,
        Output:      result.Text,
        Success:     true,
        Metadata:    result.Variables,
    }, nil
}
```

## 配置集成

### Agent配置结构

```go
// 扩展AnyiConfig
type AgentConfig struct {
    ID          string                 `mapstructure:"id" json:"id" yaml:"id"`
    Name        string                 `mapstructure:"name" json:"name" yaml:"name"`
    Role        string                 `mapstructure:"role" json:"role" yaml:"role"`
    Backstory   string                 `mapstructure:"backstory" json:"backstory" yaml:"backstory"`
    Goal        string                 `mapstructure:"goal" json:"goal" yaml:"goal"`
    Capabilities []string              `mapstructure:"capabilities" json:"capabilities" yaml:"capabilities"`
    Verbose     bool                   `mapstructure:"verbose" json:"verbose" yaml:"verbose"`
}

type CrewConfig struct {
    ID      string        `mapstructure:"id" json:"id" yaml:"id"`
    Name    string        `mapstructure:"name" json:"name" yaml:"name"`
    Agents  []string      `mapstructure:"agents" json:"agents" yaml:"agents"` // Agent IDs
    Process string        `mapstructure:"process" json:"process" yaml:"process"`
}

// 扩展AnyiConfig结构
type AnyiConfig struct {
    Clients    []llm.ClientConfig
    Flows      []FlowConfig
    Formatters []FormatterConfig
    Agents     []AgentConfig    // 新增
    Crews      []CrewConfig     // 新增
}
```

## 使用示例

### 配置示例

```yaml
agents:
  - id: "research_agent"
    name: "Research Specialist"
    role: "Senior Research Analyst"
    backstory: "Expert in gathering and analyzing information from various sources. Skilled at summarizing findings and identifying key insights."
    goal: "Provide comprehensive research reports with actionable insights"
    capabilities:
      - "web_research_flow"
      - "data_analysis_flow"
      - "report_generation_flow"
    verbose: true

  - id: "coding_agent"
    name: "Code Developer"
    role: "Senior Software Engineer"
    backstory: "Experienced full-stack developer with expertise in multiple programming languages and frameworks."
    goal: "Write clean, efficient, and well-documented code"
    capabilities:
      - "code_generation_flow"
      - "code_review_flow"
      - "test_writing_flow"

crews:
  - id: "development_team"
    name: "Software Development Team"
    agents:
      - "research_agent"
      - "coding_agent"
    process: "sequential"
```

### 代码使用示例

```go
func main() {
    // 初始化Anyi
    anyi.ConfigFromFile("config.yaml")
    
    // 创建Agent系统
    client, _ := anyi.GetDefaultClient()
    agentSystem := agent.NewAgentSystem(client)
    
    // 执行任务
    results, err := agentSystem.ExecuteObjective(
        "research_agent",
        "Research the latest trends in AI programming assistants and provide a summary report",
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        log.Printf("Task: %s", result.Task.Description)
        log.Printf("Result: %s", result.Output)
    }
}
```

## 避免循环引用的策略

1. **单向依赖**: Agent包只依赖flow和llm包，不反向依赖
2. **接口隔离**: 通过接口定义清晰的边界
3. **注册表模式**: 使用全局注册表避免直接包引用
4. **依赖注入**: 通过构造函数注入依赖

## 实施步骤

1. **创建agent包结构**
2. **实现核心数据模型** (Agent, Crew, Task)
3. **实现注册表和管理器**
4. **集成任务规划器**
5. **实现工作流执行器**
6. **扩展配置系统**
7. **添加示例和文档**
8. **测试和验证**

## 优势

1. **与现有系统无缝集成**: 重用Anyi的工作流和LLM能力
2. **灵活的agent定义**: 支持角色、背景故事、目标定制
3. **动态规划**: LLM驱动的智能任务分解
4. **可扩展性**: 易于添加新的agent类型和能力
5. **极简设计**: 保持代码简洁性和可维护性

这个设计方案将为Anyi框架提供强大的CrewAI风格agent支持，同时保持项目的简洁性和架构清晰性。
