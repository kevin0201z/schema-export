package sql

import (
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/model"
)

func TestGetDialect(t *testing.T) {
	tests := []struct {
		name string
		db   string
		want string
	}{
		{name: "oracle", db: "oracle", want: "oracle"},
		{name: "dm", db: "dm", want: "oracle"},
		{name: "sqlserver", db: "sqlserver", want: "sqlserver"},
		{name: "mysql", db: "mysql", want: "mysql"},
		{name: "postgres", db: "postgres", want: "postgres"},
		{name: "postgresql", db: "postgresql", want: "postgres"},
		{name: "default", db: "other", want: "generic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDialect(tt.db).GetName(); got != tt.want {
				t.Fatalf("GetDialect(%q).GetName() = %q, want %q", tt.db, got, tt.want)
			}
		})
	}
}

func TestExporterFactoryAndIdentity(t *testing.T) {
	exp := NewExporter()
	if got := exp.GetName(); got != "sql" {
		t.Fatalf("unexpected exporter name: %s", got)
	}
	if got := exp.GetExtension(); got != ".sql" {
		t.Fatalf("unexpected exporter extension: %s", got)
	}

	f := &Factory{}
	created, err := f.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created == nil {
		t.Fatalf("expected exporter instance")
	}
	if got := f.GetType(); got != "sql" {
		t.Fatalf("unexpected factory type: %s", got)
	}
}

func TestGenericDialect(t *testing.T) {
	d := &GenericDialect{}

	if d.GetName() != "generic" {
		t.Fatalf("unexpected name")
	}
	if got := d.QuoteIdentifier("users"); got != "\"users\"" {
		t.Fatalf("unexpected quote: %s", got)
	}

	col := &model.Column{Name: "age", DataType: "number", Precision: 10, Scale: 2, DefaultValue: "42", Comment: "age", IsNullable: false, IsPrimaryKey: true}
	if got := d.GetDataType(col); got != "DECIMAL(10,2)" {
		t.Fatalf("unexpected datatype: %s", got)
	}
	if got := d.GetDefaultValue(col); got != "DEFAULT 42" {
		t.Fatalf("unexpected default: %s", got)
	}
	if got := d.GetColumnDefinition(col); !strings.Contains(got, "\"age\" DECIMAL(10,2)") || !strings.Contains(got, "PRIMARY KEY") {
		t.Fatalf("unexpected column definition: %s", got)
	}

	cc := &model.CheckConstraint{Name: "ck_age", Definition: "age > 0"}
	if got := d.GetCheckConstraint(cc); got != "CONSTRAINT \"ck_age\" CHECK (age > 0)" {
		t.Fatalf("unexpected check constraint: %s", got)
	}
	if got := d.GetColumnCommentSQL("users", col); got != "COMMENT ON COLUMN \"users\".\"age\" IS 'age';" {
		t.Fatalf("unexpected column comment sql: %s", got)
	}
	if got := d.GetTableCommentSQL("users", "users table"); got != "COMMENT ON TABLE \"users\" IS 'users table';" {
		t.Fatalf("unexpected table comment sql: %s", got)
	}
	if got := d.GetViewCommentSQL("v_users", "users view"); got != "COMMENT ON VIEW \"v_users\" IS 'users view';" {
		t.Fatalf("unexpected view comment sql: %s", got)
	}
	if d.SupportsInlineComment() {
		t.Fatalf("generic should not support inline comment")
	}
	if !d.SupportsInlineCheck() {
		t.Fatalf("generic should support inline check")
	}
}

func TestMySQLDialect(t *testing.T) {
	d := &MySQLDialect{}

	if d.GetName() != "mysql" {
		t.Fatalf("unexpected name")
	}
	if got := d.QuoteIdentifier("users"); got != "`users`" {
		t.Fatalf("unexpected quote: %s", got)
	}

	if got := d.GetDataType(&model.Column{DataType: "varchar2", Length: 12}); got != "VARCHAR(12)" {
		t.Fatalf("unexpected datatype: %s", got)
	}
	if got := d.GetDataType(&model.Column{DataType: "number", Precision: 10, Scale: 2}); got != "DECIMAL(10,2)" {
		t.Fatalf("unexpected decimal datatype: %s", got)
	}
	if got := d.GetDefaultValue(&model.Column{DefaultValue: "sysdate"}); got != "DEFAULT CURRENT_TIMESTAMP" {
		t.Fatalf("unexpected default value: %s", got)
	}
	if got := d.GetColumnDefinition(&model.Column{Name: "name", DataType: "varchar", Length: 10, Comment: "o'clock"}); !strings.Contains(got, "COMMENT 'o\\'clock'") {
		t.Fatalf("unexpected column definition: %s", got)
	}
	if got := d.GetCheckConstraint(&model.CheckConstraint{Name: "ck", Definition: "sysdate > 0"}); !strings.Contains(got, "NOW()") {
		t.Fatalf("unexpected check constraint: %s", got)
	}
	if got := d.GetTableCommentSQL("users", "user's table"); got != "ALTER TABLE `users` COMMENT = 'user\\'s table';" {
		t.Fatalf("unexpected table comment sql: %s", got)
	}
	if got := d.GetColumnCommentSQL("users", &model.Column{Name: "name"}); got != "" {
		t.Fatalf("expected empty column comment sql, got %s", got)
	}
	if got := d.GetViewCommentSQL("v_users", "comment"); got != "" {
		t.Fatalf("expected empty view comment sql, got %s", got)
	}
	if !d.SupportsInlineComment() || !d.SupportsInlineCheck() {
		t.Fatalf("mysql should support inline comment/check")
	}
}

func TestPostgreSQLDialect(t *testing.T) {
	d := &PostgreSQLDialect{}

	if d.GetName() != "postgres" {
		t.Fatalf("unexpected name")
	}
	if got := d.QuoteIdentifier("users"); got != "\"users\"" {
		t.Fatalf("unexpected quote: %s", got)
	}
	if got := d.GetDataType(&model.Column{DataType: "varchar", Length: 12}); got != "VARCHAR(12)" {
		t.Fatalf("unexpected datatype: %s", got)
	}
	if got := d.GetDefaultValue(&model.Column{DefaultValue: "nextval('users_id_seq'::regclass)"}); got != "nextval('users_id_seq'::regclass)" {
		t.Fatalf("unexpected default value: %s", got)
	}
	if got := d.GetDefaultValue(&model.Column{DefaultValue: "my value"}); got != "DEFAULT 'my value'" {
		t.Fatalf("unexpected quoted default value: %s", got)
	}
	if got := d.GetColumnDefinition(&model.Column{Name: "id", DataType: "int", IsAutoIncrement: true, IsNullable: false}); !strings.Contains(got, "GENERATED ALWAYS AS IDENTITY") {
		t.Fatalf("unexpected column definition: %s", got)
	}
	if got := d.GetCheckConstraint(&model.CheckConstraint{Name: "ck", Definition: "(age > 0)"}); got != "CONSTRAINT \"ck\" CHECK (age > 0)" {
		t.Fatalf("unexpected check constraint: %s", got)
	}
	if got := d.GetColumnCommentSQL("users", &model.Column{Name: "name", Comment: "user's name"}); got != "COMMENT ON COLUMN \"users\".\"name\" IS 'user''s name';" {
		t.Fatalf("unexpected column comment sql: %s", got)
	}
	if got := d.GetTableCommentSQL("users", "users table"); got != "COMMENT ON TABLE \"users\" IS 'users table';" {
		t.Fatalf("unexpected table comment sql: %s", got)
	}
	if got := d.GetViewCommentSQL("v_users", "users view"); got != "COMMENT ON VIEW \"v_users\" IS 'users view';" {
		t.Fatalf("unexpected view comment sql: %s", got)
	}
	if d.SupportsInlineComment() {
		t.Fatalf("postgres should not support inline comment")
	}
	if !d.SupportsInlineCheck() {
		t.Fatalf("postgres should support inline check")
	}
}

func TestOracleDialect(t *testing.T) {
	d := &OracleDialect{}

	if d.GetName() != "oracle" {
		t.Fatalf("unexpected name")
	}
	if got := d.QuoteIdentifier("users"); got != "\"users\"" {
		t.Fatalf("unexpected quote: %s", got)
	}
	if got := d.GetDataType(&model.Column{DataType: "varchar2", Length: 12}); got != "VARCHAR2(12)" {
		t.Fatalf("unexpected datatype: %s", got)
	}
	if got := d.GetDefaultValue(&model.Column{DefaultValue: "sysdate"}); got != "DEFAULT SYSDATE" {
		t.Fatalf("unexpected default value: %s", got)
	}
	if got := d.GetColumnDefinition(&model.Column{Name: "id", DataType: "number", IsPrimaryKey: true, IsNullable: false}); !strings.Contains(got, "PRIMARY KEY") {
		t.Fatalf("unexpected column definition: %s", got)
	}
	if got := d.GetCheckConstraint(&model.CheckConstraint{Name: "ck", Definition: "age > 0"}); got != "CONSTRAINT \"ck\" CHECK (age > 0)" {
		t.Fatalf("unexpected check constraint: %s", got)
	}
	if got := d.GetColumnCommentSQL("users", &model.Column{Name: "name", Comment: "user's name"}); got != "COMMENT ON COLUMN \"users\".\"name\" IS 'user''s name';" {
		t.Fatalf("unexpected column comment sql: %s", got)
	}
	if got := d.GetTableCommentSQL("users", "users table"); got != "COMMENT ON TABLE \"users\" IS 'users table';" {
		t.Fatalf("unexpected table comment sql: %s", got)
	}
	if got := d.GetViewCommentSQL("v_users", "users view"); got != "COMMENT ON TABLE \"v_users\" IS 'users view';" {
		t.Fatalf("unexpected view comment sql: %s", got)
	}
	if d.SupportsInlineComment() {
		t.Fatalf("oracle should not support inline comment")
	}
	if !d.SupportsInlineCheck() {
		t.Fatalf("oracle should support inline check")
	}
}

func TestSQLServerDialect(t *testing.T) {
	d := &SQLServerDialect{}

	if d.GetName() != "sqlserver" {
		t.Fatalf("unexpected name")
	}
	if got := d.QuoteIdentifier("users"); got != "[users]" {
		t.Fatalf("unexpected quote: %s", got)
	}
	if got := d.GetDataType(&model.Column{DataType: "nvarchar", Length: -1}); got != "NVARCHAR(MAX)" {
		t.Fatalf("unexpected datatype: %s", got)
	}
	if got := d.GetDefaultValue(&model.Column{DefaultValue: "getdate()"}); got != "DEFAULT GETDATE()" {
		t.Fatalf("unexpected default value: %s", got)
	}
	if got := d.GetColumnDefinition(&model.Column{Name: "id", DataType: "int", IsAutoIncrement: true, IsPrimaryKey: true, IsNullable: false}); !strings.Contains(got, "IDENTITY(1,1)") || !strings.Contains(got, "PRIMARY KEY") {
		t.Fatalf("unexpected column definition: %s", got)
	}
	if got := d.GetCheckConstraint(&model.CheckConstraint{Name: "ck", Definition: "age > 0"}); got != "CONSTRAINT [ck] CHECK (age > 0)" {
		t.Fatalf("unexpected check constraint: %s", got)
	}
	if got := d.GetColumnCommentSQL("users", &model.Column{Name: "name", Comment: "user's name"}); !strings.Contains(got, "N'user''s name'") {
		t.Fatalf("unexpected column comment sql: %s", got)
	}
	if got := d.GetTableCommentSQL("users", "users table"); !strings.Contains(got, "TABLE', 'users'") {
		t.Fatalf("unexpected table comment sql: %s", got)
	}
	if got := d.GetViewCommentSQL("v_users", "users view"); !strings.Contains(got, "VIEW', 'v_users'") {
		t.Fatalf("unexpected view comment sql: %s", got)
	}
	if d.SupportsInlineComment() {
		t.Fatalf("sqlserver should not support inline comment")
	}
	if !d.SupportsInlineCheck() {
		t.Fatalf("sqlserver should support inline check")
	}
}
