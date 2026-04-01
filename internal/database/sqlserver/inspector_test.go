package sqlserver

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestGetIndexesReturnsSortedIndexes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary_key", "column_name"}).
		AddRow("IDX_B", "NONCLUSTERED", false, false, "col_b").
		AddRow("IDX_A", "NONCLUSTERED", true, false, "col_a1").
		AddRow("IDX_A", "NONCLUSTERED", true, false, "col_a2")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("TBL").WillReturnRows(rows)

	indexes, err := ins.GetIndexes(context.Background(), "TBL")
	if err != nil {
		t.Fatalf("GetIndexes failed: %v", err)
	}

	if len(indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(indexes))
	}

	if indexes[0].Name != "IDX_A" || indexes[1].Name != "IDX_B" {
		t.Fatalf("expected sorted indexes [IDX_A IDX_B], got [%s %s]", indexes[0].Name, indexes[1].Name)
	}

	if got := indexes[0].GetColumnsString(); got != "col_a1, col_a2" {
		t.Fatalf("expected IDX_A columns to preserve query order, got %q", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
