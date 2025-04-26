package anyi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/datasource"
	"github.com/jieliu2000/anyi/llm/openai"
	"github.com/stretchr/testify/assert"
)

func TestMCPExecutorInit(t *testing.T) {
	executor := &MCPExecutor{
		Endpoint: "test.mcp.com",
		APIKey:   "test-api-key",
		AgentID:  "test-agent",
		Timeout:  30,
		UseSSL:   true,
		ParamsTemplate: `{
			"query": "{{.query}}",
			"options": {"format": "json"}
		}`,
	}

	err := executor.Init()
	assert.NoError(t, err)
	assert.NotNil(t, executor.httpClient)
	assert.NotNil(t, executor.paramsTempl)
	assert.Equal(t, DataTransferDirect, executor.DataTransfer)
}

func TestMCPExecutorMissingRequiredFields(t *testing.T) {
	// 缺少端点
	executor1 := &MCPExecutor{
		APIKey:  "test-api-key",
		AgentID: "test-agent",
	}
	err := executor1.Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "端点不能为空")

	// 缺少API密钥
	executor2 := &MCPExecutor{
		Endpoint: "test.mcp.com",
		AgentID:  "test-agent",
	}
	err = executor2.Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API密钥不能为空")

	// 缺少代理ID
	executor3 := &MCPExecutor{
		Endpoint: "test.mcp.com",
		APIKey:   "test-api-key",
	}
	err = executor3.Init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "代理ID不能为空")
}

func TestMCPExecutorRun(t *testing.T) {
	// 创建临时目录用于文件数据源测试
	tempDir, err := os.MkdirTemp("", "mcp_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFilePath := filepath.Join(tempDir, "test.txt")
	testFileContent := "这是一个测试文件\n用于测试MCP执行器"
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 注册文件数据源
	fileDs := datasource.NewFileDataSource("test_files", tempDir)
	fileDs.Init()
	datasource.Register("test_files", fileDs)

	// 创建模拟MCP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法
		assert.Equal(t, "POST", r.Method)

		// 检查授权头
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// 检查内容类型
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"result": "这是MCP代理的响应结果",
			"write_data": [
				{
					"data_source": "test_files",
					"path": "output.txt",
					"data": "MCP代理生成的输出文件内容"
				}
			]
		}`)
	}))
	defer server.Close()

	// 创建MCP执行器
	executor := &MCPExecutor{
		Endpoint: server.URL,
		APIKey:   "test-api-key",
		AgentID:  "test-agent",
		Timeout:  5,
		UseSSL:   false,
		DataSources: []DataSourceConfig{
			{
				Name:          "test_files",
				Type:          "file",
				QueryTemplate: "{{.file_name}}",
			},
		},
		IncludeMetadata: true,
	}

	// 初始化执行器
	err = executor.Init()
	assert.NoError(t, err)

	// 创建流程上下文
	flowContext := flow.FlowContext{
		Variables: map[string]interface{}{
			"query":     "测试查询",
			"file_name": "test.txt",
		},
	}

	// 创建流程步骤
	client, err := openai.NewClient(openai.DefaultConfig(""))
	if err != nil {
		t.Fatal(err)
	}

	step := flow.NewStep(executor, nil, client)

	// 执行MCP步骤
	result, err := executor.Run(flowContext, step)
	assert.NoError(t, err)
	assert.Equal(t, "这是MCP代理的响应结果", result.Text)

	// 检查是否生成了输出文件
	outputPath := filepath.Join(tempDir, "output.txt")
	outputContent, err := os.ReadFile(outputPath)
	assert.NoError(t, err)
	assert.Equal(t, "MCP代理生成的输出文件内容", string(outputContent))
}

func TestMCPExecutorPrepareParams(t *testing.T) {
	// 测试带模板的参数准备
	executor1 := &MCPExecutor{
		Endpoint: "test.mcp.com",
		APIKey:   "test-api-key",
		AgentID:  "test-agent",
		ParamsTemplate: `{
			"query": "{{.query}}",
			"max_tokens": 1000,
			"options": {
				"format": "{{.format}}"
			}
		}`,
	}

	err := executor1.Init()
	assert.NoError(t, err)

	variables := map[string]interface{}{
		"query":  "测试查询",
		"format": "json",
	}

	params, err := executor1.prepareAgentParams(variables)
	assert.NoError(t, err)
	assert.Equal(t, "test-agent", params["agent_id"])
	assert.Equal(t, "测试查询", params["query"])
	assert.Equal(t, float64(1000), params["max_tokens"])
	assert.Equal(t, "json", params["options"].(map[string]interface{})["format"])

	// 测试不带模板的参数准备
	executor2 := &MCPExecutor{
		Endpoint: "test.mcp.com",
		APIKey:   "test-api-key",
		AgentID:  "test-agent",
	}

	err = executor2.Init()
	assert.NoError(t, err)

	params, err = executor2.prepareAgentParams(variables)
	assert.NoError(t, err)
	assert.Equal(t, "test-agent", params["agent_id"])
	assert.Equal(t, variables, params["input"])
}

func TestMCPExecutorCollectData(t *testing.T) {
	// 创建临时目录用于文件数据源测试
	tempDir, err := os.MkdirTemp("", "mcp_data_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFilePath := filepath.Join(tempDir, "data.json")
	testFileContent := `{"key": "value", "number": 123}`
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 注册文件数据源
	fileDs := datasource.NewFileDataSource("json_files", tempDir)
	fileDs.Init()
	datasource.Register("json_files", fileDs)

	// 创建执行器
	executor := &MCPExecutor{
		Endpoint: "test.mcp.com",
		APIKey:   "test-api-key",
		AgentID:  "test-agent",
		DataSources: []DataSourceConfig{
			{
				Name:          "json_files",
				Type:          "file",
				QueryTemplate: "{{.file_name}}",
			},
		},
		IncludeMetadata: true,
	}

	// 初始化执行器
	err = executor.Init()
	assert.NoError(t, err)

	// 收集数据
	variables := map[string]interface{}{
		"query":     "测试查询",
		"file_name": "data.json",
	}

	sourceData, err := executor.collectDataFromSources(context.Background(), variables)
	assert.NoError(t, err)
	assert.NotNil(t, sourceData)

	// 检查数据
	jsonFileData, ok := sourceData["json_files"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, jsonFileData["data"])
	assert.NotNil(t, jsonFileData["metadata"])

	// 检查JSON内容已被解析为对象
	data, ok := jsonFileData["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", data["key"])
	assert.Equal(t, float64(123), data["number"])

	// 检查元数据
	metadata, ok := jsonFileData["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "file", metadata["type"])
	assert.Equal(t, tempDir, metadata["base_dir"])
}
