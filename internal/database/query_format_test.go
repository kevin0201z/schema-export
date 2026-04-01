package database

import (
    "fmt"
    "strings"
    "testing"
)

type formatCase struct{
    name string
    tmpl string
    placeholders []string
    expectContains []string
}

func TestQueryFormatPlaceholders(t *testing.T) {
    cases := []formatCase{
        {
            name: "tables with schema - colon",
            tmpl: queryTablesAllSQL,
            placeholders: []string{ ":1" },
            expectContains: []string{"OWNER = :1"},
        },
        {
            name: "tables user - question",
            tmpl: queryTablesUserSQL,
            placeholders: []string{ "?" },
            expectContains: []string{"USER_TAB_COMMENTS"},
        },
        {
            name: "columns all - many placeholders",
            tmpl: queryColumnsAllSQL,
            placeholders: []string{":1",":2",":3",":4"},
            expectContains: []string{"ALL_TAB_COLUMNS", "ALL_COL_COMMENTS"},
        },
        {
            name: "indexes all",
            tmpl: queryIndexesAllSQL,
            placeholders: []string{":1",":2"},
            expectContains: []string{"ALL_INDEXES", "ALL_IND_COLUMNS"},
        },
        {
            name: "foreign keys all",
            tmpl: queryForeignKeysAllSQL,
            placeholders: []string{":1",":2"},
            expectContains: []string{"ALL_CONSTRAINTS", "ALL_CONS_COLUMNS"},
        },
        {
            name: "table comment all",
            tmpl: queryTableCommentAllSQL,
            placeholders: []string{":1",":2"},
            expectContains: []string{"ALL_TAB_COMMENTS"},
        },
    }

    for _, c := range cases {
        t.Run(c.name, func(t *testing.T){
            // Replace %s with placeholders sequentially
            args := make([]interface{}, len(c.placeholders))
            for i, p := range c.placeholders { args[i] = p }
            formatted := fmt.Sprintf(c.tmpl, args...)

            // ensure expected substrings are present
            for _, s := range c.expectContains {
                if !strings.Contains(formatted, s) {
                    t.Fatalf("formatted query missing %q: %s", s, formatted)
                }
            }

            // ensure all placeholders appear
            for _, p := range c.placeholders {
                if !strings.Contains(formatted, p) {
                    t.Fatalf("formatted query missing placeholder %q: %s", p, formatted)
                }
            }
        })
    }
}
