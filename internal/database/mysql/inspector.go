package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL Go 驱动

	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector MySQL 数据库 Inspector 实现
type Inspector struct {
	*database.BaseInspector
}

// NewInspector 创建 MySQL Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		BaseInspector: database.NewBaseInspector(config),
	}
}

// Connect 连接 MySQL 数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(database.DefaultMaxOpenConns)
	db.SetMaxIdleConns(database.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(database.DefaultConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, database.DefaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping mysql database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建 MySQL DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		dsn := config.DSN
		if strings.HasPrefix(dsn, "mysql://") || strings.Contains(dsn, "@tcp(") {
			return dsn
		}
		if !strings.Contains(dsn, "://") && !strings.Contains(dsn, "@tcp(") {
			dsn = "mysql://" + dsn
		}
		return dsn
	}

	params := []string{"charset=utf8mb4", "parseTime=true", "loc=Local"}
	if config.SSLMode != "" {
		params = append(params, "tls="+config.SSLMode)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		strings.Join(params, "&"),
	)
}

// GetTables 获取所有表列表
func (i *Inspector) GetTables(ctx context.Context) ([]model.Table, error) {
	query := `
		SELECT TABLE_NAME, TABLE_COMMENT 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
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

	comment, _ := i.getTableComment(ctx, tableName)
	table.Comment = comment

	columns, err := i.GetColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	indexes, err := i.GetIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Indexes = indexes

	foreignKeys, err := i.GetForeignKeys(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.ForeignKeys = foreignKeys

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
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_COMMENT,
			EXTRA LIKE '%auto_increment%' AS is_auto_increment,
			(SELECT COUNT(*) FROM information_schema.KEY_COLUMN_USAGE kcu
			 JOIN information_schema.TABLE_CONSTRAINTS tc ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
			 WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ? 
			   AND tc.CONSTRAINT_TYPE = 'PRIMARY KEY' AND kcu.COLUMN_NAME = c.COLUMN_NAME) > 0 AS is_pk
		FROM information_schema.COLUMNS c
		WHERE c.TABLE_SCHEMA = DATABASE() AND c.TABLE_NAME = ?
		ORDER BY c.ORDINAL_POSITION
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []model.Column
	for rows.Next() {
		var col model.Column
		var nullable string
		var defaultValue sql.NullString
		var comment sql.NullString
		var isAutoIncrement bool
		var isPK int64

		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&nullable,
			&defaultValue,
			&comment,
			&isAutoIncrement,
			&isPK,
		); err != nil {
			return nil, err
		}

		col.IsNullable = (nullable == "YES")
		col.IsPrimaryKey = (isPK > 0)
		col.IsAutoIncrement = isAutoIncrement
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
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
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*model.Index)
	for rows.Next() {
		var indexName string
		var nonUnique int
		var columnName string
		if err := rows.Scan(&indexName, &nonUnique, &columnName); err != nil {
			return nil, err
		}

		idx, exists := indexMap[indexName]
		if !exists {
			idx = &model.Index{
				Name:      indexName,
				IsUnique:  (nonUnique == 0),
				IsPrimary: false,
			}
			if idx.IsUnique {
				idx.Type = model.IndexTypeUnique
			} else {
				idx.Type = model.IndexTypeNormal
			}
			indexMap[indexName] = idx
		}
		idx.Columns = append(idx.Columns, columnName)
	}

	for name, idx := range indexMap {
		if name == "PRIMARY" {
			idx.IsPrimary = true
			idx.Type = model.IndexTypePrimary
		}
	}

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
			CONSTRAINT_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			DELETE_RULE
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		  AND REFERENCED_TABLE_NAME IS NOT NULL
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var foreignKeys []model.ForeignKey
	for rows.Next() {
		var fk model.ForeignKey
		if err := rows.Scan(
			&fk.Name,
			&fk.Column,
			&fk.RefTable,
			&fk.RefColumn,
			&fk.OnDelete,
		); err != nil {
			return nil, err
		}
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// GetCheckConstraints 获取表 CHECK 约束列表
func (i *Inspector) GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error) {
	query := `
		SELECT 
			cc.CONSTRAINT_NAME,
			cc.CHECK_CLAUSE,
			cu.COLUMN_NAME
		FROM information_schema.CHECK_CONSTRAINTS cc
		JOIN information_schema.TABLE_CONSTRAINTS tc ON cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
		LEFT JOIN information_schema.KEY_COLUMN_USAGE cu ON cc.CONSTRAINT_NAME = cu.CONSTRAINT_NAME
		WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ?
		ORDER BY cc.CONSTRAINT_NAME
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

	result := make([]model.CheckConstraint, 0, len(constraintMap))
	for _, cc := range constraintMap {
		result = append(result, *cc)
	}

	return result, rows.Err()
}

// getTableComment 获取表注释
func (i *Inspector) getTableComment(ctx context.Context, tableName string) (string, error) {
	query := `SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
	var comment sql.NullString
	if err := i.GetDB().QueryRowContext(ctx, query, tableName).Scan(&comment); err != nil {
		return "", err
	}
	if comment.Valid {
		return comment.String, nil
	}
	return "", nil
}

// Factory MySQL Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	return NewInspector(config), nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "mysql"
}

func init() {
	inspector.Register("mysql", &Factory{})
}
