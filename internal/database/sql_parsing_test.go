package database

import (
    "context"
    "database/sql"
    "fmt"
    "regexp"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/schema-export/schema-export/internal/inspector"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to open sqlmock: %v", err)
    }
    return db, mock
}

func TestQueryColumnsParsing(t *testing.T) {
    db, mock := setupMockDB(t)
    defer db.Close()

    cfg := inspector.ConnectionConfig{Schema: "S"}
    o := NewOracleCompatibleInspector(cfg, PlaceholderColon)
    o.SetDB(db)

    // prepare columns rows
    cols := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "DATA_LENGTH", "DATA_PRECISION", "DATA_SCALE", "NULLABLE", "DATA_DEFAULT", "COMMENTS", "IS_PK"}).
        AddRow("id", "NUMBER", 10, 10, 0, "N", nil, "id comment", 1).
        AddRow("name", "VARCHAR2", 255, nil, nil, "Y", sql.NullString{String: "def", Valid: true}, sql.NullString{String: "nm", Valid: true}, 0)

    qCols := regexp.QuoteMeta(
        fmt.Sprintf(queryColumnsAllSQL, 
            o.placeholderStr(1), o.placeholderStr(2), o.placeholderStr(3), o.placeholderStr(4)))
    mock.ExpectQuery(qCols).WithArgs("TBL", cfg.Schema, "TBL", cfg.Schema).WillReturnRows(cols)

    _, err := o.GetColumns(context.Background(), "TBL")
    if err != nil {
        t.Fatalf("GetColumns failed: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet expectations: %v", err)
    }
}

func TestQueryIndexesAndPK(t *testing.T) {
    db, mock := setupMockDB(t)
    defer db.Close()

    cfg := inspector.ConnectionConfig{Schema: "S"}
    o := NewOracleCompatibleInspector(cfg, PlaceholderColon)
    o.SetDB(db)

    // index rows
    idxRows := sqlmock.NewRows([]string{"INDEX_NAME", "UNIQUENESS", "COLUMN_NAME"}).
        AddRow("IDX_A", "UNIQUE", "col1").
        AddRow("IDX_A", "UNIQUE", "col2").
        AddRow("IDX_B", "NONUNIQUE", "col3")

    qIdx := regexp.QuoteMeta(fmt.Sprintf(queryIndexesAllSQL, o.placeholderStr(1), o.placeholderStr(2)))
    mock.ExpectQuery(qIdx).WithArgs("TBL", cfg.Schema).WillReturnRows(idxRows)

    pkRow := sqlmock.NewRows([]string{"CONSTRAINT_NAME"}).AddRow("PK_IDX")
    qPK := regexp.QuoteMeta(fmt.Sprintf(queryPKNameAllSQL, o.placeholderStr(1), o.placeholderStr(2)))
    mock.ExpectQuery(qPK).WithArgs("TBL", cfg.Schema).WillReturnRows(pkRow)

    _, err := o.GetIndexes(context.Background(), "TBL")
    if err != nil {
        t.Fatalf("GetIndexes failed: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet expectations: %v", err)
    }
}

func TestQueryForeignKeys(t *testing.T) {
    db, mock := setupMockDB(t)
    defer db.Close()

    cfg := inspector.ConnectionConfig{Schema: "S"}
    o := NewOracleCompatibleInspector(cfg, PlaceholderColon)
    o.SetDB(db)

    fkRows := sqlmock.NewRows([]string{"CONSTRAINT_NAME", "COLUMN_NAME", "REF_TABLE", "REF_COLUMN", "DELETE_RULE"}).
        AddRow("FK_A", "col1", "REF_T", "refc", "CASCADE")

    qFk := regexp.QuoteMeta(fmt.Sprintf(queryForeignKeysAllSQL, o.placeholderStr(1), o.placeholderStr(2)))
    mock.ExpectQuery(qFk).WithArgs("TBL", cfg.Schema).WillReturnRows(fkRows)

    _, err := o.GetForeignKeys(context.Background(), "TBL")
    if err != nil {
        t.Fatalf("GetForeignKeys failed: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet expectations: %v", err)
    }
}
