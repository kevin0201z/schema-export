package model

import (
	"testing"
)

func TestGetColumnsString(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		expected string
	}{
		{
			name:     "single column",
			columns:  []string{"id"},
			expected: "id",
		},
		{
			name:     "multiple columns",
			columns:  []string{"first_name", "last_name"},
			expected: "first_name, last_name",
		},
		{
			name:     "empty columns",
			columns:  []string{},
			expected: "",
		},
		{
			name:     "three columns",
			columns:  []string{"col1", "col2", "col3"},
			expected: "col1, col2, col3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := Index{Columns: tt.columns}
			result := idx.GetColumnsString()
			if result != tt.expected {
				t.Errorf("GetColumnsString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIndexTypeConstants(t *testing.T) {
	// 验证索引类型常量
	if IndexTypeNormal != "NORMAL" {
		t.Errorf("IndexTypeNormal = %q, expected 'NORMAL'", IndexTypeNormal)
	}
	if IndexTypeUnique != "UNIQUE" {
		t.Errorf("IndexTypeUnique = %q, expected 'UNIQUE'", IndexTypeUnique)
	}
	if IndexTypePrimary != "PRIMARY" {
		t.Errorf("IndexTypePrimary = %q, expected 'PRIMARY'", IndexTypePrimary)
	}
}
