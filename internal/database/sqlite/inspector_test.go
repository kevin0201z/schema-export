package sqlite

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/inspector"
)

func TestSQLiteBuildDSN(t *testing.T) {
	t.Run("prefers DSN", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{DSN: "file:test.db?mode=ro", Database: "ignored.db"}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != "file:test.db?mode=ro" {
			t.Fatalf("BuildDSN() = %q", got)
		}
	})

	t.Run("falls back to database path", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{Database: "/tmp/app.db"}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != "/tmp/app.db" {
			t.Fatalf("BuildDSN() = %q", got)
		}
	})

	t.Run("supports memory DSN", func(t *testing.T) {
		cfg := inspector.ConnectionConfig{DSN: ":memory:"}
		ins := NewInspector(cfg)
		if got := ins.BuildDSN(); got != ":memory:" {
			t.Fatalf("BuildDSN() = %q", got)
		}
	})
}

func TestSQLiteInspectorMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sample.db")
	ins := NewInspector(inspector.ConnectionConfig{DSN: dbPath})

	ctx := context.Background()
	if err := ins.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer ins.Close()

	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL CHECK (length(name) > 0),
			email TEXT UNIQUE,
			age INTEGER,
			status TEXT DEFAULT 'active',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			CHECK (age >= 0),
			CHECK (status IN ('active', 'disabled'))
		)`,
		`CREATE TABLE memberships (
			user_id INTEGER NOT NULL,
			group_id INTEGER NOT NULL,
			role TEXT DEFAULT 'member',
			PRIMARY KEY (user_id, group_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE NO ACTION
		)`,
		`CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			region_id INTEGER NOT NULL,
			amount REAL CHECK (amount >= 0),
			FOREIGN KEY (user_id, region_id) REFERENCES memberships(user_id, group_id) ON DELETE CASCADE ON UPDATE RESTRICT
		)`,
		`CREATE INDEX idx_orders_user_region ON orders(user_id, region_id)`,
		`CREATE UNIQUE INDEX idx_users_email ON users(email)`,
		`CREATE VIEW active_users AS SELECT id, name, status FROM users WHERE status = 'active'`,
		`CREATE TRIGGER trg_orders_amount_check
		 BEFORE INSERT ON orders
		 WHEN NEW.amount < 0
		 BEGIN
		   SELECT RAISE(ABORT, 'amount must be non-negative');
		 END`,
	}
	for _, stmt := range statements {
		if _, err := ins.GetDB().ExecContext(ctx, stmt); err != nil {
			t.Fatalf("exec failed for %q: %v", stmt, err)
		}
	}

	tables, err := ins.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() failed: %v", err)
	}
	if len(tables) != 3 {
		t.Fatalf("GetTables() len = %d, want 3", len(tables))
	}
	for _, table := range tables {
		if strings.HasPrefix(table.Name, "sqlite_") {
			t.Fatalf("system table should be filtered: %s", table.Name)
		}
	}

	users, err := ins.GetTable(ctx, "users")
	if err != nil {
		t.Fatalf("GetTable(users) failed: %v", err)
	}
	if len(users.Columns) != 6 {
		t.Fatalf("users columns = %d, want 6", len(users.Columns))
	}
	if !users.Columns[0].IsPrimaryKey || !users.Columns[0].IsAutoIncrement {
		t.Fatalf("users.id should be primary key autoincrement: %+v", users.Columns[0])
	}
	if users.Columns[1].CheckConstraint != "length(name) > 0" {
		t.Fatalf("users.name check = %q", users.Columns[1].CheckConstraint)
	}
	if users.Columns[4].DefaultValue != "'active'" {
		t.Fatalf("users.status default = %q", users.Columns[4].DefaultValue)
	}
	if len(users.CheckConstraints) != 2 {
		t.Fatalf("users table checks = %d, want 2", len(users.CheckConstraints))
	}

	indexes, err := ins.GetIndexes(ctx, "orders")
	if err != nil {
		t.Fatalf("GetIndexes() failed: %v", err)
	}
	if len(indexes) != 1 {
		t.Fatalf("orders indexes = %d, want 1", len(indexes))
	}
	if indexes[0].Name != "idx_orders_user_region" || strings.Join(indexes[0].Columns, ",") != "user_id,region_id" {
		t.Fatalf("unexpected index: %+v", indexes[0])
	}

	fks, err := ins.GetForeignKeys(ctx, "orders")
	if err != nil {
		t.Fatalf("GetForeignKeys() failed: %v", err)
	}
	if len(fks) != 1 {
		t.Fatalf("orders foreign keys = %d, want 1", len(fks))
	}
	if fks[0].Column != "user_id, region_id" || fks[0].RefColumn != "user_id, group_id" {
		t.Fatalf("unexpected composite foreign key: %+v", fks[0])
	}
	if fks[0].OnDelete != "CASCADE" || fks[0].OnUpdate != "RESTRICT" {
		t.Fatalf("unexpected fk actions: %+v", fks[0])
	}

	checks, err := ins.GetCheckConstraints(ctx, "orders")
	if err != nil {
		t.Fatalf("GetCheckConstraints() failed: %v", err)
	}
	if len(checks) != 0 {
		t.Fatalf("orders table checks = %d, want 0 because amount is column-level", len(checks))
	}

	views, err := ins.GetViews(ctx)
	if err != nil {
		t.Fatalf("GetViews() failed: %v", err)
	}
	if len(views) != 1 || views[0].Name != "active_users" {
		t.Fatalf("unexpected views: %+v", views)
	}
	if !strings.Contains(strings.ToUpper(views[0].Definition), "CREATE VIEW ACTIVE_USERS AS SELECT") {
		t.Fatalf("view definition should preserve raw create SQL, got: %s", views[0].Definition)
	}
	if len(views[0].Columns) != 3 {
		t.Fatalf("view columns = %d, want 3", len(views[0].Columns))
	}

	triggers, err := ins.GetTriggers(ctx, "orders")
	if err != nil {
		t.Fatalf("GetTriggers() failed: %v", err)
	}
	if len(triggers) != 1 {
		t.Fatalf("triggers = %d, want 1", len(triggers))
	}
	if triggers[0].Timing != "BEFORE" || triggers[0].Event != "INSERT" {
		t.Fatalf("unexpected trigger metadata: %+v", triggers[0])
	}

	procedures, err := ins.GetProcedures(ctx)
	if err != nil {
		t.Fatalf("GetProcedures() failed: %v", err)
	}
	if procedures == nil {
		t.Fatal("procedures should be an empty slice, not nil")
	}
	if len(procedures) != 0 {
		t.Fatalf("procedures = %d, want 0", len(procedures))
	}

	functions, err := ins.GetFunctions(ctx)
	if err != nil {
		t.Fatalf("GetFunctions() failed: %v", err)
	}
	if functions == nil {
		t.Fatal("functions should be an empty slice, not nil")
	}
	if len(functions) != 0 {
		t.Fatalf("functions = %d, want 0", len(functions))
	}

	sequences, err := ins.GetSequences(ctx)
	if err != nil {
		t.Fatalf("GetSequences() failed: %v", err)
	}
	if sequences == nil {
		t.Fatal("sequences should be an empty slice, not nil")
	}
	if len(sequences) != 0 {
		t.Fatalf("sequences = %d, want 0", len(sequences))
	}
}
