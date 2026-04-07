package exporter

import (
	"github.com/schema-export/schema-export/internal/model"
)

// Exporter 导出器接口。
//
// Exporter 定义了将数据库元数据导出为特定格式的统一接口。
// 每种支持的导出格式都需要实现该接口。
//
// 使用流程:
//  1. 创建导出器实例
//  2. 准备要导出的数据和选项
//  3. 调用 Export 方法执行导出
type Exporter interface {
	// Export 导出数据库元数据到文件。
	//
	// 参数:
	//   - tables: 表结构列表
	//   - views: 视图列表
	//   - procedures: 存储过程列表
	//   - functions: 函数列表
	//   - triggers: 触发器列表
	//   - sequences: 序列列表
	//   - options: 导出选项
	//
	// 返回值:
	//   - error: 导出失败时返回错误
	Export(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options ExportOptions) error

	// GetName 获取导出器名称。
	//
	// 返回值:
	//   - string: 导出器名称（如 "markdown", "sql", "json", "yaml"）
	GetName() string

	// GetExtension 获取导出文件的扩展名。
	//
	// 返回值:
	//   - string: 文件扩展名（如 ".md", ".sql", ".json", ".yaml"）
	GetExtension() string
}

// ExportOptions 导出选项配置。
//
// ExportOptions 控制导出的行为，包括输出路径、文件命名、内容过滤等。
type ExportOptions struct {
	OutputDir         string   // 输出目录路径
	FileName          string   // 文件名（单文件模式）
	SplitFiles        bool     // 是否按表分文件导出
	Tables            []string // 指定导出的表（空表示全部）
	Exclude           []string // 排除的表
	DbType            string   // 数据库类型（用于选择 SQL 方言）
	IncludeViews      bool     // 是否包含视图
	IncludeProcedures bool     // 是否包含存储过程
	IncludeFunctions  bool     // 是否包含函数
	IncludeTriggers   bool     // 是否包含触发器
	IncludeSequences  bool     // 是否包含序列
}

// ExporterFactory 导出器工厂接口。
//
// 工厂接口用于创建特定格式的导出器实例。
// 每种导出格式都应该实现该接口，并通过 Register 函数注册到全局注册表。
type ExporterFactory interface {
	// Create 创建导出器实例。
	//
	// 返回值:
	//   - Exporter: 新创建的导出器实例
	//   - error: 创建失败时返回错误
	Create() (Exporter, error)

	// GetType 获取导出器类型标识。
	//
	// 返回值:
	//   - string: 导出器类型（如 "markdown", "sql", "json", "yaml"）
	GetType() string
}

// Registry 导出器工厂注册表。
//
// Registry 维护了导出格式到导出器工厂的映射关系。
// 支持动态注册新的导出格式。
type Registry struct {
	factories map[string]ExporterFactory
}

// NewRegistry 创建新的工厂注册表。
//
// 返回值:
//   - *Registry: 新创建的空注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]ExporterFactory),
	}
}

// Register 注册导出器工厂到注册表。
//
// 参数:
//   - exporterType: 导出格式类型标识
//   - factory: 工厂实例
func (r *Registry) Register(exporterType string, factory ExporterFactory) {
	r.factories[exporterType] = factory
}

// Get 从注册表获取指定格式的工厂。
//
// 参数:
//   - exporterType: 导出格式类型标识
//
// 返回值:
//   - ExporterFactory: 工厂实例
//   - bool: 是否存在该格式的工厂
func (r *Registry) Get(exporterType string) (ExporterFactory, bool) {
	factory, ok := r.factories[exporterType]
	return factory, ok
}

// GetSupportedTypes 获取注册表支持的所有导出格式。
//
// 返回值:
//   - []string: 导出格式类型列表
func (r *Registry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// globalRegistry 全局导出器工厂注册表。
//
// 各导出器实现通过 init 函数自动注册到全局注册表。
var globalRegistry = NewRegistry()

// Register 注册工厂到全局注册表。
//
// 通常在导出器实现包的 init 函数中调用。
//
// 参数:
//   - exporterType: 导出格式类型标识
//   - factory: 工厂实例
func Register(exporterType string, factory ExporterFactory) {
	globalRegistry.Register(exporterType, factory)
}

// GetFactory 从全局注册表获取工厂。
//
// 参数:
//   - exporterType: 导出格式类型标识
//
// 返回值:
//   - ExporterFactory: 工厂实例
//   - bool: 是否存在该格式的工厂
func GetFactory(exporterType string) (ExporterFactory, bool) {
	return globalRegistry.Get(exporterType)
}

// GetSupportedTypes 获取全局注册表支持的所有导出格式。
//
// 返回值:
//   - []string: 导出格式类型列表
func GetSupportedTypes() []string {
	return globalRegistry.GetSupportedTypes()
}
