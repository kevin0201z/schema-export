package exportapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schema-export/schema-export/internal/config"
	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/filter"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Service 封装导出流程编排。
type Service struct {
	Config *config.Config
}

// NewService 创建导出服务。
func NewService(cfg *config.Config) *Service {
	return &Service{Config: cfg}
}

// Run 执行完整导出流程。
func (s *Service) Run() error {
	factory, ok := inspector.GetFactory(s.Config.Database.Type)
	if !ok {
		return fmt.Errorf("unsupported database type: %s", s.Config.Database.Type)
	}

	connConfig := s.Config.Database.ToConnectionConfig()
	ins, err := factory.Create(connConfig)
	if err != nil {
		return fmt.Errorf("failed to create inspector: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.DefaultExportTimeout)
	defer cancel()

	if err := ins.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer ins.Close()

	if err := ins.TestConnection(ctx); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	fmt.Printf("Connected to %s database\n", s.Config.Database.Type)

	tables, err := ins.GetTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	fmt.Printf("Found %d tables\n", len(tables))

	tableFilter, err := filter.NewTableFilter(
		s.Config.Export.Tables,
		s.Config.Export.Exclude,
		s.Config.Export.Patterns,
	)
	if err != nil {
		return fmt.Errorf("invalid table filter: %w", err)
	}

	tables = tableFilter.FilterTables(tables)
	fmt.Printf("Filtered to %d tables\n", len(tables))

	fullTables, failedTables, err := s.loadTables(ctx, ins, tables)
	if err != nil {
		return err
	}

	if len(failedTables) > 0 {
		fmt.Printf("Warning: %d tables failed to process: %v\n", len(failedTables), failedTables)
	}

	if len(fullTables) == 0 {
		return fmt.Errorf("no tables were successfully processed")
	}

	fmt.Printf("Successfully processed %d tables\n", len(fullTables))

	var views []model.View
	if s.Config.Export.IncludeViews {
		views, err = ins.GetViews(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get views: %v\n", err)
		} else {
			fmt.Printf("Found %d views\n", len(views))
		}
	}

	if err := s.ExportAllFormats(fullTables, views); err != nil {
		return err
	}

	fmt.Println("Export completed successfully!")
	return nil
}

func (s *Service) loadTables(ctx context.Context, ins inspector.Inspector, tables []model.Table) ([]model.Table, []string, error) {
	var fullTables []model.Table
	var failedTables []string

	for _, table := range tables {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("export cancelled: %w", ctx.Err())
		default:
		}

		fullTable, err := ins.GetTable(ctx, table.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get table %s: %v\n", table.Name, err)
			failedTables = append(failedTables, table.Name)
			continue
		}
		fullTables = append(fullTables, *fullTable)
	}

	return fullTables, failedTables, nil
}

// ExportAllFormats 导出所有格式。
func (s *Service) ExportAllFormats(tables []model.Table, views []model.View) error {
	var failed []string
	successCount := 0

	for _, format := range s.Config.Export.Formats {
		if err := s.exportFormat(tables, views, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting to %s: %v\n", format, err)
			failed = append(failed, fmt.Sprintf("%s (%v)", format, err))
			continue
		}
		successCount++
	}

	if len(failed) == 0 {
		return nil
	}

	if successCount == 0 {
		return fmt.Errorf("all exports failed: %s", strings.Join(failed, "; "))
	}

	return fmt.Errorf("partial export failure: %s", strings.Join(failed, "; "))
}

func (s *Service) exportFormat(tables []model.Table, views []model.View, format string) error {
	factory, ok := exporter.GetFactory(format)
	if !ok {
		return fmt.Errorf("unsupported export format: %s", format)
	}

	exp, err := factory.Create()
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	outputDir, fileName := ParseOutputPath(s.Config.Export.OutputDir, format)
	options := exporter.ExportOptions{
		OutputDir:    outputDir,
		FileName:     fileName,
		SplitFiles:   s.Config.Export.SplitFiles,
		DbType:       s.Config.Database.Type,
		IncludeViews: s.Config.Export.IncludeViews,
	}

	if err := exp.Export(tables, views, options); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Printf("Exported to %s format\n", format)
	return nil
}

// ParseOutputPath 解析输出路径，返回目录和文件名。
func ParseOutputPath(outputPath string, format string) (dir string, fileName string) {
	if outputPath == "" {
		return "./output", ""
	}

	ext := filepath.Ext(outputPath)
	if ext == "" {
		return outputPath, ""
	}

	dir = filepath.Dir(outputPath)
	fileName = filepath.Base(outputPath)
	if format == "sql" && ext != ".sql" {
		fileName = fileName[:len(fileName)-len(ext)] + ".sql"
	} else if format == "markdown" && ext != ".md" {
		fileName = fileName[:len(fileName)-len(ext)] + ".md"
	}

	return dir, fileName
}
