package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	pq "github.com/lib/pq" // PostgreSQL Go 驱动

	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector PostgreSQL 数据库 Inspector 实现
type Inspector struct {
	*database.BaseInspector
}

// NewInspector 创建 PostgreSQL Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		BaseInspector: database.NewBaseInspector(config),
	}
}

// Connect 连接 PostgreSQL 数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(database.DefaultMaxOpenConns)
	db.SetMaxIdleConns(database.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(database.DefaultConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, database.DefaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping postgres database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建 PostgreSQL DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		dsn := config.DSN
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			return dsn
		}
		if !strings.Contains(dsn, "://") {
			dsn = "postgres://" + dsn
		}
		return dsn
	}

	params := []string{"sslmode=disable"}
	if config.SSLMode != "" {
		params = []string{"sslmode=" + config.SSLMode}
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s",
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
		SELECT table_name, 
			   COALESCE(obj_description((table_schema || '.' || table_name)::regclass, 'pg_class'), '')
		FROM information_schema.tables 
		WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'
		ORDER BY table_name
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
		if comment.Valid && comment.String != "" {
			table.Comment = comment.String
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetTable 获取表详细信息
func (i *Inspector) GetTable(ctx context.Context, tableName string) (*model.Table, error) {
	query := `
		SELECT table_name,
			   COALESCE(obj_description((table_schema || '.' || table_name)::regclass, 'pg_class'), '')
		FROM information_schema.tables 
		WHERE table_schema = current_schema() AND table_name = $1
	`

	var table model.Table
	var comment sql.NullString
	err := i.GetDB().QueryRowContext(ctx, query, tableName).Scan(&table.Name, &comment)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("table %s not found", tableName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query table %s: %w", tableName, err)
	}

	table.Type = model.TableTypeTable
	if comment.Valid && comment.String != "" {
		table.Comment = comment.String
	}

	columns, err := i.GetColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	indexes, err := i.GetIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	sort.Slice(indexes, func(i, j int) bool { return indexes[i].Name < indexes[j].Name })
	table.Indexes = indexes
	for _, idx := range indexes {
		if !idx.IsPrimary {
			continue
		}
		for _, pkColumn := range idx.Columns {
			for i := range table.Columns {
				if table.Columns[i].Name == pkColumn {
					table.Columns[i].IsPrimaryKey = true
				}
			}
		}
	}

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

	return &table, nil
}

// GetColumns 获取表字段信息
func (i *Inspector) GetColumns(ctx context.Context, tableName string) ([]model.Column, error) {
	query := `
		SELECT column_name,
			   data_type,
			   COALESCE(character_maximum_length, 0),
			   COALESCE(numeric_precision, 0),
			   COALESCE(numeric_scale, 0),
			   CASE WHEN is_nullable = 'YES' THEN false ELSE true END,
			   COALESCE(column_default, ''),
			   COALESCE(col_description((table_schema || '.' || table_name)::regclass, ordinal_position), '')
		FROM information_schema.columns 
		WHERE table_schema = current_schema() AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []model.Column
	for rows.Next() {
		var col model.Column
		var charLen, precision, scale int64
		var defaultVal, comment sql.NullString
		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&charLen,
			&precision,
			&scale,
			&col.IsNullable,
			&defaultVal,
			&comment,
		); err != nil {
			return nil, err
		}

		col.Length = int(charLen)
		col.Precision = int(precision)
		col.Scale = int(scale)

		if defaultVal.Valid && defaultVal.String != "" {
			defaultValStr := defaultVal.String
			if strings.HasPrefix(defaultValStr, "nextval('") {
				col.IsAutoIncrement = true
				defaultValStr = ""
			}
			col.DefaultValue = defaultValStr
		}

		if comment.Valid && comment.String != "" {
			col.Comment = comment.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// GetIndexes 获取索引信息
func (i *Inspector) GetIndexes(ctx context.Context, tableName string) ([]model.Index, error) {
	query := `
		SELECT i.relname AS index_name,
			   am.amname AS index_type,
			   ix.indisunique AS is_unique,
			   ix.indisprimary AS is_primary,
			   array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) AS columns
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = i.relam
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE t.relname = $1 AND n.nspname = current_schema()
		GROUP BY i.relname, am.amname, ix.indisunique, ix.indisprimary
		ORDER BY i.relname
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes for table %s: %w", tableName, err)
	}
	defer rows.Close()

	indexMap := make(map[string]*model.Index)
	for rows.Next() {
		var name, indexType string
		var isUnique, isPrimary bool
		var columns pq.StringArray
		if err := rows.Scan(&name, &indexType, &isUnique, &isPrimary, &columns); err != nil {
			return nil, err
		}

		idx, exists := indexMap[name]
		if !exists {
			idx = &model.Index{
				Name:      name,
				Type:      model.IndexType(indexType),
				IsUnique:  isUnique,
				IsPrimary: isPrimary,
				Columns:   []string{},
			}
			indexMap[name] = idx
		}
		for _, col := range columns {
			found := false
			for _, c := range idx.Columns {
				if c == col {
					found = true
					break
				}
			}
			if !found {
				idx.Columns = append(idx.Columns, col)
			}
		}
	}

	result := make([]model.Index, 0, len(indexMap))
	for _, idx := range indexMap {
		result = append(result, *idx)
	}
	return result, rows.Err()
}

// GetForeignKeys 获取外键信息
func (i *Inspector) GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error) {
	query := `
		SELECT tc.constraint_name,
			   kcu.column_name,
			   ccu.table_name AS foreign_table,
			   ccu.column_name AS foreign_column,
			   rc.delete_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
		JOIN information_schema.referential_constraints rc ON rc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		  AND tc.table_schema = current_schema() 
		  AND tc.table_name = $1
		ORDER BY tc.constraint_name
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var fks []model.ForeignKey
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
		if onDelete.Valid {
			fk.OnDelete = onDelete.String
		}
		fks = append(fks, fk)
	}

	return fks, rows.Err()
}

// GetCheckConstraints 获取 CHECK 约束
func (i *Inspector) GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error) {
	query := `
		SELECT con.conname AS constraint_name,
			   pg_get_constraintdef(con.oid) AS definition,
			   array_agg(att.attname ORDER BY array_position(con.conkey, att.attnum)) FILTER (WHERE att.attname IS NOT NULL) AS columns
		FROM pg_constraint con
		JOIN pg_class rel ON rel.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
		LEFT JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
		WHERE con.contype = 'c' 
		  AND rel.relname = $1 
		  AND nsp.nspname = current_schema()
		GROUP BY con.conname, pg_get_constraintdef(con.oid)
		ORDER BY con.conname
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query check constraints for table %s: %w", tableName, err)
	}
	defer rows.Close()

	constraintMap := make(map[string]*model.CheckConstraint)
	for rows.Next() {
		var name, definition string
		var columns pq.StringArray
		if err := rows.Scan(&name, &definition, &columns); err != nil {
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
		for _, col := range columns {
			if col != "" {
				cc.Columns = append(cc.Columns, col)
			}
		}
	}

	result := make([]model.CheckConstraint, 0, len(constraintMap))
	for _, cc := range constraintMap {
		result = append(result, *cc)
	}
	return result, rows.Err()
}

// GetViews 获取视图列表
func (i *Inspector) GetViews(ctx context.Context) ([]model.View, error) {
	query := `
		SELECT table_name,
			   COALESCE(obj_description((table_schema || '.' || table_name)::regclass, 'pg_class'), ''),
			   view_definition
		FROM information_schema.views 
		WHERE table_schema = current_schema()
		ORDER BY table_name
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
		if comment.Valid && comment.String != "" {
			view.Comment = comment.String
		}
		if definition.Valid && definition.String != "" {
			view.Definition = definition.String
		}
		views = append(views, view)
	}

	return views, rows.Err()
}

// GetProcedures 获取存储过程列表
func (i *Inspector) GetProcedures(ctx context.Context) ([]model.Procedure, error) {
	query := `
		SELECT p.proname,
			   COALESCE(obj_description(p.oid, 'pg_proc'), ''),
			   pg_get_functiondef(p.oid)
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE p.prokind = 'p' AND n.nspname = current_schema()
		ORDER BY p.proname
	`

	rows, err := i.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query procedures: %w", err)
	}
	defer rows.Close()

	var procedures []model.Procedure
	for rows.Next() {
		var proc model.Procedure
		var comment, definition sql.NullString
		if err := rows.Scan(&proc.Name, &comment, &definition); err != nil {
			return nil, err
		}
		if comment.Valid && comment.String != "" {
			proc.Comment = comment.String
		}
		if definition.Valid && definition.String != "" {
			proc.Definition = definition.String
		}
		procedures = append(procedures, proc)
	}

	return procedures, rows.Err()
}

// GetFunctions 获取函数列表
func (i *Inspector) GetFunctions(ctx context.Context) ([]model.Function, error) {
	query := `
		SELECT p.proname,
			   COALESCE(obj_description(p.oid, 'pg_proc'), ''),
			   pg_get_functiondef(p.oid),
			   format_type(p.prorettype, NULL)
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE p.prokind = 'f' AND n.nspname = current_schema()
		ORDER BY p.proname
	`

	rows, err := i.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query functions: %w", err)
	}
	defer rows.Close()

	var functions []model.Function
	for rows.Next() {
		var fn model.Function
		var comment, definition, returnType sql.NullString
		if err := rows.Scan(&fn.Name, &comment, &definition, &returnType); err != nil {
			return nil, err
		}
		if comment.Valid && comment.String != "" {
			fn.Comment = comment.String
		}
		if definition.Valid && definition.String != "" {
			fn.Definition = definition.String
		}
		if returnType.Valid && returnType.String != "" {
			fn.ReturnType = returnType.String
		}
		functions = append(functions, fn)
	}

	return functions, rows.Err()
}

// GetTriggers 获取触发器列表
func (i *Inspector) GetTriggers(ctx context.Context, tableName string) ([]model.Trigger, error) {
	query := `
		SELECT tgname,
			   event_manipulation,
			   action_timing,
			   CASE WHEN tgenabled = 'O' THEN 'ENABLED' ELSE 'DISABLED' END,
			   pg_get_triggerdef(t.oid)
		FROM pg_trigger t
		JOIN pg_class c ON c.oid = t.tgrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE NOT tgisinternal 
		  AND c.relname = $1 
		  AND n.nspname = current_schema()
		ORDER BY tgname
	`

	rows, err := i.GetDB().QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query triggers for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var triggers []model.Trigger
	for rows.Next() {
		var tr model.Trigger
		var event, timing, status, definition sql.NullString
		if err := rows.Scan(&tr.Name, &event, &timing, &status, &definition); err != nil {
			return nil, err
		}
		tr.TableName = tableName
		if event.Valid {
			tr.Event = event.String
		}
		if timing.Valid {
			tr.Timing = timing.String
		}
		if status.Valid {
			tr.Status = status.String
		}
		if definition.Valid {
			tr.Definition = definition.String
		}
		triggers = append(triggers, tr)
	}

	return triggers, rows.Err()
}

// GetSequences 获取序列列表
func (i *Inspector) GetSequences(ctx context.Context) ([]model.Sequence, error) {
	query := `
		SELECT sequence_name,
			   minimum_value::bigint,
			   maximum_value::bigint,
			   increment::bigint,
			   cycle_flag,
			   cache_size,
			   last_value
		FROM information_schema.sequences 
		WHERE sequence_schema = current_schema()
		ORDER BY sequence_name
	`

	rows, err := i.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sequences: %w", err)
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
		seq.Cycle = (cycleFlag == "YES")
		sequences = append(sequences, seq)
	}

	return sequences, rows.Err()
}

// Factory PostgreSQL Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	return NewInspector(config), nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "postgres"
}

func init() {
	inspector.Register("postgres", &Factory{})
}
