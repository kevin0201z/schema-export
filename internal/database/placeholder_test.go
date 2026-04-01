package database

import (
    "testing"

    "github.com/schema-export/schema-export/internal/inspector"
)

func TestPlaceholderStr(t *testing.T) {
    o := NewOracleCompatibleInspector(inspector.ConnectionConfig{}, PlaceholderColon)
    if got := o.placeholderStr(1); got != ":1" {
        t.Fatalf("expected :1, got %s", got)
    }

    o2 := NewOracleCompatibleInspector(inspector.ConnectionConfig{}, PlaceholderQuestion)
    if got := o2.placeholderStr(1); got != "?" {
        t.Fatalf("expected ?, got %s", got)
    }
}
