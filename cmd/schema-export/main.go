package main

import (
	"fmt"
	"os"

	"github.com/schema-export/schema-export/internal/cli"
	"github.com/spf13/cobra"

	// 导入驱动注册
	_ "github.com/schema-export/schema-export/internal/database/dm"
	_ "github.com/schema-export/schema-export/internal/database/oracle"
	_ "github.com/schema-export/schema-export/internal/database/sqlserver"

	// 导入导出器注册
	_ "github.com/schema-export/schema-export/internal/exporter/markdown"
	_ "github.com/schema-export/schema-export/internal/exporter/sql"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "schema-export",
		Short: "Database schema export tool",
		Long: `A cross-database schema export tool that currently supports DM, Oracle, and SQL Server.
		
Generate database structure documentation in Markdown and SQL DDL formats.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	rootCmd.AddCommand(newExportCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newExportCmd() *cobra.Command {
	cmd := cli.NewExportCommand()

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export database schema",
		Long:  `Connect to database and export table structures to Markdown or SQL format.`,
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.Run()
		},
	}

	// 数据库连接参数
	exportCmd.Flags().StringVar(&cmd.Config.Database.Type, "type", "dm", "Database type (dm, oracle, sqlserver)")
	exportCmd.Flags().StringVar(&cmd.Config.Database.Host, "host", "", "Database host")
	exportCmd.Flags().IntVar(&cmd.Config.Database.Port, "port", 0, "Database port")
	exportCmd.Flags().StringVar(&cmd.Config.Database.Database, "database", "", "Database name")
	exportCmd.Flags().StringVar(&cmd.Config.Database.Username, "username", "", "Database username")
	exportCmd.Flags().StringVar(&cmd.Config.Database.Password, "password", "", "Database password")
	exportCmd.Flags().StringVar(&cmd.Config.Database.DSN, "dsn", "", "Database DSN connection string")
	exportCmd.Flags().StringVar(&cmd.Config.Database.Schema, "schema", "", "Database schema")

	// 导出参数
	exportCmd.Flags().StringVar(&cmd.Config.Export.OutputDir, "output", "./output", "Output directory")
	exportCmd.Flags().StringSliceVar(&cmd.Config.Export.Formats, "formats", []string{"markdown"}, "Export formats (markdown, sql)")
	exportCmd.Flags().BoolVar(&cmd.Config.Export.SplitFiles, "split", false, "Split output into separate files per table")
	exportCmd.Flags().StringSliceVar(&cmd.Config.Export.Tables, "tables", nil, "Tables to export (comma-separated)")
	exportCmd.Flags().StringSliceVar(&cmd.Config.Export.Exclude, "exclude", nil, "Tables to exclude (comma-separated)")
	exportCmd.Flags().StringSliceVar(&cmd.Config.Export.Patterns, "patterns", nil, "Table name patterns to match (regex)")

	return exportCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("schema-export version %s\n", version)
			fmt.Printf("  Commit: %s\n", commit)
			fmt.Printf("  Built:  %s\n", date)
		},
	}
}
