package sql

// GetDialect 根据数据库类型获取方言
func GetDialect(dbType string) Dialect {
	switch dbType {
	case "oracle", "dm":
		return &OracleDialect{}
	case "sqlserver":
		return &SQLServerDialect{}
	case "mysql":
		return &MySQLDialect{}
	case "postgres", "postgresql":
		return &PostgreSQLDialect{}
	default:
		return &GenericDialect{}
	}
}
