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
	PlaceholderColon                            // :1, :2 占位符 (Oracle)
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

// QueryTablesInput 查询表输入参数
type QueryTablesInput struct {
	Schema string
}

// QueryTables 查询表列表
func (o *OracleCompatibleInspector) QueryTables(ctx context.Context, input QueryTablesInput) ([]model.Table, error) {
	var query string
	var args []interface{}

	if input.Schema != "" {
		query = `
			SELECT TABLE_NAME, COMMENTS 
			FROM ALL_TAB_COMMENTS 
			WHERE TABLE_TYPE = 'TABLE' AND OWNER = ` + o.placeholderStr(1) + `
			ORDER BY TABLE_NAME
		`
		args = append(args, input.Schema)
	} else {
		query = `
			SELECT TABLE_NAME, COMMENTS 
			FROM USER_TAB_COMMENTS 
			WHERE TABLE_TYPE = 'TABLE'
			ORDER BY TABLE_NAME
		`
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

// QueryColumnsInput 查询字段输入参数
type QueryColumnsInput struct {
	TableName string
	Schema    string
}

// QueryColumns 查询表字段
func (o *OracleCompatibleInspector) QueryColumns(ctx context.Context, input QueryColumnsInput) ([]model.Column, error) {
	var query string
	var args []interface{}

	if input.Schema != "" {
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
				WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.OWNER = ` + o.placeholderStr(2) + ` AND cons.CONSTRAINT_TYPE = 'P'
			) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
			WHERE c.TABLE_NAME = ` + o.placeholderStr(3) + ` AND c.OWNER = ` + o.placeholderStr(4) + `
			ORDER BY c.COLUMN_ID
		`
		args = append(args, input.TableName, input.Schema, input.TableName, input.Schema)
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
				WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.CONSTRAINT_TYPE = 'P'
			) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
			WHERE c.TABLE_NAME = ` + o.placeholderStr(2) + `
			ORDER BY c.COLUMN_ID
		`
		args = append(args, input.TableName, input.TableName)
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

// QueryIndexesInput 查询索引输入参数
type QueryIndexesInput struct {
	TableName string
	Schema    string
}

// QueryIndexes 查询表索引
func (o *OracleCompatibleInspector) QueryIndexes(ctx context.Context, input QueryIndexesInput) ([]model.Index, error) {
	var query string
	var args []interface{}

	if input.Schema != "" {
		query = `
			SELECT 
				i.INDEX_NAME,
				i.UNIQUENESS,
				ic.COLUMN_NAME
			FROM ALL_INDEXES i
			JOIN ALL_IND_COLUMNS ic ON i.OWNER = ic.INDEX_OWNER AND i.INDEX_NAME = ic.INDEX_NAME
			WHERE i.TABLE_NAME = ` + o.placeholderStr(1) + ` AND i.OWNER = ` + o.placeholderStr(2) + `
			ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
		`
		args = append(args, input.TableName, input.Schema)
	} else {
		query = `
			SELECT 
				i.INDEX_NAME,
				i.UNIQUENESS,
				ic.COLUMN_NAME
			FROM USER_INDEXES i
			JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
			WHERE i.TABLE_NAME = ` + o.placeholderStr(1) + `
			ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
		`
		args = append(args, input.TableName)
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
	if input.Schema != "" {
		pkQuery = `
			SELECT cons.CONSTRAINT_NAME
			FROM ALL_CONSTRAINTS cons
			WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.OWNER = ` + o.placeholderStr(2) + ` AND cons.CONSTRAINT_TYPE = 'P'
		`
		pkArgs = append(pkArgs, input.TableName, input.Schema)
	} else {
		pkQuery = `
			SELECT cons.CONSTRAINT_NAME
			FROM USER_CONSTRAINTS cons
			WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.CONSTRAINT_TYPE = 'P'
		`
		pkArgs = append(pkArgs, input.TableName)
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

// QueryForeignKeysInput 查询外键输入参数
type QueryForeignKeysInput struct {
	TableName string
	Schema    string
}

// QueryForeignKeys 查询表外键
func (o *OracleCompatibleInspector) QueryForeignKeys(ctx context.Context, input QueryForeignKeysInput) ([]model.ForeignKey, error) {
	var query string
	var args []interface{}

	if input.Schema != "" {
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
			WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.OWNER = ` + o.placeholderStr(2) + ` AND cons.CONSTRAINT_TYPE = 'R'
		`
		args = append(args, input.TableName, input.Schema)
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
			WHERE cons.TABLE_NAME = ` + o.placeholderStr(1) + ` AND cons.CONSTRAINT_TYPE = 'R'
		`
		args = append(args, input.TableName)
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

// QueryTableCommentInput 查询表注释输入参数
type QueryTableCommentInput struct {
	TableName string
	Schema    string
}

// QueryTableComment 查询表注释
func (o *OracleCompatibleInspector) QueryTableComment(ctx context.Context, input QueryTableCommentInput) (string, error) {
	var query string
	var args []interface{}
	if input.Schema != "" {
		query = `SELECT COMMENTS FROM ALL_TAB_COMMENTS WHERE TABLE_NAME = ` + o.placeholderStr(1) + ` AND OWNER = ` + o.placeholderStr(2) + ``
		args = append(args, input.TableName, input.Schema)
	} else {
		query = `SELECT COMMENTS FROM USER_TAB_COMMENTS WHERE TABLE_NAME = ` + o.placeholderStr(1) + ``
		args = append(args, input.TableName)
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
