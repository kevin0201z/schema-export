package model

import (
	"testing"
)

func TestGetPrimaryKey(t *testing.T) {
	tests := []struct {
		name         string
		columns      []Column
		expectedName string
		expectedNil  bool
	}{
		{
			name: "single primary key",
			columns: []Column{
				{Name: "id", IsPrimaryKey: true},
				{Name: "name", IsPrimaryKey: false},
			},
			expectedName: "id",
			expectedNil:  false,
		},
		{
			name: "no primary key",
			columns: []Column{
				{Name: "name", IsPrimaryKey: false},
				{Name: "email", IsPrimaryKey: false},
			},
			expectedName: "",
			expectedNil:  true,
		},
		{
			name: "multiple columns no pk",
			columns: []Column{
				{Name: "col1", IsPrimaryKey: false},
				{Name: "col2", IsPrimaryKey: false},
				{Name: "col3", IsPrimaryKey: false},
			},
			expectedName: "",
			expectedNil:  true,
		},
		{
			name:         "empty columns",
			columns:      []Column{},
			expectedName: "",
			expectedNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := Table{Columns: tt.columns}
			pk := table.GetPrimaryKey()

			if tt.expectedNil {
				if pk != nil {
					t.Errorf("expected nil, got %v", pk)
				}
			} else {
				if pk == nil {
					t.Errorf("expected non-nil, got nil")
				} else if pk.Name != tt.expectedName {
					t.Errorf("expected name %q, got %q", tt.expectedName, pk.Name)
				}
			}
		})
	}
}

func TestGetColumnByName(t *testing.T) {
	columns := []Column{
		{Name: "id", DataType: "INT"},
		{Name: "username", DataType: "VARCHAR"},
		{Name: "email", DataType: "VARCHAR"},
	}
	table := Table{Columns: columns}

	tests := []struct {
		name         string
		searchName   string
		expectedNil  bool
		expectedType string
	}{
		{
			name:         "existing column",
			searchName:   "username",
			expectedNil:  false,
			expectedType: "VARCHAR",
		},
		{
			name:        "non-existing column",
			searchName:  "password",
			expectedNil: true,
		},
		{
			name:         "first column",
			searchName:   "id",
			expectedNil:  false,
			expectedType: "INT",
		},
		{
			name:        "empty name",
			searchName:  "",
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := table.GetColumnByName(tt.searchName)

			if tt.expectedNil {
				if col != nil {
					t.Errorf("expected nil for %q, got %v", tt.searchName, col)
				}
			} else {
				if col == nil {
					t.Errorf("expected non-nil for %q, got nil", tt.searchName)
				} else if col.DataType != tt.expectedType {
					t.Errorf("expected type %q, got %q", tt.expectedType, col.DataType)
				}
			}
		})
	}
}
