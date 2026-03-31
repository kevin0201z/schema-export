package config

import (
	"testing"

	"github.com/schema-export/schema-export/internal/model"
)

func TestNewTableFilter(t *testing.T) {
	tests := []struct {
		name      string
		include   []string
		exclude   []string
		patterns  []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "empty filter",
			wantErr: false,
		},
		{
			name:    "with include list",
			include: []string{"users", "orders"},
			wantErr: false,
		},
		{
			name:    "with exclude list",
			exclude: []string{"temp_", "log_"},
			wantErr: false,
		},
		{
			name:     "with valid pattern",
			patterns: []string{"^sys_.*"},
			wantErr:  false,
		},
		{
			name:      "with invalid pattern",
			patterns:  []string{"[invalid"},
			wantErr:   true,
			errSubstr: "invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewTableFilter(tt.include, tt.exclude, tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTableFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errSubstr != "" {
				if !contains(err.Error(), tt.errSubstr) {
					t.Errorf("error message '%s' should contain '%s'", err.Error(), tt.errSubstr)
				}
			}
			if err == nil && filter == nil {
				t.Error("expected filter to be non-nil")
			}
		})
	}
}

func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name      string
		include   []string
		exclude   []string
		patterns  []string
		tableName string
		want      bool
	}{
		{
			name:      "empty filter includes all",
			tableName: "any_table",
			want:      true,
		},
		{
			name:      "include list match",
			include:   []string{"users", "orders"},
			tableName: "users",
			want:      true,
		},
		{
			name:      "include list no match",
			include:   []string{"users", "orders"},
			tableName: "products",
			want:      false,
		},
		{
			name:      "exclude list match",
			exclude:   []string{"temp_data", "log_table"},
			tableName: "temp_data",
			want:      false,
		},
		{
			name:      "exclude list no match",
			exclude:   []string{"temp_data", "log_table"},
			tableName: "users",
			want:      true,
		},
		{
			name:      "pattern match",
			patterns:  []string{"^sys_.*"},
			tableName: "sys_config",
			want:      true,
		},
		{
			name:      "pattern no match",
			patterns:  []string{"^sys_.*"},
			tableName: "users",
			want:      false,
		},
		{
			name:      "include and exclude combined",
			include:   []string{"users", "orders", "temp_users"},
			exclude:   []string{"temp_users"},
			tableName: "temp_users",
			want:      false,
		},
		{
			name:      "include and exclude combined - included",
			include:   []string{"users", "orders"},
			exclude:   []string{"temp_"},
			tableName: "users",
			want:      true,
		},
		{
			name:      "case insensitive include",
			include:   []string{"USERS", "Orders"},
			tableName: "users",
			want:      true,
		},
		{
			name:      "case insensitive exclude",
			exclude:   []string{"TEMP_DATA"},
			tableName: "temp_data",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewTableFilter(tt.include, tt.exclude, tt.patterns)
			if err != nil {
				t.Fatalf("NewTableFilter() failed: %v", err)
			}

			got := filter.ShouldInclude(tt.tableName)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.tableName, got, tt.want)
			}
		})
	}
}

func TestFilterTables(t *testing.T) {
	tables := []model.Table{
		{Name: "users"},
		{Name: "orders"},
		{Name: "products"},
		{Name: "temp_data"},
	}

	filter, err := NewTableFilter(
		[]string{"users", "orders", "products"},
		[]string{"temp_data"},
		nil,
	)
	if err != nil {
		t.Fatalf("NewTableFilter() failed: %v", err)
	}

	filtered := filter.FilterTables(tables)

	if len(filtered) != 3 {
		t.Errorf("expected 3 tables, got %d", len(filtered))
	}

	// 验证 temp_data 被排除
	for _, table := range filtered {
		if table.Name == "temp_data" {
			t.Error("temp_data should be excluded")
		}
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
