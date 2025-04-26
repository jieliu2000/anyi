package datasource

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"encoding/base64"
	"encoding/json"
)

// FileDataSource 文件系统数据源
type FileDataSource struct {
	name    string
	baseDir string
}

// NewFileDataSource 创建新的文件数据源
func NewFileDataSource(name, baseDir string) *FileDataSource {
	return &FileDataSource{
		name:    name,
		baseDir: baseDir,
	}
}

// Init 初始化数据源
func (ds *FileDataSource) Init() error {
	// 确保基础目录存在
	if _, err := os.Stat(ds.baseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(ds.baseDir, 0755); err != nil {
			return fmt.Errorf("创建数据源基础目录失败: %w", err)
		}
	}
	return nil
}

// GetData 获取文件数据
func (ds *FileDataSource) GetData(ctx context.Context, query string, options map[string]interface{}) (interface{}, error) {
	path := filepath.Join(ds.baseDir, query)

	// 检查路径安全性，防止目录遍历攻击
	if !ds.isPathSafe(path) {
		return nil, fmt.Errorf("不安全的文件路径: %s", query)
	}

	// 检查文件是否存在
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	// 如果是目录，返回目录内容列表
	if info.IsDir() {
		return ds.listDirectory(path)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 检查文件类型并处理
	return ds.processFileContent(path, data, options)
}

// WriteData 写入文件数据
func (ds *FileDataSource) WriteData(ctx context.Context, path string, data interface{}, options map[string]interface{}) error {
	fullPath := filepath.Join(ds.baseDir, path)

	// 检查路径安全性
	if !ds.isPathSafe(fullPath) {
		return fmt.Errorf("不安全的文件路径: %s", path)
	}

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 将数据转换为字节
	var bytes []byte
	switch d := data.(type) {
	case string:
		bytes = []byte(d)
	case []byte:
		bytes = d
	default:
		// 尝试将其他类型转换为字符串
		bytes = []byte(fmt.Sprintf("%v", data))
	}

	// 写入文件
	return ioutil.WriteFile(fullPath, bytes, 0644)
}

// GetMetadata 获取数据源元数据
func (ds *FileDataSource) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"type":     "file",
		"base_dir": ds.baseDir,
	}
}

// Name 返回数据源名称
func (ds *FileDataSource) Name() string {
	return ds.name
}

// 辅助方法

// isPathSafe 检查路径是否安全（防止目录遍历）
func (ds *FileDataSource) isPathSafe(path string) bool {
	// 获取规范化的路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absBaseDir, err := filepath.Abs(ds.baseDir)
	if err != nil {
		return false
	}

	// 确保路径在基础目录下
	return strings.HasPrefix(absPath, absBaseDir)
}

// listDirectory 列出目录内容
func (ds *FileDataSource) listDirectory(path string) (interface{}, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		fileInfo := map[string]interface{}{
			"name":  file.Name(),
			"size":  file.Size(),
			"isDir": file.IsDir(),
			"time":  file.ModTime().Format(time.RFC3339),
		}
		result = append(result, fileInfo)
	}

	return result, nil
}

// processFileContent 处理文件内容
func (ds *FileDataSource) processFileContent(path string, data []byte, options map[string]interface{}) (interface{}, error) {
	// 根据文件扩展名处理内容
	ext := strings.ToLower(filepath.Ext(path))

	// 检查是否请求了原始字节
	if rawBytes, _ := options["raw_bytes"].(bool); rawBytes {
		return data, nil
	}

	switch ext {
	case ".json":
		// 解析JSON文件
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return string(data), nil // 如果解析失败，返回原始字符串
		}
		return jsonData, nil

	case ".txt", ".md", ".go", ".py", ".js", ".html", ".css", ".xml", ".csv", ".yaml", ".yml":
		// 文本文件直接返回字符串
		return string(data), nil

	default:
		// 默认情况下，尝试作为文本返回，如果包含二进制数据则返回base64编码
		if utf8.Valid(data) {
			return string(data), nil
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
}
