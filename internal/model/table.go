package model

// TableType 表类型
type TableType string

const (
	TableTypeTable TableType = "TABLE" // 普通表
	TableTypeView  TableType = "VIEW"  // 视图
)

// Table 表结构元数据
type Table struct {
	Name        string       // 表名
	Comment     string       // 表注释
	Type        TableType    // 表类型（表/视图）
	Columns     []Column     // 字段列表
	Indexes     []Index      // 索引列表
	ForeignKeys []ForeignKey // 外键列表
	primaryKey  *Column      // 缓存主键字段
}

// GetPrimaryKey 获取主键字段（带缓存）
func (t *Table) GetPrimaryKey() *Column {
	if t.primaryKey == nil {
		for i := range t.Columns {
			if t.Columns[i].IsPrimaryKey {
				t.primaryKey = &t.Columns[i]
				break
			}
		}
	}
	return t.primaryKey
}

// GetColumnByName 根据名称获取字段
func (t *Table) GetColumnByName(name string) *Column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}
