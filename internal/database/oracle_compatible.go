package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

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

	// CHECK 约束查询
	queryCheckConstraintsAllSQL = `
		SELECT 
			cons.CONSTRAINT_NAME,
			cons.SEARCH_CONDITION,
			cc.COLUMN_NAME
		FROM ALL_CONSTRAINTS cons
		LEFT JOIN ALL_CONS_COLUMNS cc ON cons.OWNER = cc.OWNER AND cons.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE cons.TABLE_NAME = %s AND cons.OWNER = %s AND cons.CONSTRAINT_TYPE = 'C'
		ORDER BY cons.CONSTRAINT_NAME
	`
	queryCheckConstraintsUserSQL = `
		SELECT 
			cons.CONSTRAINT_NAME,
			cons.SEARCH_CONDITION,
			cc.COLUMN_NAME
		FROM USER_CONSTRAINTS cons
		LEFT JOIN USER_CONS_COLUMNS cc ON cons.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE cons.TABLE_NAME = %s AND cons.CONSTRAINT_TYPE = 'C'
		ORDER BY cons.CONSTRAINT_NAME
	`

	// 视图查询
	queryViewsAllSQL = `
		SELECT v.VIEW_NAME, c.COMMENTS, v.TEXT
		FROM ALL_VIEWS v
		LEFT JOIN ALL_TAB_COMMENTS c ON v.OWNER = c.OWNER AND v.VIEW_NAME = c.TABLE_NAME AND c.TABLE_TYPE = 'VIEW'
		WHERE v.OWNER = %s
		ORDER BY v.VIEW_NAME
	`
	queryViewsUserSQL = `
		SELECT v.VIEW_NAME, c.COMMENTS, v.TEXT
		FROM USER_VIEWS v
		LEFT JOIN USER_TAB_COMMENTS c ON v.VIEW_NAME = c.TABLE_NAME AND c.TABLE_TYPE = 'VIEW'
		ORDER BY v.VIEW_NAME
	`

	// 存储过程查询
	queryProceduresAllSQL = `
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM ALL_PROCEDURES p
		LEFT JOIN ALL_TAB_COMMENTS c ON p.OWNER = c.OWNER AND p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OWNER = %s AND p.OBJECT_TYPE = 'PROCEDURE'
		ORDER BY p.OBJECT_NAME
	`
	queryProceduresUserSQL = `
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM USER_PROCEDURES p
		LEFT JOIN USER_TAB_COMMENTS c ON p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OBJECT_TYPE = 'PROCEDURE'
		ORDER BY p.OBJECT_NAME
	`

	// 函数查询
	queryFunctionsAllSQL = `
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM ALL_PROCEDURES p
		LEFT JOIN ALL_TAB_COMMENTS c ON p.OWNER = c.OWNER AND p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OWNER = %s AND p.OBJECT_TYPE = 'FUNCTION'
		ORDER BY p.OBJECT_NAME
	`
	queryFunctionsUserSQL = `
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM USER_PROCEDURES p
		LEFT JOIN USER_TAB_COMMENTS c ON p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OBJECT_TYPE = 'FUNCTION'
		ORDER BY p.OBJECT_NAME
	`

	// 源代码查询
	querySourceAllSQL = `
		SELECT TEXT FROM ALL_SOURCE 
		WHERE OWNER = %s AND NAME = %s AND TYPE = %s
		ORDER BY LINE
	`
	querySourceUserSQL = `
		SELECT TEXT FROM USER_SOURCE 
		WHERE NAME = %s AND TYPE = %s
		ORDER BY LINE
	`

	// 触发器查询
	queryTriggersAllSQL = `
		SELECT 
			t.TRIGGER_NAME,
			t.TABLE_NAME,
			t.TRIGGERING_EVENT,
			t.STATUS,
			t.TRIGGER_TYPE
		FROM ALL_TRIGGERS t
		WHERE t.TABLE_NAME = %s AND t.OWNER = %s
		ORDER BY t.TRIGGER_NAME
	`
	queryTriggersUserSQL = `
		SELECT 
			t.TRIGGER_NAME,
			t.TABLE_NAME,
			t.TRIGGERING_EVENT,
			t.STATUS,
			t.TRIGGER_TYPE
		FROM USER_TRIGGERS t
		WHERE t.TABLE_NAME = %s
		ORDER BY t.TRIGGER_NAME
	`

	// 序列查询
	querySequencesAllSQL = `
		SELECT 
			SEQUENCE_NAME,
			MIN_VALUE,
			MAX_VALUE,
			INCREMENT_BY,
			CYCLE_FLAG,
			CACHE_SIZE,
			LAST_NUMBER
		FROM ALL_SEQUENCES
		WHERE SEQUENCE_OWNER = %s
		ORDER BY SEQUENCE_NAME
	`
	querySequencesUserSQL = `
		SELECT 
			SEQUENCE_NAME,
			MIN_VALUE,
			MAX_VALUE,
			INCREMENT_BY,
			CYCLE_FLAG,
			CACHE_SIZE,
			LAST_NUMBER
		FROM USER_SEQUENCES
		ORDER BY SEQUENCE_NAME
	`
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

	// 获取 CHECK 约束
	checkConstraints, err := o.GetCheckConstraints(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.CheckConstraints = checkConstraints

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

// GetCheckConstraints 获取表 CHECK 约束列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error) {
	config := o.GetConfig()
	return o.queryCheckConstraints(ctx, tableName, config.Schema)
}

// GetViews 获取视图列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetViews(ctx context.Context) ([]model.View, error) {
	config := o.GetConfig()
	return o.queryViews(ctx, config.Schema)
}

// GetProcedures 获取存储过程列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetProcedures(ctx context.Context) ([]model.Procedure, error) {
	config := o.GetConfig()
	return o.queryProcedures(ctx, config.Schema)
}

// GetFunctions 获取函数列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetFunctions(ctx context.Context) ([]model.Function, error) {
	config := o.GetConfig()
	return o.queryFunctions(ctx, config.Schema)
}

// GetTriggers 获取触发器列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetTriggers(ctx context.Context, tableName string) ([]model.Trigger, error) {
	config := o.GetConfig()
	return o.queryTriggers(ctx, tableName, config.Schema)
}

// GetSequences 获取序列列表（实现 inspector.Inspector 接口）
func (o *OracleCompatibleInspector) GetSequences(ctx context.Context) ([]model.Sequence, error) {
	config := o.GetConfig()
	return o.querySequences(ctx, config.Schema)
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

// queryCheckConstraints 查询表 CHECK 约束
func (o *OracleCompatibleInspector) queryCheckConstraints(ctx context.Context, tableName, schema string) ([]model.CheckConstraint, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryCheckConstraintsAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, schema)
	} else {
		query = fmt.Sprintf(queryCheckConstraintsUserSQL, o.placeholderStr(1))
		args = append(args, tableName)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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

// placeholderStr 获取占位符字符串
func (o *OracleCompatibleInspector) placeholderStr(index int) string {
	if o.placeholder == PlaceholderColon {
		return fmt.Sprintf(":%d", index)
	}
	return "?"
}

// queryViews 查询视图列表
func (o *OracleCompatibleInspector) queryViews(ctx context.Context, schema string) ([]model.View, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryViewsAllSQL, o.placeholderStr(1))
		args = append(args, schema)
	} else {
		query = queryViewsUserSQL
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []model.View
	for rows.Next() {
		var view model.View
		var comment sql.NullString
		var text sql.NullString
		if err := rows.Scan(&view.Name, &comment, &text); err != nil {
			return nil, err
		}
		if comment.Valid {
			view.Comment = comment.String
		}
		if text.Valid {
			view.Definition = text.String
		}
		views = append(views, view)
	}

	return views, rows.Err()
}

// queryProcedures 查询存储过程列表
func (o *OracleCompatibleInspector) queryProcedures(ctx context.Context, schema string) ([]model.Procedure, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryProceduresAllSQL, o.placeholderStr(1))
		args = append(args, schema)
	} else {
		query = queryProceduresUserSQL
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var procedures []model.Procedure
	for rows.Next() {
		var proc model.Procedure
		var comment sql.NullString
		if err := rows.Scan(&proc.Name, &comment); err != nil {
			return nil, err
		}
		if comment.Valid {
			proc.Comment = comment.String
		}
		procedures = append(procedures, proc)
	}

	for i := range procedures {
		def, err := o.querySource(ctx, procedures[i].Name, "PROCEDURE", schema)
		if err == nil {
			procedures[i].Definition = def
		}
	}

	return procedures, rows.Err()
}

// queryFunctions 查询函数列表
func (o *OracleCompatibleInspector) queryFunctions(ctx context.Context, schema string) ([]model.Function, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryFunctionsAllSQL, o.placeholderStr(1))
		args = append(args, schema)
	} else {
		query = queryFunctionsUserSQL
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var functions []model.Function
	for rows.Next() {
		var fn model.Function
		var comment sql.NullString
		if err := rows.Scan(&fn.Name, &comment); err != nil {
			return nil, err
		}
		if comment.Valid {
			fn.Comment = comment.String
		}
		functions = append(functions, fn)
	}

	for i := range functions {
		def, err := o.querySource(ctx, functions[i].Name, "FUNCTION", schema)
		if err == nil {
			functions[i].Definition = def
		}
	}

	return functions, rows.Err()
}

// querySource 查询源代码
func (o *OracleCompatibleInspector) querySource(ctx context.Context, name, objType, schema string) (string, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(querySourceAllSQL, o.placeholderStr(1), o.placeholderStr(2), o.placeholderStr(3))
		args = append(args, schema, name, objType)
	} else {
		query = fmt.Sprintf(querySourceUserSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, name, objType)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var source string
	for rows.Next() {
		var text sql.NullString
		if err := rows.Scan(&text); err != nil {
			return "", err
		}
		if text.Valid {
			source += text.String
		}
	}

	return source, rows.Err()
}

// queryTriggers 查询触发器列表
func (o *OracleCompatibleInspector) queryTriggers(ctx context.Context, tableName, schema string) ([]model.Trigger, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(queryTriggersAllSQL, o.placeholderStr(1), o.placeholderStr(2))
		args = append(args, tableName, schema)
	} else {
		query = fmt.Sprintf(queryTriggersUserSQL, o.placeholderStr(1))
		args = append(args, tableName)
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []model.Trigger
	for rows.Next() {
		var name, tblName, event, status, triggerType string
		if err := rows.Scan(&name, &tblName, &event, &status, &triggerType); err != nil {
			return nil, err
		}

		tr := model.Trigger{
			Name:      name,
			TableName: tblName,
			Event:     event,
			Status:    status,
			Timing:    parseTriggerTiming(triggerType),
		}
		triggers = append(triggers, tr)
	}

	for i := range triggers {
		def, err := o.querySource(ctx, triggers[i].Name, "TRIGGER", schema)
		if err == nil && def != "" {
			triggers[i].Definition = def
		}
	}

	return triggers, rows.Err()
}

// parseTriggerTiming 从 TRIGGER_TYPE 解析触发时机
func parseTriggerTiming(triggerType string) string {
	if triggerType == "" {
		return ""
	}
	upperType := strings.ToUpper(triggerType)
	if strings.HasPrefix(upperType, "BEFORE") {
		return "BEFORE"
	}
	if strings.HasPrefix(upperType, "AFTER") {
		return "AFTER"
	}
	if strings.HasPrefix(upperType, "INSTEAD OF") {
		return "INSTEAD OF"
	}
	return ""
}

// querySequences 查询序列列表
func (o *OracleCompatibleInspector) querySequences(ctx context.Context, schema string) ([]model.Sequence, error) {
	var query string
	var args []interface{}

	if schema != "" {
		query = fmt.Sprintf(querySequencesAllSQL, o.placeholderStr(1))
		args = append(args, schema)
	} else {
		query = querySequencesUserSQL
	}

	rows, err := o.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sequences []model.Sequence
	for rows.Next() {
		var seq model.Sequence
		var minVal, maxVal, incrBy, cacheSize, lastNum sql.NullInt64
		var cycleFlag string
		if err := rows.Scan(
			&seq.Name,
			&minVal,
			&maxVal,
			&incrBy,
			&cycleFlag,
			&cacheSize,
			&lastNum,
		); err != nil {
			return nil, err
		}
		if minVal.Valid {
			seq.MinValue = minVal.Int64
		}
		if maxVal.Valid {
			seq.MaxValue = maxVal.Int64
		}
		if incrBy.Valid {
			seq.IncrementBy = incrBy.Int64
		}
		if cacheSize.Valid {
			seq.CacheSize = cacheSize.Int64
		}
		if lastNum.Valid {
			seq.LastValue = lastNum.Int64
		}
		seq.Cycle = (cycleFlag == "Y" || cycleFlag == "YES")
		sequences = append(sequences, seq)
	}

	return sequences, rows.Err()
}
