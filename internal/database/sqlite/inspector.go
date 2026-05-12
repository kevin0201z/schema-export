package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	_ "modernc.org/sqlite"

	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

var triggerPattern = regexp.MustCompile(`(?is)create\s+trigger(?:\s+if\s+not\s+exists)?\s+(?:"[^"]+"|` + "`[^`]+`" + `|\[[^\]]+\]|\S+)\s+(before|after|instead\s+of)\s+(insert|update|delete)`)

// Inspector SQLite 元数据读取实现。
type Inspector struct {
	*database.BaseInspector
}

type checkConstraintDef struct {
	Name       string
	Definition string
	Columns    []string
	ColumnName string
}

type pragmaTableInfoRow struct {
	CID        int
	Name       string
	DataType   string
	NotNull    int
	DefaultVal sql.NullString
	PK         int
}

// NewInspector 创建 SQLite Inspector。
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{BaseInspector: database.NewBaseInspector(config)}
}

// Connect 连接 SQLite 数据库。
func (i *Inspector) Connect(ctx context.Context) error {
	db, err := sql.Open("sqlite", i.BuildDSN())
	if err != nil {
		return fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(database.DefaultConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, database.DefaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建 SQLite DSN。
func (i *Inspector) BuildDSN() string {
	cfg := i.GetConfig()
	if strings.TrimSpace(cfg.DSN) != "" {
		return strings.TrimSpace(cfg.DSN)
	}
	return strings.TrimSpace(cfg.Database)
}

// GetTables 获取普通表列表。
func (i *Inspector) GetTables(ctx context.Context) ([]model.Table, error) {
	rows, err := i.GetDB().QueryContext(ctx, `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite tables: %w", err)
	}
	defer rows.Close()

	var tables []model.Table
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, model.Table{Name: name, Type: model.TableTypeTable})
	}
	return tables, rows.Err()
}

// GetTable 获取单表完整元数据。
func (i *Inspector) GetTable(ctx context.Context, tableName string) (*model.Table, error) {
	exists, err := i.objectExists(ctx, "table", tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("table %s not found", tableName)
	}

	table := &model.Table{Name: tableName, Type: model.TableTypeTable}

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

	checks, err := i.GetCheckConstraints(ctx, tableName)
	if err != nil {
		return nil, err
	}
	table.CheckConstraints = checks

	return table, nil
}

// GetColumns 获取字段信息。
func (i *Inspector) GetColumns(ctx context.Context, tableName string) ([]model.Column, error) {
	createSQL, err := i.getObjectSQL(ctx, "table", tableName)
	if err != nil {
		return nil, err
	}

	rows, err := i.queryTableInfo(ctx, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := parseSQLiteChecks(tableName, createSQL)
	columnChecks := make(map[string]string)
	for _, check := range checks {
		if check.ColumnName != "" {
			columnChecks[check.ColumnName] = check.Definition
		}
	}

	autoIncrementCols := sqliteAutoincrementColumns(createSQL)

	var columns []model.Column
	for rows.Next() {
		var row pragmaTableInfoRow
		if err := rows.Scan(&row.CID, &row.Name, &row.DataType, &row.NotNull, &row.DefaultVal, &row.PK); err != nil {
			return nil, err
		}

		col := model.Column{
			Name:         row.Name,
			DataType:     row.DataType,
			IsNullable:   row.NotNull == 0,
			IsPrimaryKey: row.PK > 0,
		}
		if row.DefaultVal.Valid {
			col.DefaultValue = row.DefaultVal.String
		}
		if expr, ok := columnChecks[row.Name]; ok {
			col.CheckConstraint = expr
		}
		if autoIncrementCols[row.Name] {
			col.IsAutoIncrement = true
		}
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// GetIndexes 获取索引信息。
func (i *Inspector) GetIndexes(ctx context.Context, tableName string) ([]model.Index, error) {
	rows, err := i.GetDB().QueryContext(ctx, fmt.Sprintf("PRAGMA index_list(%s)", quoteSQLiteIdentifier(tableName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite indexes for table %s: %w", tableName, err)
	}
	defer rows.Close()

	type indexMeta struct {
		name   string
		unique bool
	}

	var metas []indexMeta
	for rows.Next() {
		var seq int
		var name, origin, partial string
		var unique int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		if origin == "pk" || strings.HasPrefix(name, "sqlite_autoindex_") {
			continue
		}
		metas = append(metas, indexMeta{name: name, unique: unique == 1})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	indexes := make([]model.Index, 0, len(metas))
	for _, meta := range metas {
		indexColumns, err := i.getIndexColumns(ctx, meta.name)
		if err != nil {
			return nil, err
		}

		idx := model.Index{
			Name:     meta.name,
			Columns:  indexColumns,
			IsUnique: meta.unique,
			Type:     model.IndexTypeNormal,
		}
		if idx.IsUnique {
			idx.Type = model.IndexTypeUnique
		}
		indexes = append(indexes, idx)
	}

	sort.Slice(indexes, func(a, b int) bool { return indexes[a].Name < indexes[b].Name })
	return indexes, nil
}

// GetForeignKeys 获取外键信息。
func (i *Inspector) GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error) {
	rows, err := i.GetDB().QueryContext(ctx, fmt.Sprintf("PRAGMA foreign_key_list(%s)", quoteSQLiteIdentifier(tableName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite foreign keys for table %s: %w", tableName, err)
	}
	defer rows.Close()

	type fkPart struct {
		seq      int
		from     string
		to       string
		refTable string
		onUpdate string
		onDelete string
	}

	partsByID := make(map[int][]fkPart)
	for rows.Next() {
		var id int
		var part fkPart
		var match string
		if err := rows.Scan(&id, &part.seq, &part.refTable, &part.from, &part.to, &part.onUpdate, &part.onDelete, &match); err != nil {
			return nil, err
		}
		partsByID[id] = append(partsByID[id], part)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var ids []int
	for id := range partsByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var foreignKeys []model.ForeignKey
	for _, id := range ids {
		parts := partsByID[id]
		sort.Slice(parts, func(a, b int) bool { return parts[a].seq < parts[b].seq })

		fromCols := make([]string, 0, len(parts))
		toCols := make([]string, 0, len(parts))
		for _, part := range parts {
			fromCols = append(fromCols, part.from)
			toCols = append(toCols, part.to)
		}

		foreignKeys = append(foreignKeys, model.ForeignKey{
			Name:      fmt.Sprintf("fk_%s_%d", tableName, id+1),
			Column:    strings.Join(fromCols, ", "),
			RefTable:  parts[0].refTable,
			RefColumn: strings.Join(toCols, ", "),
			OnDelete:  parts[0].onDelete,
			OnUpdate:  parts[0].onUpdate,
		})
	}

	return foreignKeys, nil
}

// GetCheckConstraints 获取表级 CHECK 约束。
func (i *Inspector) GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error) {
	createSQL, err := i.getObjectSQL(ctx, "table", tableName)
	if err != nil {
		return nil, err
	}

	parsed := parseSQLiteChecks(tableName, createSQL)
	constraints := make([]model.CheckConstraint, 0, len(parsed))
	for _, item := range parsed {
		if item.ColumnName != "" {
			continue
		}
		constraints = append(constraints, model.CheckConstraint{
			Name:       item.Name,
			Definition: item.Definition,
			Columns:    item.Columns,
		})
	}

	return constraints, nil
}

// GetViews 获取视图列表。
func (i *Inspector) GetViews(ctx context.Context) ([]model.View, error) {
	rows, err := i.GetDB().QueryContext(ctx, `
		SELECT name, COALESCE(sql, '')
		FROM sqlite_master
		WHERE type = 'view'
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite views: %w", err)
	}
	defer rows.Close()

	var views []model.View
	for rows.Next() {
		var view model.View
		if err := rows.Scan(&view.Name, &view.Definition); err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for idx := range views {
		columns, err := i.getViewColumns(ctx, views[idx].Name)
		if err != nil {
			return nil, err
		}
		views[idx].Columns = columns
	}

	return views, nil
}

// GetProcedures 返回空集合，SQLite 无存储过程。
func (i *Inspector) GetProcedures(context.Context) ([]model.Procedure, error) {
	return []model.Procedure{}, nil
}

// GetFunctions 返回空集合，SQLite 无独立函数对象。
func (i *Inspector) GetFunctions(context.Context) ([]model.Function, error) {
	return []model.Function{}, nil
}

// GetTriggers 获取指定表的触发器列表。
func (i *Inspector) GetTriggers(ctx context.Context, tableName string) ([]model.Trigger, error) {
	rows, err := i.GetDB().QueryContext(ctx, `
		SELECT name, tbl_name, COALESCE(sql, '')
		FROM sqlite_master
		WHERE type = 'trigger' AND tbl_name = ?
		ORDER BY name
	`, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite triggers for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var triggers []model.Trigger
	for rows.Next() {
		var tr model.Trigger
		if err := rows.Scan(&tr.Name, &tr.TableName, &tr.Definition); err != nil {
			return nil, err
		}
		tr.Status = "ENABLED"
		tr.Timing, tr.Event = parseSQLiteTriggerParts(tr.Definition)
		triggers = append(triggers, tr)
	}

	return triggers, rows.Err()
}

// GetSequences 返回空集合，SQLite 无序列对象。
func (i *Inspector) GetSequences(context.Context) ([]model.Sequence, error) {
	return []model.Sequence{}, nil
}

func (i *Inspector) queryTableInfo(ctx context.Context, tableName string) (*sql.Rows, error) {
	rows, err := i.GetDB().QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quoteSQLiteIdentifier(tableName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite columns for table %s: %w", tableName, err)
	}
	return rows, nil
}

func (i *Inspector) getIndexColumns(ctx context.Context, indexName string) ([]string, error) {
	rows, err := i.GetDB().QueryContext(ctx, fmt.Sprintf("PRAGMA index_xinfo(%s)", quoteSQLiteIdentifier(indexName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite index columns for index %s: %w", indexName, err)
	}
	defer rows.Close()

	type indexCol struct {
		seqno int
		name  sql.NullString
		key   int
	}

	var cols []indexCol
	for rows.Next() {
		var col indexCol
		var cid, desc int
		var coll sql.NullString
		if err := rows.Scan(&col.seqno, &cid, &col.name, &desc, &coll, &col.key); err != nil {
			return nil, err
		}
		if col.key == 1 && col.name.Valid {
			cols = append(cols, col)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(cols, func(a, b int) bool { return cols[a].seqno < cols[b].seqno })
	result := make([]string, 0, len(cols))
	for _, col := range cols {
		result = append(result, col.name.String)
	}
	return result, nil
}

func (i *Inspector) getViewColumns(ctx context.Context, viewName string) ([]model.Column, error) {
	rows, err := i.GetDB().QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quoteSQLiteIdentifier(viewName)))
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite view columns for view %s: %w", viewName, err)
	}
	defer rows.Close()

	var columns []model.Column
	for rows.Next() {
		var row pragmaTableInfoRow
		if err := rows.Scan(&row.CID, &row.Name, &row.DataType, &row.NotNull, &row.DefaultVal, &row.PK); err != nil {
			return nil, err
		}
		columns = append(columns, model.Column{
			Name:         row.Name,
			DataType:     row.DataType,
			IsNullable:   row.NotNull == 0,
			IsPrimaryKey: row.PK > 0,
		})
	}

	return columns, rows.Err()
}

func (i *Inspector) getObjectSQL(ctx context.Context, objectType, objectName string) (string, error) {
	var sqlText sql.NullString
	err := i.GetDB().QueryRowContext(ctx, `
		SELECT sql
		FROM sqlite_master
		WHERE type = ? AND name = ?
	`, objectType, objectName).Scan(&sqlText)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("%s %s not found", objectType, objectName)
	}
	if err != nil {
		return "", fmt.Errorf("failed to query sqlite object SQL for %s %s: %w", objectType, objectName, err)
	}
	if !sqlText.Valid {
		return "", nil
	}
	return sqlText.String, nil
}

func (i *Inspector) objectExists(ctx context.Context, objectType, objectName string) (bool, error) {
	var count int
	err := i.GetDB().QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM sqlite_master
		WHERE type = ? AND name = ?
	`, objectType, objectName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query sqlite object existence for %s %s: %w", objectType, objectName, err)
	}
	return count > 0, nil
}

func parseSQLiteChecks(tableName, createSQL string) []checkConstraintDef {
	body := extractCreateTableBody(createSQL)
	if body == "" {
		return nil
	}

	segments := splitSQLiteTopLevel(body, ',')
	columnNames := make([]string, 0, len(segments))
	for _, seg := range segments {
		if isSQLiteTableConstraint(seg) {
			continue
		}
		if name := sqliteDefinitionName(seg); name != "" {
			columnNames = append(columnNames, name)
		}
	}

	var checks []checkConstraintDef
	checkIndex := 1
	for _, seg := range segments {
		if seg == "" {
			continue
		}

		columnName := ""
		if !isSQLiteTableConstraint(seg) {
			columnName = sqliteDefinitionName(seg)
		}

		for _, expr := range extractSQLiteCheckExprs(seg) {
			checks = append(checks, checkConstraintDef{
				Name:       fmt.Sprintf("ck_%s_%d", tableName, checkIndex),
				Definition: expr,
				Columns:    detectSQLiteCheckColumns(expr, columnNames),
				ColumnName: columnName,
			})
			checkIndex++
		}
	}

	return checks
}

func extractCreateTableBody(createSQL string) string {
	start := strings.Index(createSQL, "(")
	end := strings.LastIndex(createSQL, ")")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return createSQL[start+1 : end]
}

func splitSQLiteTopLevel(s string, delimiter rune) []string {
	var result []string
	var current strings.Builder
	depth := 0
	var quote rune

	for _, r := range s {
		switch {
		case quote != 0:
			current.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"' || r == '`':
			quote = r
			current.WriteRune(r)
		case r == '[':
			quote = ']'
			current.WriteRune(r)
		case r == '(':
			depth++
			current.WriteRune(r)
		case r == ')':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		case r == delimiter && depth == 0:
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		result = append(result, strings.TrimSpace(current.String()))
	}
	return result
}

func extractSQLiteCheckExprs(segment string) []string {
	upper := strings.ToUpper(segment)
	var exprs []string
	searchStart := 0
	for {
		idx := strings.Index(upper[searchStart:], "CHECK")
		if idx == -1 {
			return exprs
		}
		idx += searchStart
		openRel := strings.Index(segment[idx:], "(")
		if openRel == -1 {
			return exprs
		}
		openIdx := idx + openRel
		expr, endIdx := readSQLiteParenthesized(segment, openIdx)
		if endIdx == -1 {
			return exprs
		}
		exprs = append(exprs, strings.TrimSpace(expr))
		searchStart = endIdx + 1
	}
}

func readSQLiteParenthesized(s string, openIdx int) (string, int) {
	depth := 0
	var quote rune
	var body strings.Builder

	for i, r := range s[openIdx:] {
		switch {
		case quote != 0:
			if depth > 0 {
				body.WriteRune(r)
			}
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"' || r == '`':
			if depth > 0 {
				body.WriteRune(r)
			}
			quote = r
		case r == '[':
			if depth > 0 {
				body.WriteRune(r)
			}
			quote = ']'
		case r == '(':
			depth++
			if depth > 1 {
				body.WriteRune(r)
			}
		case r == ')':
			depth--
			if depth == 0 {
				return body.String(), openIdx + i
			}
			body.WriteRune(r)
		default:
			if depth > 0 {
				body.WriteRune(r)
			}
		}
	}

	return "", -1
}

func isSQLiteTableConstraint(segment string) bool {
	upper := strings.ToUpper(strings.TrimSpace(segment))
	keywords := []string{"CONSTRAINT ", "PRIMARY KEY", "UNIQUE", "FOREIGN KEY", "CHECK"}
	for _, keyword := range keywords {
		if strings.HasPrefix(upper, keyword) {
			return true
		}
	}
	return false
}

func sqliteDefinitionName(segment string) string {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return ""
	}

	if segment[0] == '"' {
		if idx := strings.Index(segment[1:], `"`); idx != -1 {
			return strings.ReplaceAll(segment[1:idx+1], `""`, `"`)
		}
	}
	if segment[0] == '`' {
		if idx := strings.Index(segment[1:], "`"); idx != -1 {
			return segment[1 : idx+1]
		}
	}
	if segment[0] == '[' {
		if idx := strings.Index(segment[1:], "]"); idx != -1 {
			return segment[1 : idx+1]
		}
	}

	end := len(segment)
	for i, r := range segment {
		if unicode.IsSpace(r) {
			end = i
			break
		}
	}
	return segment[:end]
}

func detectSQLiteCheckColumns(expr string, columnNames []string) []string {
	exprLower := strings.ToLower(expr)
	var cols []string
	for _, name := range columnNames {
		if name == "" {
			continue
		}
		if containsSQLiteIdentifier(exprLower, strings.ToLower(name)) {
			cols = append(cols, name)
		}
	}
	return cols
}

func containsSQLiteIdentifier(exprLower, nameLower string) bool {
	if strings.Contains(exprLower, `"`+nameLower+`"`) || strings.Contains(exprLower, "`"+nameLower+"`") || strings.Contains(exprLower, "["+nameLower+"]") {
		return true
	}

	idx := strings.Index(exprLower, nameLower)
	for idx != -1 {
		beforeOK := idx == 0 || !isSQLiteIdentChar(rune(exprLower[idx-1]))
		afterIdx := idx + len(nameLower)
		afterOK := afterIdx == len(exprLower) || !isSQLiteIdentChar(rune(exprLower[afterIdx]))
		if beforeOK && afterOK {
			return true
		}

		nextStart := idx + len(nameLower)
		if nextStart >= len(exprLower) {
			break
		}
		offset := strings.Index(exprLower[nextStart:], nameLower)
		if offset == -1 {
			break
		}
		idx = nextStart + offset
	}

	return false
}

func isSQLiteIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

func sqliteAutoincrementColumns(createSQL string) map[string]bool {
	body := extractCreateTableBody(createSQL)
	result := make(map[string]bool)
	for _, seg := range splitSQLiteTopLevel(body, ',') {
		if strings.Contains(strings.ToUpper(seg), "AUTOINCREMENT") {
			if name := sqliteDefinitionName(seg); name != "" {
				result[name] = true
			}
		}
	}
	return result
}

func parseSQLiteTriggerParts(definition string) (string, string) {
	matches := triggerPattern.FindStringSubmatch(definition)
	if len(matches) != 3 {
		return "", ""
	}
	return strings.ToUpper(strings.TrimSpace(matches[1])), strings.ToUpper(strings.TrimSpace(matches[2]))
}

func quoteSQLiteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// Factory SQLite Inspector 工厂。
type Factory struct{}

// Create 创建 SQLite Inspector。
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	return NewInspector(config), nil
}

// GetType 获取工厂类型。
func (f *Factory) GetType() string {
	return "sqlite"
}

func init() {
	inspector.Register("sqlite", &Factory{})
}
