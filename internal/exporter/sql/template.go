package sql

// tableTemplate SQL DDL 表模板 - 优化大模型理解
const tableTemplate = `-- --------------------------------------------------------
-- Table: {{.Name}}
-- --------------------------------------------------------
{{if .Comment}}-- Description: {{.Comment}}{{end}}
-- Type: {{.Type}}
-- Columns: {{len .Columns}}
-- Indexes: {{len .Indexes}}
-- Foreign Keys: {{len .ForeignKeys}}

CREATE TABLE {{.Name}} (
{{- range $i, $col := .Columns }}
    -- {{$col.Name}}: {{if $col.Comment}}{{$col.Comment}}{{else}}No description{{end}}
    {{$col.Name}} {{$col.GetFullDataType}}{{if not $col.IsNullable}} NOT NULL{{end}}{{if $col.DefaultValue}} DEFAULT {{$col.DefaultValue}}{{end}}{{if $col.IsPrimaryKey}} PRIMARY KEY{{end}}{{if $col.IsAutoIncrement}} AUTO_INCREMENT{{end}}{{if lt (add $i 1) (len $.Columns)}},{{end}}
{{- end }}
);

{{if .Indexes -}}
-- --------------------------------------------------------
-- Indexes for {{.Name}}
-- --------------------------------------------------------
{{- range .Indexes }}
{{if .IsPrimary -}}
-- Primary Key: {{.Name}} on ({{.GetColumnsString}})
{{else -}}
CREATE {{if .IsUnique}}UNIQUE {{end}}INDEX {{.Name}} ON {{$.Name}} ({{.GetColumnsString}});
{{end -}}
{{- end }}
{{end -}}

{{if .ForeignKeys -}}
-- --------------------------------------------------------
-- Foreign Keys for {{.Name}}
-- --------------------------------------------------------
{{- range .ForeignKeys }}
-- Relationship: {{$.Name}}.{{.Column}} -> {{.RefTable}}.{{.RefColumn}}
ALTER TABLE {{$.Name}} ADD CONSTRAINT {{.Name}} FOREIGN KEY ({{.Column}}) REFERENCES {{.RefTable}}({{.RefColumn}}){{if .OnDelete}} ON DELETE {{.OnDelete}}{{end}};
{{- end }}
{{end -}}
`
