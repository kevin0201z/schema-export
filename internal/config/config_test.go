package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Database.Type != "dm" {
		t.Errorf("expected default type 'dm', got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Port != 5236 {
		t.Errorf("expected default port 5236, got %d", cfg.Database.Port)
	}

	if cfg.Export.OutputDir != "./output" {
		t.Errorf("expected default output dir './output', got '%s'", cfg.Export.OutputDir)
	}

	if len(cfg.Export.Formats) != 1 || cfg.Export.Formats[0] != "markdown" {
		t.Errorf("expected default format ['markdown'], got %v", cfg.Export.Formats)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 保存原始环境变量
	origType := os.Getenv("DB_TYPE")
	origHost := os.Getenv("DB_HOST")
	origPort := os.Getenv("DB_PORT")
	origOutput := os.Getenv("EXPORT_OUTPUT")
	origFormats := os.Getenv("EXPORT_FORMATS")
	origSchema := os.Getenv("DB_SCHEMA")

	// 测试后恢复
	defer func() {
		os.Setenv("DB_TYPE", origType)
		os.Setenv("DB_HOST", origHost)
		os.Setenv("DB_PORT", origPort)
		os.Setenv("EXPORT_OUTPUT", origOutput)
		os.Setenv("EXPORT_FORMATS", origFormats)
		os.Setenv("DB_SCHEMA", origSchema)
	}()

	// 设置测试环境变量
	os.Setenv("DB_TYPE", "oracle")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "1521")
	os.Setenv("EXPORT_OUTPUT", "./testoutput")
	os.Setenv("EXPORT_FORMATS", "MARKDOWN, sql ,")
	os.Setenv("DB_SCHEMA", "test_schema")

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	if cfg.Database.Type != "oracle" {
		t.Errorf("expected type 'oracle', got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Host != "testhost" {
		t.Errorf("expected host 'testhost', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != 1521 {
		t.Errorf("expected port 1521, got %d", cfg.Database.Port)
	}

	if cfg.Export.OutputDir != "./testoutput" {
		t.Errorf("expected output dir './testoutput', got '%s'", cfg.Export.OutputDir)
	}

	if len(cfg.Export.Formats) != 2 || cfg.Export.Formats[0] != "markdown" || cfg.Export.Formats[1] != "sql" {
		t.Errorf("expected normalized formats ['markdown', 'sql'], got %v", cfg.Export.Formats)
	}

	if cfg.Database.Schema != "test_schema" {
		t.Errorf("expected schema from env to remain unchanged before validate, got '%s'", cfg.Database.Schema)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with DSN",
			config: &Config{
				Database: DatabaseConfig{
					Type: "dm",
					DSN:  "dm://user:pass@localhost:5236?schema=test_schema",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with host and username",
			config: &Config{
				Database: DatabaseConfig{
					Type:     "oracle",
					Host:     "localhost",
					Username: "scott",
				},
			},
			wantErr: false,
		},
		{
			name: "missing type",
			config: &Config{
				Database: DatabaseConfig{
					Host: "localhost",
				},
			},
			wantErr: true,
		},
		{
			name: "missing host and DSN",
			config: &Config{
				Database: DatabaseConfig{
					Type: "dm",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: &Config{
				Database: DatabaseConfig{
					Type: "dm",
					Host: "localhost",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNormalizesSchemaAndFormats(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Type: "oracle",
			DSN:  "oracle://user:pass@localhost:1521/ORCL?schema=test_schema",
		},
		Export: ExportConfig{
			Formats: []string{"MARKDOWN", " sql ", ""},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if cfg.Database.Schema != "TEST_SCHEMA" {
		t.Fatalf("expected normalized schema TEST_SCHEMA, got %q", cfg.Database.Schema)
	}

	if len(cfg.Export.Formats) != 2 || cfg.Export.Formats[0] != "markdown" || cfg.Export.Formats[1] != "sql" {
		t.Fatalf("expected normalized formats [markdown sql], got %v", cfg.Export.Formats)
	}
}

func TestNormalizeSchema(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		schema   string
		expected string
	}{
		{name: "oracle uppercases", dbType: "oracle", schema: "app", expected: "APP"},
		{name: "dm uppercases", dbType: "dm", schema: "test_schema", expected: "TEST_SCHEMA"},
		{name: "quoted schema preserved", dbType: "oracle", schema: "\"app\"", expected: "\"app\""},
		{name: "other database unchanged", dbType: "postgres", schema: "public", expected: "public"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSchema(tt.dbType, tt.schema); got != tt.expected {
				t.Fatalf("normalizeSchema(%q, %q) = %q, want %q", tt.dbType, tt.schema, got, tt.expected)
			}
		})
	}
}

func TestExtractSchemaFromDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected string
	}{
		{
			dsn:      "dm://user:pass@localhost:5236?schema=TEST_SCHEMA",
			expected: "TEST_SCHEMA",
		},
		{
			dsn:      "dm://user:pass@localhost:5236?schema=TEST_SCHEMA&other=value",
			expected: "TEST_SCHEMA",
		},
		{
			dsn:      "dm://user:pass@localhost:5236",
			expected: "",
		},
		{
			dsn:      "oracle://user:pass@localhost:1521/ORCL?schema=OTHER",
			expected: "OTHER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			result := extractSchemaFromDSN(tt.dsn)
			if result != tt.expected {
				t.Errorf("extractSchemaFromDSN() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestToConnectionConfig(t *testing.T) {
	dbConfig := DatabaseConfig{
		Type:     "dm",
		Host:     "localhost",
		Port:     5236,
		Database: "DAMENG",
		Username: "SYSDBA",
		Password: "password",
		DSN:      "dm://user:pass@host",
		Schema:   "TEST",
	}

	connConfig := dbConfig.ToConnectionConfig()

	if connConfig.Type != dbConfig.Type {
		t.Errorf("Type mismatch")
	}
	if connConfig.Host != dbConfig.Host {
		t.Errorf("Host mismatch")
	}
	if connConfig.Port != dbConfig.Port {
		t.Errorf("Port mismatch")
	}
	if connConfig.Username != dbConfig.Username {
		t.Errorf("Username mismatch")
	}
	if connConfig.Password != dbConfig.Password {
		t.Errorf("Password mismatch")
	}
	if connConfig.DSN != dbConfig.DSN {
		t.Errorf("DSN mismatch")
	}
	if connConfig.Schema != dbConfig.Schema {
		t.Errorf("Schema mismatch")
	}
}
