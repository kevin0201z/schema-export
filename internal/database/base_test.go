package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestBaseInspectorBuildDSN(t *testing.T) {
	ins := NewBaseInspector(inspector.ConnectionConfig{})
	if got := ins.BuildDSN(); got != "" {
		t.Fatalf("expected empty DSN, got %q", got)
	}

	ins = NewBaseInspector(inspector.ConnectionConfig{DSN: "driver://user:pass@host"})
	if got := ins.BuildDSN(); got != "driver://user:pass@host" {
		t.Fatalf("unexpected DSN: %q", got)
	}
}

func TestBaseInspectorConnect(t *testing.T) {
	ins := NewBaseInspector(inspector.ConnectionConfig{})
	if err := ins.Connect(context.Background()); err == nil || err.Error() != "not implemented" {
		t.Fatalf("expected not implemented error, got %v", err)
	}
}

func TestBaseInspectorClose(t *testing.T) {
	t.Run("nil db", func(t *testing.T) {
		ins := NewBaseInspector(inspector.ConnectionConfig{})
		if err := ins.Close(); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})

	t.Run("real db", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}

		ins := NewBaseInspector(inspector.ConnectionConfig{})
		ins.SetDB(db)
		mock.ExpectClose()
		if err := ins.Close(); err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}

func TestBaseInspectorTestConnection(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		ins := NewBaseInspector(inspector.ConnectionConfig{})
		if err := ins.TestConnection(context.Background()); err == nil || err.Error() != "database not connected" {
			t.Fatalf("expected not connected error, got %v", err)
		}
	})

	t.Run("ping success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectPing()

		ins := NewBaseInspector(inspector.ConnectionConfig{})
		ins.SetDB(db)

		if err := ins.TestConnection(context.Background()); err != nil {
			t.Fatalf("TestConnection failed: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})

	t.Run("ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectPing().WillReturnError(errors.New("ping failed"))

		ins := NewBaseInspector(inspector.ConnectionConfig{})
		ins.SetDB(db)

		if err := ins.TestConnection(context.Background()); err == nil || err.Error() != "ping failed" {
			t.Fatalf("expected ping failure, got %v", err)
		}
	})
}
