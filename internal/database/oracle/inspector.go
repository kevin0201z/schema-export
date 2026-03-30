package oracle

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/godror/godror" // Oracle 驱动
	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector Oracle 数据库 Inspector 实现
type Inspector struct {
	*database.BaseInspector
}

// NewInspector 创建 Oracle Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		BaseInspector: database.NewBaseInspector(config),
	}
}

// Connect 连接 Oracle 数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()
	db, err := sql.Open("godror", dsn)
	if err != nil {
		return fmt.Errorf("failed to open oracle connection: %w", err)
	}
	
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping oracle database: %w", err)
	}
	
	i.SetDB(db)
	return nil
}

// BuildDSN 构建 Oracle DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		return config.DSN
	}
	
	// Oracle DSN 格式: user/password@host:port/service_name
	serviceName := config.Database
	if serviceName == "" {
		serviceName = config.Schema
	}
	
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		serviceName,
	)
	
	return dsn
}

// GetTables 获取所有表列表
func (i *Inspector) GetTables(ctx context.Context) ([]model.Table, error) {
	query := `
		SELECT TABLE_NAME, COMMENTS 
		FROM USER_TAB_COMMENTS 
		WHERE TABLE_TYPE = 'TABLE'
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
	
	// 获取表注释
	commentQuery := `SELECT COMMENTS FROM USER_TAB_COMMENTS WHERE TABLE_NAME = :1`
	var comment sql.NullString
	if err := i.GetDB().QueryRowContext(ctx, commentQuery, tableName).Scan(&comment); err == nil && comment.Valid {
		table.Comment = comment.String
	}
	
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
	
	return table, nil
}

// GetColumns 获取表字段列表
func (i *Inspector) GetColumns(ctx context.Context, tableName string) ([]model.Column, error) {
	query := `
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.DATA_LENGTH,
			c.DATA_PRECISION,
			c.DATA_SCALE,
			c.NULLABLE,
			c.DATA_DEFAULT,
			cc.COMMENTS,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as IS_PK
		FROM USER_TAB_COLUMNS c
		LEFT JOIN USER_COL_COMMENTS cc ON c.TABLE_NAME = cc.TABLE_NAME AND c.COLUMN_NAME = cc.COLUMN_NAME
		LEFT JOIN (
			SELECT col.COLUMN_NAME
			FROM USER_CONSTRAINTS cons
			JOIN USER_CONS_COLUMNS col ON cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
			WHERE cons.TABLE_NAME = :1 AND cons.CONSTRAINT_TYPE = 'P'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = :2
		ORDER BY c.COLUMN_ID
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
		var isPK int
		var precision, scale sql.NullInt64

		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.Length,
			&precision,
			&scale,
			&nullable,
			&defaultValue,
			&comment,
			&isPK,
		); err != nil {
			return nil, err
		}

		if precision.Valid {
			col.Precision = int(precision.Int64)
		}
		if scale.Valid {
			col.Scale = int(scale.Int64)
		}
		col.IsNullable = (nullable == "Y")
		col.IsPrimaryKey = (isPK == 1)
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
		SELECT 
			i.INDEX_NAME,
			i.UNIQUENESS,
			ic.COLUMN_NAME
		FROM USER_INDEXES i
		JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_NAME = :1
		ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
	`
	
	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()
	
	indexMap := make(map[string]*model.Index)
	for rows.Next() {
		var indexName, uniqueness, columnName string
		if err := rows.Scan(&indexName, &uniqueness, &columnName); err != nil {
			return nil, err
		}
		
		idx, exists := indexMap[indexName]
		if !exists {
			idx = &model.Index{
				Name:      indexName,
				IsUnique:  (uniqueness == "UNIQUE"),
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
	
	// 检查主键索引
	pkQuery := `
		SELECT cons.CONSTRAINT_NAME
		FROM USER_CONSTRAINTS cons
		WHERE cons.TABLE_NAME = :1 AND cons.CONSTRAINT_TYPE = 'P'
	`
	var pkName string
	if err := i.GetDB().QueryRowContext(ctx, pkQuery, tableName).Scan(&pkName); err == nil {
		if idx, exists := indexMap[pkName]; exists {
			idx.IsPrimary = true
			idx.Type = model.IndexTypePrimary
		}
	}
	
	indexes := make([]model.Index, 0, len(indexMap))
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	
	return indexes, rows.Err()
}

// GetForeignKeys 获取表外键列表
func (i *Inspector) GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error) {
	query := `
		SELECT 
			cons.CONSTRAINT_NAME,
			col.COLUMN_NAME,
			refCons.TABLE_NAME as REF_TABLE,
			refCol.COLUMN_NAME as REF_COLUMN,
			cons.DELETE_RULE
		FROM USER_CONSTRAINTS cons
		JOIN USER_CONS_COLUMNS col ON cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
		JOIN USER_CONSTRAINTS refCons ON cons.R_CONSTRAINT_NAME = refCons.CONSTRAINT_NAME
		JOIN USER_CONS_COLUMNS refCol ON refCons.CONSTRAINT_NAME = refCol.CONSTRAINT_NAME AND col.POSITION = refCol.POSITION
		WHERE cons.TABLE_NAME = :1 AND cons.CONSTRAINT_TYPE = 'R'
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

// Factory Oracle Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	ins := NewInspector(config)
	return ins, nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "oracle"
}

func init() {
	inspector.Register("oracle", &Factory{})
}
