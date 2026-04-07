package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func newOracleCompatibleMockInspector(t *testing.T, schema string, placeholder PlaceholderType) (*OracleCompatibleInspector, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	ins := NewOracleCompatibleInspector(inspector.ConnectionConfig{Schema: schema}, placeholder)
	ins.SetDB(db)

	cleanup := func() {
		db.Close()
	}

	return ins, mock, cleanup
}

func TestOracleCompatibleHelpers(t *testing.T) {
	insColon := NewOracleCompatibleInspector(inspector.ConnectionConfig{}, PlaceholderColon)
	if got := insColon.placeholderStr(2); got != ":2" {
		t.Fatalf("unexpected colon placeholder: %s", got)
	}

	insQuestion := NewOracleCompatibleInspector(inspector.ConnectionConfig{}, PlaceholderQuestion)
	if got := insQuestion.placeholderStr(2); got != "?" {
		t.Fatalf("unexpected question placeholder: %s", got)
	}

	cases := []struct {
		input string
		want  string
	}{
		{"BEFORE EACH ROW", "BEFORE"},
		{"after insert", "AFTER"},
		{"instead of update", "INSTEAD OF"},
		{"", ""},
		{"something else", ""},
	}
	for _, tc := range cases {
		if got := parseTriggerTiming(tc.input); got != tc.want {
			t.Fatalf("parseTriggerTiming(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestOracleCompatibleQuerySource(t *testing.T) {
	t.Run("user branch concatenates rows", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TEXT FROM USER_SOURCE 
		WHERE NAME = ? AND TYPE = ?
		ORDER BY LINE
	`)).WithArgs("PROC_A", "PROCEDURE").WillReturnRows(sqlmock.NewRows([]string{"TEXT"}).
			AddRow("BEGIN\n").
			AddRow("NULL; END;"))

		got, err := ins.querySource(context.Background(), "PROC_A", "PROCEDURE", "")
		if err != nil {
			t.Fatalf("querySource failed: %v", err)
		}
		if got != "BEGIN\nNULL; END;" {
			t.Fatalf("unexpected source: %q", got)
		}
	})

	t.Run("schema branch concatenates rows", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "APP", PlaceholderColon)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TEXT FROM ALL_SOURCE 
		WHERE OWNER = :1 AND NAME = :2 AND TYPE = :3
		ORDER BY LINE
	`)).WithArgs("APP", "FN_A", "FUNCTION").WillReturnRows(sqlmock.NewRows([]string{"TEXT"}).
			AddRow("RETURN ").
			AddRow("1;"))

		got, err := ins.querySource(context.Background(), "FN_A", "FUNCTION", "APP")
		if err != nil {
			t.Fatalf("querySource failed: %v", err)
		}
		if got != "RETURN 1;" {
			t.Fatalf("unexpected source: %q", got)
		}
	})
}

func TestOracleCompatibleUserBranchQueries(t *testing.T) {
	t.Run("tables and table comment", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TABLE_NAME, COMMENTS 
		FROM USER_TAB_COMMENTS 
		WHERE TABLE_TYPE = 'TABLE'
		ORDER BY TABLE_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "COMMENTS"}).AddRow("USERS", "用户表"))

		tables, err := ins.GetTables(context.Background())
		if err != nil {
			t.Fatalf("GetTables failed: %v", err)
		}
		if len(tables) != 1 || tables[0].Name != "USERS" {
			t.Fatalf("unexpected tables: %#v", tables)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COMMENTS FROM USER_TAB_COMMENTS WHERE TABLE_NAME = ?`)).
			WithArgs("USERS").
			WillReturnRows(sqlmock.NewRows([]string{"COMMENTS"}).AddRow("用户表"))

		comment, err := ins.queryTableComment(context.Background(), "USERS", "")
		if err != nil {
			t.Fatalf("queryTableComment failed: %v", err)
		}
		if comment != "用户表" {
			t.Fatalf("unexpected comment: %q", comment)
		}
	})

	t.Run("views", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT v.VIEW_NAME, c.COMMENTS, v.TEXT
		FROM USER_VIEWS v
		LEFT JOIN USER_TAB_COMMENTS c ON v.VIEW_NAME = c.TABLE_NAME AND c.TABLE_TYPE = 'VIEW'
		ORDER BY v.VIEW_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"VIEW_NAME", "COMMENTS", "TEXT"}).
			AddRow("V_USERS", "用户视图", "SELECT * FROM USERS"))

		views, err := ins.GetViews(context.Background())
		if err != nil {
			t.Fatalf("GetViews failed: %v", err)
		}
		if len(views) != 1 || views[0].Definition != "SELECT * FROM USERS" {
			t.Fatalf("unexpected views: %#v", views)
		}
	})

	t.Run("procedures", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM USER_PROCEDURES p
		LEFT JOIN USER_TAB_COMMENTS c ON p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OBJECT_TYPE = 'PROCEDURE'
		ORDER BY p.OBJECT_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"OBJECT_NAME", "COMMENTS"}).AddRow("SP_CLEAN", "过程"))

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TEXT FROM USER_SOURCE 
		WHERE NAME = ? AND TYPE = ?
		ORDER BY LINE
	`)).WithArgs("SP_CLEAN", "PROCEDURE").WillReturnRows(sqlmock.NewRows([]string{"TEXT"}).
			AddRow("BEGIN ").
			AddRow("NULL; END;"))

		procs, err := ins.GetProcedures(context.Background())
		if err != nil {
			t.Fatalf("GetProcedures failed: %v", err)
		}
		if len(procs) != 1 || procs[0].Definition != "BEGIN NULL; END;" {
			t.Fatalf("unexpected procedures: %#v", procs)
		}
	})

	t.Run("functions", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.OBJECT_NAME, c.COMMENTS
		FROM USER_PROCEDURES p
		LEFT JOIN USER_TAB_COMMENTS c ON p.OBJECT_NAME = c.TABLE_NAME
		WHERE p.OBJECT_TYPE = 'FUNCTION'
		ORDER BY p.OBJECT_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{"OBJECT_NAME", "COMMENTS"}).AddRow("FN_TOTAL", "函数"))

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TEXT FROM USER_SOURCE 
		WHERE NAME = ? AND TYPE = ?
		ORDER BY LINE
	`)).WithArgs("FN_TOTAL", "FUNCTION").WillReturnRows(sqlmock.NewRows([]string{"TEXT"}).
			AddRow("RETURN ").
			AddRow("1;"))

		fns, err := ins.GetFunctions(context.Background())
		if err != nil {
			t.Fatalf("GetFunctions failed: %v", err)
		}
		if len(fns) != 1 || fns[0].Definition != "RETURN 1;" {
			t.Fatalf("unexpected functions: %#v", fns)
		}
	})

	t.Run("triggers", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			t.TRIGGER_NAME,
			t.TABLE_NAME,
			t.TRIGGERING_EVENT,
			t.STATUS,
			t.TRIGGER_TYPE
		FROM USER_TRIGGERS t
		WHERE t.TABLE_NAME = ?
		ORDER BY t.TRIGGER_NAME
	`)).WithArgs("USERS").WillReturnRows(sqlmock.NewRows([]string{"TRIGGER_NAME", "TABLE_NAME", "TRIGGERING_EVENT", "STATUS", "TRIGGER_TYPE"}).
			AddRow("TR_USERS", "USERS", "INSERT", "ENABLED", "BEFORE EACH ROW"))

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT TEXT FROM USER_SOURCE 
		WHERE NAME = ? AND TYPE = ?
		ORDER BY LINE
	`)).WithArgs("TR_USERS", "TRIGGER").WillReturnRows(sqlmock.NewRows([]string{"TEXT"}).
			AddRow("BEGIN ").
			AddRow("NULL; END;"))

		triggers, err := ins.GetTriggers(context.Background(), "USERS")
		if err != nil {
			t.Fatalf("GetTriggers failed: %v", err)
		}
		if len(triggers) != 1 || triggers[0].Timing != "BEFORE" || triggers[0].Definition != "BEGIN NULL; END;" {
			t.Fatalf("unexpected triggers: %#v", triggers)
		}
	})

	t.Run("sequences", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "", PlaceholderQuestion)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			SEQUENCE_NAME,
			MIN_VALUE,
			MAX_VALUE,
			INCREMENT_BY,
			CYCLE_FLAG,
			CACHE_SIZE,
			LAST_NUMBER
		FROM USER_SEQUENCES
		ORDER BY SEQUENCE_NAME
	`)).WillReturnRows(sqlmock.NewRows([]string{
			"SEQUENCE_NAME", "MIN_VALUE", "MAX_VALUE", "INCREMENT_BY", "CYCLE_FLAG", "CACHE_SIZE", "LAST_NUMBER",
		}).AddRow("SEQ_USERS", int64(1), int64(100), int64(5), "Y", int64(20), int64(11)))

		sequences, err := ins.GetSequences(context.Background())
		if err != nil {
			t.Fatalf("GetSequences failed: %v", err)
		}
		if len(sequences) != 1 || !sequences[0].Cycle || sequences[0].LastValue != 11 {
			t.Fatalf("unexpected sequences: %#v", sequences)
		}
	})
}

func TestOracleCompatibleSchemaBranchQueries(t *testing.T) {
	t.Run("columns", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "APP", PlaceholderColon)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.DATA_LENGTH,
			c.DATA_PRECISION,
			c.DATA_SCALE,
			c.NULLABLE,
			c.DATA_DEFAULT,
			cc.COMMENTS,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as IS_PK
		FROM ALL_TAB_COLUMNS c
		LEFT JOIN ALL_COL_COMMENTS cc ON c.OWNER = cc.OWNER AND c.TABLE_NAME = cc.TABLE_NAME AND c.COLUMN_NAME = cc.COLUMN_NAME
		LEFT JOIN (
			SELECT col.COLUMN_NAME
			FROM ALL_CONSTRAINTS cons
			JOIN ALL_CONS_COLUMNS col ON cons.OWNER = col.OWNER AND cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
			WHERE cons.TABLE_NAME = :1 AND cons.OWNER = :2 AND cons.CONSTRAINT_TYPE = 'P'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = :3 AND c.OWNER = :4
		ORDER BY c.COLUMN_ID
	`)).WithArgs("USERS", "APP", "USERS", "APP").WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "DATA_TYPE", "DATA_LENGTH", "DATA_PRECISION", "DATA_SCALE", "NULLABLE", "DATA_DEFAULT", "COMMENTS", "IS_PK",
		}).AddRow("ID", "NUMBER", 20, 10, 0, "N", "((1))", "主键", 1))

		cols, err := ins.GetColumns(context.Background(), "USERS")
		if err != nil {
			t.Fatalf("GetColumns failed: %v", err)
		}
		if len(cols) != 1 || !cols[0].IsPrimaryKey || cols[0].Length != 20 {
			t.Fatalf("unexpected columns: %#v", cols)
		}
	})

	t.Run("indexes", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "APP", PlaceholderColon)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.INDEX_NAME,
			i.UNIQUENESS,
			ic.COLUMN_NAME
		FROM ALL_INDEXES i
		JOIN ALL_IND_COLUMNS ic ON i.OWNER = ic.INDEX_OWNER AND i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_NAME = :1 AND i.OWNER = :2
		ORDER BY i.INDEX_NAME, ic.COLUMN_POSITION
	`)).WithArgs("USERS", "APP").WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "UNIQUENESS", "COLUMN_NAME"}).
			AddRow("IDX_A", "UNIQUE", "ID").
			AddRow("PRIMARY", "UNIQUE", "ID"))

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT cons.CONSTRAINT_NAME
		FROM ALL_CONSTRAINTS cons
		WHERE cons.TABLE_NAME = :1 AND cons.OWNER = :2 AND cons.CONSTRAINT_TYPE = 'P'
	`)).WithArgs("USERS", "APP").WillReturnRows(sqlmock.NewRows([]string{"CONSTRAINT_NAME"}).AddRow("PRIMARY"))

		indexes, err := ins.GetIndexes(context.Background(), "USERS")
		if err != nil {
			t.Fatalf("GetIndexes failed: %v", err)
		}
		if len(indexes) != 2 || !indexes[1].IsPrimary {
			t.Fatalf("unexpected indexes: %#v", indexes)
		}
	})

	t.Run("foreign keys", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "APP", PlaceholderColon)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cons.CONSTRAINT_NAME,
			col.COLUMN_NAME,
			refCons.TABLE_NAME as REF_TABLE,
			refCol.COLUMN_NAME as REF_COLUMN,
			cons.DELETE_RULE
		FROM ALL_CONSTRAINTS cons
		JOIN ALL_CONS_COLUMNS col ON cons.OWNER = col.OWNER AND cons.CONSTRAINT_NAME = col.CONSTRAINT_NAME
		JOIN ALL_CONSTRAINTS refCons ON cons.R_OWNER = refCons.OWNER AND cons.R_CONSTRAINT_NAME = refCons.CONSTRAINT_NAME
		JOIN ALL_CONS_COLUMNS refCol ON refCons.OWNER = refCol.OWNER AND refCons.CONSTRAINT_NAME = refCol.CONSTRAINT_NAME AND col.POSITION = refCol.POSITION
		WHERE cons.TABLE_NAME = :1 AND cons.OWNER = :2 AND cons.CONSTRAINT_TYPE = 'R'
	`)).WithArgs("USERS", "APP").WillReturnRows(sqlmock.NewRows([]string{
			"CONSTRAINT_NAME", "COLUMN_NAME", "REF_TABLE", "REF_COLUMN", "DELETE_RULE",
		}).AddRow("FK_USERS_GROUPS", "GROUP_ID", "GROUPS", "ID", "CASCADE"))

		fks, err := ins.GetForeignKeys(context.Background(), "USERS")
		if err != nil {
			t.Fatalf("GetForeignKeys failed: %v", err)
		}
		if len(fks) != 1 || fks[0].OnDelete != "CASCADE" {
			t.Fatalf("unexpected foreign keys: %#v", fks)
		}
	})

	t.Run("check constraints", func(t *testing.T) {
		ins, mock, cleanup := newOracleCompatibleMockInspector(t, "APP", PlaceholderColon)
		defer cleanup()

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cons.CONSTRAINT_NAME,
			cons.SEARCH_CONDITION,
			cc.COLUMN_NAME
		FROM ALL_CONSTRAINTS cons
		LEFT JOIN ALL_CONS_COLUMNS cc ON cons.OWNER = cc.OWNER AND cons.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE cons.TABLE_NAME = :1 AND cons.OWNER = :2 AND cons.CONSTRAINT_TYPE = 'C'
		ORDER BY cons.CONSTRAINT_NAME
	`)).WithArgs("USERS", "APP").WillReturnRows(sqlmock.NewRows([]string{"CONSTRAINT_NAME", "SEARCH_CONDITION", "COLUMN_NAME"}).
			AddRow("CK_USERS_AGE", "AGE > 0", "AGE").
			AddRow("CK_USERS_AGE", "AGE > 0", "STATUS"))

		checks, err := ins.GetCheckConstraints(context.Background(), "USERS")
		if err != nil {
			t.Fatalf("GetCheckConstraints failed: %v", err)
		}
		if len(checks) != 1 || checks[0].GetColumnsString() != "AGE, STATUS" {
			t.Fatalf("unexpected check constraints: %#v", checks)
		}
	})
}
