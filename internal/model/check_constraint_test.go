package model

import (
	"testing"
)

func TestCheckConstraintGetColumnsString(t *testing.T) {
	tests := []struct {
		name     string
		cc       CheckConstraint
		expected string
	}{
		{
			name: "single column",
			cc: CheckConstraint{
				Name:       "chk_age",
				Definition: "age >= 0",
				Columns:    []string{"age"},
			},
			expected: "age",
		},
		{
			name: "multiple columns",
			cc: CheckConstraint{
				Name:       "chk_date_range",
				Definition: "end_date >= start_date",
				Columns:    []string{"start_date", "end_date"},
			},
			expected: "start_date, end_date",
		},
		{
			name: "empty columns",
			cc: CheckConstraint{
				Name:       "chk_valid",
				Definition: "1=1",
				Columns:    []string{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cc.GetColumnsString(); got != tt.expected {
				t.Errorf("GetColumnsString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
