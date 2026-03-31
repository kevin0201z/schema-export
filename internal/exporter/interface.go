package exporter

import (
	"github.com/schema-export/schema-export/internal/model"
)

// Exporter 导出器接口
type Exporter interface {
	// Export 导出表结构
	Export(tables []model.Table, options ExportOptions) error

	// GetName 获取导出器名称
	GetName() string

	// GetExtension 获取文件扩展名
	GetExtension() string
}

// ExportOptions 导出选项
type ExportOptions struct {
	OutputDir  string   // 输出目录
	FileName   string   // 文件名（单文件模式）
	SplitFiles bool     // 是否分文件导出
	Tables     []string // 指定导出的表（空表示全部）
	Exclude    []string // 排除的表
}

// ExporterFactory 导出器工厂接口
type ExporterFactory interface {
	// Create 创建 Exporter 实例
	Create() (Exporter, error)

	// GetType 获取导出器类型
	GetType() string
}

// Registry Exporter 工厂注册表
type Registry struct {
	factories map[string]ExporterFactory
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]ExporterFactory),
	}
}

// Register 注册工厂
func (r *Registry) Register(exporterType string, factory ExporterFactory) {
	r.factories[exporterType] = factory
}

// Get 获取工厂
func (r *Registry) Get(exporterType string) (ExporterFactory, bool) {
	factory, ok := r.factories[exporterType]
	return factory, ok
}

// GetSupportedTypes 获取支持的导出器类型列表
func (r *Registry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// 全局注册表
var globalRegistry = NewRegistry()

// Register 注册工厂到全局注册表
func Register(exporterType string, factory ExporterFactory) {
	globalRegistry.Register(exporterType, factory)
}

// GetFactory 从全局注册表获取工厂
func GetFactory(exporterType string) (ExporterFactory, bool) {
	return globalRegistry.Get(exporterType)
}

// GetSupportedTypes 获取全局注册表支持的导出器类型
func GetSupportedTypes() []string {
	return globalRegistry.GetSupportedTypes()
}
