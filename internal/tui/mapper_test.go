package tui

import (
	"os"
	"strings"
	"testing"
)

func TestToConfig_DSNMode(t *testing.T) {
	s := TuiState{
		DBType:           "mysql",
		ConnectionMethod: ConnectionMethodDSN,
		DSN:              "mysql://user:pass@localhost:3306/db",
		Schema:           "myapp",
		IncludeViews:     true,
		IncludeFunctions: true,
		Tables:           "users,orders",
		OutputDir:        "./output",
		Formats:          []string{"markdown", "sql"},
		SplitFiles:       true,
	}

	cfg := s.ToConfig()

	if cfg.Database.Type != "mysql" {
		t.Fatalf("Type: got %q, want mysql", cfg.Database.Type)
	}
	if cfg.Database.DSN != "mysql://user:pass@localhost:3306/db" {
		t.Fatalf("DSN: got %q", cfg.Database.DSN)
	}
	if cfg.Database.Schema != "myapp" {
		t.Fatalf("Schema (mysql): got %q, want myapp", cfg.Database.Schema)
	}
	// DSN 模式下分离参数应为空
	if cfg.Database.Host != "" {
		t.Fatalf("Host should be empty in DSN mode, got %q", cfg.Database.Host)
	}
	if cfg.Database.Username != "" {
		t.Fatalf("Username should be empty in DSN mode, got %q", cfg.Database.Username)
	}

	if !cfg.Export.IncludeViews {
		t.Fatal("IncludeViews should be true")
	}
	if !cfg.Export.IncludeFunctions {
		t.Fatal("IncludeFunctions should be true")
	}
	if len(cfg.Export.Tables) != 2 || cfg.Export.Tables[0] != "users" || cfg.Export.Tables[1] != "orders" {
		t.Fatalf("Tables: got %v", cfg.Export.Tables)
	}
	if cfg.Export.OutputDir != "./output" {
		t.Fatalf("OutputDir: got %q", cfg.Export.OutputDir)
	}
	if len(cfg.Export.Formats) != 2 || cfg.Export.Formats[0] != "markdown" || cfg.Export.Formats[1] != "sql" {
		t.Fatalf("Formats: got %v", cfg.Export.Formats)
	}
	if !cfg.Export.SplitFiles {
		t.Fatal("SplitFiles should be true")
	}
}

func TestToConfig_ParamsMode(t *testing.T) {
	s := TuiState{
		DBType:            "postgres",
		ConnectionMethod:  ConnectionMethodParams,
		Host:              "db.example.com",
		Port:              "5432",
		Database:          "mydb",
		Username:          "admin",
		Password:          "secret",
		Schema:            "public",
		IncludeProcedures: true,
		IncludeSequences:  true,
		Exclude:           "temp_",
		Patterns:          "^app_.*",
		OutputDir:         "./docs",
		Formats:           []string{"json", "yaml"},
	}

	cfg := s.ToConfig()

	if cfg.Database.Host != "db.example.com" {
		t.Fatalf("Host: got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("Port: got %d", cfg.Database.Port)
	}
	if cfg.Database.Database != "mydb" {
		t.Fatalf("Database: got %q", cfg.Database.Database)
	}
	if cfg.Database.Username != "admin" {
		t.Fatalf("Username: got %q", cfg.Database.Username)
	}
	if cfg.Database.Password != "secret" {
		t.Fatalf("Password: got %q", cfg.Database.Password)
	}
	// Params 模式下 DSN 应为空
	if cfg.Database.DSN != "" {
		t.Fatalf("DSN should be empty in params mode, got %q", cfg.Database.DSN)
	}
	if !cfg.Export.IncludeProcedures {
		t.Fatal("IncludeProcedures should be true")
	}
	if !cfg.Export.IncludeSequences {
		t.Fatal("IncludeSequences should be true")
	}
	if len(cfg.Export.Exclude) != 1 || cfg.Export.Exclude[0] != "temp_" {
		t.Fatalf("Exclude: got %v", cfg.Export.Exclude)
	}
	if len(cfg.Export.Patterns) != 1 || cfg.Export.Patterns[0] != "^app_.*" {
		t.Fatalf("Patterns: got %v", cfg.Export.Patterns)
	}
}

func TestToConfig_PortConversion(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		portStr  string
		expected int
	}{
		{name: "explicit port", dbType: "mysql", portStr: "3306", expected: 3306},
		{name: "zero port", dbType: "mysql", portStr: "0", expected: 0},
		{name: "empty mysql port uses mysql default", dbType: "mysql", portStr: "", expected: 3306},
		{name: "empty postgres port uses postgres default", dbType: "postgres", portStr: "", expected: 5432},
		{name: "empty sqlserver port uses sqlserver default", dbType: "sqlserver", portStr: "", expected: 1433},
		{name: "invalid port falls back to db default", dbType: "mysql", portStr: "abc", expected: 3306},
		{name: "large numeric port is preserved", dbType: "mysql", portStr: "99999", expected: 99999},
	}

	for _, tt := range tests {
		s := TuiState{
			DBType:           tt.dbType,
			ConnectionMethod: ConnectionMethodParams,
			Host:             "localhost",
			Username:         "root",
			Port:             tt.portStr,
		}
		cfg := s.ToConfig()
		if cfg.Database.Port != tt.expected {
			t.Fatalf("%s: Port(%q): got %d, want %d", tt.name, tt.portStr, cfg.Database.Port, tt.expected)
		}
	}
}

func TestToConfig_SchemaNormalization(t *testing.T) {
	tests := []struct {
		dbType   string
		schema   string
		expected string
	}{
		{"oracle", "myschema", "MYSCHEMA"},
		{"dm", "myschema", "MYSCHEMA"},
		{"postgres", "myschema", "myschema"},
		{"mysql", "MySchema", "MySchema"},
	}

	for _, tt := range tests {
		s := TuiState{
			DBType:           tt.dbType,
			ConnectionMethod: ConnectionMethodDSN,
			DSN:              tt.dbType + "://user:pass@host",
			Schema:           tt.schema,
		}
		cfg := s.ToConfig()
		if cfg.Database.Schema != tt.expected {
			t.Fatalf("Schema(%s/%s): got %q, want %q", tt.dbType, tt.schema, cfg.Database.Schema, tt.expected)
		}
	}
}

func TestToConfig_BooleanDefaults(t *testing.T) {
	s := TuiState{
		DBType:           "dm",
		ConnectionMethod: ConnectionMethodDSN,
		DSN:              "dm://user:pass@host:5236",
	}
	cfg := s.ToConfig()

	if cfg.Export.IncludeViews {
		t.Fatal("IncludeViews should default to false")
	}
	if cfg.Export.IncludeProcedures {
		t.Fatal("IncludeProcedures should default to false")
	}
}

func TestToConfig_EnvVarNotOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("DB_HOST", "envhost")
	defer os.Unsetenv("DB_HOST")

	// TuiState 中 Host 为空，通过 ToConfig 也不应该被 env 覆盖
	// 因为 ToConfig 不调用 LoadFromEnv
	s := TuiState{
		DBType:           "mysql",
		ConnectionMethod: ConnectionMethodParams,
		Host:             "",
		Username:         "root",
	}
	cfg := s.ToConfig()

	// Host 应保持空（不调用 LoadFromEnv）
	if cfg.Database.Host != "" {
		t.Fatalf("Host should remain empty (env not loaded), got %q", cfg.Database.Host)
	}
}

func TestToConfig_ValidateDefaults(t *testing.T) {
	s := TuiState{
		DBType:           "dm",
		ConnectionMethod: ConnectionMethodDSN,
		DSN:              "dm://user:pass@host:5236",
		OutputDir:        "",  // 空 → Validate 设置默认
		Formats:          nil, // nil → Validate 设置默认
	}
	cfg := s.ToConfig()

	if cfg.Export.OutputDir != "./output" {
		t.Fatalf("OutputDir should default to ./output, got %q", cfg.Export.OutputDir)
	}
	if len(cfg.Export.Formats) != 1 || cfg.Export.Formats[0] != "markdown" {
		t.Fatalf("Formats should default to [markdown], got %v", cfg.Export.Formats)
	}
}

func TestToConfig_FormatsPreserved(t *testing.T) {
	s := TuiState{
		DBType:           "dm",
		ConnectionMethod: ConnectionMethodDSN,
		DSN:              "dm://user:pass@host:5236",
		Formats:          []string{"JSON", " SQL ", "yaml"},
	}
	cfg := s.ToConfig()

	// Validate 会规范化格式为小写
	if len(cfg.Export.Formats) != 3 {
		t.Fatalf("Formats count: got %d, want 3", len(cfg.Export.Formats))
	}
	if cfg.Export.Formats[0] != "json" || cfg.Export.Formats[1] != "sql" || cfg.Export.Formats[2] != "yaml" {
		t.Fatalf("Formats not normalized: got %v", cfg.Export.Formats)
	}
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"  ", nil},
		{"users", []string{"users"}},
		{"users,orders", []string{"users", "orders"}},
		{" users , orders ", []string{"users", "orders"}},
		{"a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		result := parseCommaSeparated(tt.input)
		if tt.expected == nil && result != nil {
			t.Fatalf("parseCommaSeparated(%q): got %v, want nil", tt.input, result)
		}
		if tt.expected != nil {
			if len(result) != len(tt.expected) {
				t.Fatalf("parseCommaSeparated(%q): got %v, want %v", tt.input, result, tt.expected)
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Fatalf("parseCommaSeparated(%q): got %v, want %v", tt.input, result, tt.expected)
				}
			}
		}
	}
}

func TestApplyContentCheckboxes(t *testing.T) {
	s := &TuiState{}
	cb := checkboxModel{
		items: []checkboxItem{
			{key: "views", selected: true},
			{key: "procedures", selected: false},
			{key: "functions", selected: true},
			{key: "triggers", selected: false},
			{key: "sequences", selected: true},
		},
	}
	s.applyContentCheckboxes(cb)

	if !s.IncludeViews {
		t.Fatal("IncludeViews should be true")
	}
	if s.IncludeProcedures {
		t.Fatal("IncludeProcedures should be false")
	}
	if !s.IncludeFunctions {
		t.Fatal("IncludeFunctions should be true")
	}
	if s.IncludeTriggers {
		t.Fatal("IncludeTriggers should be false")
	}
	if !s.IncludeSequences {
		t.Fatal("IncludeSequences should be true")
	}
}

func TestApplyFormatCheckboxes(t *testing.T) {
	s := &TuiState{}
	cb := checkboxModel{
		items: []checkboxItem{
			{key: "markdown", selected: true},
			{key: "sql", selected: false},
			{key: "json", selected: true},
		},
	}
	s.applyFormatCheckboxes(cb)

	if len(s.Formats) != 2 || s.Formats[0] != "markdown" || s.Formats[1] != "json" {
		t.Fatalf("Formats: got %v, want [markdown, json]", s.Formats)
	}
}

func TestDefaultTuiState_EnvVars(t *testing.T) {
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_HOST", "db.local")
	os.Setenv("EXPORT_INCLUDE_VIEWS", "true")
	os.Setenv("EXPORT_FORMATS", "sql,json")
	defer func() {
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("EXPORT_INCLUDE_VIEWS")
		os.Unsetenv("EXPORT_FORMATS")
	}()

	s := defaultTuiState()

	if s.DBType != "postgres" {
		t.Fatalf("DBType: got %q, want postgres", s.DBType)
	}
	if s.Host != "db.local" {
		t.Fatalf("Host: got %q, want db.local", s.Host)
	}
	if !s.IncludeViews {
		t.Fatal("IncludeViews should be true from env")
	}
	if len(s.Formats) != 2 || s.Formats[0] != "sql" || s.Formats[1] != "json" {
		t.Fatalf("Formats: got %v", s.Formats)
	}
}

func TestDefaultTuiState_Defaults(t *testing.T) {
	// 不设环境变量，测试默认值
	s := defaultTuiState()

	if s.DBType != "dm" {
		t.Fatalf("DBType: got %q, want dm", s.DBType)
	}
	if s.OutputDir != "./output" {
		t.Fatalf("OutputDir: got %q, want ./output", s.OutputDir)
	}
	if len(s.Formats) != 1 || s.Formats[0] != "markdown" {
		t.Fatalf("Formats: got %v", s.Formats)
	}
	if s.ConnectionMethod != ConnectionMethodDSN {
		t.Fatalf("ConnectionMethod should default to DSN")
	}
}

func TestPageInputCount(t *testing.T) {
	if n := pageInputCount(PageDSNForm); n != 2 {
		t.Fatalf("DSNForm count: got %d, want 2", n)
	}
	if n := pageInputCount(PageParamsForm); n != 6 {
		t.Fatalf("ParamsForm count: got %d, want 6", n)
	}
	if n := pageInputCount(PageWelcome); n != 0 {
		t.Fatalf("Welcome count: got %d, want 0", n)
	}
}

func TestNextPage(t *testing.T) {
	if p := nextPage(PageWelcome); p != PageDBType {
		t.Fatalf("Welcome next: got %v, want DBType", p)
	}
	if p := nextPage(PageConfirm); p != PageExecution {
		t.Fatalf("Confirm next: got %v, want Execution", p)
	}
}

func TestCheckboxModel_SelectedKeys(t *testing.T) {
	cb := checkboxModel{
		items: []checkboxItem{
			{key: "a", selected: true},
			{key: "b", selected: false},
			{key: "c", selected: true},
		},
	}
	keys := cb.selectedKeys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "c" {
		t.Fatalf("selectedKeys: got %v, want [a, c]", keys)
	}
}

func TestCheckboxModel_View(t *testing.T) {
	cb := checkboxModel{
		items: []checkboxItem{
			{title: "Option 1", selected: true},
			{title: "Option 2", selected: false},
		},
		cursor: 0,
	}
	view := cb.view()
	if !strings.Contains(view, "[x]") {
		t.Fatal("view should contain [x] for selected item")
	}
	if !strings.Contains(view, "[ ]") {
		t.Fatal("view should contain [ ] for unselected item")
	}
	if !strings.Contains(view, "Option 1") || !strings.Contains(view, "Option 2") {
		t.Fatal("view should contain option titles")
	}
}
