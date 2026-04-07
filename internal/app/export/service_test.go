package exportapp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/config"
	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

type mockInspector struct {
	connectErr    error
	testConnErr   error
	getTablesErr  error
	getTableErr   map[string]error
	viewsErr      error
	proceduresErr error
	functionsErr  error
	triggersErr   map[string]error
	sequencesErr  error

	tablesByName  map[string]model.Table
	tables        []model.Table
	views         []model.View
	procedures    []model.Procedure
	functions     []model.Function
	triggersByTbl map[string][]model.Trigger
	sequences     []model.Sequence

	connectCalls       int
	closeCalls         int
	testConnCalls      int
	getTablesCalls     int
	getTableCalls      map[string]int
	getViewsCalls      int
	getProceduresCalls int
	getFunctionsCalls  int
	getTriggersCalls   map[string]int
	getSequencesCalls  int
}

func (m *mockInspector) Connect(context.Context) error {
	m.connectCalls++
	return m.connectErr
}

func (m *mockInspector) Close() error {
	m.closeCalls++
	return nil
}

func (m *mockInspector) TestConnection(context.Context) error {
	m.testConnCalls++
	return m.testConnErr
}

func (m *mockInspector) GetTables(context.Context) ([]model.Table, error) {
	m.getTablesCalls++
	if m.getTablesErr != nil {
		return nil, m.getTablesErr
	}
	return m.tables, nil
}

func (m *mockInspector) GetTable(_ context.Context, tableName string) (*model.Table, error) {
	if m.getTableCalls == nil {
		m.getTableCalls = make(map[string]int)
	}
	m.getTableCalls[tableName]++
	if m.getTableErr != nil {
		if err := m.getTableErr[tableName]; err != nil {
			return nil, err
		}
	}
	if m.tablesByName != nil {
		if table, ok := m.tablesByName[tableName]; ok {
			copied := table
			return &copied, nil
		}
	}
	return nil, fmt.Errorf("table %s not found", tableName)
}

func (m *mockInspector) GetColumns(context.Context, string) ([]model.Column, error) {
	return nil, nil
}

func (m *mockInspector) GetIndexes(context.Context, string) ([]model.Index, error) {
	return nil, nil
}

func (m *mockInspector) GetForeignKeys(context.Context, string) ([]model.ForeignKey, error) {
	return nil, nil
}

func (m *mockInspector) GetCheckConstraints(context.Context, string) ([]model.CheckConstraint, error) {
	return nil, nil
}

func (m *mockInspector) GetViews(context.Context) ([]model.View, error) {
	m.getViewsCalls++
	if m.viewsErr != nil {
		return nil, m.viewsErr
	}
	return m.views, nil
}

func (m *mockInspector) GetProcedures(context.Context) ([]model.Procedure, error) {
	m.getProceduresCalls++
	if m.proceduresErr != nil {
		return nil, m.proceduresErr
	}
	return m.procedures, nil
}

func (m *mockInspector) GetFunctions(context.Context) ([]model.Function, error) {
	m.getFunctionsCalls++
	if m.functionsErr != nil {
		return nil, m.functionsErr
	}
	return m.functions, nil
}

func (m *mockInspector) GetTriggers(_ context.Context, tableName string) ([]model.Trigger, error) {
	if m.getTriggersCalls == nil {
		m.getTriggersCalls = make(map[string]int)
	}
	m.getTriggersCalls[tableName]++
	if m.triggersErr != nil {
		if err := m.triggersErr[tableName]; err != nil {
			return nil, err
		}
	}
	return m.triggersByTbl[tableName], nil
}

func (m *mockInspector) GetSequences(context.Context) ([]model.Sequence, error) {
	m.getSequencesCalls++
	if m.sequencesErr != nil {
		return nil, m.sequencesErr
	}
	return m.sequences, nil
}

type mockInspectorFactory struct {
	inspector *mockInspector
	createErr error
}

func (f *mockInspectorFactory) Create(inspector.ConnectionConfig) (inspector.Inspector, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.inspector, nil
}

func (f *mockInspectorFactory) GetType() string { return "mock" }

type mockExporter struct {
	exportErr  error
	called     bool
	tables     []model.Table
	views      []model.View
	procedures []model.Procedure
	functions  []model.Function
	triggers   []model.Trigger
	sequences  []model.Sequence
	options    exporter.ExportOptions
}

func (e *mockExporter) Export(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	e.called = true
	e.tables = append([]model.Table(nil), tables...)
	e.views = append([]model.View(nil), views...)
	e.procedures = append([]model.Procedure(nil), procedures...)
	e.functions = append([]model.Function(nil), functions...)
	e.triggers = append([]model.Trigger(nil), triggers...)
	e.sequences = append([]model.Sequence(nil), sequences...)
	e.options = options
	return e.exportErr
}

func (e *mockExporter) GetName() string      { return "mock" }
func (e *mockExporter) GetExtension() string { return ".mock" }

type mockExporterFactory struct {
	exporter  *mockExporter
	createErr error
}

func (f *mockExporterFactory) Create() (exporter.Exporter, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.exporter, nil
}

func (f *mockExporterFactory) GetType() string { return "mock" }

func registerMockInspector(t *testing.T, dbType string, ins *mockInspector, createErr error) {
	t.Helper()
	inspector.Register(dbType, &mockInspectorFactory{inspector: ins, createErr: createErr})
}

func registerMockExporter(t *testing.T, format string, exp *mockExporter, createErr error) {
	t.Helper()
	exporter.Register(format, &mockExporterFactory{exporter: exp, createErr: createErr})
}

func uniqueName(t *testing.T, prefix string) string {
	t.Helper()
	return fmt.Sprintf("%s-%s", prefix, strings.ReplaceAll(t.Name(), "/", "-"))
}

func TestRunUnsupportedDatabaseType(t *testing.T) {
	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{
			Type: "unknown",
		},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "unsupported database type") {
		t.Fatalf("expected unsupported database type error, got %v", err)
	}
}

func TestRunInspectorFactoryCreateFailure(t *testing.T) {
	dbType := uniqueName(t, "mock-create-fail")
	registerMockInspector(t, dbType, nil, errors.New("create boom"))

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export:   config.ExportConfig{Formats: []string{"mock"}},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "failed to create inspector") {
		t.Fatalf("expected create inspector error, got %v", err)
	}
}

func TestRunConnectFailure(t *testing.T) {
	dbType := uniqueName(t, "mock-connect-fail")
	ins := &mockInspector{connectErr: errors.New("connect boom")}
	registerMockInspector(t, dbType, ins, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export:   config.ExportConfig{Formats: []string{"mock"}},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "failed to connect to database") {
		t.Fatalf("expected connect error, got %v", err)
	}
	if ins.closeCalls != 0 {
		t.Fatalf("expected close not called on connect failure, got %d", ins.closeCalls)
	}
}

func TestRunTestConnectionFailure(t *testing.T) {
	dbType := uniqueName(t, "mock-testconn-fail")
	ins := &mockInspector{testConnErr: errors.New("test connection boom")}
	registerMockInspector(t, dbType, ins, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export:   config.ExportConfig{Formats: []string{"mock"}},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "database connection test failed") {
		t.Fatalf("expected test connection error, got %v", err)
	}
	if ins.closeCalls != 1 {
		t.Fatalf("expected close called once after connect, got %d", ins.closeCalls)
	}
}

func TestRunGetTablesFailure(t *testing.T) {
	dbType := uniqueName(t, "mock-gettables-fail")
	ins := &mockInspector{getTablesErr: errors.New("get tables boom")}
	registerMockInspector(t, dbType, ins, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export:   config.ExportConfig{Formats: []string{"mock"}},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "failed to get tables") {
		t.Fatalf("expected get tables error, got %v", err)
	}
	if ins.closeCalls != 1 {
		t.Fatalf("expected close called once, got %d", ins.closeCalls)
	}
}

func TestRunNoTablesWereSuccessfullyProcessed(t *testing.T) {
	dbType := uniqueName(t, "mock-no-tables")
	ins := &mockInspector{
		tables: []model.Table{{Name: "users"}, {Name: "orders"}},
		getTableErr: map[string]error{
			"users":  errors.New("boom"),
			"orders": errors.New("boom"),
		},
	}
	registerMockInspector(t, dbType, ins, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export:   config.ExportConfig{Formats: []string{"mock"}},
	})

	err := svc.Run()
	if err == nil || !strings.Contains(err.Error(), "no tables were successfully processed") {
		t.Fatalf("expected no tables error, got %v", err)
	}
}

func TestLoadTablesContinuesAfterSingleTableFailure(t *testing.T) {
	ins := &mockInspector{
		tablesByName: map[string]model.Table{
			"users":  {Name: "users", Comment: "users table"},
			"orders": {Name: "orders", Comment: "orders table"},
		},
		getTableErr: map[string]error{
			"orders": errors.New("orders boom"),
		},
		getTableCalls: make(map[string]int),
	}
	svc := NewService(&config.Config{})

	tables, failed, err := svc.loadTables(context.Background(), ins, []model.Table{{Name: "users"}, {Name: "orders"}})
	if err != nil {
		t.Fatalf("loadTables returned error: %v", err)
	}
	if len(tables) != 1 || tables[0].Name != "users" {
		t.Fatalf("expected one successful table, got %#v", tables)
	}
	if len(failed) != 1 || failed[0] != "orders" {
		t.Fatalf("expected one failed table orders, got %#v", failed)
	}
	if ins.getTableCalls["users"] != 1 || ins.getTableCalls["orders"] != 1 {
		t.Fatalf("expected both tables to be attempted, got %#v", ins.getTableCalls)
	}
}

func TestLoadTablesRespectsContextCancellation(t *testing.T) {
	ins := &mockInspector{}
	svc := NewService(&config.Config{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := svc.loadTables(ctx, ins, []model.Table{{Name: "users"}})
	if err == nil || !strings.Contains(err.Error(), "export cancelled") {
		t.Fatalf("expected cancellation error, got %v", err)
	}
}

func TestRunIncludesSupplementalObjects(t *testing.T) {
	dbType := uniqueName(t, "mock-full")
	format := uniqueName(t, "mock-format")

	ins := &mockInspector{
		tables: []model.Table{
			{Name: "users", Comment: "users"},
			{Name: "orders", Comment: "orders"},
		},
		tablesByName: map[string]model.Table{
			"users":  {Name: "users", Comment: "users", Columns: []model.Column{{Name: "id", IsPrimaryKey: true}}},
			"orders": {Name: "orders", Comment: "orders", Columns: []model.Column{{Name: "id", IsPrimaryKey: true}}},
		},
		getTableCalls:    make(map[string]int),
		getTriggersCalls: make(map[string]int),
		views:            []model.View{{Name: "active_users", Definition: "select 1"}},
		procedures:       []model.Procedure{{Name: "refresh_stats"}},
		functions:        []model.Function{{Name: "format_name"}},
		sequences:        []model.Sequence{{Name: "user_seq"}},
		triggersByTbl: map[string][]model.Trigger{
			"users":  {{Name: "users_trigger"}},
			"orders": {{Name: "orders_trigger"}},
		},
	}
	registerMockInspector(t, dbType, ins, nil)

	exp := &mockExporter{}
	registerMockExporter(t, format, exp, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export: config.ExportConfig{
			Formats:           []string{format},
			OutputDir:         "./docs/schema.md",
			SplitFiles:        true,
			IncludeViews:      true,
			IncludeProcedures: true,
			IncludeFunctions:  true,
			IncludeTriggers:   true,
			IncludeSequences:  true,
		},
	})

	err := svc.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}
	if !exp.called {
		t.Fatalf("expected exporter to be called")
	}
	if len(exp.tables) != 2 || len(exp.views) != 1 || len(exp.procedures) != 1 || len(exp.functions) != 1 || len(exp.triggers) != 2 || len(exp.sequences) != 1 {
		t.Fatalf("unexpected exported object counts: tables=%d views=%d procedures=%d functions=%d triggers=%d sequences=%d",
			len(exp.tables), len(exp.views), len(exp.procedures), len(exp.functions), len(exp.triggers), len(exp.sequences))
	}
	if !exp.options.SplitFiles || !exp.options.IncludeViews || !exp.options.IncludeProcedures || !exp.options.IncludeFunctions || !exp.options.IncludeTriggers || !exp.options.IncludeSequences {
		t.Fatalf("expected export options to be forwarded, got %#v", exp.options)
	}
	if exp.options.DbType != dbType {
		t.Fatalf("expected DbType %q, got %q", dbType, exp.options.DbType)
	}
	if exp.options.OutputDir != "docs" || exp.options.FileName != "schema.md" {
		t.Fatalf("expected ParseOutputPath to split output path, got dir=%q file=%q", exp.options.OutputDir, exp.options.FileName)
	}
	if ins.getViewsCalls != 1 || ins.getProceduresCalls != 1 || ins.getFunctionsCalls != 1 || ins.getSequencesCalls != 1 {
		t.Fatalf("expected supplemental object queries to run once, got views=%d procedures=%d functions=%d sequences=%d",
			ins.getViewsCalls, ins.getProceduresCalls, ins.getFunctionsCalls, ins.getSequencesCalls)
	}
	if ins.getTriggersCalls["users"] != 1 || ins.getTriggersCalls["orders"] != 1 {
		t.Fatalf("expected trigger queries for each table, got %#v", ins.getTriggersCalls)
	}
}

func TestRunContinuesWhenSupplementalQueriesFail(t *testing.T) {
	dbType := uniqueName(t, "mock-warning")
	format := uniqueName(t, "mock-warning-format")

	ins := &mockInspector{
		tables: []model.Table{{Name: "users", Columns: []model.Column{{Name: "id", IsPrimaryKey: true}}}},
		tablesByName: map[string]model.Table{
			"users": {Name: "users", Columns: []model.Column{{Name: "id", IsPrimaryKey: true}}},
		},
		getTableCalls:    make(map[string]int),
		getTriggersCalls: make(map[string]int),
		viewsErr:         errors.New("views boom"),
		proceduresErr:    errors.New("procedures boom"),
		functionsErr:     errors.New("functions boom"),
		sequencesErr:     errors.New("sequences boom"),
		triggersErr: map[string]error{
			"users": errors.New("triggers boom"),
		},
	}
	registerMockInspector(t, dbType, ins, nil)

	exp := &mockExporter{}
	registerMockExporter(t, format, exp, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: dbType},
		Export: config.ExportConfig{
			Formats:           []string{format},
			IncludeViews:      true,
			IncludeProcedures: true,
			IncludeFunctions:  true,
			IncludeTriggers:   true,
			IncludeSequences:  true,
		},
	})

	if err := svc.Run(); err != nil {
		t.Fatalf("Run() should continue despite supplemental failures, got %v", err)
	}
	if !exp.called {
		t.Fatalf("expected exporter to be called")
	}
	if len(exp.views) != 0 || len(exp.procedures) != 0 || len(exp.functions) != 0 || len(exp.triggers) != 0 || len(exp.sequences) != 0 {
		t.Fatalf("expected failed supplemental lookups to export empty slices, got views=%d procedures=%d functions=%d triggers=%d sequences=%d",
			len(exp.views), len(exp.procedures), len(exp.functions), len(exp.triggers), len(exp.sequences))
	}
}

func TestExportAllFormatsSuccessPartialAndAllFailure(t *testing.T) {
	successFormat := uniqueName(t, "mock-success")
	partialFailFormat := uniqueName(t, "mock-partial-fail")
	allFailFormat := uniqueName(t, "mock-all-fail")

	registerMockExporter(t, successFormat, &mockExporter{}, nil)
	registerMockExporter(t, partialFailFormat, &mockExporter{}, errors.New("boom"))
	registerMockExporter(t, allFailFormat, &mockExporter{}, errors.New("boom"))

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: "dm"},
		Export:   config.ExportConfig{Formats: []string{successFormat}},
	})
	if err := svc.ExportAllFormats([]model.Table{{Name: "users"}}, nil, nil, nil, nil, nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	svc.Config.Export.Formats = []string{successFormat, partialFailFormat}
	err := svc.ExportAllFormats([]model.Table{{Name: "users"}}, nil, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "partial export failure") {
		t.Fatalf("expected partial export failure, got %v", err)
	}

	svc.Config.Export.Formats = []string{allFailFormat}
	err = svc.ExportAllFormats([]model.Table{{Name: "users"}}, nil, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "all exports failed") {
		t.Fatalf("expected all exports failed, got %v", err)
	}
}

func TestExportFormatRejectsUnsupportedAndFactoryErrors(t *testing.T) {
	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: "mysql"},
		Export:   config.ExportConfig{OutputDir: "./docs/schema.txt", SplitFiles: true},
	})

	err := svc.exportFormat(nil, nil, nil, nil, nil, nil, "unsupported")
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected unsupported format error, got %v", err)
	}

	format := uniqueName(t, "mock-factory-error")
	registerMockExporter(t, format, nil, errors.New("factory boom"))
	err = svc.exportFormat(nil, nil, nil, nil, nil, nil, format)
	if err == nil || !strings.Contains(err.Error(), "failed to create exporter") {
		t.Fatalf("expected create exporter error, got %v", err)
	}
}

func TestExportFormatForwardsOptions(t *testing.T) {
	format := "markdown"
	exp := &mockExporter{}
	registerMockExporter(t, format, exp, nil)

	svc := NewService(&config.Config{
		Database: config.DatabaseConfig{Type: "oracle"},
		Export: config.ExportConfig{
			OutputDir:         "./docs/schema.txt",
			SplitFiles:        true,
			IncludeViews:      true,
			IncludeProcedures: true,
			IncludeFunctions:  true,
			IncludeTriggers:   true,
			IncludeSequences:  true,
		},
	})

	tables := []model.Table{{Name: "users"}}
	views := []model.View{{Name: "active_users"}}
	procedures := []model.Procedure{{Name: "refresh_stats"}}
	functions := []model.Function{{Name: "format_name"}}
	triggers := []model.Trigger{{Name: "users_trigger"}}
	sequences := []model.Sequence{{Name: "user_seq"}}

	if err := svc.exportFormat(tables, views, procedures, functions, triggers, sequences, format); err != nil {
		t.Fatalf("exportFormat() failed: %v", err)
	}
	if !exp.called {
		t.Fatalf("expected exporter to be called")
	}
	if exp.options.OutputDir != "docs" || exp.options.FileName != "schema.md" {
		t.Fatalf("expected parsed output path to be forwarded, got dir=%q file=%q", exp.options.OutputDir, exp.options.FileName)
	}
	if exp.options.DbType != "oracle" {
		t.Fatalf("expected DbType oracle, got %q", exp.options.DbType)
	}
	if !exp.options.SplitFiles || !exp.options.IncludeViews || !exp.options.IncludeProcedures || !exp.options.IncludeFunctions || !exp.options.IncludeTriggers || !exp.options.IncludeSequences {
		t.Fatalf("expected options to be forwarded, got %#v", exp.options)
	}
	if len(exp.tables) != 1 || len(exp.views) != 1 || len(exp.procedures) != 1 || len(exp.functions) != 1 || len(exp.triggers) != 1 || len(exp.sequences) != 1 {
		t.Fatalf("unexpected forwarded payload sizes: tables=%d views=%d procedures=%d functions=%d triggers=%d sequences=%d",
			len(exp.tables), len(exp.views), len(exp.procedures), len(exp.functions), len(exp.triggers), len(exp.sequences))
	}
	if exp.options.OutputDir == "" || exp.options.FileName == "" {
		t.Fatalf("expected output path to be parsed, got %#v", exp.options)
	}
}
