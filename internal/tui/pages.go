package tui

import (
	"errors"
	"fmt"
	"strings"
)

// --- 欢迎页 ---

func (m model) welcomeView() string {
	s := styles.Title.Render("数据库结构导出工具 - TUI 模式")
	s += "\n\n"
	s += "通过交互式向导完成数据库导出，无需记忆命令行参数。\n\n"
	s += fmt.Sprintf("默认输出目录: %s\n", m.state.OutputDir)
	s += fmt.Sprintf("默认数据库类型: %s\n", m.state.DBType)
	s += "\n"
	s += styles.Hint.Render("按 Enter 开始配置 · Ctrl+C 退出")
	return s
}

// --- 数据库类型页 ---

func (m model) dbTypeView() string {
	s := styles.Title.Render("选择数据库类型")
	s += "\n\n"
	s += m.dbTypeList.View()
	s += "\n"
	s += styles.Hint.Render("↑↓ 选择 · Enter 确认 · Esc 返回")
	return s
}

// --- 连接方式页 ---

func (m model) connMethodView() string {
	s := styles.Title.Render("选择连接方式")
	s += "\n\n"
	s += m.connMethodList.View()
	s += "\n"
	s += styles.Hint.Render("↑↓ 选择 · Enter 确认 · Esc 返回")
	return s
}

// --- DSN 表单页 ---

func (m model) dsnFormView() string {
	s := styles.Title.Render("DSN 连接配置")
	s += "\n\n"
	s += "DSN 连接字符串:\n"
	s += m.dsnInput.View()
	s += "\n\n"
	s += "Schema（可选）:\n"
	s += m.schemaInput.View()
	s += "\n\n"
	if m.pageError != "" {
		s += styles.Error.Render(m.pageError) + "\n\n"
	}
	// 显示焦点提示
	s += styles.Hint.Render(focusHint(m.focusedInput, 2))
	return s
}

// --- 分离参数表单页 ---

func (m model) paramsFormView() string {
	s := styles.Title.Render("分离参数连接配置")
	s += "\n"

	fields := []struct {
		label string
		input *textinputAdapter
	}{
		{"主机地址 (Host)", &textinputAdapter{m.hostInput}},
		{"端口 (Port)", &textinputAdapter{m.portInput}},
		{"数据库名 (Database)", &textinputAdapter{m.dbNameInput}},
		{"用户名 (Username)", &textinputAdapter{m.usernameInput}},
		{"密码 (Password)", &textinputAdapter{m.passwordInput}},
		{"Schema（可选）", &textinputAdapter{m.schemaInput}},
	}

	for i, f := range fields {
		prefix := "  "
		if i == m.focusedInput {
			prefix = styles.Selected.Render("> ")
		}
		s += fmt.Sprintf("\n%s%s:\n", prefix, f.label)
		s += "  " + f.input.View()
	}

	s += "\n\n"
	if m.pageError != "" {
		s += styles.Error.Render(m.pageError) + "\n\n"
	}
	s += styles.Hint.Render(focusHint(m.focusedInput, len(fields)))
	return s
}

// textinputAdapter 包装 textinput.Model 以便统一引用。
type textinputAdapter struct {
	m interface{ View() string }
}

func (a textinputAdapter) View() string {
	return a.m.View()
}

// --- 导出内容页 ---

func (m model) exportContentView() string {
	s := styles.Title.Render("选择导出对象")
	s += "\n\n"
	s += "已默认包含表导出。选择需要导出的其他对象：\n\n"
	s += m.contentCheckboxes.view()
	s += "\n"
	s += styles.Hint.Render("↑↓ 移动 · 空格 选择 · Enter 确认 · Esc 返回")
	return s
}

// --- 表过滤页 ---

func (m model) tableFilterView() string {
	s := styles.Title.Render("表过滤规则（可选）")
	s += "\n\n"

	fields := []struct {
		label string
		input interface{ View() string }
	}{
		{"指定表名（逗号分隔）", m.tablesInput},
		{"排除表名（逗号分隔）", m.excludeInput},
		{"正则匹配模式（逗号分隔）", m.patternsInput},
	}

	for i, f := range fields {
		prefix := "  "
		if i == m.focusedInput {
			prefix = styles.Selected.Render("> ")
		}
		s += fmt.Sprintf("\n%s%s:\n", prefix, f.label)
		s += "  " + f.input.View()
	}

	s += "\n\n"
	if m.pageError != "" {
		s += styles.Error.Render(m.pageError) + "\n\n"
	}
	s += styles.Hint.Render(focusHint(m.focusedInput, len(fields)))
	return s
}

// --- 输出设置页 ---

func (m model) outputSettingsView() string {
	s := styles.Title.Render("输出设置")
	s += "\n"

	// 输出目录输入
	prefix := "  "
	if m.focusedInput == 0 {
		prefix = styles.Selected.Render("> ")
	}
	s += fmt.Sprintf("\n%s输出目录:\n", prefix)
	s += "  " + m.outputDirInput.View()

	// 格式多选
	s += "\n\n导出格式:\n"
	if m.focusedInput == 1 {
		s += styles.Selected.Render("> (多选模式)")
	} else {
		s += "  (多选模式)"
	}
	s += "\n"
	s += m.formatCheckboxes.view()

	// 分文件开关
	splitLabel := "[ ] 按表分文件导出"
	if m.state.SplitFiles {
		splitLabel = styles.Selected.Render("[x] 按表分文件导出")
	}
	if m.focusedInput == 2 {
		splitLabel = styles.Selected.Render("> " + splitLabel)
	}
	s += "\n" + splitLabel

	s += "\n\n"
	if m.pageError != "" {
		s += styles.Error.Render(m.pageError) + "\n\n"
	}
	s += styles.Hint.Render("Tab/Shift+Tab 切换字段 · ↑↓ 空格 操作多选 · Enter 确认 · Esc 返回")
	return s
}

// --- 确认页 ---

func (m model) confirmView() string {
	s := styles.Title.Render("确认导出配置")
	s += "\n\n"
	s += m.sanitizedSummary
	s += "\n\n"
	s += styles.Hint.Render("Enter 开始导出 · Esc 返回修改 · Ctrl+C 退出")
	return s
}

// --- 执行页 ---

func (m model) executionView() string {
	s := styles.Title.Render("执行导出")
	s += "\n\n"

	if !m.exportDone {
		s += m.spinner.View() + " 正在导出，请稍候..."
		s += "\n\n"
		s += styles.Hint.Render("导出过程中请勿关闭终端")
		return s
	}

	if m.exportError != nil {
		s += styles.Error.Render("导出失败")
		s += "\n\n"
		s += fmt.Sprintf("错误信息: %s\n", m.exportError.Error())
		s += "\n"
		s += styles.Hint.Render("Enter 返回修改配置 · Ctrl+C 退出")
		return s
	}

	s += styles.Success.Render("导出成功")
	s += "\n\n"
	if m.exportResult != "" {
		s += m.exportResult
	}
	s += "\n"
	s += styles.Hint.Render("Enter 返回首页 · Ctrl+C 退出")
	return s
}

// focusHint 生成焦点提示文本。
func focusHint(focused, total int) string {
	return fmt.Sprintf("Tab/Shift+Tab 切换字段 (%d/%d) · Enter 确认 · Esc 返回", focused+1, total)
}

// --- 校验 ---

// validateCurrentPage 校验当前页的输入。
func (m model) validateCurrentPage() error {
	switch m.currentPage {
	case PageDSNForm:
		if strings.TrimSpace(m.dsnInput.Value()) == "" {
			return errors.New(errMsgEmptyDSN)
		}
	case PageParamsForm:
		if strings.TrimSpace(m.hostInput.Value()) == "" {
			return errors.New(errMsgEmptyHost)
		}
		if strings.TrimSpace(m.usernameInput.Value()) == "" {
			return errors.New(errMsgEmptyUser)
		}
		if port := strings.TrimSpace(m.portInput.Value()); port != "" {
			p, err := parsePort(port)
			if err != nil || p < 1 || p > 65535 {
				return errors.New(errMsgInvalidPort)
			}
		}
	case PageOutputSettings:
		if len(m.formatCheckboxes.selectedKeys()) == 0 {
			return errors.New(errMsgNoFormat)
		}
	}
	return nil
}

// nextPage 返回下一页。
func nextPage(p Page) Page {
	switch p {
	case PageWelcome:
		return PageDBType
	case PageDBType:
		return PageConnectionMethod
	case PageConnectionMethod:
		return PageDSNForm // 默认 DSN 模式，可能在 handleNext 中被修改
	case PageDSNForm, PageParamsForm:
		return PageExportContent
	case PageExportContent:
		return PageTableFilter
	case PageTableFilter:
		return PageOutputSettings
	case PageOutputSettings:
		return PageConfirm
	case PageConfirm:
		return PageExecution
	default:
		return PageWelcome
	}
}

// pageInputCount 返回当前页的输入字段数量。
func pageInputCount(p Page) int {
	switch p {
	case PageDSNForm:
		return 2 // DSN, Schema
	case PageParamsForm:
		return 6 // host, port, db, user, pass, schema
	case PageTableFilter:
		return 3 // tables, exclude, patterns
	case PageOutputSettings:
		return 3 // output dir, formats checkbox, split
	default:
		return 0
	}
}
