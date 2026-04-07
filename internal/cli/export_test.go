package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	exportapp "github.com/schema-export/schema-export/internal/app/export"
	"github.com/schema-export/schema-export/internal/config"
	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

type stubExportService struct {
	runCalled bool
	cfg       *config.Config
	runErr    error
}

func (s *stubExportService) Run() error {
	s.runCalled = true
	return s.runErr
}

func clearExportEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"DB_TYPE",
		"DB_HOST",
		"DB_PORT",
		"DB_DATABASE",
		"DB_USERNAME",
		"DB_PASSWORD",
		"DB_DSN",
		"DB_SCHEMA",
		"EXPORT_OUTPUT",
		"EXPORT_FORMATS",
		"EXPORT_SPLIT",
		"EXPORT_INCLUDE_VIEWS",
		"EXPORT_INCLUDE_PROCEDURES",
		"EXPORT_INCLUDE_FUNCTIONS",
		"EXPORT_INCLUDE_TRIGGERS",
		"EXPORT_INCLUDE_SEQUENCES",
	} {
		t.Setenv(key, "")
	}
}

func TestParseFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{"markdown"},
		},
		{
			name:     "single format",
			input:    "sql",
			expected: []string{"sql"},
		},
		{
			name:     "multiple formats",
			input:    "markdown,sql",
			expected: []string{"markdown", "sql"},
		},
		{
			name:     "with spaces",
			input:    "markdown , sql",
			expected: []string{"markdown", "sql"},
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "MARKDOWN,SQL",
			expected: []string{"markdown", "sql"},
		},
		{
			name:     "empty parts ignored",
			input:    "markdown,,sql",
			expected: []string{"markdown", "sql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFormats(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func TestParseTables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single table",
			input:    "users",
			expected: []string{"users"},
		},
		{
			name:     "multiple tables",
			input:    "users,orders,products",
			expected: []string{"users", "orders", "products"},
		},
		{
			name:     "with spaces",
			input:    "users , orders",
			expected: []string{"users", "orders"},
		},
		{
			name:     "empty parts ignored",
			input:    "users,,orders",
			expected: []string{"users", "orders"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTables(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, result)
					return
				}
			}
		})
	}
}

func TestParseOutputPath(t *testing.T) {
	tests := []struct {
		name         string
		outputPath   string
		format       string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "empty path",
			outputPath:   "",
			format:       "markdown",
			expectedDir:  "./output",
			expectedFile: "",
		},
		{
			name:         "directory only",
			outputPath:   "./docs",
			format:       "markdown",
			expectedDir:  "./docs",
			expectedFile: "",
		},
		{
			name:         "file with md extension",
			outputPath:   "./docs/schema.md",
			format:       "markdown",
			expectedDir:  "docs",
			expectedFile: "schema.md",
		},
		{
			name:         "file with sql extension",
			outputPath:   "./docs/schema.sql",
			format:       "sql",
			expectedDir:  "docs",
			expectedFile: "schema.sql",
		},
		{
			name:         "markdown format changes txt to md",
			outputPath:   "./docs/schema.txt",
			format:       "markdown",
			expectedDir:  "docs",
			expectedFile: "schema.md",
		},
		{
			name:         "sql format changes txt to sql",
			outputPath:   "./docs/schema.txt",
			format:       "sql",
			expectedDir:  "docs",
			expectedFile: "schema.sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, file := exportapp.ParseOutputPath(tt.outputPath, tt.format)

			if dir != tt.expectedDir {
				t.Errorf("expected dir %q, got %q", tt.expectedDir, dir)
			}
			if file != tt.expectedFile {
				t.Errorf("expected file %q, got %q", tt.expectedFile, file)
			}
		})
	}
}

type testExporter struct {
	exportErr error
}

func (e *testExporter) Export(_ []model.Table, _ []model.View, _ []model.Procedure, _ []model.Function, _ []model.Trigger, _ []model.Sequence, _ exporter.ExportOptions) error {
	return e.exportErr
}
func (e *testExporter) GetName() string      { return "test" }
func (e *testExporter) GetExtension() string { return ".test" }

type testExporterFactory struct {
	exportErr error
}

func (f *testExporterFactory) Create() (exporter.Exporter, error) {
	return &testExporter{exportErr: f.exportErr}, nil
}

func (f *testExporterFactory) GetType() string { return "test" }

func registerTestExporterFactory(err error) string {
	name := fmt.Sprintf("test-exporter-%d", time.Now().UnixNano())
	exporter.Register(name, &testExporterFactory{exportErr: err})
	return name
}

func TestExportAllFormats(t *testing.T) {
	successFormat := registerTestExporterFactory(nil)
	failedFormat := registerTestExporterFactory(errors.New("boom"))

	tests := []struct {
		name    string
		formats []string
		wantErr string
	}{
		{
			name:    "all succeed",
			formats: []string{successFormat},
		},
		{
			name:    "partial failure",
			formats: []string{successFormat, failedFormat},
			wantErr: "partial export failure",
		},
		{
			name:    "all fail",
			formats: []string{failedFormat},
			wantErr: "all exports failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewExportCommand()
			cmd.SetFormats(tt.formats)

			err := exportapp.NewService(cmd.Config).ExportAllFormats([]model.Table{{Name: "users"}}, nil, nil, nil, nil, nil)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestExportCommandRunValidatesConfig(t *testing.T) {
	clearExportEnv(t)

	cmd := NewExportCommand()
	cmd.Config.Database.Type = ""

	err := cmd.Run()
	if err == nil || !strings.Contains(err.Error(), "database type is required") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestExportCommandRunInvokesService(t *testing.T) {
	clearExportEnv(t)

	cmd := NewExportCommand()
	cmd.Config.Database.Type = "mysql"
	cmd.Config.Database.Host = "127.0.0.1"
	cmd.Config.Database.Username = "root"
	cmd.Config.Export.Formats = []string{"markdown"}

	originalNewService := newService
	defer func() { newService = originalNewService }()

	stub := &stubExportService{}
	newService = func(cfg *config.Config) exportService {
		stub.cfg = cfg
		return stub
	}

	if err := cmd.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}
	if !stub.runCalled {
		t.Fatalf("expected service Run to be called")
	}
	if stub.cfg != cmd.Config {
		t.Fatalf("expected config pointer to be forwarded")
	}
}
