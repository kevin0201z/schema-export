package tui

import (
	"testing"
)

func TestNewCommand_HasUse(t *testing.T) {
	cmd := NewCommand()
	if cmd.Use != "tui" {
		t.Fatalf("Use: got %q, want tui", cmd.Use)
	}
	if cmd.Short == "" {
		t.Fatal("Short should not be empty")
	}
}

func TestNewCommand_IsRunnable(t *testing.T) {
	cmd := NewCommand()
	if cmd.RunE == nil {
		t.Fatal("RunE should not be nil")
	}
}
