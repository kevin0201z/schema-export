package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// PlaceholderType 参数占位符类型
type PlaceholderType int

const (
	PlaceholderQuestion PlaceholderType = iota // ? 占位符 (DM, MySQL)
	PlaceholderColon                           // :1, :2 占位符 (Oracle)
)

// Oracle 兼容数据库 SQL 查询常量
const (
	// 表列表查询
	queryTablesAllSQL = `
		SELECT TABLE_NAME, COMMENTS 
		FROM ALL_TAB_COMMENTS 
		WHERE TABLE_TYPE = 'TABLE' AND OWNER = %s
		ORDER BY TABLE_NAME
	`
	queryTablesUserSQL = `
		SELECT TABLE_NAME, COMMENTS 
		FROM USER_TAB_COMMENTS 
		WHERE TABLE_TYPE = 'TABLE'
		ORDER BY TABLE_NAME
	`

	// 字段查询
	queryColumnsAllSQL = `
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
			WHERE cons.TABLE_NAME = %s AND cons.OWNER = %s AND cons.CONSTRAINT_TYPE = 'P'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = %s AND c.OWNER = %s
		ORDER BY c.COLUMN_ID
	`
	queryColumnsUserSQL = `
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
			WHERE cons.TABLE_NAME = %s AND cons.CONSTRAINT_TYPE = 'P'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = %s
		ORDER BY c.COLUMN_ID
	`

	// 索引查询
	queryIndexesAllSQL = `
		SELECT 
			i.INDEX_NAME,
			i.UNIQUENESS,
			ic.COLUMN_NAME
		FROM ALL_INDEXES i
		JOIN ALL_IND_COLUMNS ic ON i.OWNER = ic.INDEX_OWNER AND i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_NAME = %s AND i.OWNER = %s
		ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
	`
	queryIndexesUserSQL = `
		SELECT 
			i.INDEX_NAME,
			i.UNIQUENESS,
			ic.COLUMN_NAME
		FROM USER_INDEXES i
		JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_NAME = %s
		ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
	`
	queryPKNameAllSQL = `
		SELECT cons.CONSTRAINT_NAME
		FROM ALL_CONSTRAINTS cons
		WHERE cons.TABLE_NAME = %s AND cons.OWNER = %s AND cons.CONSTRAINT_TYPE = 'P'
	`
	queryPKNameUserSQL = `
		SELECT cons.CONSTRAINT_NAME
		FROM USER_CONSTRAINTS cons
		WHERE cons.TABLE_NAME = %s AND cons.CONSTRAINT_TYPE = 'P'
	`

	// 外键查询
	queryForeignKeysAllSQL = `
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
		WHERE cons.TABLE_NAME = %s AND cons.OWNER = %s AND cons.CONSTRAINT_TYPE = 'R'
	`
	queryForeignKeysUserSQL = `
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
		WHERE cons.TABLE_NAME = %s AND cons.CONSTRAINT_TYPE = 'R'
	`

	// 表注释查询
	queryTableCommentAllSQL  = `SELECT COMMENTS FROM ALL_TAB_COMMENTS WHERE TABLE_NAME = %s AND OWNER = %s`
	queryTableCommentUserSQL = `SELECT COMMENTS FROM USER_TAB_COMMENTS WHERE TABLE_NAME = %s`
)

// OracleCompatibleInspector Oracle 兼容数据库 Inspector 基础实现
type OracleCompatibleInspector struct {
	*BaseInspector
	placeholder PlaceholderType
}

// NewOracleCompatibleInspector 创建 Oracle 兼容 Inspector
func NewOracleCompatibleInspector(config inspector.ConnectionConfig, placeholder PlaceholderType) *OracleCompatibleInspector {
	return &OracleCompatibleInspector{
		BaseInspector: NewBaseInspector(config),
		placeholder:   placeholder,
	}
}

// GetTables 获取所有表列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetTables(ctx context.Context) ([]model.Table, error) {
	config := o.GetConfig()
	return o.queryTables(ctx, config.Schema)
}

// GetTable 获取单个表的完整元数据（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetTable(ctx context.Context, tableName string) (*model.Table, error) {
	config := o.GetConfig()
	schema := config.Schema

	table := &model.Table{
		Name: tableName,
		Type: model.TableTypeTable,
	}

	// 获取表注释
	comment, _ := o.queryTableComment(ctx, tableName, schema)
	table.Comment = comment

	// 获取字段
	columns, err := o.GetColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	// 获取索引
	indexes, err := o.GetIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Indexes = indexes

	// 获取外键
	foreignKeys, err := o.GetForeignKeys(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.ForeignKeys = foreignKeys

	return table, nil
}

// GetColumns 获取表字段列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetColumns(ctx context.Context, tableName string) ([]model.Column, error) {
	config := o.GetConfig()
	return o.queryColumns(ctx, tableName, config.Schema)
}

// GetIndexes 获取表索引列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetIndexes(ctx context.Context, tableName string) ([]model.Index, error) {
	config := o.GetConfig()
	return o.queryIndexes(ctx, tableName, config.Schema)
}

// GetForeignKeys 获取表外键列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error) {
	config := o.GetConfig()
	return o.queryForeignKeys(ctx, tableName, config.Schema)
}

// QueryInput 通用查询输入参数
type QueryInput struct {
	TableName string // 可选，表名
	Schema    string // 可选，Schema
}

// queryTables 查询表列表
func (o *OracleCompatibleInspector) queryTables(ctx context.Context, schema string) ([]model.Table, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryTablesAllSQL, o.placeholderStr(1))
		args = append(args, schema)
	} else {
		query = queryTablesUserSQL
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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

// queryColumns 查询表字段
func (o *OracleCompatibleInspector) queryColumns(ctx context.Context, tableName, schema string) ([]model.Column, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryColumnsAllSQL,
			o.placeholderStr(1), o.placeholderStr(2),
			o.placeholderStr(3), o.placeholderStr(4))
		args = append(args, tableName, schema, tableName, schema)
	} else {
		query = fmt.Sprintf(queryColumnsUserSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, tableName)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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

// queryIndexes 查询表索引
func (o *OracleCompatibleInspector) queryIndexes(ctx context.Context, tableName, schema string) ([]model.Index, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryIndexesAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, schema)
	} else {
		query = fmt.Sprintf(queryIndexesUserSQL, o.placeholderStr(1))
		args = append(args, tableName)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
		pkQuery = fmt.Sprintf(queryPKNameAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		pkArgs = append(pkArgs, tableName, schema)
	} else {
		pkQuery = fmt.Sprintf(queryPKNameUserSQL, o.placeholderStr(1))
		pkArgs = append(pkArgs, tableName)
	}

	var pkName string
	if err := o.GetDB().QueryRowContext(ctx, pkQuery, pkArgs...).Scan(&pkName); err == nil {
		if idx, exists := indexMap[pkName]; exists {
			idx.IsPrimary = true
			idx.Type = model.IndexTypePrimary
		}
	}

	// 按名称排序，确保输出稳定
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

// queryForeignKeys 查询表外键
func (o *OracleCompatibleInspector) queryForeignKeys(ctx context.Context, tableName, schema string) ([]model.ForeignKey, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryForeignKeysAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, schema)
	} else {
		query = fmt.Sprintf(queryForeignKeysUserSQL, o.placeholderStr(1))
		args = append(args, tableName)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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

// queryTableComment 查询表注释
func (o *OracleCompatibleInspector) queryTableComment(ctx context.Context, tableName, schema string) (string, error) {
	var query string
	var args []interface{}
	if schema != "" {
		query = fmt.Sprintf(queryTableCommentAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, schema)
	} else {
		query = fmt.Sprintf(queryTableCommentUserSQL, o.placeholderStr(1))
		args = append(args, tableName)
	}
	var comment sql.NullString
	if err := o.GetDB().QueryRowContext(ctx, query, args...).Scan(&comment); err != nil {
		return "", err
	}
	if comment.Valid {
		return comment.String, nil
	}
	return "", nil
}

// placeholderStr 获取占位符字符串
func (o *OracleCompatibleInspector) placeholderStr(index int) string {
	if o.placeholder == PlaceholderColon {
		return fmt.Sprintf(":%d", index)
	}
	return "?"
}
