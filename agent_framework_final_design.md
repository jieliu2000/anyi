# Anyi Agent Framework 最终设计

## 设计理念

创建一个类似CrewAI的Agent框架，保持极简性、功能完整性和无循环引用问题。

## 核心架构

### 包结构
```
anyi/
├── agent/           # Agent核心模块
│   ├── agent.go     # Agent接口和基础实现
│   ├── planner.go   # 规划器组件
│   └── executor.go  # 执行器组件
├── registry/        # 独立注册表模块
│   └── registry.go  # 全局注册表实现
├── flow/            # 现有Flow模块（保持不变）
└── llm/             # 现有LLM模块（保持不变）
```

## 核心接口设计

### Agent接口（极简但完整）
```go
// agent/agent.go
package agent

// AgentContext 执行上下文
type AgentContext struct {
    Variables map[string]interface{}
    Memory    interface{}
    History   []string
}

// FlowGetter 依赖接口 - 解决循环引用
type FlowGetter interface {
    GetFlow(name string) (interface{}, error)
}
```

## 基础实现

### BasicAgent实现
```go
// agent/agent.go
type Agent struct {
    Role          string
    BackStory     string
    AvailableFlows []string  // 使用anyi标准命名
    Config        Config
    getFlow       FlowGetter // 依赖注入
}

func NewAgent(role, backstory string, availableFlows []string, getFlow FlowGetter) *Agent {
    return &Agent{
        Role:          role,
        BackStory:     backstory,
        AvailableFlows: availableFlows,
        getFlow:       getFlow,
        Config:        DefaultConfig(),
    }
}

func (a *BasicAgent) Execute(task string, ctx *AgentContext) (string, *AgentContext, error) {
    if ctx == nil {
        ctx = &AgentContext{Variables: make(map[string]interface{})}
    }
    
    // 1. 智能规划
    plan := a.planExecution(task, ctx)
    
    // 2. 执行计划
    result := task
    for _, step := range plan.Steps {
        flow, err := a.getFlow.GetFlow(step.Tool)
        if err != nil {
            return "", ctx, err
        }
        
        // 执行Flow
        if executable, ok := flow.(interface {
            Execute(input string, ctx map[string]interface{}) (string, map[string]interface{}, error)
        }); ok {
            result, ctx.Variables, err = executable.Execute(result, ctx.Variables)
            if err != nil {
                return "", ctx, err
            }
        }
        
        ctx.History = append(ctx.History, result)
    }
    
    return result, ctx, nil
}
```

## 注册表集成（无循环引用）

```go
// registry/registry.go
package registry

import (
    "sync"
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
    Agents:     make(map[string]agent.Agent),
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
researchAgent := agent.NewBasicAgent(
    "Research Assistant",
    "Expert at information analysis", 
    []string{"web_search", "data_analysis", "report_writer"},
    registry.Global, // 注入registry实现FlowGetter
)

// 注册Agent
registry.RegisterAgent("researcher", researchAgent)

// 执行任务
result, newCtx, err := researchAgent.Execute(
    "Research AI applications in healthcare",
    &agent.AgentContext{
        Variables: map[string]interface{}{"depth": "detailed", "sources": 10},
    },
)
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
    
    result, newCtx, err := agent.Execute(s.Task, &agent.AgentContext{
        Variables: ctx.Variables,
    })
    
    ctx.Text = result
    ctx.Variables = newCtx.Variables
    return ctx, nil
}
```

## 设计优势

### ✅ 无循环引用
- 通过接口隔离和依赖注入彻底解决
- `agent`包不直接导入`registry`
- `registry`只依赖`agent`接口

### ✅ 极简性
- 对外接口简单直观
- 使用体验类似CrewAI
- 学习成本低

### ✅ 功能完整
- 智能规划能力
- 完整的上下文管理
- 错误处理和重试机制

### ✅ 易于扩展
- 支持不同的Agent实现
- 可配置的规划策略
- 无缝集成现有Flow体系

## 迁移计划

1. **第一阶段**：创建`agent`和`registry`模块
2. **第二阶段**：实现基础Agent和规划器
3. **第三阶段**：集成测试和文档
4. **第四阶段**：逐步替换现有代码

这个设计为anyi提供了强大的Agent能力，同时保持了架构的清晰和可维护性。