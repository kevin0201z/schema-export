package database_test

import (
	"testing"

	"github.com/schema-export/schema-export/internal/database/postgres"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestPostgresBuildDSN(t *testing.T) {
	cfg := inspector.ConnectionConfig{DSN: "postgres://u:p@h:5432/db"}
	ins := postgres.NewInspector(cfg)
	if got := ins.BuildDSN(); got != cfg.DSN {
		t.Fatalf("expected %s got %s", cfg.DSN, got)
	}

	cfg2 := inspector.ConnectionConfig{DSN: "postgresql://u:p@h:5432/db"}
	ins2 := postgres.NewInspector(cfg2)
	if got := ins2.BuildDSN(); got != cfg2.DSN {
		t.Fatalf("expected %s got %s", cfg2.DSN, got)
	}

	cfg3 := inspector.ConnectionConfig{DSN: "u:p@h:5432/db"}
	ins3 := postgres.NewInspector(cfg3)
	if got := ins3.BuildDSN(); got != "postgres://"+cfg3.DSN {
		t.Fatalf("expected prefixed DSN, got %s", got)
	}

	cfg4 := inspector.ConnectionConfig{Username: "u", Password: "p", Host: "h", Port: 5432, Database: "db"}
	ins4 := postgres.NewInspector(cfg4)
	got := ins4.BuildDSN()
	if got == "" {
		t.Fatalf("expected non-empty DSN built from components")
	}
	expectedPrefix := "postgres://u:p@h:5432/db"
	if len(got) < len(expectedPrefix) || got[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("expected DSN to start with %s, got %s", expectedPrefix, got)
	}
}

func TestPostgresBuildDSNWithSSL(t *testing.T) {
	cfg := inspector.ConnectionConfig{
		Username: "u",
		Password: "p",
		Host:     "h",
		Port:     5432,
		Database: "db",
		SSLMode:  "require",
	}
	ins := postgres.NewInspector(cfg)
	got := ins.BuildDSN()
	if !contains(got, "sslmode=require") {
		t.Fatalf("expected sslmode=require in DSN, got %s", got)
	}
}

func TestPostgresFactory(t *testing.T) {
	factory := &postgres.Factory{}
	if factory.GetType() != "postgres" {
		t.Fatalf("expected type 'postgres', got '%s'", factory.GetType())
	}

	cfg := inspector.ConnectionConfig{Type: "postgres"}
	ins, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ins == nil {
		t.Fatal("expected non-nil inspector")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
