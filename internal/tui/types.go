package tui

// Page 标识当前所在的向导页面。
type Page int

const (
	PageWelcome          Page = iota // 欢迎页
	PageDBType                       // 选择数据库类型
	PageConnectionMethod             // 选择连接方式
	PageDSNForm                      // DSN 表单
	PageParamsForm                   // 分离参数表单
	PageExportContent                // 选择导出对象类型
	PageTableFilter                  // 表过滤规则
	PageOutputSettings               // 输出设置
	PageConfirm                      // 确认摘要
	PageExecution                    // 执行导出 + 结果展示
)

// ConnectionMethod 数据库连接方式。
type ConnectionMethod int

const (
	ConnectionMethodDSN    ConnectionMethod = iota // DSN 连接字符串
	ConnectionMethodParams                         // 分离参数
)

// TuiState 存储所有表单输入值。
type TuiState struct {
	// 数据库连接
	DBType           string           // 数据库类型: dm/oracle/sqlserver/mysql/postgres
	ConnectionMethod ConnectionMethod // 连接方式
	DSN              string           // DSN 连接字符串
	Host             string           // 主机地址
	Port             string           // 端口（字符串存储，映射时转换）
	Database         string           // 数据库名称
	Username         string           // 用户名
	Password         string           // 密码
	Schema           string           // Schema

	// 导出内容
	IncludeViews      bool // 包含视图
	IncludeProcedures bool // 包含存储过程
	IncludeFunctions  bool // 包含函数
	IncludeTriggers   bool // 包含触发器
	IncludeSequences  bool // 包含序列

	// 表过滤（逗号分隔字符串，映射时解析为 []string）
	Tables   string // 指定表名
	Exclude  string // 排除表名
	Patterns string // 正则匹配模式

	// 输出设置
	OutputDir  string   // 输出目录
	Formats    []string // 导出格式: markdown/sql/json/yaml
	SplitFiles bool     // 是否按表分文件导出
}

// 校验错误消息。
const (
	errMsgEmptyDSN    = "DSN 不能为空"
	errMsgEmptyHost   = "主机地址不能为空"
	errMsgEmptyUser   = "用户名不能为空"
	errMsgInvalidPort = "端口必须为 1-65535 之间的数字"
	errMsgNoFormat    = "请至少选择一种导出格式"
)

// 数据库类型选项。
var dbTypeOptions = []string{"dm", "oracle", "sqlserver", "mysql", "postgres"}

// 导出格式选项。
var formatOptions = []string{"markdown", "sql", "json", "yaml"}

// 导出内容选项（label 用于显示，key 用于映射到 TuiState 字段）。
type contentOption struct {
	label string
	key   string
}

var contentOptions = []contentOption{
	{label: "视图 (Views)", key: "views"},
	{label: "存储过程 (Procedures)", key: "procedures"},
	{label: "函数 (Functions)", key: "functions"},
	{label: "触发器 (Triggers)", key: "triggers"},
	{label: "序列 (Sequences)", key: "sequences"},
}
