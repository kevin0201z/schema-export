package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/microsoft/go-mssqldb" // SQL Server Go 驱动

	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector SQL Server 数据库 Inspector 实现
type Inspector struct {
	*database.BaseInspector
}

// NewInspector 创建 SQL Server Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		BaseInspector: database.NewBaseInspector(config),
	}
}

// Connect 连接 SQL Server 数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sqlserver connection: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(database.DefaultMaxOpenConns)
	db.SetMaxIdleConns(database.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(database.DefaultConnMaxLifetime)

	// 使用带超时的 context
	pingCtx, cancel := context.WithTimeout(ctx, database.DefaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping sqlserver database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建 SQL Server DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		dsn := config.DSN
		// 如果 DSN 中已经有 sqlserver:// 前缀，直接返回
		if strings.HasPrefix(dsn, "sqlserver://") {
			return dsn
		}
		// 如果 DSN 中没有 sqlserver:// 前缀，添加它
		if !strings.Contains(dsn, "://") {
			dsn = "sqlserver://" + dsn
		}
		return dsn
	}

	// SQL Server DSN 格式: sqlserver://user:password@host:port?database=dbname
	params := ""
	if config.Database != "" {
		params = fmt.Sprintf("?database=%s", config.Database)
	}

	return fmt.Sprintf("sqlserver://%s:%s@%s:%d%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		params,
	)
}

// GetTables 获取所有表列表
func (i *Inspector) GetTables(ctx context.Context) ([]model.Table, error) {
	query := `
		SELECT 
			t.name AS table_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.is_ms_shipped = 0
		ORDER BY t.name
	`

	rows, err := i.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []model.Table
	for rows.Next() {
		var table model.Table
		var comment sql.NullString
		if err := rows.Scan(&table.Name, &comment); err != nil {
			return nil, err
		}
		table.Type = model.TableTypeTable
		if comment.Valid {
			table.Comment = comment.String
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetTable 获取单个表的完整元数据
func (i *Inspector) GetTable(ctx context.Context, tableName string) (*model.Table, error) {
	table := &model.Table{
		Name: tableName,
		Type: model.TableTypeTable,
	}

	// 获取表注释
	comment, _ := i.getTableComment(ctx, tableName)
	table.Comment = comment

	// 获取字段
	columns, err := i.GetColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	// 获取索引
	indexes, err := i.GetIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Indexes = indexes

	// 获取外键
	foreignKeys, err := i.GetForeignKeys(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.ForeignKeys = foreignKeys

	// 获取 CHECK 约束
	checkConstraints, err := i.GetCheckConstraints(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.CheckConstraints = checkConstraints

	return table, nil
}

// GetColumns 获取表字段列表
func (i *Inspector) GetColumns(ctx context.Context, tableName string) ([]model.Column, error) {
	query := `
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []model.Column
	for rows.Next() {
		var col model.Column
		var maxLength int
		var precision, scale sql.NullInt64
		var nullable bool
		var defaultValue, comment sql.NullString
		var isPrimaryKey int
		var isAutoIncrement bool

		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&maxLength,
			&precision,
			&scale,
			&nullable,
			&defaultValue,
			&comment,
			&isPrimaryKey,
			&isAutoIncrement,
		); err != nil {
			return nil, err
		}

		// SQL Server 的 max_length 对于 Unicode 类型需要除以 2
		if isUnicodeType(col.DataType) {
			col.Length = maxLength / 2
		} else {
			col.Length = maxLength
		}

		if precision.Valid {
			col.Precision = int(precision.Int64)
		}
		if scale.Valid {
			col.Scale = int(scale.Int64)
		}
		col.IsNullable = nullable
		col.IsPrimaryKey = (isPrimaryKey == 1)
		col.IsAutoIncrement = isAutoIncrement
		if defaultValue.Valid {
			col.DefaultValue = cleanDefaultValue(defaultValue.String)
		}
		if comment.Valid {
			col.Comment = comment.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// GetIndexes 获取表索引列表
func (i *Inspector) GetIndexes(ctx context.Context, tableName string) ([]model.Index, error) {
	query := `
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*model.Index)
	for rows.Next() {
		var indexName, indexType, columnName string
		var isUnique, isPrimaryKey bool
		if err := rows.Scan(&indexName, &indexType, &isUnique, &isPrimaryKey, &columnName); err != nil {
			return nil, err
		}

		idx, exists := indexMap[indexName]
		if !exists {
			idx = &model.Index{
				Name:      indexName,
				IsUnique:  isUnique,
				IsPrimary: isPrimaryKey,
			}
			if isPrimaryKey {
				idx.Type = model.IndexTypePrimary
			} else if isUnique {
				idx.Type = model.IndexTypeUnique
			} else {
				idx.Type = model.IndexTypeNormal
			}
			indexMap[indexName] = idx
		}
		idx.Columns = append(idx.Columns, columnName)
	}

	// 转换为切片
	names := make([]string, 0, len(indexMap))
	for name := range indexMap {
		names = append(names, name)
	}
	sort.Strings(names)

	indexes := make([]model.Index, 0, len(indexMap))
	for _, name := range names {
		indexes = append(indexes, *indexMap[name])
	}

	return indexes, rows.Err()
}

// GetForeignKeys 获取表外键列表
func (i *Inspector) GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error) {
	query := `
		SELECT 
			fk.name AS fk_name,
			pc.name AS column_name,
			rt.name AS ref_table,
			rc.name AS ref_column,
			fk.delete_referential_action_desc AS on_delete
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables pt ON fk.parent_object_id = pt.object_id
		INNER JOIN sys.columns pc ON fkc.parent_object_id = pc.object_id AND fkc.parent_column_id = pc.column_id
		INNER JOIN sys.tables rt ON fk.referenced_object_id = rt.object_id
		INNER JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
		WHERE pt.name = @p1
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []model.ForeignKey
	for rows.Next() {
		var fk model.ForeignKey
		var onDelete sql.NullString
		if err := rows.Scan(
			&fk.Name,
			&fk.Column,
			&fk.RefTable,
			&fk.RefColumn,
			&onDelete,
		); err != nil {
			return nil, err
		}
		if onDelete.Valid && onDelete.String != "NO_ACTION" {
			fk.OnDelete = onDelete.String
		}
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// GetCheckConstraints 获取表 CHECK 约束列表
func (i *Inspector) GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error) {
	query := `
		SELECT 
			cc.name AS constraint_name,
			cc.definition AS definition,
			c.name AS column_name
		FROM sys.check_constraints cc
		LEFT JOIN sys.columns c ON cc.parent_column_id = c.column_id AND cc.parent_object_id = c.object_id
		INNER JOIN sys.tables t ON cc.parent_object_id = t.object_id
		WHERE t.name = @p1
		ORDER BY cc.name
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query check constraints: %w", err)
	}
	defer rows.Close()

	constraintMap := make(map[string]*model.CheckConstraint)
	for rows.Next() {
		var name, definition string
		var columnName sql.NullString
		if err := rows.Scan(&name, &definition, &columnName); err != nil {
			return nil, err
		}

		cc, exists := constraintMap[name]
		if !exists {
			cc = &model.CheckConstraint{
				Name:       name,
				Definition: definition,
				Columns:    []string{},
			}
			constraintMap[name] = cc
		}
		if columnName.Valid && columnName.String != "" {
			cc.Columns = append(cc.Columns, columnName.String)
		}
	}

	// 转换为切片
	result := make([]model.CheckConstraint, 0, len(constraintMap))
	for _, cc := range constraintMap {
		result = append(result, *cc)
	}

	return result, rows.Err()
}

// GetViews 获取视图列表
func (i *Inspector) GetViews(ctx context.Context) ([]model.View, error) {
	query := `
		SELECT 
			v.name AS view_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS view_comment,
			CAST(m.definition AS NVARCHAR(MAX)) AS view_definition
		FROM sys.views v
		LEFT JOIN sys.sql_modules m ON v.object_id = m.object_id
		LEFT JOIN sys.extended_properties ep 
			ON v.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE v.is_ms_shipped = 0
		ORDER BY v.name
	`

	rows, err := i.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []model.View
	for rows.Next() {
		var view model.View
		var comment, definition sql.NullString
		if err := rows.Scan(&view.Name, &comment, &definition); err != nil {
			return nil, err
		}
		if comment.Valid {
			view.Comment = comment.String
		}
		if definition.Valid {
			view.Definition = definition.String
		}
		views = append(views, view)
	}

	return views, rows.Err()
}

// getTableComment 获取表注释
func (i *Inspector) getTableComment(ctx context.Context, tableName string) (string, error) {
	query := `
		SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1
	`
	var comment sql.NullString
	if err := i.GetDB().QueryRowContext(ctx, query, tableName).Scan(&comment); err != nil {
		return "", err
	}
	if comment.Valid {
		return comment.String, nil
	}
	return "", nil
}

// isUnicodeType 判断是否为 Unicode 类型
func isUnicodeType(dataType string) bool {
	unicodeTypes := map[string]bool{
		"nvarchar": true,
		"nchar":    true,
		"ntext":    true,
	}
	return unicodeTypes[strings.ToLower(dataType)]
}

// cleanDefaultValue 清理默认值中的括号
func cleanDefaultValue(value string) string {
	// SQL Server 默认值通常带有括号，如 ((0)) 或 ('default')
	value = strings.TrimSpace(value)
	for strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		value = strings.TrimPrefix(value, "(")
		value = strings.TrimSuffix(value, ")")
		value = strings.TrimSpace(value)
	}
	return value
}

// Factory SQL Server Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	return NewInspector(config), nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "sqlserver"
}

func init() {
	inspector.Register("sqlserver", &Factory{})
}

// Forwarding methods to ensure embedded BaseInspector methods are accessible
func (i *Inspector) SetDB(db *sql.DB) {
	i.BaseInspector.SetDB(db)
}

func (i *Inspector) GetDB() *sql.DB {
	return i.BaseInspector.GetDB()
}

func (i *Inspector) GetConfig() inspector.ConnectionConfig {
	return i.BaseInspector.GetConfig()
}
