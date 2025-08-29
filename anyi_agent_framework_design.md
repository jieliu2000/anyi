# Anyi Agent Framework 最终设计

## 设计理念

创建一个类似CrewAI的Agent框架，保持极简性、功能完整性和无循环引用问题，使用值类型的AgentContext确保安全性。

## 核心架构

### 包结构
```
anyi/
├── agent/           # Agent核心模块
│   ├── agent.go     # Agent类型和实现
│   ├── planner.go   # 规划器组件
│   └── context.go   # 上下文定义
├── registry/        # 独立注册表模块
│   └── registry.go  # 全局注册表实现
├── flow/            # 现有Flow模块（保持不变）
└── llm/             # 现有LLM模块（保持不变）
```

## 核心类型设计

### AgentContext值类型
```go
// agent/context.go
package agent

// AgentContext 执行上下文 - 使用值类型确保安全性
type AgentContext struct {
    Variables map[string]interface{}
    Memory    interface{}
    History   []string
}
```

### Agent具体类型
```go
// agent/agent.go
package agent

// FlowGetter 依赖接口 - 解决循环引用
type FlowGetter interface {
    GetFlow(name string) (interface{}, error)
}

// Agent 具体类型 - 不需要接口
type Agent struct {
    Role          string
    BackStory     string
    AvailableFlows []string  // 可用的Flow名称列表
    Config        Config
    getFlow       FlowGetter // 依赖注入
}

// NewAgent 创建新Agent
func NewAgent(role, backstory string, availableFlows []string, getFlow FlowGetter) *Agent {
    return &Agent{
        Role:          role,
        BackStory:     backstory,
        AvailableFlows: availableFlows,
        getFlow:       getFlow,
        Config:        DefaultConfig(),
    }
}

// Execute 执行任务 - 使用值类型的AgentContext
func (a *Agent) Execute(task string, ctx AgentContext) (string, AgentContext, error) {
    if ctx.Variables == nil {
        ctx.Variables = make(map[string]interface{})
    }
    
    // 使用副本进行操作，避免修改原始上下文
    resultCtx := ctx
    result := task
    
    // 1. 智能规划
    plan := a.planExecution(task, resultCtx)
    
    // 2. 执行计划
    for _, step := range plan.Steps {
        flow, err := a.getFlow.GetFlow(step.FlowName)
        if err != nil {
            return "", ctx, err
        }
        
        // 执行Flow
        if executable, ok := flow.(interface {
            Execute(input string, ctx map[string]interface{}) (string, map[string]interface{}, error)
        }); ok {
            result, resultCtx.Variables, err = executable.Execute(result, resultCtx.Variables)
            if err != nil {
                return "", ctx, err
            }
        }
        
        resultCtx.History = append(resultCtx.History, result)
        
        // 检查是否完成目标
        if a.isTaskCompleted(result, task) {
            break
        }
    }
    
    return result, resultCtx, nil
}
```

## 注册表集成

```go
// registry/registry.go
package registry

import (
    "sync"
    "fmt"
    "github.com/jieliu2000/anyi/flow"
    "github.com/jieliu2000/anyi/llm"
    "github.com/jieliu2000/anyi/agent"
)

// Registry 统一注册表
type Registry struct {
    mu         sync.RWMutex
    Clients    map[string]llm.Client
    Flows      map[string]*flow.Flow
    Agents     map[string]*agent.Agent // 使用具体类型指针
    Executors  map[string]flow.StepExecutor
    Validators map[string]flow.StepValidator
}

var Global = &Registry{
    Clients:    make(map[string]llm.Client),
    Flows:      make(map[string]*flow.Flow),
    Agents:     make(map[string]*agent.Agent),
    Executors:  make(map[string]flow.StepExecutor),
    Validators: make(map[string]flow.StepValidator),
}

// 实现agent.FlowGetter接口
func (r *Registry) GetFlow(name string) (interface{}, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    flow, exists := r.Flows[name]
    if !exists {
        return nil, fmt.Errorf("flow %s not found", name)
    }
    return flow, nil
}

// Agent注册函数
func RegisterAgent(name string, agent *agent.Agent) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.Agents[name]; exists {
        return fmt.Errorf("agent %s already exists", name)
    }
    
    r.Agents[name] = agent
    return nil
}

func GetAgent(name string) (*agent.Agent, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    agent, exists := r.Agents[name]
    if !exists {
        return nil, fmt.Errorf("agent %s not found", name)
    }
    return agent, nil
}
```

## 使用示例

```go
// 创建Agent（依赖注入解决循环引用）
researchAgent := agent.NewAgent(
    "Research Assistant",
    "Expert at information analysis and report generation",
    []string{"web_search", "data_analysis", "report_writer"},
    registry.Global, // 注入registry实现FlowGetter
)

// 注册Agent
if err := registry.RegisterAgent("researcher", researchAgent); err != nil {
    log.Fatal(err)
}

// 创建初始上下文
initialCtx := agent.AgentContext{
    Variables: map[string]interface{}{
        "depth":   "detailed",
        "sources": 10,
        "format":  "markdown",
    },
}

// 执行任务 - 使用值类型，安全无忧
result, updatedCtx, err := researchAgent.Execute(
    "Research AI applications in healthcare and write a comprehensive report",
    initialCtx, // 值传递，不会修改originalCtx
)

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Result: %s\n", result)
fmt.Printf("Updated variables: %+v\n", updatedCtx.Variables)

// 如果需要更新原始上下文，可以手动赋值
// initialCtx = updatedCtx
```

## Flow集成示例

```go
// 在Flow中使用Agent
type AgentStep struct {
    Agent string `mapstructure:"agent"`
    Task  string `mapstructure:"task"`
}

func (s *AgentStep) Execute(ctx flow.Context) (flow.Context, error) {
    agent, err := registry.GetAgent(s.Agent)
    if err != nil {
        return ctx, err
    }
    
    // 准备Agent上下文
    agentCtx := agent.AgentContext{
        Variables: ctx.Variables,
    }
    
    result, updatedAgentCtx, err := agent.Execute(s.Task, agentCtx)
    if err != nil {
        return ctx, err
    }
    
    // 更新Flow上下文
    ctx.Text = result
    ctx.Variables = updatedAgentCtx.Variables
    
    return ctx, nil
}
```

## 设计优势

### ✅ 无循环引用
- 通过`FlowGetter`接口依赖注入
- `agent`包不直接导入`registry`
- 编译安全，架构清晰

### ✅ 值类型安全性
- `AgentContext`使用值类型，避免意外修改
- 执行过程不会影响传入的上下文
- 线程安全，支持并发访问

### ✅ 极简性
- 不需要接口抽象层
- 代码直观易懂，学习成本低
- 使用体验类似CrewAI

### ✅ 功能完整
- 智能规划能力
- 完整的上下文管理
- 错误处理和重试机制

### ✅ 易于扩展
- 通过`AvailableFlows`配置支持各种Agent变体
- 依赖注入支持不同的Flow获取策略
- 易于添加新的规划算法

## 迁移计划

1. **第一阶段**：创建`agent`和`registry`模块
2. **第二阶段**：实现基础Agent和规划器
3. **第三阶段**：集成测试和文档
4. **第四阶段**：逐步替换现有代码

这个设计为anyi提供了强大的Agent能力，同时保持了架构的清晰性、安全性和可维护性。