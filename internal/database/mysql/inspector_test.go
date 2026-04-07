package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestMySQLBuildDSN(t *testing.T) {
	t.Run("returns raw DSN when already in mysql driver form", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{DSN: "root:password@tcp(localhost:3306)/mydb"}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != cfg.DSN {
			t.Fatalf("expected %s got %s", cfg.DSN, got)
		}
	})

	t.Run("adds mysql scheme for raw DSN", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{DSN: "root:password@localhost:3306/mydb"}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != "mysql://root:password@localhost:3306/mydb" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("builds DSN from components with SSL mode", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{
			Username: "root",
			Password: "password",
			Host:     "localhost",
			Port:     3306,
			Database: "mydb",
			SSLMode:  "custom",
		}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true&loc=Local&tls=custom" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})
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

func TestMySQLGetTableSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("用户表"))

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
	`)).WithArgs("users", "users").WillReturnRows(sqlmock.NewRows([]string{
		"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "COLUMN_COMMENT", "is_auto_increment", "is_pk",
	}).AddRow("id", "int(11)", "NO", "((0))", "主键ID", true, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "COLUMN_NAME"}).
		AddRow("PRIMARY", 0, "id"))

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
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
		"CONSTRAINT_NAME", "COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "DELETE_RULE",
	}).AddRow("fk_user_group", "group_id", "groups", "id", "CASCADE"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.CONSTRAINT_NAME,
			cc.CHECK_CLAUSE,
			cu.COLUMN_NAME
		FROM information_schema.CHECK_CONSTRAINTS cc
		JOIN information_schema.TABLE_CONSTRAINTS tc ON cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
		LEFT JOIN information_schema.KEY_COLUMN_USAGE cu ON cc.CONSTRAINT_NAME = cu.CONSTRAINT_NAME
		WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ?
		ORDER BY cc.CONSTRAINT_NAME
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"CONSTRAINT_NAME", "CHECK_CLAUSE", "COLUMN_NAME"}).
		AddRow("ck_users_age", "age > 0", "age"))

	table, err := ins.GetTable(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTable failed: %v", err)
	}
	if table.Comment != "用户表" {
		t.Fatalf("unexpected table comment: %q", table.Comment)
	}
	if len(table.Columns) != 1 || !table.Columns[0].IsPrimaryKey || table.Columns[0].DefaultValue != "((0))" {
		t.Fatalf("unexpected columns: %#v", table.Columns)
	}
	if len(table.Indexes) != 1 || !table.Indexes[0].IsPrimary {
		t.Fatalf("unexpected indexes: %#v", table.Indexes)
	}
	if len(table.ForeignKeys) != 1 || table.ForeignKeys[0].OnDelete != "CASCADE" {
		t.Fatalf("unexpected foreign keys: %#v", table.ForeignKeys)
	}
	if len(table.CheckConstraints) != 1 || table.CheckConstraints[0].Name != "ck_users_age" {
		t.Fatalf("unexpected check constraints: %#v", table.CheckConstraints)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMySQLGetTableSubcallFailures(t *testing.T) {
	tests := []struct {
		name           string
		setupExpect    func(sqlmock.Sqlmock)
		expectedErrSub string
	}{
		{
			name: "columns failure",
			setupExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
					WithArgs("users").
					WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("用户表"))
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
	`)).WithArgs("users", "users").WillReturnError(sql.ErrConnDone)
			},
			expectedErrSub: "failed to query columns",
		},
		{
			name: "indexes failure",
			setupExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
					WithArgs("users").
					WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("用户表"))
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
	`)).WithArgs("users", "users").WillReturnRows(sqlmock.NewRows([]string{
					"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "COLUMN_COMMENT", "is_auto_increment", "is_pk",
				}).AddRow("id", "int(11)", "NO", nil, nil, false, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`)).WithArgs("users").WillReturnError(sql.ErrConnDone)
			},
			expectedErrSub: "failed to query indexes",
		},
		{
			name: "foreign keys failure",
			setupExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
					WithArgs("users").
					WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("用户表"))
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
	`)).WithArgs("users", "users").WillReturnRows(sqlmock.NewRows([]string{
					"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "COLUMN_COMMENT", "is_auto_increment", "is_pk",
				}).AddRow("id", "int(11)", "NO", nil, nil, false, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "COLUMN_NAME"}).
					AddRow("PRIMARY", 0, "id"))
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
	`)).WithArgs("users").WillReturnError(sql.ErrConnDone)
			},
			expectedErrSub: "failed to query foreign keys",
		},
		{
			name: "check constraints failure",
			setupExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
					WithArgs("users").
					WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("用户表"))
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
	`)).WithArgs("users", "users").WillReturnRows(sqlmock.NewRows([]string{
					"COLUMN_NAME", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT", "COLUMN_COMMENT", "is_auto_increment", "is_pk",
				}).AddRow("id", "int(11)", "NO", nil, nil, false, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "COLUMN_NAME"}).
					AddRow("PRIMARY", 0, "id"))
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
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
					"CONSTRAINT_NAME", "COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "DELETE_RULE",
				}).AddRow("fk_user_group", "group_id", "groups", "id", "CASCADE"))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.CONSTRAINT_NAME,
			cc.CHECK_CLAUSE,
			cu.COLUMN_NAME
		FROM information_schema.CHECK_CONSTRAINTS cc
		JOIN information_schema.TABLE_CONSTRAINTS tc ON cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
		LEFT JOIN information_schema.KEY_COLUMN_USAGE cu ON cc.CONSTRAINT_NAME = cu.CONSTRAINT_NAME
		WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ?
		ORDER BY cc.CONSTRAINT_NAME
	`)).WithArgs("users").WillReturnError(sql.ErrConnDone)
			},
			expectedErrSub: "failed to query check constraints",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer db.Close()

			ins := NewInspector(inspector.ConnectionConfig{})
			ins.SetDB(db)
			tc.setupExpect(mock)

			_, err = ins.GetTable(context.Background(), "users")
			if err == nil || !regexp.MustCompile(tc.expectedErrSub).MatchString(err.Error()) {
				t.Fatalf("expected error containing %q, got %v", tc.expectedErrSub, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestMySQLOtherMetadataQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.CONSTRAINT_NAME,
			cc.CHECK_CLAUSE,
			cu.COLUMN_NAME
		FROM information_schema.CHECK_CONSTRAINTS cc
		JOIN information_schema.TABLE_CONSTRAINTS tc ON cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
		LEFT JOIN information_schema.KEY_COLUMN_USAGE cu ON cc.CONSTRAINT_NAME = cu.CONSTRAINT_NAME
		WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ?
		ORDER BY cc.CONSTRAINT_NAME
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"CONSTRAINT_NAME", "CHECK_CLAUSE", "COLUMN_NAME"}).
		AddRow("ck_users_age", "age > 0", "age"))

	checks, err := ins.GetCheckConstraints(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetCheckConstraints failed: %v", err)
	}
	if len(checks) != 1 || checks[0].GetColumnsString() != "age" {
		t.Fatalf("unexpected checks: %#v", checks)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			TABLE_NAME,
			TABLE_COMMENT,
			VIEW_DEFINITION
		FROM information_schema.VIEWS
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_COMMENT", "VIEW_DEFINITION"}).
		AddRow("v_users", "用户视图", "SELECT * FROM users"))

	views, err := ins.GetViews(context.Background())
	if err != nil {
		t.Fatalf("GetViews failed: %v", err)
	}
	if len(views) != 1 || views[0].Definition != "SELECT * FROM users" {
		t.Fatalf("unexpected views: %#v", views)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			ROUTINE_NAME,
			ROUTINE_COMMENT,
			ROUTINE_DEFINITION
		FROM information_schema.ROUTINES
		WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_TYPE = 'PROCEDURE'
		ORDER BY ROUTINE_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"ROUTINE_NAME", "ROUTINE_COMMENT", "ROUTINE_DEFINITION"}).
		AddRow("sp_cleanup", "过程", "BEGIN END"))

	procedures, err := ins.GetProcedures(context.Background())
	if err != nil {
		t.Fatalf("GetProcedures failed: %v", err)
	}
	if len(procedures) != 1 || procedures[0].Name != "sp_cleanup" {
		t.Fatalf("unexpected procedures: %#v", procedures)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			ROUTINE_NAME,
			ROUTINE_COMMENT,
			ROUTINE_DEFINITION,
			DTD_IDENTIFIER
		FROM information_schema.ROUTINES
		WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_TYPE = 'FUNCTION'
		ORDER BY ROUTINE_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"ROUTINE_NAME", "ROUTINE_COMMENT", "ROUTINE_DEFINITION", "DTD_IDENTIFIER"}).
		AddRow("fn_total", "函数", "RETURN 1", "int"))

	functions, err := ins.GetFunctions(context.Background())
	if err != nil {
		t.Fatalf("GetFunctions failed: %v", err)
	}
	if len(functions) != 1 || functions[0].ReturnType != "int" {
		t.Fatalf("unexpected functions: %#v", functions)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			TRIGGER_NAME,
			EVENT_MANIPULATION,
			EVENT_OBJECT_TABLE,
			ACTION_TIMING,
			ACTION_STATEMENT
		FROM information_schema.TRIGGERS
		WHERE TRIGGER_SCHEMA = DATABASE() AND EVENT_OBJECT_TABLE = ?
		ORDER BY TRIGGER_NAME
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"TRIGGER_NAME", "EVENT_MANIPULATION", "EVENT_OBJECT_TABLE", "ACTION_TIMING", "ACTION_STATEMENT"}).
		AddRow("tr_users", "INSERT", "users", "BEFORE", "SET NEW.updated_at = NOW()"))

	triggers, err := ins.GetTriggers(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTriggers failed: %v", err)
	}
	if len(triggers) != 1 || triggers[0].Timing != "BEFORE" {
		t.Fatalf("unexpected triggers: %#v", triggers)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow(nil))

	comment, err := ins.getTableComment(context.Background(), "users")
	if err != nil {
		t.Fatalf("getTableComment failed: %v", err)
	}
	if comment != "" {
		t.Fatalf("expected empty comment, got %q", comment)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT TABLE_NAME, TABLE_COMMENT 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`)).WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_COMMENT"}))

	tables, err := ins.GetTables(context.Background())
	if err != nil {
		t.Fatalf("GetTables failed: %v", err)
	}
	if len(tables) != 0 {
		t.Fatalf("expected empty tables, got %#v", tables)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			TRIGGER_NAME,
			EVENT_MANIPULATION,
			EVENT_OBJECT_TABLE,
			ACTION_TIMING,
			ACTION_STATEMENT
		FROM information_schema.TRIGGERS
		WHERE TRIGGER_SCHEMA = DATABASE() AND EVENT_OBJECT_TABLE = ?
		ORDER BY TRIGGER_NAME`)).WithArgs("empty").WillReturnRows(sqlmock.NewRows([]string{"TRIGGER_NAME", "EVENT_MANIPULATION", "EVENT_OBJECT_TABLE", "ACTION_TIMING", "ACTION_STATEMENT"}))

	triggers, err = ins.GetTriggers(context.Background(), "empty")
	if err != nil {
		t.Fatalf("GetTriggers empty failed: %v", err)
	}
	if len(triggers) != 0 {
		t.Fatalf("expected empty triggers, got %#v", triggers)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
