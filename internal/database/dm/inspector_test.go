package dm

import (
	"testing"

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
