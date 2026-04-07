package oracle

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestOracleBuildDSN(t *testing.T) {
	t.Run("returns prefixed dsn as-is", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "oracle://user:pass@host:1521/service"})
		if got := ins.BuildDSN(); got != "oracle://user:pass@host:1521/service" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("adds prefix for traditional dsn", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "user/pass@host:1521/service"})
		if got := ins.BuildDSN(); got != "oracle://user/pass@host:1521/service" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("builds from components using database", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{
			Username: "scott",
			Password: "tiger",
			Host:     "localhost",
			Port:     1521,
			Database: "orclpdb1",
		})
		if got := ins.BuildDSN(); got != "oracle://scott:tiger@localhost:1521/orclpdb1" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("builds from schema when database empty", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{
			Username: "scott",
			Password: "tiger",
			Host:     "localhost",
			Port:     1521,
			Schema:   "orcl",
		})
		if got := ins.BuildDSN(); got != "oracle://scott:tiger@localhost:1521/orcl" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})
}

func TestOracleFactory(t *testing.T) {
	f := &Factory{}
	if got := f.GetType(); got != "oracle" {
		t.Fatalf("unexpected type: %s", got)
	}

	ins, err := f.Create(inspector.ConnectionConfig{Host: "localhost"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ins == nil {
		t.Fatalf("expected inspector instance")
	}
}

func TestOracleConnect(t *testing.T) {
	originalOpen := openDB
	defer func() { openDB = originalOpen }()

	t.Run("open failure", func(t *testing.T) {
		openDB = func(string, string) (*sql.DB, error) {
			return nil, errors.New("open failed")
		}

		ins := NewInspector(inspector.ConnectionConfig{})
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to open oracle connection: open failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		openDB = func(string, string) (*sql.DB, error) {
			return db, nil
		}
		mock.ExpectPing().WillReturnError(errors.New("ping failed"))

		ins := NewInspector(inspector.ConnectionConfig{})
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to ping oracle database: ping failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		openDB = func(string, string) (*sql.DB, error) {
			return db, nil
		}
		mock.ExpectPing()

		ins := NewInspector(inspector.ConnectionConfig{DSN: "oracle://user:pass@host:1521/service"})
		if err := ins.Connect(context.Background()); err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		if ins.GetDB() == nil {
			t.Fatalf("expected db to be set")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}
