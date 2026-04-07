package dm

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestDMBuildDSN(t *testing.T) {
	t.Run("returns prefixed dsn as-is", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "dm://user:pass@host:5236"})
		if got := ins.BuildDSN(); got != "dm://user:pass@host:5236" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("adds prefix for raw dsn", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "user:pass@host:5236"})
		if got := ins.BuildDSN(); got != "dm://user:pass@host:5236" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("builds from components", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{
			Username: "sysdba",
			Password: "dameng",
			Host:     "127.0.0.1",
			Port:     5236,
		})
		if got := ins.BuildDSN(); got != "dm://sysdba:dameng@127.0.0.1:5236" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})
}

func TestDMFactory(t *testing.T) {
	f := &Factory{}
	if got := f.GetType(); got != "dm" {
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

func TestDMConnect(t *testing.T) {
	originalOpen := openDB
	defer func() { openDB = originalOpen }()

	t.Run("open failure", func(t *testing.T) {
		openDB = func(string, string) (*sql.DB, error) {
			return nil, errors.New("open failed")
		}

		ins := NewInspector(inspector.ConnectionConfig{})
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to open dm connection: open failed" {
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
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to ping dm database: ping failed" {
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

		ins := NewInspector(inspector.ConnectionConfig{DSN: "dm://user:pass@host:5236"})
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
