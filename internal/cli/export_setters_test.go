package cli

import "testing"

func TestExportCommandSetters(t *testing.T) {
	cmd := NewExportCommand()

	cmd.SetDatabaseType("mysql")
	cmd.SetDatabaseHost("127.0.0.1")
	cmd.SetDatabasePort(3306)
	cmd.SetDatabaseName("app")
	cmd.SetDatabaseUsername("root")
	cmd.SetDatabasePassword("secret")
	cmd.SetDatabaseDSN("mysql://root:secret@tcp(127.0.0.1:3306)/app")
	cmd.SetDatabaseSchema("public")
	cmd.SetOutputDir("./out")
	cmd.SetFormats([]string{"sql", "markdown"})
	cmd.SetSplitFiles(true)
	cmd.SetTables([]string{"users", "orders"})
	cmd.SetExclude([]string{"audit"})
	cmd.SetPatterns([]string{"^app_.*"})

	if cmd.Config.Database.Type != "mysql" {
		t.Fatalf("unexpected type: %q", cmd.Config.Database.Type)
	}
	if cmd.Config.Database.Host != "127.0.0.1" || cmd.Config.Database.Port != 3306 {
		t.Fatalf("unexpected host/port: %#v", cmd.Config.Database)
	}
	if cmd.Config.Database.Database != "app" || cmd.Config.Database.Username != "root" || cmd.Config.Database.Password != "secret" {
		t.Fatalf("unexpected database config: %#v", cmd.Config.Database)
	}
	if cmd.Config.Database.DSN != "mysql://root:secret@tcp(127.0.0.1:3306)/app" {
		t.Fatalf("unexpected DSN: %q", cmd.Config.Database.DSN)
	}
	if cmd.Config.Database.Schema != "public" {
		t.Fatalf("unexpected schema: %q", cmd.Config.Database.Schema)
	}
	if cmd.Config.Export.OutputDir != "./out" {
		t.Fatalf("unexpected output dir: %q", cmd.Config.Export.OutputDir)
	}
	if len(cmd.Config.Export.Formats) != 2 || !cmd.Config.Export.SplitFiles {
		t.Fatalf("unexpected export config: %#v", cmd.Config.Export)
	}
	if len(cmd.Config.Export.Tables) != 2 || len(cmd.Config.Export.Exclude) != 1 || len(cmd.Config.Export.Patterns) != 1 {
		t.Fatalf("unexpected list config: %#v", cmd.Config.Export)
	}
}
