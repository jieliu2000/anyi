package anyi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/datasource"
	"github.com/jieliu2000/anyi/internal/utils"
)

// DataTransferOption 数据传输选项
type DataTransferOption string

const (
	// DataTransferDirect 直接传输数据
	DataTransferDirect DataTransferOption = "direct"

	// DataTransferReference 通过引用传输（URL）
	DataTransferReference DataTransferOption = "reference"

	// DataTransferObjectStorage 通过临时对象存储（如S3预签名URL）
	DataTransferObjectStorage DataTransferOption = "object_storage"
)

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	// 数据源名称
	Name string `mapstructure:"name" json:"name" yaml:"name"`

	// 数据源类型
	Type string `mapstructure:"type" json:"type" yaml:"type"`

	// 数据源查询模板
	QueryTemplate string `mapstructure:"query_template" json:"query_template" yaml:"query_template"`

	// 数据源配置参数
	Config map[string]interface{} `mapstructure:"config" json:"config" yaml:"config"`
}

// MCPExecutor MCP代理执行器
type MCPExecutor struct {
	// MCP服务器地址
	Endpoint string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`

	// API密钥
	APIKey string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`

	// 要使用的代理ID
	AgentID string `mapstructure:"agent_id" json:"agent_id" yaml:"agent_id"`

	// 代理参数模板
	ParamsTemplate string `mapstructure:"params_template" json:"params_template" yaml:"params_template"`

	// 代理请求的超时时间（秒）
	Timeout int `mapstructure:"timeout" json:"timeout" yaml:"timeout"`

	// 是否使用SSL连接
	UseSSL bool `mapstructure:"use_ssl" json:"use_ssl" yaml:"use_ssl"`

	// 数据源配置
	DataSources []DataSourceConfig `mapstructure:"data_sources" json:"data_sources" yaml:"data_sources"`

	// 是否在请求中包含数据源元数据
	IncludeMetadata bool `mapstructure:"include_metadata" json:"include_metadata" yaml:"include_metadata"`

	// 数据传输方式
	DataTransfer DataTransferOption `mapstructure:"data_transfer" json:"data_transfer" yaml:"data_transfer"`

	// 对象存储配置（如果使用对象存储传输）
	ObjectStorageConfig map[string]interface{} `mapstructure:"object_storage_config" json:"object_storage_config" yaml:"object_storage_config"`

	// HTTP客户端
	httpClient *http.Client

	// 参数模板
	paramsTempl *template.Template
}

// Init 初始化MCP执行器
func (e *MCPExecutor) Init() error {
	// 检查必要配置
	if e.Endpoint == "" {
		return errors.New("MCP端点不能为空")
	}

	// 替换环境变量
	if strings.HasPrefix(e.APIKey, "$") {
		envVar := strings.TrimPrefix(e.APIKey, "$")
		e.APIKey = os.Getenv(envVar)
	}

	if e.APIKey == "" {
		return errors.New("API密钥不能为空")
	}

	if e.AgentID == "" {
		return errors.New("代理ID不能为空")
	}

	// 设置默认值
	if e.Timeout <= 0 {
		e.Timeout = 60
	}

	if e.DataTransfer == "" {
		e.DataTransfer = DataTransferDirect
	}

	// 初始化HTTP客户端
	e.httpClient = &http.Client{
		Timeout: time.Duration(e.Timeout) * time.Second,
	}

	// 初始化参数模板
	if e.ParamsTemplate != "" {
		tmpl, err := template.New("mcp_params").Parse(e.ParamsTemplate)
		if err != nil {
			return fmt.Errorf("解析参数模板失败: %w", err)
		}
		e.paramsTempl = tmpl
	}

	// 初始化数据源
	for _, dsConfig := range e.DataSources {
		// 检查数据源是否已注册
		if _, ok := datasource.Get(dsConfig.Name); !ok {
			// 尝试创建数据源
			if err := e.createDataSource(dsConfig); err != nil {
				log.Warnf("创建数据源失败: %s, 错误: %v", dsConfig.Name, err)
			}
		}
	}

	return nil
}

// createDataSource 创建数据源
func (e *MCPExecutor) createDataSource(config DataSourceConfig) error {
	if config.Type == "" {
		return fmt.Errorf("数据源类型不能为空")
	}

	var ds datasource.DataSource

	switch config.Type {
	case "file":
		baseDir, _ := config.Config["base_dir"].(string)
		if baseDir == "" {
			baseDir = "."
		}
		ds = datasource.NewFileDataSource(config.Name, baseDir)
	default:
		return fmt.Errorf("不支持的数据源类型: %s", config.Type)
	}

	// 初始化数据源
	if err := ds.Init(); err != nil {
		return err
	}

	// 注册数据源
	datasource.Register(config.Name, ds)
	return nil
}

// Run 执行MCP代理调用
func (e *MCPExecutor) Run(flowContext flow.FlowContext, step *flow.Step) (*flow.FlowContext, error) {
	// 准备上下文数据
	ctx := context.Background()

	// 1. 收集数据源数据
	sourceData, err := e.collectDataFromSources(ctx, flowContext.Variables)
	if err != nil {
		return &flowContext, fmt.Errorf("收集数据源数据失败: %w", err)
	}

	// 2. 准备代理参数
	params, err := e.prepareAgentParams(flowContext.Variables)
	if err != nil {
		return &flowContext, fmt.Errorf("准备代理参数失败: %w", err)
	}

	// 如果有数据源数据，添加到参数中
	if len(sourceData) > 0 {
		params["data_sources"] = sourceData
	}

	// 3. 调用MCP代理
	result, err := e.callMCPAgent(ctx, params)
	if err != nil {
		return &flowContext, fmt.Errorf("调用MCP代理失败: %w", err)
	}

	// 4. 处理返回结果中的数据写入指令
	processedResult, err := e.processAgentResponse(result)
	if err != nil {
		log.Warnf("处理代理响应时出错: %v", err)
	}

	// 5. 更新流程上下文
	flowContext.Text = processedResult
	return &flowContext, nil
}

// collectDataFromSources 从数据源收集数据
func (e *MCPExecutor) collectDataFromSources(ctx context.Context, variables map[string]interface{}) (map[string]interface{}, error) {
	sourceData := make(map[string]interface{})

	for _, dsConfig := range e.DataSources {
		// 处理查询模板
		query, err := utils.ProcessTemplate(dsConfig.QueryTemplate, variables)
		if err != nil {
			return nil, err
		}

		// 获取数据源
		ds, ok := datasource.Get(dsConfig.Name)
		if !ok {
			return nil, &datasource.ErrDataSourceNotFound{Name: dsConfig.Name}
		}

		// 查询数据
		options := make(map[string]interface{})
		if dsConfig.Config != nil {
			for k, v := range dsConfig.Config {
				options[k] = v
			}
		}

		data, err := ds.GetData(ctx, query, options)
		if err != nil {
			return nil, err
		}

		// 添加数据
		sourceItem := map[string]interface{}{
			"data": data,
		}

		// 添加元数据
		if e.IncludeMetadata {
			sourceItem["metadata"] = ds.GetMetadata()
		}

		sourceData[dsConfig.Name] = sourceItem
	}

	// 根据传输方式处理数据
	switch e.DataTransfer {
	case DataTransferDirect:
		// 直接返回数据，不做处理
		return sourceData, nil

	case DataTransferReference, DataTransferObjectStorage:
		// 这些高级传输方式需要额外实现
		log.Warn("暂不支持的数据传输方式:", e.DataTransfer)
		return sourceData, nil

	default:
		// 默认为直接传输
		return sourceData, nil
	}
}

// prepareAgentParams 准备代理参数
func (e *MCPExecutor) prepareAgentParams(variables map[string]interface{}) (map[string]interface{}, error) {
	// 默认参数
	params := map[string]interface{}{
		"agent_id": e.AgentID,
	}

	// 如果有参数模板，使用模板处理
	if e.paramsTempl != nil {
		var buf bytes.Buffer
		if err := e.paramsTempl.Execute(&buf, variables); err != nil {
			return nil, err
		}

		// 解析模板输出为JSON
		var templateParams map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &templateParams); err != nil {
			// 如果不是有效的JSON，作为字符串输入
			params["input"] = buf.String()
		} else {
			// 合并模板参数
			for k, v := range templateParams {
				params[k] = v
			}
		}
	} else {
		// 没有模板，直接使用变量作为输入
		params["input"] = variables
	}

	return params, nil
}

// callMCPAgent 调用MCP代理
func (e *MCPExecutor) callMCPAgent(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// 准备URL
	protocol := "https"
	if !e.UseSSL {
		protocol = "http"
	}
	url := fmt.Sprintf("%s://%s/agent", protocol, e.Endpoint)

	// 准备请求体
	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// 添加头信息
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.APIKey))

	// 发送请求
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP服务返回错误: %s (%d)", string(body), resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果不是JSON，返回原始响应
		return map[string]interface{}{
			"result": string(body),
		}, nil
	}

	return result, nil
}

// processAgentResponse 处理代理响应，包括处理数据写入指令
func (e *MCPExecutor) processAgentResponse(response map[string]interface{}) (string, error) {
	// 提取结果文本
	var result string
	if res, ok := response["result"]; ok {
		switch r := res.(type) {
		case string:
			result = r
		case map[string]interface{}:
			// 尝试编码为JSON字符串
			jsonBytes, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				result = fmt.Sprintf("%v", r)
			} else {
				result = string(jsonBytes)
			}
		default:
			result = fmt.Sprintf("%v", r)
		}
	}

	// 处理数据写入指令
	if writes, ok := response["write_data"].([]interface{}); ok {
		for _, writeOp := range writes {
			writeMap, ok := writeOp.(map[string]interface{})
			if !ok {
				continue
			}

			dsName, _ := writeMap["data_source"].(string)
			if dsName == "" {
				continue
			}

			path, _ := writeMap["path"].(string)
			if path == "" {
				continue
			}

			data := writeMap["data"]
			if data == nil {
				continue
			}

			options, _ := writeMap["options"].(map[string]interface{})

			// 获取数据源
			ds, ok := datasource.Get(dsName)
			if !ok {
				log.Warnf("数据源不存在: %s", dsName)
				continue
			}

			// 检查是否支持写入
			writableDS, ok := ds.(datasource.WritableDataSource)
			if !ok {
				log.Warnf("数据源不支持写入: %s", dsName)
				continue
			}

			// 执行写入
			if err := writableDS.WriteData(context.Background(), path, data, options); err != nil {
				log.Warnf("写入数据失败: %v", err)
			} else {
				log.Infof("成功写入数据到 %s:%s", dsName, path)
			}
		}
	}

	return result, nil
}

// 注册MCP执行器
func init() {
	RegisterExecutor("mcp", &MCPExecutor{})
}
