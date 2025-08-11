# 组件参考

本文档提供 Anyi 框架中所有内置组件的全面参考，包括执行器、验证器和其他可重用组件。

## 概述

Anyi 组件是工作流的构建块。它们提供可以组合在一起创建复杂 AI 应用程序的特定功能。组件设计为：

- **可重用**：可以在不同工作流中使用
- **可配置**：接受参数来自定义行为
- **可组合**：可以组合创建复杂逻辑
- **可扩展**：可以扩展或替换为自定义实现

## 执行器

执行器是在工作流步骤内执行特定任务的组件。它们处理输入、执行操作并产生输出。

### LLMExecutor

最常用的执行器，向 LLM 提供商发送提示并处理响应。

#### 配置

```yaml
executor:
  type: "llm"
  withconfig:
    template: "分析以下文本：{{.Text}}"
    systemMessage: "你是一个专业分析师。"
    temperature: 0.7
    maxTokens: 1000
    extractThink: true
```

#### 配置选项

- `template`（字符串）：使用 Go 模板语法的提示模板
- `systemMessage`（字符串，可选）：LLM 的系统消息
- `temperature`（float64，可选）：温度覆盖（0.0-2.0）
- `maxTokens`（int，可选）：生成的最大令牌数
- `topP`（float64，可选）：Top-p 采样参数
- `frequencyPenalty`（float64，可选）：频率惩罚（-2.0 到 2.0）
- `presencePenalty`（float64，可选）：存在惩罚（-2.0 到 2.0）
- `stop`（[]string，可选）：停止序列
- `extractThink`（bool，可选）：从响应中提取 `<think>` 标签

#### 模板变量

模板可以从 FlowContext 访问以下变量：

- `{{.Text}}`：当前文本内容
- `{{.Memory}}`：结构化内存数据
- `{{.Memory.FieldName}}`：特定内存字段
- `{{.Think}}`：之前的思考过程
- `{{.Images}}`：图像 URL 数组

#### 模板函数

内置模板函数：

- `{{len .Text}}`：获取文本长度
- `{{contains .Text "substring"}}`：检查文本是否包含子字符串
- `{{upper .Text}}`：转换为大写
- `{{lower .Text}}`：转换为小写
- `{{trim .Text}}`：删除空白字符

#### 使用示例

```go
// 程序化使用
executor := &executors.LLMExecutor{
    Template: "用 3 个要点总结这个文本：\n\n{{.Text}}",
    SystemMessage: "你是一个专业的总结专家。",
    Temperature: 0.3,
    MaxTokens: 500,
}
```

### SetContextExecutor

直接修改流程上下文而不调用外部 API。

#### 配置

```yaml
executor:
  type: "setcontext"
  withconfig:
    text: "新的文本内容"
    memory:
      key1: "value1"
      key2: "value2"
    think: "初始思考过程"
    images:
      - "https://example.com/image1.jpg"
      - "https://example.com/image2.jpg"
```

#### 配置选项

- `text`（字符串，可选）：设置上下文文本
- `memory`（映射，可选）：设置结构化内存数据
- `think`（字符串，可选）：设置思考内容
- `images`（[]string，可选）：设置图像 URL
- `append`（bool，可选）：追加到现有内容而不是替换

#### 用例

- 初始化工作流上下文
- 为复杂工作流设置结构化数据
- 在工作流阶段之间重置上下文
- 为后续步骤准备数据

#### 使用示例

```go
// 使用结构化数据初始化工作流
executor := &executors.SetContextExecutor{
    Memory: map[string]interface{}{
        "task": "content_analysis",
        "requirements": []string{"准确性", "简洁性", "清晰性"},
        "status": "initialized",
    },
}
```

### ConditionalFlowExecutor

在工作流中基于条件启用分支逻辑。

#### 配置

```yaml
executor:
  type: "conditional"
  withconfig:
    condition: '{{contains .Text "错误"}}'
    trueFlow: "错误处理器"
    falseFlow: "正常处理器"
    trueSteps:
      - name: "记录错误"
        executor:
          type: "llm"
          withconfig:
            template: "记录这个错误：{{.Text}}"
    falseSteps:
      - name: "继续处理"
        executor:
          type: "llm"
          withconfig:
            template: "正常处理：{{.Text}}"
```

#### 配置选项

- `condition`（字符串）：计算为布尔值的 Go 模板表达式
- `trueFlow`（字符串，可选）：条件为真时执行的流程
- `falseFlow`（字符串，可选）：条件为假时执行的流程
- `trueSteps`（[]StepConfig，可选）：条件为真时执行的步骤
- `falseSteps`（[]StepConfig，可选）：条件为假时执行的步骤

#### 条件表达式

支持的条件模式：

- `{{eq .Memory.Status "complete"}}`：相等检查
- `{{ne .Text ""}}`：不等检查
- `{{gt (len .Text) 100}}`：大于比较
- `{{lt .Memory.Count 5}}`：小于比较
- `{{contains .Text "关键词"}}`：字符串包含检查
- `{{and (gt (len .Text) 50) (lt (len .Text) 500)}}`：逻辑 AND
- `{{or (eq .Memory.Type "urgent") (eq .Memory.Type "critical")}}`：逻辑 OR

### RunCommandExecutor

执行 shell 命令并捕获其输出。

#### 配置

```yaml
executor:
  type: "command"
  withconfig:
    command: "python"
    args:
      - "script.py"
      - "{{.Text}}"
    workingDir: "/path/to/scripts"
    timeout: 30
    captureOutput: true
    captureError: true
```

#### 配置选项

- `command`（字符串）：要执行的命令
- `args`（[]string，可选）：命令参数（支持模板）
- `workingDir`（字符串，可选）：工作目录
- `timeout`（int，可选）：超时秒数（默认：30）
- `captureOutput`（bool，可选）：捕获标准输出（默认：true）
- `captureError`（bool，可选）：捕获标准错误（默认：false）
- `env`（map[string]string，可选）：环境变量

#### 安全考虑

- 在命令中使用输入之前始终验证输入
- 对允许的命令使用白名单
- 以最小权限运行
- 考虑使用容器化进行隔离

#### 使用示例

```go
// 执行 Python 数据处理脚本
executor := &anyi.RunCommandExecutor{
    Command: "python",
    Args: []string{"process_data.py", "--input", "{{.Text}}"},
    WorkingDir: "/opt/scripts",
    Timeout: 60,
}
```

## 验证器

验证器确保执行器输出在进入下一步之前满足特定条件。

### StringValidator

基于各种字符串条件验证文本输出。

#### 配置

```yaml
validator:
  type: "string"
  withconfig:
    minLength: 100
    maxLength: 2000
    contains: "必需短语"
    notContains: "禁止词汇"
    matchRegex: "^[A-Z].*\\.$"
    notMatchRegex: "\\b(垃圾|欺诈)\\b"
    startsWith: "摘要："
    endsWith: "。"
```

#### 配置选项

- `minLength`（int，可选）：最小字符串长度
- `maxLength`（int，可选）：最大字符串长度
- `contains`（字符串，可选）：必需子字符串
- `notContains`（字符串，可选）：禁止子字符串
- `matchRegex`（字符串，可选）：必需正则表达式模式
- `notMatchRegex`（字符串，可选）：禁止正则表达式模式
- `startsWith`（字符串，可选）：必需前缀
- `endsWith`（字符串，可选）：必需后缀

#### 使用示例

```go
// 验证输出是适当的摘要
validator := &anyi.StringValidator{
    MinLength: 50,
    MaxLength: 500,
    StartsWith: "摘要：",
    EndsWith: "。",
    NotContains: "我无法",
}
```

### JsonValidator

验证输出是有效 JSON，并可选择根据 JSON Schema 进行验证。

#### 配置

```yaml
validator:
  type: "json"
  withconfig:
    schema: |
      {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "summary": {"type": "string"},
          "tags": {
            "type": "array",
            "items": {"type": "string"}
          },
          "confidence": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          }
        },
        "required": ["title", "summary"]
      }
    requiredFields:
      - "title"
      - "summary"
```

#### 配置选项

- `schema`（字符串，可选）：用于验证的 JSON Schema
- `requiredFields`（[]string，可选）：必需字段名称列表
- `allowEmpty`（bool，可选）：允许空 JSON 对象（默认：false）

#### 使用示例

```go
// 验证结构化分析输出
validator := &anyi.JsonValidator{
    RequiredFields: []string{"analysis", "recommendations", "confidence"},
    Schema: `{
        "type": "object",
        "properties": {
            "analysis": {"type": "string", "minLength": 100},
            "recommendations": {
                "type": "array",
                "items": {"type": "string"},
                "minItems": 1
            },
            "confidence": {"type": "number", "minimum": 0, "maximum": 1}
        }
    }`,
}
```

### RegexValidator

根据正则表达式模式验证输出。

#### 配置

```yaml
validator:
  type: "regex"
  withconfig:
    pattern: "^\\d{3}-\\d{2}-\\d{4}$"
    flags: "i"
    multiline: true
    dotAll: true
```

#### 配置选项

- `pattern`（字符串）：正则表达式模式
- `flags`（字符串，可选）：正则表达式标志（i=不区分大小写，m=多行，s=点匹配所有）
- `multiline`（bool，可选）：启用多行模式
- `dotAll`（bool，可选）：启用点匹配所有模式

#### 使用示例

```go
// 验证邮箱格式
validator := &anyi.RegexValidator{
    Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
    Flags: "i",
}
```

### CustomValidator

创建自定义验证器的基础接口。

#### 接口

```go
type Validator interface {
    Validate(context *FlowContext) error
}
```

#### 实现示例

```go
type WordCountValidator struct {
    MinWords int
    MaxWords int
}

func (v *WordCountValidator) Validate(context *FlowContext) error {
    words := strings.Fields(context.Text)
    count := len(words)

    if v.MinWords > 0 && count < v.MinWords {
        return fmt.Errorf("文本有 %d 个词，最少需要：%d", count, v.MinWords)
    }

    if v.MaxWords > 0 && count > v.MaxWords {
        return fmt.Errorf("文本有 %d 个词，最多允许：%d", count, v.MaxWords)
    }

    return nil
}
```

## 模板系统

### 模板语法

Anyi 使用 Go 的 `text/template` 包以及附加函数。

#### 基本语法

```go
// 变量访问
{{.Text}}
{{.Memory.FieldName}}

// 条件语句
{{if .Text}}文本存在{{end}}
{{if eq .Memory.Status "ready"}}准备好了！{{else}}未准备好{{end}}

// 循环
{{range .Memory.Items}}
- {{.}}
{{end}}

// 函数
{{len .Text}}
{{contains .Text "关键词"}}
```

#### 自定义函数

模板中可用的附加函数：

- `contains`：检查字符串是否包含子字符串
- `hasPrefix`：检查字符串是否有前缀
- `hasSuffix`：检查字符串是否有后缀
- `upper`：转换为大写
- `lower`：转换为小写
- `title`：转换为标题格式
- `trim`：删除空白字符
- `join`：用分隔符连接数组元素
- `split`：将字符串分割为数组
- `replace`：替换子字符串

#### 模板示例

```yaml
# 分析模板
template: |
  分析以下{{.Memory.ContentType}}：

  {{.Text}}

  请提供：
  {{range .Memory.Requirements}}
  - {{.}}
  {{end}}

  {{if .Think}}
  之前的分析：{{.Think}}
  {{end}}

# 条件处理
template: |
  {{if gt (len .Text) 1000}}
  这是一个长文档。请提供详细分析。
  {{else}}
  这是一个短文档。请提供简洁分析。
  {{end}}

  内容：{{.Text}}
```

## 错误处理

### 内置错误类型

组件返回的常见错误类型：

- `ValidationError`：验证失败
- `ExecutionError`：执行失败
- `TemplateError`：模板处理失败
- `ConfigurationError`：配置无效

### 错误处理模式

```go
// 检查特定错误类型
if validationErr, ok := err.(*anyi.ValidationError); ok {
    log.Printf("验证失败：%s", validationErr.Message)
    // 处理验证失败
}

// 指数退避重试逻辑
maxRetries := 3
backoff := time.Second

for i := 0; i < maxRetries; i++ {
    result, err := executor.Execute(context)
    if err == nil {
        break
    }

    if i < maxRetries-1 {
        time.Sleep(backoff)
        backoff *= 2 // 指数退避
    }
}
```

## 性能优化

### 模板缓存

```go
// 缓存编译的模板
var templateCache = make(map[string]*template.Template)

func getTemplate(templateString string) (*template.Template, error) {
    if tmpl, exists := templateCache[templateString]; exists {
        return tmpl, nil
    }

    tmpl, err := template.New("").Parse(templateString)
    if err != nil {
        return nil, err
    }

    templateCache[templateString] = tmpl
    return tmpl, nil
}
```

### 内存管理

```go
// 在长时间运行的工作流中清理上下文
func cleanupContext(ctx *FlowContext) {
    // 清理大型内存对象
    if len(ctx.Text) > 10000 {
        ctx.Text = ctx.Text[:1000] + "...[截断]"
    }

    // 清理旧的内存条目
    if len(ctx.Memory) > 100 {
        // 保留最重要的条目
        important := make(map[string]interface{})
        for _, key := range []string{"task", "status", "result"} {
            if val, exists := ctx.Memory[key]; exists {
                important[key] = val
            }
        }
        ctx.Memory = important
    }
}
```

## 最佳实践

### 组件设计

1. **保持组件简单**：每个组件应该有单一职责
2. **使配置可验证**：实现配置验证
3. **提供有用的错误消息**：帮助用户调试问题
4. **支持测试**：使组件易于单元测试

### 模板设计

1. **使用描述性变量名**：`{{.UserQuery}}` 比 `{{.Text}}` 更清晰
2. **提供默认值**：`{{.Memory.Language | default "中文"}}`
3. **验证输入**：在模板中检查必需的变量
4. **保持模板简洁**：复杂逻辑应该在执行器中处理

### 错误处理

1. **优雅降级**：当可能时提供备用选项
2. **记录详细信息**：包含足够的上下文进行调试
3. **重试策略**：对瞬时错误实现重试
4. **用户友好消息**：向最终用户提供清晰的错误消息

### 测试

```go
func TestLLMExecutor(t *testing.T) {
    executor := &anyi.LLMExecutor{
        Template: "总结：{{.Text}}",
        MaxTokens: 100,
    }

    mockClient := &MockClient{
        Response: &chat.Message{Content: "这是一个测试摘要"},
    }

    context := &FlowContext{
        Text: "这是要总结的长文本...",
    }

    result, err := executor.Execute(context, mockClient)
    assert.NoError(t, err)
    assert.Contains(t, result.Text, "摘要")
}
```

## 扩展组件

### 创建自定义执行器

```go
type CustomAnalysisExecutor struct {
    AnalysisType string
    Parameters   map[string]interface{}
}

func (e *CustomAnalysisExecutor) Execute(ctx *FlowContext, client Client) (*FlowContext, error) {
    // 实现自定义分析逻辑
    analysis := performAnalysis(ctx.Text, e.AnalysisType, e.Parameters)

    return &FlowContext{
        Text: analysis,
        Memory: map[string]interface{}{
            "analysisType": e.AnalysisType,
            "originalText": ctx.Text,
        },
    }, nil
}
```

### 注册自定义组件

```go
// 注册自定义执行器
anyi.RegisterExecutor("custom-analysis", func(config map[string]interface{}) (Executor, error) {
    return &CustomAnalysisExecutor{
        AnalysisType: config["type"].(string),
        Parameters:   config["parameters"].(map[string]interface{}),
    }, nil
})

// 在配置中使用
executor:
  type: "custom-analysis"
  withconfig:
    type: "sentiment"
    parameters:
      language: "zh"
      detailed: true
```
