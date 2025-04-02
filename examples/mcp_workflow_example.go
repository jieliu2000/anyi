package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/mcp"
)

// 定义工作流程的内存结构
type TaskAnalysis struct {
	UserQuery   string
	Analysis    string
	ActionPlan  []string
	Completed   bool
	CurrentStep int
}

func main() {
	// 初始化anyi
	anyi.Init()

	// 创建MCP客户端
	config := mcp.DefaultConfigWithModel(os.Getenv("MCP_API_KEY"), os.Getenv("MCP_MODEL_ID"))
	config.Endpoint = os.Getenv("MCP_ENDPOINT")
	if config.Endpoint == "" {
		config.Endpoint = "https://example.com/api/mcp"
		log.Println("警告: 使用默认MCP端点，在生产环境中请设置MCP_ENDPOINT环境变量")
	}

	// 注册客户端
	mcpClient, err := anyi.NewClient("mcp", config)
	if err != nil {
		log.Fatalf("创建MCP客户端失败: %v", err)
	}

	// 创建用于分析任务的模板
	taskAnalysisTemplate := `
你是一个任务分析助手。分析以下用户查询，并提供详细的分析和行动计划。

用户查询: {{.Memory.UserQuery}}

请提供:
1. 对查询的详细分析
2. 解决查询所需的行动步骤列表
`

	// 创建一个步骤处理器
	taskAnalyzerExecutor, err := anyi.NewPromptTemplateFormatter("task-analyzer", taskAnalysisTemplate)
	if err != nil {
		log.Fatalf("创建模板格式化器失败: %v", err)
	}

	// 创建用于提取结构化信息的模板
	extractInfoTemplate := `
将以下分析转换为结构化格式:

分析: {{.Text}}

提取以下信息:
1. 分析摘要
2. 行动步骤列表

以JSON格式返回这些信息，使用以下结构:
{
  "analysis": "分析摘要",
  "actionPlan": ["步骤1", "步骤2", ...]
}
`

	extractorExecutor, err := anyi.NewPromptTemplateFormatter("info-extractor", extractInfoTemplate)
	if err != nil {
		log.Fatalf("创建信息提取模板失败: %v", err)
	}

	// 创建用于执行步骤的模板
	executeStepTemplate := `
你正在执行以下任务的步骤:

任务分析: {{.Memory.Analysis}}

当前正在执行步骤 {{.Memory.CurrentStep}} (共{{len .Memory.ActionPlan}}步):
{{index .Memory.ActionPlan .Memory.CurrentStep}}

请执行此步骤并详细报告结果。
`

	stepExecutorFormatter, err := anyi.NewPromptTemplateFormatter("step-executor", executeStepTemplate)
	if err != nil {
		log.Fatalf("创建步骤执行模板失败: %v", err)
	}

	// 创建用于总结的模板
	summaryTemplate := `
以下是任务的总结:

用户查询: {{.Memory.UserQuery}}
分析: {{.Memory.Analysis}}

执行的步骤:
{{range $index, $step := .Memory.ActionPlan}}
{{$index}}. {{$step}}
{{end}}

请提供简洁的执行总结。
`

	summaryFormatter, err := anyi.NewPromptTemplateFormatter("summary", summaryTemplate)
	if err != nil {
		log.Fatalf("创建总结模板失败: %v", err)
	}

	// 注册LLM执行器
	analyzer := anyi.NewLLMStepExecutorWithFormatter("analyzer", taskAnalyzerExecutor, "你是一个专业任务分析师，提供详细的任务分析和行动计划", mcpClient)
	extractor := anyi.NewLLMStepExecutorWithFormatter("extractor", extractorExecutor, "你是一个数据提取专家，将文本内容转换为结构化数据", mcpClient)
	stepExecutor := anyi.NewLLMStepExecutorWithFormatter("step-executor", stepExecutorFormatter, "你是一个执行专家，执行给定的任务步骤", mcpClient)
	summarizer := anyi.NewLLMStepExecutorWithFormatter("summarizer", summaryFormatter, "你是一个专业总结师，提供清晰简洁的总结", mcpClient)

	// 创建JSON验证器
	jsonValidator := &anyi.JsonValidator{}
	err = jsonValidator.Init()
	if err != nil {
		log.Fatalf("初始化JSON验证器失败: %v", err)
	}

	// 创建自定义执行器来解析JSON并更新流程上下文
	updateContextExecutor := &UpdateTaskContextExecutor{}

	// 创建执行步骤的执行器
	executeTaskStepExecutor := &ExecuteTaskStepExecutor{}

	// 注册自定义执行器
	anyi.RegisterExecutor("updateTaskContext", updateContextExecutor)
	anyi.RegisterExecutor("executeTaskStep", executeTaskStepExecutor)

	// 创建工作流程步骤
	steps := []flow.Step{
		{
			Name:     "分析任务",
			Executor: analyzer,
		},
		{
			Name:      "提取结构化信息",
			Executor:  extractor,
			Validator: jsonValidator,
		},
		{
			Name:     "更新任务上下文",
			Executor: updateContextExecutor,
		},
		{
			Name:     "执行步骤",
			Executor: executeTaskStepExecutor,
		},
		{
			Name:     "执行任务步骤",
			Executor: stepExecutor,
		},
		{
			Name:     "总结任务",
			Executor: summarizer,
		},
	}

	// 创建工作流程
	taskFlow, err := anyi.NewFlow("mcp-task-flow", mcpClient, steps...)
	if err != nil {
		log.Fatalf("创建工作流程失败: %v", err)
	}

	// 创建任务
	task := TaskAnalysis{
		UserQuery: "帮我创建一个简单的待办事项应用的设计",
		Completed: false,
	}

	// 创建流程上下文
	flowContext := flow.FlowContext{
		Text:   "",
		Memory: task,
	}

	// 运行工作流程
	resultContext, err := taskFlow.Run(flowContext)
	if err != nil {
		log.Fatalf("执行工作流程失败: %v", err)
	}

	// 展示结果
	fmt.Println("工作流程执行结果:")
	fmt.Println("----------------")
	fmt.Println(resultContext.Text)
}

// UpdateTaskContextExecutor 解析JSON并更新任务上下文
type UpdateTaskContextExecutor struct{}

func (e *UpdateTaskContextExecutor) Init() error {
	return nil
}

func (e *UpdateTaskContextExecutor) Run(context flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// 解析前一步的JSON输出
	var result struct {
		Analysis   string   `json:"analysis"`
		ActionPlan []string `json:"actionPlan"`
	}

	err := json.Unmarshal([]byte(context.Text), &result)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 获取并更新任务数据
	taskData, ok := context.Memory.(TaskAnalysis)
	if !ok {
		return nil, fmt.Errorf("任务数据类型不正确")
	}

	// 更新任务数据
	taskData.Analysis = result.Analysis
	taskData.ActionPlan = result.ActionPlan
	taskData.CurrentStep = 0

	// 更新上下文内存
	context.Memory = taskData
	context.Text = fmt.Sprintf("更新了任务上下文。分析: %s, 步骤数: %d", taskData.Analysis, len(taskData.ActionPlan))

	return &context, nil
}

// ExecuteTaskStepExecutor 检查是否有更多步骤要执行，如果没有则设置为已完成
type ExecuteTaskStepExecutor struct{}

func (e *ExecuteTaskStepExecutor) Init() error {
	return nil
}

// finishedStepIndex 用于记录完成步骤的索引值
var finishedStepIndex = 5 // 总结任务的索引值

func (e *ExecuteTaskStepExecutor) Run(context flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// 获取任务数据
	taskData, ok := context.Memory.(TaskAnalysis)
	if !ok {
		return nil, fmt.Errorf("任务数据类型不正确")
	}

	// 检查是否已完成所有步骤
	if taskData.CurrentStep >= len(taskData.ActionPlan) {
		taskData.Completed = true
		context.Memory = taskData

		// 将文本设置为完成信息，实际流程控制需要在工作流配置中设置
		context.Text = "所有步骤已完成，跳转到总结步骤"

		return &context, nil
	}

	// 继续执行下一步
	context.Text = fmt.Sprintf("执行步骤 %d: %s", taskData.CurrentStep+1, taskData.ActionPlan[taskData.CurrentStep])

	return &context, nil
}
