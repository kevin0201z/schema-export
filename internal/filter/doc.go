// Package filter 提供表名过滤功能。
//
// 该包实现了灵活的表名过滤机制，支持多种过滤规则：
//   - 白名单模式：只导出指定的表
//   - 黑名单模式：排除指定的表
//   - 正则匹配：使用正则表达式匹配表名
//
// 过滤规则优先级：
//  1. 白名单检查：如果设置了白名单，表必须在白名单中
//  2. 黑名单检查：如果在黑名单中，则排除
//  3. 正则匹配：如果设置了正则模式，表名必须匹配至少一个模式
//
// 使用示例:
//
//	filter, err := filter.NewTableFilter(
//	    []string{"users", "orders"},  // 白名单
//	    []string{"temp_*"},            // 黑名单
//	    []string{"^app_.*"},           // 正则模式
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	tables := []model.Table{...}
//	filtered := filter.FilterTables(tables)
package filter
