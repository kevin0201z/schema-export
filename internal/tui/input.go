package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// styles 集中管理 TUI 样式。
var styles = struct {
	Title    lipgloss.Style
	Error    lipgloss.Style
	Hint     lipgloss.Style
	Success  lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
}{
	Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
	Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
	Hint:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	Success:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
	Normal:   lipgloss.NewStyle(),
}

// item 实现 list.Item 接口。
type item struct {
	text string
}

func (i item) FilterValue() string { return i.text }
func (i item) Title() string       { return i.text }
func (i item) Description() string { return "" }

// newItemDelegate 创建列表项委托。
func newItemDelegate() list.ItemDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("10")).BorderLeftForeground(lipgloss.Color("10"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("10"))
	return d
}

// newDBTypeList 创建数据库类型选择列表。
func newDBTypeList(width, height int) list.Model {
	items := make([]list.Item, len(dbTypeOptions))
	for i, opt := range dbTypeOptions {
		items[i] = item{text: opt}
	}
	l := list.New(items, newItemDelegate(), width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return l
}

// newConnMethodList 创建连接方式选择列表。
func newConnMethodList(width, height int) list.Model {
	items := []list.Item{
		item{text: "DSN 连接字符串"},
		item{text: "分离参数（主机/端口/用户名/密码）"},
	}
	l := list.New(items, newItemDelegate(), width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return l
}

// newTextInput 创建文本输入组件。
func newTextInput(placeholder, value string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	ti.CharLimit = 512
	return ti
}

// newPasswordInput 创建密码输入组件。
func newPasswordInput(placeholder, value string) textinput.Model {
	ti := newTextInput(placeholder, value)
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	return ti
}

// newSpinner 创建加载动画组件。
func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	return s
}

// defaultTuiState 从环境变量读取默认值。
func defaultTuiState() TuiState {
	s := TuiState{
		DBType:            orDefault(os.Getenv("DB_TYPE"), "dm"),
		OutputDir:         orDefault(os.Getenv("EXPORT_OUTPUT"), "./output"),
		Formats:           []string{"markdown"},
		SplitFiles:        os.Getenv("EXPORT_SPLIT") == "true" || os.Getenv("EXPORT_SPLIT") == "1",
		IncludeViews:      os.Getenv("EXPORT_INCLUDE_VIEWS") == "true" || os.Getenv("EXPORT_INCLUDE_VIEWS") == "1",
		IncludeProcedures: os.Getenv("EXPORT_INCLUDE_PROCEDURES") == "true" || os.Getenv("EXPORT_INCLUDE_PROCEDURES") == "1",
		IncludeFunctions:  os.Getenv("EXPORT_INCLUDE_FUNCTIONS") == "true" || os.Getenv("EXPORT_INCLUDE_FUNCTIONS") == "1",
		IncludeTriggers:   os.Getenv("EXPORT_INCLUDE_TRIGGERS") == "true" || os.Getenv("EXPORT_INCLUDE_TRIGGERS") == "1",
		IncludeSequences:  os.Getenv("EXPORT_INCLUDE_SEQUENCES") == "true" || os.Getenv("EXPORT_INCLUDE_SEQUENCES") == "1",
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		s.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		s.Port = v
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		s.Database = v
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		s.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		s.Password = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		s.DSN = v
	}
	if v := os.Getenv("DB_SCHEMA"); v != "" {
		s.Schema = v
	}
	if v := os.Getenv("EXPORT_FORMATS"); v != "" {
		s.Formats = parseFormatEnv(v)
	}

	s.ConnectionMethod = ConnectionMethodDSN
	s.normalizeContentOptionsForDBType()
	return s
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func parseFormatEnv(v string) []string {
	parts := strings.Split(v, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return []string{"markdown"}
	}
	return result
}

// --- checkboxModel 多选组件 ---

// checkboxItem 多选项。
type checkboxItem struct {
	title    string
	selected bool
	key      string
}

// checkboxModel 自定义多选组件。
type checkboxModel struct {
	items  []checkboxItem
	cursor int
}

// newFormatCheckboxes 创建格式多选组件。
func newFormatCheckboxes(selected []string) checkboxModel {
	selSet := make(map[string]bool)
	for _, s := range selected {
		selSet[s] = true
	}
	items := make([]checkboxItem, len(formatOptions))
	for i, opt := range formatOptions {
		items[i] = checkboxItem{
			title:    opt,
			selected: selSet[opt],
			key:      opt,
		}
	}
	return checkboxModel{items: items}
}

// newContentCheckboxes 创建导出内容多选组件。
func newContentCheckboxes(s TuiState) checkboxModel {
	s.normalizeContentOptionsForDBType()
	selMap := map[string]bool{
		"views":      s.IncludeViews,
		"procedures": s.IncludeProcedures,
		"functions":  s.IncludeFunctions,
		"triggers":   s.IncludeTriggers,
		"sequences":  s.IncludeSequences,
	}
	options := contentOptionsForDBType(s.DBType)
	items := make([]checkboxItem, len(options))
	for i, opt := range options {
		items[i] = checkboxItem{
			title:    opt.label,
			selected: selMap[opt.key],
			key:      opt.key,
		}
	}
	return checkboxModel{items: items}
}

// update 处理多选组件的键盘事件。
func (c checkboxModel) update(msg tea.Msg) (checkboxModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.cursor > 0 {
				c.cursor--
			}
		case "down", "j":
			if c.cursor < len(c.items)-1 {
				c.cursor++
			}
		case " ":
			c.items[c.cursor].selected = !c.items[c.cursor].selected
		}
	}
	return c, nil
}

// view 渲染多选组件。
func (c checkboxModel) view() string {
	var b strings.Builder
	for i, item := range c.items {
		cursor := "  "
		if i == c.cursor {
			cursor = "> "
		}
		checked := "[ ]"
		if item.selected {
			checked = styles.Selected.Render("[x]")
		}
		line := cursor + checked + " " + item.title
		if i == c.cursor {
			line = styles.Selected.Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// selectedKeys 返回已选项的 key 列表。
func (c checkboxModel) selectedKeys() []string {
	var keys []string
	for _, item := range c.items {
		if item.selected {
			keys = append(keys, item.key)
		}
	}
	return keys
}
