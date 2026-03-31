package model

import (
	"testing"
)

func TestGetFullDataType(t *testing.T) {
	tests := []struct {
		name     string
		column   Column
		expected string
	}{
		{
			name:     "simple type",
			column:   Column{DataType: "INT"},
			expected: "INT",
		},
		{
			name:     "varchar with length",
			column:   Column{DataType: "VARCHAR", Length: 255},
			expected: "VARCHAR(255)",
		},
		{
			name:     "decimal with precision",
			column:   Column{DataType: "DECIMAL", Precision: 10},
			expected: "DECIMAL(10)",
		},
		{
			name:     "decimal with precision and scale",
			column:   Column{DataType: "DECIMAL", Precision: 10, Scale: 2},
			expected: "DECIMAL(10,2)",
		},
		{
			name:     "length takes precedence over precision",
			column:   Column{DataType: "VARCHAR", Length: 100, Precision: 10},
			expected: "VARCHAR(100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.column.GetFullDataType()
			if result != tt.expected {
				t.Errorf("GetFullDataType() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected bool
	}{
		{"INT", "INT", true},
		{"INTEGER", "INTEGER", true},
		{"BIGINT", "BIGINT", true},
		{"SMALLINT", "SMALLINT", true},
		{"TINYINT", "TINYINT", true},
		{"DECIMAL", "DECIMAL", true},
		{"NUMERIC", "NUMERIC", true},
		{"FLOAT", "FLOAT", true},
		{"DOUBLE", "DOUBLE", true},
		{"REAL", "REAL", true},
		{"NUMBER", "NUMBER", true},
		{"VARCHAR", "VARCHAR", false},
		{"CHAR", "CHAR", false},
		{"TEXT", "TEXT", false},
		{"DATE", "DATE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := Column{DataType: tt.dataType}
			result := col.IsNumeric()
			if result != tt.expected {
				t.Errorf("IsNumeric() for %s = %v, expected %v", tt.dataType, result, tt.expected)
			}
		})
	}
}

func TestIsString(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected bool
	}{
		{"VARCHAR", "VARCHAR", true},
		{"VARCHAR2", "VARCHAR2", true},
		{"CHAR", "CHAR", true},
		{"NCHAR", "NCHAR", true},
		{"NVARCHAR", "NVARCHAR", true},
		{"NVARCHAR2", "NVARCHAR2", true},
		{"TEXT", "TEXT", true},
		{"CLOB", "CLOB", true},
		{"NCLOB", "NCLOB", true},
		{"INT", "INT", false},
		{"NUMBER", "NUMBER", false},
		{"DATE", "DATE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := Column{DataType: tt.dataType}
			result := col.IsString()
			if result != tt.expected {
				t.Errorf("IsString() for %s = %v, expected %v", tt.dataType, result, tt.expected)
			}
		})
	}
}
