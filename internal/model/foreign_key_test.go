package model

import (
	"testing"
)

func TestGetReferenceString(t *testing.T) {
	tests := []struct {
		name       string
		fk         ForeignKey
		expected   string
	}{
		{
			name: "normal reference",
			fk: ForeignKey{
				RefTable:  "users",
				RefColumn: "id",
			},
			expected: "users(id)",
		},
		{
			name: "self reference",
			fk: ForeignKey{
				RefTable:  "employees",
				RefColumn: "manager_id",
			},
			expected: "employees(manager_id)",
		},
		{
			name: "empty reference",
			fk: ForeignKey{
				RefTable:  "",
				RefColumn: "",
			},
			expected: "()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fk.GetReferenceString()
			if result != tt.expected {
				t.Errorf("GetReferenceString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetOnDeleteRule(t *testing.T) {
	tests := []struct {
		name     string
		onDelete string
		expected string
	}{
		{
			name:     "cascade",
			onDelete: "CASCADE",
			expected: "CASCADE",
		},
		{
			name:     "set null",
			onDelete: "SET NULL",
			expected: "SET NULL",
		},
		{
			name:     "restrict",
			onDelete: "RESTRICT",
			expected: "RESTRICT",
		},
		{
			name:     "no action",
			onDelete: "NO ACTION",
			expected: "NO ACTION",
		},
		{
			name:     "empty defaults to no action",
			onDelete: "",
			expected: "NO ACTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fk := ForeignKey{OnDelete: tt.onDelete}
			result := fk.GetOnDeleteRule()
			if result != tt.expected {
				t.Errorf("GetOnDeleteRule() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetOnUpdateRule(t *testing.T) {
	tests := []struct {
		name     string
		onUpdate string
		expected string
	}{
		{
			name:     "cascade",
			onUpdate: "CASCADE",
			expected: "CASCADE",
		},
		{
			name:     "set null",
			onUpdate: "SET NULL",
			expected: "SET NULL",
		},
		{
			name:     "restrict",
			onUpdate: "RESTRICT",
			expected: "RESTRICT",
		},
		{
			name:     "empty defaults to no action",
			onUpdate: "",
			expected: "NO ACTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fk := ForeignKey{OnUpdate: tt.onUpdate}
			result := fk.GetOnUpdateRule()
			if result != tt.expected {
				t.Errorf("GetOnUpdateRule() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
