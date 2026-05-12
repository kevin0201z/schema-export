package tui

import (
	"strings"
	"testing"
)

func TestSanitizeDSN_Empty(t *testing.T) {
	if result := SanitizeDSN(""); result != "" {
		t.Fatalf("got %q, want empty", result)
	}
}

func TestSanitizeDSN_Standard(t *testing.T) {
	input := "dm://SYSDBA:password@localhost:5236"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("DSN should contain ******, got %q", result)
	}
	if strings.Contains(result, "password") {
		t.Fatalf("DSN should not contain password, got %q", result)
	}
	if !strings.Contains(result, "SYSDBA") {
		t.Fatalf("DSN should contain username, got %q", result)
	}
}

func TestSanitizeDSN_Oracle(t *testing.T) {
	input := "oracle://user:secret@host:1521/ORCL"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("got %q, want masked DSN", result)
	}
	if strings.Contains(result, "secret") {
		t.Fatalf("got %q, password should be masked", result)
	}
}

func TestSanitizeDSN_MySQL(t *testing.T) {
	input := "mysql://root:admin@tcp(localhost:3306)/testdb"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("got %q, want masked DSN", result)
	}
}

func TestSanitizeDSN_NoPassword(t *testing.T) {
	input := "dm://user@host:5236"
	result := SanitizeDSN(input)

	if result != input {
		t.Fatalf("no-password DSN should be unchanged, got %q", result)
	}
}

func TestSanitizeDSN_Fallback(t *testing.T) {
	// 无法解析为 URL 的 DSN 使用 fallback 脱敏
	input := "user:pass@localhost"
	result := sanitizeDSNFallback(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("fallback should mask password, got %q", result)
	}
	if strings.Contains(result, "pass") {
		t.Fatalf("fallback should not contain password, got %q", result)
	}
}

func TestSanitizeDSN_FallbackNoAt(t *testing.T) {
	input := "just-a-string"
	result := sanitizeDSNFallback(input)
	if result != input {
		t.Fatalf("fallback without @ should be unchanged, got %q", result)
	}
}

func TestSanitizeDSN_FallbackNoColon(t *testing.T) {
	input := "user@host"
	result := sanitizeDSNFallback(input)
	if result != input {
		t.Fatalf("fallback without colon should be unchanged, got %q", result)
	}
}

func TestSanitizeDSN_ProtocolColon(t *testing.T) {
	// 确保协议中的冒号不被当作密码分隔符
	input := "postgres://user:pass@host:5432/db"
	result := sanitizeDSNFallback(input)
	if !strings.Contains(result, "******") {
		t.Fatalf("got %q, want masked", result)
	}
}

func TestSanitizeDSN_MySQLRawTCP(t *testing.T) {
	// MySQL 原始 DSN 格式: url.Parse 成功但不填充 u.User
	input := "root:password@tcp(localhost:3306)/testdb"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("MySQL raw DSN should be masked, got %q", result)
	}
	if strings.Contains(result, "password") {
		t.Fatalf("MySQL raw DSN should not contain password, got %q", result)
	}
}

func TestSanitizeDSN_SQLServerRaw(t *testing.T) {
	// SQL Server 原始 DSN 格式
	input := "sqlserver://sa:Admin123@localhost:1433?database=mydb"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("SQL Server DSN should be masked, got %q", result)
	}
	if strings.Contains(result, "Admin123") {
		t.Fatalf("SQL Server DSN should not contain password, got %q", result)
	}
}

func TestSanitizeDSN_RawDSNNoProtocol(t *testing.T) {
	// 无 scheme:// 前缀的原始 user:pass@host 格式
	input := "user:mysecret@localhost:5432/dbname"
	result := SanitizeDSN(input)

	if !strings.Contains(result, "******") {
		t.Fatalf("raw DSN without protocol should be masked, got %q", result)
	}
	if strings.Contains(result, "mysecret") {
		t.Fatalf("raw DSN should not contain password, got %q", result)
	}
}

func TestSanitizeDSN_UserOnlyNoPassword(t *testing.T) {
	// 有 @ 但无密码——应原样返回
	input := "root@localhost"
	result := SanitizeDSN(input)
	if result != input {
		t.Fatalf("user-only DSN should be unchanged, got %q", result)
	}
}

func TestSanitizeSummary_DSNMode(t *testing.T) {
	s := &TuiState{
		DBType:           "dm",
		ConnectionMethod: ConnectionMethodDSN,
		DSN:              "dm://admin:secret@host:5236",
		Schema:           "APP",
		IncludeViews:     true,
		IncludeFunctions: true,
		OutputDir:        "./output",
		Formats:          []string{"markdown"},
	}

	summary := s.SanitizeSummary()

	if strings.Contains(summary, "secret") {
		t.Fatal("summary should not contain password")
	}
	if !strings.Contains(summary, "******") {
		t.Fatal("summary should contain masked password")
	}
	if !strings.Contains(summary, "DSN") {
		t.Fatal("summary should mention DSN connection")
	}
}

func TestSanitizeSummary_ParamsMode(t *testing.T) {
	s := &TuiState{
		DBType:           "oracle",
		ConnectionMethod: ConnectionMethodParams,
		Host:             "db.example.com",
		Port:             "1521",
		Username:         "sysdba",
		Password:         "secret123",
		IncludeSequences: true,
		Tables:           "users,orders",
		Formats:          []string{"sql", "json"},
		SplitFiles:       true,
	}

	summary := s.SanitizeSummary()

	if strings.Contains(summary, "secret123") {
		t.Fatal("summary should not contain password")
	}
	if !strings.Contains(summary, "******") {
		t.Fatal("summary should contain masked password")
	}
	if !strings.Contains(summary, "分离参数") {
		t.Fatal("summary should mention params connection")
	}
	if !strings.Contains(summary, "users,orders") {
		t.Fatal("summary should contain tables filter")
	}
	if !strings.Contains(summary, "sql, json") {
		t.Fatal("summary should contain formats")
	}
}

func TestSanitizeSummary_AllContentTypes(t *testing.T) {
	s := &TuiState{
		DBType:            "mysql",
		ConnectionMethod:  ConnectionMethodDSN,
		DSN:               "mysql://user@host",
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
		Formats:           []string{"markdown"},
	}

	summary := s.SanitizeSummary()

	for _, want := range []string{"视图: 是", "存储过程: 是", "函数: 是", "触发器: 是", "序列: 是"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary should contain %q", want)
		}
	}
}

func TestSanitizeSummary_SQLiteOmitsUnsupportedContentTypes(t *testing.T) {
	s := &TuiState{
		DBType:            "sqlite",
		ConnectionMethod:  ConnectionMethodDSN,
		DSN:               "./app.db",
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
		Formats:           []string{"markdown"},
	}

	summary := s.SanitizeSummary()

	for _, want := range []string{"视图: 是", "触发器: 是"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary should contain %q", want)
		}
	}
	for _, unwanted := range []string{"存储过程:", "函数:", "序列:"} {
		if strings.Contains(summary, unwanted) {
			t.Fatalf("sqlite summary should omit %q, got: %s", unwanted, summary)
		}
	}
}
