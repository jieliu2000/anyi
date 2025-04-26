package datasource

import (
	"context"
	"fmt"
	"sync"
)

// DataSource 定义数据源接口
type DataSource interface {
	// GetData 根据查询获取数据
	GetData(ctx context.Context, query string, options map[string]interface{}) (interface{}, error)

	// GetMetadata 获取数据源元数据
	GetMetadata() map[string]interface{}

	// Name 返回数据源名称
	Name() string

	// Init 初始化数据源
	Init() error
}

// WritableDataSource 可写数据源接口
type WritableDataSource interface {
	DataSource

	// WriteData 写入数据
	WriteData(ctx context.Context, path string, data interface{}, options map[string]interface{}) error
}

// registry 数据源注册表
var registry = struct {
	sync.RWMutex
	sources map[string]DataSource
}{
	sources: make(map[string]DataSource),
}

// Register 注册数据源
func Register(name string, ds DataSource) {
	registry.Lock()
	defer registry.Unlock()
	registry.sources[name] = ds
}

// Get 获取数据源
func Get(name string) (DataSource, bool) {
	registry.RLock()
	defer registry.RUnlock()
	ds, ok := registry.sources[name]
	return ds, ok
}

// ListDataSources 获取所有已注册的数据源名称
func ListDataSources() []string {
	registry.RLock()
	defer registry.RUnlock()

	names := make([]string, 0, len(registry.sources))
	for name := range registry.sources {
		names = append(names, name)
	}
	return names
}

// ErrDataSourceNotFound 数据源未找到错误
type ErrDataSourceNotFound struct {
	Name string
}

func (e ErrDataSourceNotFound) Error() string {
	return fmt.Sprintf("数据源未找到: %s", e.Name)
}

// ErrDataSourceNoWriteSupport 数据源不支持写入错误
type ErrDataSourceNoWriteSupport struct {
	Name string
}

func (e ErrDataSourceNoWriteSupport) Error() string {
	return fmt.Sprintf("数据源不支持写入: %s", e.Name)
}
