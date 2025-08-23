# Anyi Agent 系统设计方案

## 1. 概述

本设计提出在 Anyi 框架基础上添加类似 CrewAI 的 Agent 支持，实现 Agent 角色定义、工作流分配和自主任务规划执行功能。

## 2. 当前架构分析

### 2.1 现有架构优势

- **全局注册表设计**: `GlobalRegistry`统一管理 Clients、Flows、Validators、Executors、Formatters
- **模块化架构**: 清晰的包结构，flow、llm、chat 等功能分离
- **灵活的执行器系统**: 支持多种执行器类型（LLM、条件分支、命令执行等）
- **工作流系统**: 完整的 Flow-Step-Context 执行模型

### 2.2 潜在的循环引用问题

当前`GlobalRegistry`在根包`anyi`中，而 Agent 需要：

1. 注册到注册表中（Agent → Registry）
2. 从注册表中读取 Flow（Agent → Registry → Flow）

如果 Agent 也在根包中，可能导致包内循环引用。

## 3. 解决方案设计

### 3.1 包结构重构

为避免循环引用，将注册表功能移至独立包：

```
anyi/
├── registry/              # 新增：注册表包
│   ├── registry.go       # 全局注册表
│   └── registry_test.go
├── agent/                # 新增：Agent系统包
│   ├── agent.go          # Agent定义和执行
│   ├── crew.go           # Crew团队管理
│   ├── planner.go        # 任务规划器
│   ├── executor.go       # Agent执行器
│   └── types.go          # Agent相关类型定义
├── flow/                 # 现有：工作流包
├── llm/                  # 现有：LLM客户端包
└── anyi.go              # 根包：保持向后兼容的API
```

### 3.2 核心组件设计

#### 3.2.1 Agent 定义

```go
// agent/types.go
package agent

import (
    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/llm"
)

type Agent struct {
    Name          string                 `json:"name" yaml:"name"`
    Role          string                 `json:"role" yaml:"role"`
    Backstory     string                 `json:"backstory" yaml:"backstory"`
    Goal          string                 `json:"goal" yaml:"goal"`
    AllowedFlows  []string              `json:"allowedFlows" yaml:"allowedFlows"`
    Client        llm.Client            `json:"-" yaml:"-"`
    ClientName    string                `json:"clientName" yaml:"clientName"`
    MaxIterations int                   `json:"maxIterations" yaml:"maxIterations"`
    Variables     map[string]any        `json:"variables" yaml:"variables"`
}

type Task struct {
    ID          string     `json:"id"`
    Description string     `json:"description"`
    AgentName   string     `json:"agentName"`
    Context     string     `json:"context"`
    ExpectedOutput string  `json:"expectedOutput"`
    Status      TaskStatus `json:"status"`
    Result      string     `json:"result"`
    Error       error      `json:"error,omitempty"`
}

type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
)

type ExecutionPlan struct {
    Tasks []PlanTask `json:"tasks"`
}

type PlanTask struct {
    StepID      string            `json:"stepId"`
    FlowName    string            `json:"flowName"`
    Description string            `json:"description"`
    Inputs      map[string]any    `json:"inputs"`
    Dependencies []string         `json:"dependencies"`
}
```

#### 3.2.2 注册表重构

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
    mu          sync.RWMutex
    clients     map[string]llm.Client
    flows       map[string]*flow.Flow
    validators  map[string]flow.StepValidator
    executors   map[string]flow.StepExecutor
    formatters  map[string]chat.PromptFormatter
    agents      map[string]*Agent    // 新增：Agent注册
    crews       map[string]*Crew     // 新增：Crew注册
    defaultClientName string
}

var Global *Registry = NewRegistry()

func NewRegistry() *Registry {
    return &Registry{
        clients:    make(map[string]llm.Client),
        flows:      make(map[string]*flow.Flow),
        validators: make(map[string]flow.StepValidator),
        executors:  make(map[string]flow.StepExecutor),
        formatters: make(map[string]chat.PromptFormatter),
        agents:     make(map[string]*Agent),
        crews:      make(map[string]*Crew),
    }
}

// Agent相关注册方法
func (r *Registry) RegisterAgent(name string, agent *Agent) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.agents[name]; exists {
        return fmt.Errorf("agent with name %q already exists", name)
    }

    r.agents[name] = agent
    return nil
}

func (r *Registry) GetAgent(name string) (*Agent, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    agent, ok := r.agents[name]
    if !ok {
        return nil, fmt.Errorf("agent %q not found", name)
    }
    return agent, nil
}

func (r *Registry) ListAgents() map[string]*Agent {
    r.mu.RLock()
    defer r.mu.RUnlock()

    agents := make(map[string]*Agent)
    for k, v := range r.agents {
        agents[k] = v
    }
    return agents
}
```

#### 3.2.3 Agent 执行器

```go
// agent/executor.go
package agent

import (
    "context"
    "fmt"
    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/registry"
)

type AgentExecutor struct {
    AgentName string `json:"agentName" yaml:"agentName"`
    TaskDescription string `json:"taskDescription" yaml:"taskDescription"`
    Context string `json:"context" yaml:"context"`
}

func (e *AgentExecutor) Init() error {
    if e.AgentName == "" {
        return fmt.Errorf("agent name is required")
    }

    // 验证Agent是否存在
    _, err := registry.Global.GetAgent(e.AgentName)
    return err
}

func (e *AgentExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    agent, err := registry.Global.GetAgent(e.AgentName)
    if err != nil {
        return &flowContext, err
    }

    // 创建任务
    task := &Task{
        ID:          generateTaskID(),
        Description: e.TaskDescription,
        AgentName:   e.AgentName,
        Context:     e.Context,
        Status:      TaskStatusPending,
    }

    // 执行任务
    result, err := e.executeTask(agent, task, flowContext)
    if err != nil {
        task.Status = TaskStatusFailed
        task.Error = err
        return &flowContext, err
    }

    task.Status = TaskStatusCompleted
    task.Result = result

    // 更新流上下文
    flowContext.Text = result
    flowContext.SetVariable("lastAgentTask", task)

    return &flowContext, nil
}

func (e *AgentExecutor) executeTask(agent *Agent, task *Task, context flow.FlowContext) (string, error) {
    planner := NewAgentPlanner(agent)

    // 生成执行计划
    plan, err := planner.GeneratePlan(task.Description, context)
    if err != nil {
        return "", fmt.Errorf("failed to generate plan: %w", err)
    }

    // 执行计划
    return e.executePlan(agent, plan, context)
}

func (e *AgentExecutor) executePlan(agent *Agent, plan *ExecutionPlan, context flow.FlowContext) (string, error) {
    var results []string
    currentContext := context

    for _, planTask := range plan.Tasks {
        // 检查依赖
        if err := e.checkDependencies(planTask.Dependencies, results); err != nil {
            return "", err
        }

        // 获取Flow
        targetFlow, err := registry.Global.GetFlow(planTask.FlowName)
        if err != nil {
            return "", fmt.Errorf("failed to get flow %s: %w", planTask.FlowName, err)
        }

        // 验证Agent是否有权限使用此Flow
        if !e.isFlowAllowed(agent, planTask.FlowName) {
            return "", fmt.Errorf("agent %s is not allowed to use flow %s", agent.Name, planTask.FlowName)
        }

        // 准备输入
        flowInput := e.prepareFlowInput(planTask, currentContext)

        // 执行Flow
        result, err := targetFlow.RunWithInput(flowInput)
        if err != nil {
            return "", fmt.Errorf("failed to execute flow %s: %w", planTask.FlowName, err)
        }

        results = append(results, result.Text)
        currentContext = *result
    }

    // 合并结果
    return e.combineResults(results), nil
}

func (e *AgentExecutor) isFlowAllowed(agent *Agent, flowName string) bool {
    for _, allowedFlow := range agent.AllowedFlows {
        if allowedFlow == flowName || allowedFlow == "*" {
            return true
        }
    }
    return false
}

func (e *AgentExecutor) prepareFlowInput(planTask PlanTask, context flow.FlowContext) string {
    // 根据planTask的输入配置和当前上下文准备Flow输入
    input := planTask.Description
    if context.Text != "" {
        input = fmt.Sprintf("%s\n\nContext: %s", input, context.Text)
    }
    return input
}

func (e *AgentExecutor) combineResults(results []string) string {
    if len(results) == 0 {
        return ""
    }
    if len(results) == 1 {
        return results[0]
    }

    combined := "Task execution completed with the following results:\n\n"
    for i, result := range results {
        combined += fmt.Sprintf("Step %d: %s\n\n", i+1, result)
    }
    return combined
}

func (e *AgentExecutor) checkDependencies(dependencies []string, completedTasks []string) error {
    // 简化的依赖检查实现
    if len(dependencies) > len(completedTasks) {
        return fmt.Errorf("dependencies not satisfied")
    }
    return nil
}

func generateTaskID() string {
    // 生成唯一任务ID的实现
    return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
```

#### 3.2.4 任务规划器

```go
// agent/planner.go
package agent

import (
    "encoding/json"
    "fmt"
    "strings"
    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/llm/chat"
    "github.com/jieliu2000/anyi/registry"
)

type AgentPlanner struct {
    agent *Agent
}

func NewAgentPlanner(agent *Agent) *AgentPlanner {
    return &AgentPlanner{agent: agent}
}

func (p *AgentPlanner) GeneratePlan(taskDescription string, context flow.FlowContext) (*ExecutionPlan, error) {
    // 获取Agent可用的Flow列表
    availableFlows, err := p.getAvailableFlows()
    if err != nil {
        return nil, err
    }

    // 构建规划提示
    prompt := p.buildPlanningPrompt(taskDescription, availableFlows, context)

    // 使用LLM生成计划
    client := p.agent.Client
    if client == nil {
        if p.agent.ClientName != "" {
            client, err = registry.Global.GetClient(p.agent.ClientName)
            if err != nil {
                return nil, fmt.Errorf("failed to get client %s: %w", p.agent.ClientName, err)
            }
        } else {
            client, err = registry.Global.GetDefaultClient()
            if err != nil {
                return nil, fmt.Errorf("no client available for agent: %w", err)
            }
        }
    }

    messages := []chat.Message{
        {Role: "system", Content: p.buildSystemMessage()},
        {Role: "user", Content: prompt},
    }

    response, err := client.Chat(messages, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate plan: %w", err)
    }

    // 解析LLM响应为执行计划
    plan, err := p.parsePlanFromResponse(response.Choices[0].Message.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse plan: %w", err)
    }

    return plan, nil
}

func (p *AgentPlanner) getAvailableFlows() (map[string]string, error) {
    availableFlows := make(map[string]string)

    for _, flowName := range p.agent.AllowedFlows {
        if flowName == "*" {
            // 如果允许所有Flow，获取所有已注册的Flow
            allFlows := registry.Global.ListFlows()
            for name, flow := range allFlows {
                availableFlows[name] = p.getFlowDescription(flow)
            }
            break
        } else {
            flow, err := registry.Global.GetFlow(flowName)
            if err != nil {
                continue // 跳过不存在的Flow
            }
            availableFlows[flowName] = p.getFlowDescription(flow)
        }
    }

    return availableFlows, nil
}

func (p *AgentPlanner) getFlowDescription(flow *flow.Flow) string {
    // 从Flow的步骤中推断功能描述
    if len(flow.Steps) == 0 {
        return "No description available"
    }

    var descriptions []string
    for _, step := range flow.Steps {
        if step.Name != "" {
            descriptions = append(descriptions, step.Name)
        }
    }

    if len(descriptions) > 0 {
        return strings.Join(descriptions, " -> ")
    }

    return fmt.Sprintf("Flow with %d steps", len(flow.Steps))
}

func (p *AgentPlanner) buildSystemMessage() string {
    return fmt.Sprintf(`You are %s, an AI agent with the following characteristics:

Role: %s
Backstory: %s
Goal: %s

Your task is to create an execution plan using available workflows. You must analyze the user's task and break it down into steps that can be executed using the available workflows.

Response format must be valid JSON with the following structure:
{
    "tasks": [
        {
            "stepId": "step_1",
            "flowName": "workflow_name",
            "description": "What this step accomplishes",
            "inputs": {"key": "value"},
            "dependencies": []
        }
    ]
}

Important guidelines:
1. Only use workflows from the available list
2. Break complex tasks into smaller, manageable steps
3. Ensure logical dependencies between steps
4. Provide clear descriptions for each step
5. Keep the plan focused and efficient`,
        p.agent.Name, p.agent.Role, p.agent.Backstory, p.agent.Goal)
}

func (p *AgentPlanner) buildPlanningPrompt(taskDescription string, availableFlows map[string]string, context flow.FlowContext) string {
    flowsList := make([]string, 0, len(availableFlows))
    for name, desc := range availableFlows {
        flowsList = append(flowsList, fmt.Sprintf("- %s: %s", name, desc))
    }

    prompt := fmt.Sprintf(`Task: %s

Available Workflows:
%s`, taskDescription, strings.Join(flowsList, "\n"))

    if context.Text != "" {
        prompt += fmt.Sprintf("\n\nContext: %s", context.Text)
    }

    if len(context.Variables) > 0 {
        prompt += "\n\nAvailable Variables:"
        for k, v := range context.Variables {
            prompt += fmt.Sprintf("\n- %s: %v", k, v)
        }
    }

    prompt += "\n\nPlease create an execution plan to accomplish this task."

    return prompt
}

func (p *AgentPlanner) parsePlanFromResponse(response string) (*ExecutionPlan, error) {
    // 尝试提取JSON部分
    response = strings.TrimSpace(response)

    // 查找JSON开始和结束
    start := strings.Index(response, "{")
    end := strings.LastIndex(response, "}")

    if start == -1 || end == -1 || start >= end {
        return nil, fmt.Errorf("no valid JSON found in response")
    }

    jsonStr := response[start : end+1]

    var plan ExecutionPlan
    if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
        return nil, fmt.Errorf("failed to unmarshal plan JSON: %w", err)
    }

    // 验证计划的有效性
    if err := p.validatePlan(&plan); err != nil {
        return nil, fmt.Errorf("invalid plan: %w", err)
    }

    return &plan, nil
}

func (p *AgentPlanner) validatePlan(plan *ExecutionPlan) error {
    if len(plan.Tasks) == 0 {
        return fmt.Errorf("plan must contain at least one task")
    }

    // 验证所有Flow都在允许列表中
    for _, task := range plan.Tasks {
        if task.FlowName == "" {
            return fmt.Errorf("task %s missing flow name", task.StepID)
        }

        allowed := false
        for _, allowedFlow := range p.agent.AllowedFlows {
            if allowedFlow == "*" || allowedFlow == task.FlowName {
                allowed = true
                break
            }
        }

        if !allowed {
            return fmt.Errorf("flow %s not allowed for agent %s", task.FlowName, p.agent.Name)
        }

        // 验证Flow是否存在
        _, err := registry.Global.GetFlow(task.FlowName)
        if err != nil {
            return fmt.Errorf("flow %s not found: %w", task.FlowName, err)
        }
    }

    return nil
}
```

#### 3.2.5 Crew 团队管理

```go
// agent/crew.go
package agent

import (
    "context"
    "fmt"
    "sync"
    "github.com/jieliu2000/anyi/flow"
)

type Crew struct {
    Name        string             `json:"name" yaml:"name"`
    Description string             `json:"description" yaml:"description"`
    Agents      []*Agent          `json:"agents" yaml:"agents"`
    Process     ProcessType       `json:"process" yaml:"process"`
    MaxRounds   int               `json:"maxRounds" yaml:"maxRounds"`
    Variables   map[string]any    `json:"variables" yaml:"variables"`
}

type ProcessType string

const (
    ProcessSequential ProcessType = "sequential"
    ProcessHierarchical ProcessType = "hierarchical"
    ProcessConsensus ProcessType = "consensus"
)

type CrewExecutor struct {
    CrewName string `json:"crewName" yaml:"crewName"`
    TaskDescription string `json:"taskDescription" yaml:"taskDescription"`
    MaxRounds int `json:"maxRounds" yaml:"maxRounds"`
}

func (e *CrewExecutor) Init() error {
    if e.CrewName == "" {
        return fmt.Errorf("crew name is required")
    }

    _, err := registry.Global.GetCrew(e.CrewName)
    return err
}

func (e *CrewExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
    crew, err := registry.Global.GetCrew(e.CrewName)
    if err != nil {
        return &flowContext, err
    }

    maxRounds := e.MaxRounds
    if maxRounds <= 0 {
        maxRounds = crew.MaxRounds
        if maxRounds <= 0 {
            maxRounds = 3 // 默认最大轮次
        }
    }

    result, err := e.executeCrew(crew, e.TaskDescription, flowContext, maxRounds)
    if err != nil {
        return &flowContext, err
    }

    flowContext.Text = result
    return &flowContext, nil
}

func (e *CrewExecutor) executeCrew(crew *Crew, taskDescription string, context flow.FlowContext, maxRounds int) (string, error) {
    switch crew.Process {
    case ProcessSequential:
        return e.executeSequential(crew, taskDescription, context)
    case ProcessHierarchical:
        return e.executeHierarchical(crew, taskDescription, context, maxRounds)
    case ProcessConsensus:
        return e.executeConsensus(crew, taskDescription, context, maxRounds)
    default:
        return e.executeSequential(crew, taskDescription, context)
    }
}

func (e *CrewExecutor) executeSequential(crew *Crew, taskDescription string, context flow.FlowContext) (string, error) {
    currentContext := context
    var results []string

    for i, agent := range crew.Agents {
        agentExecutor := &AgentExecutor{
            AgentName: agent.Name,
            TaskDescription: fmt.Sprintf("Step %d: %s", i+1, taskDescription),
            Context: currentContext.Text,
        }

        result, err := agentExecutor.Run(currentContext, nil)
        if err != nil {
            return "", fmt.Errorf("agent %s failed: %w", agent.Name, err)
        }

        results = append(results, result.Text)
        currentContext = *result
    }

    return e.combineResults(results), nil
}

func (e *CrewExecutor) executeHierarchical(crew *Crew, taskDescription string, context flow.FlowContext, maxRounds int) (string, error) {
    if len(crew.Agents) == 0 {
        return "", fmt.Errorf("no agents in crew")
    }

    // 第一个Agent作为主管
    supervisor := crew.Agents[0]
    workers := crew.Agents[1:]

    currentContext := context
    var finalResult string

    for round := 0; round < maxRounds; round++ {
        // 主管分配任务
        supervisorExecutor := &AgentExecutor{
            AgentName: supervisor.Name,
            TaskDescription: fmt.Sprintf("Round %d - Plan and coordinate: %s", round+1, taskDescription),
            Context: currentContext.Text,
        }

        planResult, err := supervisorExecutor.Run(currentContext, nil)
        if err != nil {
            return "", fmt.Errorf("supervisor failed: %w", err)
        }

        // 工作Agent并行执行
        var wg sync.WaitGroup
        workerResults := make([]string, len(workers))
        workerErrors := make([]error, len(workers))

        for i, worker := range workers {
            wg.Add(1)
            go func(idx int, agent *Agent) {
                defer wg.Done()

                workerExecutor := &AgentExecutor{
                    AgentName: agent.Name,
                    TaskDescription: fmt.Sprintf("Execute assigned task: %s", planResult.Text),
                    Context: currentContext.Text,
                }

                result, err := workerExecutor.Run(currentContext, nil)
                if err != nil {
                    workerErrors[idx] = err
                    return
                }

                workerResults[idx] = result.Text
            }(i, worker)
        }

        wg.Wait()

        // 检查工作Agent错误
        for i, err := range workerErrors {
            if err != nil {
                return "", fmt.Errorf("worker %s failed: %w", workers[i].Name, err)
            }
        }

        // 主管整合结果
        consolidatedInput := fmt.Sprintf("Task: %s\n\nWorker Results:\n%s",
            taskDescription, strings.Join(workerResults, "\n---\n"))

        consolidateExecutor := &AgentExecutor{
            AgentName: supervisor.Name,
            TaskDescription: "Consolidate and finalize results",
            Context: consolidatedInput,
        }

        finalResult, err := consolidateExecutor.Run(flow.FlowContext{Text: consolidatedInput}, nil)
        if err != nil {
            return "", fmt.Errorf("consolidation failed: %w", err)
        }

        currentContext = *finalResult

        // 如果主管认为任务完成，提前结束
        if e.isTaskComplete(finalResult.Text) {
            break
        }
    }

    return finalResult, nil
}

func (e *CrewExecutor) executeConsensus(crew *Crew, taskDescription string, context flow.FlowContext, maxRounds int) (string, error) {
    var proposals []string

    for round := 0; round < maxRounds; round++ {
        // 所有Agent提出方案
        var wg sync.WaitGroup
        roundProposals := make([]string, len(crew.Agents))
        roundErrors := make([]error, len(crew.Agents))

        for i, agent := range crew.Agents {
            wg.Add(1)
            go func(idx int, a *Agent) {
                defer wg.Done()

                contextText := context.Text
                if len(proposals) > 0 {
                    contextText += fmt.Sprintf("\n\nPrevious proposals:\n%s",
                        strings.Join(proposals, "\n---\n"))
                }

                agentExecutor := &AgentExecutor{
                    AgentName: a.Name,
                    TaskDescription: fmt.Sprintf("Propose solution for: %s", taskDescription),
                    Context: contextText,
                }

                result, err := agentExecutor.Run(flow.FlowContext{Text: contextText}, nil)
                if err != nil {
                    roundErrors[idx] = err
                    return
                }

                roundProposals[idx] = result.Text
            }(i, agent)
        }

        wg.Wait()

        // 检查错误
        for i, err := range roundErrors {
            if err != nil {
                return "", fmt.Errorf("agent %s failed in consensus round %d: %w",
                    crew.Agents[i].Name, round+1, err)
            }
        }

        proposals = append(proposals, roundProposals...)

        // 检查是否达成共识
        if consensus := e.checkConsensus(roundProposals); consensus != "" {
            return consensus, nil
        }
    }

    // 如果没有达成共识，返回最后一轮的合并结果
    return e.combineResults(proposals[len(proposals)-len(crew.Agents):]), nil
}

func (e *CrewExecutor) isTaskComplete(result string) bool {
    // 简单的完成检查，可以根据需要扩展
    completeMarkers := []string{"COMPLETE", "FINISHED", "DONE", "完成"}
    resultUpper := strings.ToUpper(result)

    for _, marker := range completeMarkers {
        if strings.Contains(resultUpper, marker) {
            return true
        }
    }

    return false
}

func (e *CrewExecutor) checkConsensus(proposals []string) string {
    // 简单的共识检查：如果所有提案足够相似，认为达成共识
    if len(proposals) <= 1 {
        return ""
    }

    // 这里可以实现更复杂的相似度检查
    // 现在简单地检查是否有重复的关键词

    return "" // 暂时返回空，表示未达成共识
}

func (e *CrewExecutor) combineResults(results []string) string {
    if len(results) == 0 {
        return ""
    }
    if len(results) == 1 {
        return results[0]
    }

    combined := "Combined Results:\n\n"
    for i, result := range results {
        combined += fmt.Sprintf("Result %d:\n%s\n\n", i+1, result)
    }
    return combined
}
```

### 3.3 向后兼容性支持

在根包中保留原有的 API，并提供新的 Agent 相关 API：

```go
// anyi.go - 添加Agent支持的API
package anyi

import (
    "github.com/jieliu2000/anyi/agent"
    "github.com/jieliu2000/anyi/registry"
)

// Agent相关的便民函数
func RegisterAgent(name string, role, backstory, goal string, allowedFlows []string, clientName string) (*agent.Agent, error) {
    agent := &agent.Agent{
        Name:         name,
        Role:         role,
        Backstory:    backstory,
        Goal:         goal,
        AllowedFlows: allowedFlows,
        ClientName:   clientName,
        MaxIterations: 10,
        Variables:    make(map[string]any),
    }

    err := registry.Global.RegisterAgent(name, agent)
    if err != nil {
        return nil, err
    }

    return agent, nil
}

func GetAgent(name string) (*agent.Agent, error) {
    return registry.Global.GetAgent(name)
}

func RegisterCrew(name string, agents []string, process agent.ProcessType) (*agent.Crew, error) {
    agentList := make([]*agent.Agent, len(agents))
    for i, agentName := range agents {
        a, err := registry.Global.GetAgent(agentName)
        if err != nil {
            return nil, fmt.Errorf("agent %s not found: %w", agentName, err)
        }
        agentList[i] = a
    }

    crew := &agent.Crew{
        Name:        name,
        Agents:      agentList,
        Process:     process,
        MaxRounds:   3,
        Variables:   make(map[string]any),
    }

    err := registry.Global.RegisterCrew(name, crew)
    if err != nil {
        return nil, err
    }

    return crew, nil
}

func NewAgentStep(agentName, taskDescription string) (*flow.Step, error) {
    executor := &agent.AgentExecutor{
        AgentName:       agentName,
        TaskDescription: taskDescription,
    }

    return flow.NewStep(executor, nil, nil), nil
}

func NewCrewStep(crewName, taskDescription string) (*flow.Step, error) {
    executor := &agent.CrewExecutor{
        CrewName:        crewName,
        TaskDescription: taskDescription,
    }

    return flow.NewStep(executor, nil, nil), nil
}

// 保持现有的GlobalRegistry兼容性
var GlobalRegistry = registry.Global

// 迁移现有的注册表函数
func RegisterClient(name string, client llm.Client) error {
    return registry.Global.RegisterClient(name, client)
}

func GetClient(name string) (llm.Client, error) {
    return registry.Global.GetClient(name)
}

func RegisterFlow(name string, flow *flow.Flow) error {
    return registry.Global.RegisterFlow(name, flow)
}

func GetFlow(name string) (*flow.Flow, error) {
    return registry.Global.GetFlow(name)
}

// 其他现有函数保持不变...
```

### 3.4 配置支持

扩展现有的配置系统支持 Agent：

```go
// config.go - 添加Agent配置支持
type AgentConfig struct {
    Name          string   `mapstructure:"name" json:"name" yaml:"name"`
    Role          string   `mapstructure:"role" json:"role" yaml:"role"`
    Backstory     string   `mapstructure:"backstory" json:"backstory" yaml:"backstory"`
    Goal          string   `mapstructure:"goal" json:"goal" yaml:"goal"`
    AllowedFlows  []string `mapstructure:"allowedFlows" json:"allowedFlows" yaml:"allowedFlows"`
    ClientName    string   `mapstructure:"clientName" json:"clientName" yaml:"clientName"`
    MaxIterations int      `mapstructure:"maxIterations" json:"maxIterations" yaml:"maxIterations"`
    Variables     map[string]any `mapstructure:"variables" json:"variables" yaml:"variables"`
}

type CrewConfig struct {
    Name        string                `mapstructure:"name" json:"name" yaml:"name"`
    Description string                `mapstructure:"description" json:"description" yaml:"description"`
    Agents      []string              `mapstructure:"agents" json:"agents" yaml:"agents"`
    Process     agent.ProcessType     `mapstructure:"process" json:"process" yaml:"process"`
    MaxRounds   int                   `mapstructure:"maxRounds" json:"maxRounds" yaml:"maxRounds"`
    Variables   map[string]any        `mapstructure:"variables" json:"variables" yaml:"variables"`
}

// 扩展AnyiConfig
type AnyiConfig struct {
    Clients    []llm.ClientConfig    `mapstructure:"clients" json:"clients" yaml:"clients"`
    Flows      []FlowConfig          `mapstructure:"flows" json:"flows" yaml:"flows"`
    Formatters []FormatterConfig     `mapstructure:"formatters" json:"formatters" yaml:"formatters"`
    Agents     []AgentConfig         `mapstructure:"agents" json:"agents" yaml:"agents"`
    Crews      []CrewConfig          `mapstructure:"crews" json:"crews" yaml:"crews"`
}
```

## 4. 使用示例

### 4.1 基本 Agent 使用

```go
package main

import (
    "github.com/jieliu2000/anyi"
    "github.com/jieliu2000/anyi/agent"
)

func main() {
    // 初始化Anyi
    anyi.Init()

    // 创建LLM客户端
    client, _ := anyi.NewClientFromConfigFile("", "config.yaml")

    // 创建一些工作流
    anyi.NewFlow("data_analysis", client,
        anyi.NewLLMStep("分析提供的数据: {{.Text}}", "你是数据分析专家", client))

    anyi.NewFlow("report_generation", client,
        anyi.NewLLMStep("基于分析结果生成报告: {{.Text}}", "你是报告写作专家", client))

    // 注册Agent
    analyst, _ := anyi.RegisterAgent(
        "data_analyst",
        "数据分析师",
        "你是一名经验丰富的数据分析师，擅长从复杂数据中发现洞察",
        "准确分析数据并提供有价值的见解",
        []string{"data_analysis", "report_generation"},
        "default",
    )

    // 创建Agent执行步骤
    agentStep, _ := anyi.NewAgentStep("data_analyst", "分析销售数据并生成月度报告")

    // 创建包含Agent的工作流
    agentFlow, _ := anyi.NewFlow("agent_workflow", client, *agentStep)

    // 执行
    result, _ := agentFlow.RunWithInput("这是本月的销售数据...")
    fmt.Println(result.Text)
}
```

### 4.2 Crew 团队协作

```go
func main() {
    anyi.Init()

    // 注册多个Agent
    anyi.RegisterAgent("researcher", "研究员", "专门负责信息收集和研究", "收集准确全面的信息", []string{"*"}, "default")
    anyi.RegisterAgent("writer", "撰写员", "专门负责内容创作和编辑", "创作高质量的内容", []string{"*"}, "default")
    anyi.RegisterAgent("reviewer", "审核员", "专门负责质量检查和改进建议", "确保内容质量和准确性", []string{"*"}, "default")

    // 创建Crew团队
    crew, _ := anyi.RegisterCrew("content_team",
        []string{"researcher", "writer", "reviewer"},
        agent.ProcessSequential)

    // 创建Crew执行步骤
    crewStep, _ := anyi.NewCrewStep("content_team", "创建关于AI发展趋势的深度分析文章")

    // 执行团队任务
    crewFlow, _ := anyi.NewFlow("crew_workflow", client, *crewStep)
    result, _ := crewFlow.RunWithInput("请创建一篇关于2024年AI发展趋势的文章")

    fmt.Println(result.Text)
}
```

### 4.3 配置文件方式

```yaml
# config.yaml
clients:
  - name: "gpt4"
    type: "openai"
    default: true
    apiKey: "your-api-key"
    model: "gpt-4"

flows:
  - name: "research"
    clientName: "gpt4"
    steps:
      - executor:
          type: "llm"
          withconfig:
            template: "Research the topic: {{.Text}}"
            systemMessage: "You are a research specialist"

  - name: "writing"
    clientName: "gpt4"
    steps:
      - executor:
          type: "llm"
          withconfig:
            template: "Write content based on: {{.Text}}"
            systemMessage: "You are a content writer"

agents:
  - name: "researcher"
    role: "Research Specialist"
    backstory: "Expert in information gathering and analysis"
    goal: "Provide comprehensive and accurate research"
    allowedFlows: ["research"]
    clientName: "gpt4"
    maxIterations: 5

  - name: "writer"
    role: "Content Writer"
    backstory: "Skilled in creating engaging and informative content"
    goal: "Create high-quality written content"
    allowedFlows: ["writing"]
    clientName: "gpt4"
    maxIterations: 3

crews:
  - name: "content_crew"
    description: "Team for content creation"
    agents: ["researcher", "writer"]
    process: "sequential"
    maxRounds: 3
```

## 5. 实施计划

### 5.1 第一阶段：核心架构

1. 创建`registry`包，迁移现有注册表功能
2. 创建`agent`包基础结构
3. 实现基本的 Agent 定义和注册
4. 更新根包，保持向后兼容性

### 5.2 第二阶段：执行引擎

1. 实现 AgentExecutor
2. 实现 AgentPlanner
3. 集成到现有的执行器系统
4. 添加配置支持

### 5.3 第三阶段：高级功能

1. 实现 Crew 团队管理
2. 添加多种协作模式（Sequential, Hierarchical, Consensus）
3. 优化任务规划算法
4. 添加监控和日志功能

### 5.4 第四阶段：完善和优化

1. 性能优化
2. 错误处理改进
3. 文档完善
4. 示例和测试用例

## 6. 优势特点

1. **无循环引用**: 通过独立的 registry 包解决依赖问题
2. **保持简洁**: 代码结构清晰，易于理解和维护
3. **向后兼容**: 现有代码无需修改即可继续使用
4. **灵活扩展**: 支持多种 Agent 协作模式
5. **配置驱动**: 支持通过配置文件定义 Agent 和 Crew
6. **类型安全**: 充分利用 Go 的类型系统确保安全性

这个设计充分利用了 Anyi 现有的优秀架构，在不破坏现有功能的前提下，添加了强大的 Agent 系统支持，为构建复杂的 AI 应用提供了坚实的基础。
