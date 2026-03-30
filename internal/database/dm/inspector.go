package dm

import (
	"context"
	"database/sql"
	"fmt"

	_ "gitee.com/chunanyong/dm" // 达梦 Go 驱动
	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector 达梦数据库 Inspector 实现
// 使用 dm-go-driver 纯 Go 驱动，无需安装 ODBC
type Inspector struct {
	*database.BaseInspector
}

// NewInspector 创建达梦 Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		BaseInspector: database.NewBaseInspector(config),
	}
}

// Connect 连接达梦数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	// 使用 dm-go-driver 连接
	db, err := sql.Open("dm", dsn)
	if err != nil {
		return fmt.Errorf("failed to open dm connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping dm database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建达梦 DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		return config.DSN
	}

	// 达梦 DSN 格式: dm://user:password@host:port?schema=schema_name
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	if config.Schema != "" {
		dsn = dsn + "?schema=" + config.Schema
	}

	return dsn
}

// GetTables 获取所有表列表
func (i *Inspector) GetTables(ctx context.Context) ([]model.Table, error) {
	config := i.GetConfig()
	schema := config.Schema

	var query string
	var args []interface{}

	if schema != "" {
		// 查询指定 schema 的表
		query = `
			SELECT TABLE_NAME, COMMENTS 
			FROM ALL_TAB_COMMENTS 
			WHERE TABLE_TYPE = 'TABLE' AND OWNER = ?
			ORDER BY TABLE_NAME
		`
		args = append(args, schema)
	} else {
		// 查询当前用户的表
		query = `
			SELECT TABLE_NAME, COMMENTS 
			FROM USER_TAB_COMMENTS 
			WHERE TABLE_TYPE = 'TABLE'
			ORDER BY TABLE_NAME
		`
	}

	rows, err := i.GetDB().QueryContext(ctx, query, args...)
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
	config := i.GetConfig()
	schema := config.Schema

	table := &model.Table{
		Name: tableName,
		Type: model.TableTypeTable,
	}

	// 获取表注释
	var commentQuery string
	var commentArgs []interface{}
	if schema != "" {
		commentQuery = `SELECT COMMENTS FROM ALL_TAB_COMMENTS WHERE TABLE_NAME = ? AND OWNER = ?`
		commentArgs = append(commentArgs, tableName, schema)
	} else {
		commentQuery = `SELECT COMMENTS FROM USER_TAB_COMMENTS WHERE TABLE_NAME = ?`
		commentArgs = append(commentArgs, tableName)
	}
	var comment sql.NullString
	if err := i.GetDB().QueryRowContext(ctx, commentQuery, commentArgs...).Scan(&comment); err == nil && comment.Valid {
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
	config := i.GetConfig()
	schema := config.Schema

	var query string
	var args []interface{}

	if schema != "" {
		query = `
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
			FROM ALL_TAB_COLUMNS c
			LEFT JOIN ALL_COL_COMMENTS cc ON c.OWNER = cc.OWNER AND c.TABLE_NAME = cc.TABLE_NAME AND c.COLUMN_NAME = cc.COLUMN_NAME
			LEFT JOIN (
				SELECT col.COLUMN_NAME
				FROM ALL_CONSTRAINTS cons
				JOIN ALL_CONS_COLUMNS col ON cons.OWNER = col.OWNER AND cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
				WHERE cons.TABLE_NAME = ? AND cons.OWNER = ? AND cons.CONSTRAINT_TYPE = 'P'
			) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
			WHERE c.TABLE_NAME = ? AND c.OWNER = ?
			ORDER BY c.COLUMN_ID
		`
		args = append(args, tableName, schema, tableName, schema)
	} else {
		query = `
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
				WHERE cons.TABLE_NAME = ? AND cons.CONSTRAINT_TYPE = 'P'
			) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
			WHERE c.TABLE_NAME = ?
			ORDER BY c.COLUMN_ID
		`
		args = append(args, tableName, tableName)
	}

	rows, err := i.GetDB().QueryContext(ctx, query, args...)
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
	config := i.GetConfig()
	schema := config.Schema
	
	var query string
	var args []interface{}
	
	if schema != "" {
		query = `
			SELECT 
				i.INDEX_NAME,
				i.UNIQUENESS,
				ic.COLUMN_NAME
			FROM ALL_INDEXES i
			JOIN ALL_IND_COLUMNS ic ON i.OWNER = ic.INDEX_OWNER AND i.INDEX_NAME = ic.INDEX_NAME
			WHERE i.TABLE_NAME = ? AND i.OWNER = ?
			ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
		`
		args = append(args, tableName, schema)
	} else {
		query = `
			SELECT 
				i.INDEX_NAME,
				i.UNIQUENESS,
				ic.COLUMN_NAME
			FROM USER_INDEXES i
			JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
			WHERE i.TABLE_NAME = ?
			ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
		`
		args = append(args, tableName)
	}
	
	rows, err := i.GetDB().QueryContext(ctx, query, args...)
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
	var pkQuery string
	var pkArgs []interface{}
	if schema != "" {
		pkQuery = `
			SELECT cons.CONSTRAINT_NAME
			FROM ALL_CONSTRAINTS cons
			WHERE cons.TABLE_NAME = ? AND cons.OWNER = ? AND cons.CONSTRAINT_TYPE = 'P'
		`
		pkArgs = append(pkArgs, tableName, schema)
	} else {
		pkQuery = `
			SELECT cons.CONSTRAINT_NAME
			FROM USER_CONSTRAINTS cons
			WHERE cons.TABLE_NAME = ? AND cons.CONSTRAINT_TYPE = 'P'
		`
		pkArgs = append(pkArgs, tableName)
	}
	
	var pkName string
	if err := i.GetDB().QueryRowContext(ctx, pkQuery, pkArgs...).Scan(&pkName); err == nil {
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
	config := i.GetConfig()
	schema := config.Schema
	
	var query string
	var args []interface{}
	
	if schema != "" {
		query = `
			SELECT 
				cons.CONSTRAINT_NAME,
				col.COLUMN_NAME,
				refCons.TABLE_NAME as REF_TABLE,
				refCol.COLUMN_NAME as REF_COLUMN,
				cons.DELETE_RULE
			FROM ALL_CONSTRAINTS cons
			JOIN ALL_CONS_COLUMNS col ON cons.OWNER = col.OWNER AND cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
			JOIN ALL_CONSTRAINTS refCons ON cons.R_OWNER = refCons.OWNER AND cons.R_CONSTRAINT_NAME = refCons.CONSTRAINT_NAME
			JOIN ALL_CONS_COLUMNS refCol ON refCons.OWNER = refCol.OWNER AND refCons.CONSTRAINT_NAME = refCol.CONSTRAINT_NAME AND col.POSITION = refCol.POSITION
			WHERE cons.TABLE_NAME = ? AND cons.OWNER = ? AND cons.CONSTRAINT_TYPE = 'R'
		`
		args = append(args, tableName, schema)
	} else {
		query = `
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
			WHERE cons.TABLE_NAME = ? AND cons.CONSTRAINT_TYPE = 'R'
		`
		args = append(args, tableName)
	}
	
	rows, err := i.GetDB().QueryContext(ctx, query, args...)
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

// Factory 达梦 Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	ins := NewInspector(config)
	return ins, nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "dm"
}

func init() {
	inspector.Register("dm", &Factory{})
}
