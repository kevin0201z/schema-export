package mysql

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestMySQLBuildDSN(t *testing.T) {
	cfg := inspector.ConnectionConfig{DSN: "root:password@tcp(localhost:3306)/mydb"}
	ins := NewInspector(cfg)
	if got := ins.BuildDSN(); got != cfg.DSN {
		t.Fatalf("expected %s got %s", cfg.DSN, got)
	}

	cfg2 := inspector.ConnectionConfig{
		Username: "root",
		Password: "password",
		Host:     "localhost",
		Port:     3306,
		Database: "mydb",
	}
	ins2 := NewInspector(cfg2)
	got := ins2.BuildDSN()
	if got == "" {
		t.Fatalf("expected non-empty DSN built from components")
	}
}

func TestMySQLGetTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_COMMENT"}).
		AddRow("users", "用户表").
		AddRow("orders", "订单表")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TABLE_NAME, TABLE_COMMENT 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`)).WillReturnRows(rows)

	tables, err := ins.GetTables(context.Background())
	if err != nil {
		t.Fatalf("GetTables failed: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "users" {
		t.Fatalf("expected first table to be 'users', got %s", tables[0].Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMySQLGetColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{
		"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE",
		"COLUMN_DEFAULT", "COLUMN_COMMENT", "is_auto_increment", "is_pk",
	}).
		AddRow("id", "int(11)", "NO", nil, "主键ID", false, 1).
		AddRow("name", "varchar(255)", "YES", nil, "名称", false, 0)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).WithArgs("users", "users").WillReturnRows(rows)

	columns, err := ins.GetColumns(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetColumns failed: %v", err)
	}
	if len(columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(columns))
	}
	if columns[0].Name != "id" {
		t.Fatalf("expected first column to be 'id', got %s", columns[0].Name)
	}
	if !columns[0].IsPrimaryKey {
		t.Fatalf("expected 'id' to be primary key")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMySQLGetIndexes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "COLUMN_NAME"}).
		AddRow("PRIMARY", 0, "id").
		AddRow("idx_name", 0, "name")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`)).WithArgs("users").WillReturnRows(rows)

	indexes, err := ins.GetIndexes(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetIndexes failed: %v", err)
	}
	if len(indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(indexes))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMySQLGetForeignKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{
		"CONSTRAINT_NAME", "COLUMN_NAME",
		"REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "DELETE_RULE",
	}).
		AddRow("fk_user_order", "user_id", "users", "id", "CASCADE")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			CONSTRAINT_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			DELETE_RULE
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		  AND REFERENCED_TABLE_NAME IS NOT NULL
	`)).WithArgs("orders").WillReturnRows(rows)

	fks, err := ins.GetForeignKeys(context.Background(), "orders")
	if err != nil {
		t.Fatalf("GetForeignKeys failed: %v", err)
	}
	if len(fks) != 1 {
		t.Fatalf("expected 1 foreign key, got %d", len(fks))
	}
	if fks[0].Name != "fk_user_order" {
		t.Fatalf("expected foreign key name 'fk_user_order', got %s", fks[0].Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
