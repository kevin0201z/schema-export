package markdown

// tableTemplate Markdown 表模板 - 中文显示
const tableTemplate = `## 表: {{.Name}}

{{if .Comment}}**描述:** {{.Comment}}

{{end -}}
**表类型:** {{.Type}}

### 基本信息

| 属性 | 值 |
|------|-----|
| 表名 | {{.Name}} |
| 类型 | {{.Type}} |
{{- if .Comment }}
| 描述 | {{.Comment}} |
{{- end }}
| 字段数 | {{len .Columns}} |
| 索引数 | {{len .Indexes}} |
| 外键数 | {{len .ForeignKeys}} |

### 字段详情

| 字段名 | 数据类型 | 长度/精度 | 可空 | 默认值 | 约束 | 注释 |
|--------|----------|-----------|------|--------|------|------|
{{- range .Columns }}
| {{.Name}} | {{.DataType}} | {{if gt .Length 0}}{{.Length}}{{else if gt .Precision 0}}{{if gt .Scale 0}}{{.Precision}},{{.Scale}}{{else}}{{.Precision}}{{end}}{{else}}-{{end}} | {{if .IsNullable}}是{{else}}否{{end}} | {{if .DefaultValue}}{{.DefaultValue}}{{else}}-{{end}} | {{if .IsPrimaryKey}}主键 {{end}}{{if .IsAutoIncrement}}自增 {{end}}{{if not .IsNullable}}非空{{end}} | {{if .Comment}}{{.Comment}}{{else}}-{{end}} |
{{- end }}

### 约束

#### 主键
{{$hasPK := false}}
{{- range .Columns }}
{{- if .IsPrimaryKey }}
{{$hasPK = true}}
- **{{.Name}}**: {{.DataType}}{{if gt .Length 0}}({{.Length}}){{else if gt .Precision 0}}({{if gt .Scale 0}}{{.Precision}},{{.Scale}}{{else}}{{.Precision}}{{end}}){{end}} - {{if .Comment}}{{.Comment}}{{else}}主键{{end}}
{{- end }}
{{- end }}
{{if not $hasPK}}
*未定义主键*
{{end}}

#### 唯一约束
{{$hasUnique := false}}
{{- range .Indexes }}
{{- if .IsUnique }}
{{$hasUnique = true}}
- **{{.Name}}**: {{.GetColumnsString}}
{{- end }}
{{- end }}
{{if not $hasUnique}}
*未定义唯一约束*
{{end}}

### 索引

{{if .Indexes -}}
| 索引名 | 类型 | 字段 | 是否唯一 | 是否主键 |
|--------|------|------|----------|----------|
{{- range .Indexes }}
| {{.Name}} | {{.Type}} | {{.GetColumnsString}} | {{if .IsUnique}}是{{else}}否{{end}} | {{if .IsPrimary}}是{{else}}否{{end}} |
{{- end }}
{{else -}}
*未定义索引*
{{end}}

### 外键

{{if .ForeignKeys -}}
| 外键名 | 字段 | 引用表 | 引用字段 | 删除规则 | 更新规则 |
|--------|------|--------|----------|----------|----------|
{{- range .ForeignKeys }}
| {{.Name}} | {{.Column}} | {{.RefTable}} | {{.RefColumn}} | {{.GetOnDeleteRule}} | {{.GetOnUpdateRule}} |
{{- end }}

#### 关联关系
{{- range .ForeignKeys }}
- **{{.Name}}**: ` + "`" + `{{$.Name}}.{{.Column}}` + "`" + ` → ` + "`" + `{{.RefTable}}.{{.RefColumn}}` + "`" + ` (删除时{{.GetOnDeleteRule}})
{{- end }}
{{else -}}
*未定义外键*

此表没有与其他表的外键关联关系。
{{end}}

---
`
