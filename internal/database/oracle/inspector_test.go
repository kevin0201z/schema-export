package oracle

import (
	"testing"

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
