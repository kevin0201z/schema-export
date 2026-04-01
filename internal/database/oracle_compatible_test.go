package database_test

import (
	"testing"

	"github.com/schema-export/schema-export/internal/database/dm"
	"github.com/schema-export/schema-export/internal/database/oracle"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestPlaceholderStr(t *testing.T) {
	t.Skip("placeholder test is in package 'database' (placeholder_test.go)")
}

func TestOracleBuildDSN(t *testing.T) {
	// DSN provided with oracle:// prefix should return as-is
	cfg := inspector.ConnectionConfig{DSN: "oracle://u:p@h:1521/svc"}
	ins := oracle.NewInspector(cfg)
	if got := ins.BuildDSN(); got != cfg.DSN {
		t.Fatalf("expected %s got %s", cfg.DSN, got)
	}

	// DSN provided without prefix but with @ should get prefixed
	cfg2 := inspector.ConnectionConfig{DSN: "u/p@h:1521/svc"}
	ins2 := oracle.NewInspector(cfg2)
	if got := ins2.BuildDSN(); got != "oracle://"+cfg2.DSN {
		t.Fatalf("expected prefixed DSN, got %s", got)
	}

	// Build from components
	cfg3 := inspector.ConnectionConfig{Username: "u", Password: "p", Host: "h", Port: 1521, Database: "db"}
	ins3 := oracle.NewInspector(cfg3)
	got := ins3.BuildDSN()
	if got == "" {
		t.Fatalf("expected non-empty DSN built from components")
	}
}

func TestDMBuildDSN(t *testing.T) {
	// DSN with dm:// prefix stays
	cfg := inspector.ConnectionConfig{DSN: "dm://u:p@h:5236"}
	ins := dm.NewInspector(cfg)
	if got := ins.BuildDSN(); got != cfg.DSN {
		t.Fatalf("expected %s got %s", cfg.DSN, got)
	}

	// DSN without prefix gets prefixed
	cfg2 := inspector.ConnectionConfig{DSN: "u:p@h:5236"}
	ins2 := dm.NewInspector(cfg2)
	if got := ins2.BuildDSN(); got != "dm://"+cfg2.DSN {
		t.Fatalf("expected prefixed DM DSN, got %s", got)
	}

	// Build from components
	cfg3 := inspector.ConnectionConfig{Username: "u", Password: "p", Host: "h", Port: 5236}
	ins3 := dm.NewInspector(cfg3)
	if got := ins3.BuildDSN(); got == "" {
		t.Fatalf("expected non-empty DM DSN from components")
	}
}
