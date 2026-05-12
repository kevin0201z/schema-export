package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	exportapp "github.com/schema-export/schema-export/internal/app/export"
	"github.com/schema-export/schema-export/internal/config"
)

// model 是 Bubble Tea 的主模型。
type model struct {
	// 表单数据
	state TuiState

	// 导航
	currentPage Page
	pageStack   []Page // 导航栈，支持可靠的后退导航

	// 输入焦点
	focusedInput int

	// Bubble Tea 组件
	dbTypeList        list.Model
	connMethodList    list.Model
	dsnInput          textinput.Model
	schemaInput       textinput.Model
	hostInput         textinput.Model
	portInput         textinput.Model
	dbNameInput       textinput.Model
	usernameInput     textinput.Model
	passwordInput     textinput.Model
	tablesInput       textinput.Model
	excludeInput      textinput.Model
	patternsInput     textinput.Model
	outputDirInput    textinput.Model
	contentCheckboxes checkboxModel
	formatCheckboxes  checkboxModel
	spinner           spinner.Model

	// 页面状态
	pageError        string
	sanitizedSummary string

	// 执行状态
	exporting    bool
	exportDone   bool
	exportResult string
	exportError  error

	// 窗口尺寸
	width, height int

	// 导出回调（可替换用于测试）
	exportFunc func(cfg *config.Config) (string, error)
}

// exportFinishedMsg 导出完成消息。
type exportFinishedMsg struct {
	err error
	log string
}

// initialModel 创建初始模型。
func initialModel() model {
	state := defaultTuiState()
	w, h := 80, 24

	dbList := newDBTypeList(w, h-4)
	// 预设默认数据库类型
	for i, listItem := range dbList.Items() {
		if it, ok := listItem.(item); ok && it.text == state.DBType {
			dbList.Select(i)
			break
		}
	}

	connList := newConnMethodList(w, h-4)

	m := model{
		state:             state,
		currentPage:       PageWelcome,
		pageStack:         nil,
		width:             w,
		height:            h,
		dbTypeList:        dbList,
		connMethodList:    connList,
		dsnInput:          newTextInput("例如: dm://user:password@host:5236", state.DSN),
		schemaInput:       newTextInput("Schema 名称（可选）", state.Schema),
		hostInput:         newTextInput("例如: localhost", state.Host),
		portInput:         newTextInput("例如: 5236", state.Port),
		dbNameInput:       newTextInput("数据库名称", state.Database),
		usernameInput:     newTextInput("用户名", state.Username),
		passwordInput:     newPasswordInput("密码", state.Password),
		tablesInput:       newTextInput("表名（逗号分隔，可选）", state.Tables),
		excludeInput:      newTextInput("排除表名（逗号分隔，可选）", state.Exclude),
		patternsInput:     newTextInput("正则匹配（逗号分隔，可选）", state.Patterns),
		outputDirInput:    newTextInput("输出目录", state.OutputDir),
		contentCheckboxes: newContentCheckboxes(state),
		formatCheckboxes:  newFormatCheckboxes(state.Formats),
		spinner:           newSpinner(),
		exportFunc:        defaultExportFunc,
	}

	// 默认 DSN 模式第一个输入获得焦点
	m.dsnInput.Focus()

	return m
}

// defaultExportFunc 默认导出函数，使用 buffer 捕获输出。
var defaultExportFunc = func(cfg *config.Config) (string, error) {
	var buf strings.Builder
	svc := exportapp.NewServiceWithWriters(cfg, &buf, &buf)
	err := svc.Run()
	return buf.String(), err
}

// Init 实现 tea.Model 接口。
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

// Update 实现 tea.Model 接口。
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dbTypeList.SetSize(msg.Width, msg.Height-4)
		m.connMethodList.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.handleBack()
		case "enter":
			return m.handleNext()
		case "tab":
			m.nextFocus()
			return m, m.focusCurrentField()
		case "shift+tab":
			m.prevFocus()
			return m, m.focusCurrentField()
		}

		// 页面特定按键处理
		return m.handlePageKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case exportFinishedMsg:
		m.exportDone = true
		m.exportError = msg.err
		m.exporting = false
		if msg.err == nil {
			m.exportResult = "导出完成！文件输出至: " + m.state.OutputDir + "\n\n" + msg.log
		}
		return m, nil
	}

	return m, nil
}

// View 实现 tea.Model 接口。
func (m model) View() string {
	var pageView string
	switch m.currentPage {
	case PageWelcome:
		pageView = m.welcomeView()
	case PageDBType:
		pageView = m.dbTypeView()
	case PageConnectionMethod:
		pageView = m.connMethodView()
	case PageDSNForm:
		pageView = m.dsnFormView()
	case PageParamsForm:
		pageView = m.paramsFormView()
	case PageExportContent:
		pageView = m.exportContentView()
	case PageTableFilter:
		pageView = m.tableFilterView()
	case PageOutputSettings:
		pageView = m.outputSettingsView()
	case PageConfirm:
		pageView = m.confirmView()
	case PageExecution:
		pageView = m.executionView()
	default:
		pageView = m.welcomeView()
	}
	return pageView
}

// handleNext 处理 Enter 键——提交当前页并进入下一页。
func (m model) handleNext() (tea.Model, tea.Cmd) {
	// 同步组件值到状态
	m.syncStateFromInputs()

	// 校验当前页
	if err := m.validateCurrentPage(); err != nil {
		m.pageError = err.Error()
		return m, nil
	}
	m.pageError = ""

	// 执行页的完成状态处理
	if m.currentPage == PageExecution && m.exportDone {
		if m.exportError != nil {
			// 弹出栈顶的 Confirm 页，使 Esc 能回到可编辑页面（OutputSettings）
			m.popPage()
			m.currentPage = PageConfirm
			return m, nil
		}
		// 成功——清空导航栈，返回欢迎页
		m.pageStack = nil
		m.currentPage = PageWelcome
		return m, nil
	}

	// 向前导航时压栈当前页
	m.pushPage(m.currentPage)

	// 连接方式页：决定下一页
	if m.currentPage == PageConnectionMethod {
		if m.state.ConnectionMethod == ConnectionMethodDSN {
			m.currentPage = PageDSNForm
		} else {
			m.currentPage = PageParamsForm
		}
	} else if m.currentPage == PageConfirm {
		m.currentPage = PageExecution
		m.resetExportState()
		// 启动导出
		return m, m.startExport()
	} else {
		m.currentPage = nextPage(m.currentPage)
	}

	// 进入确认页时预构建脱敏摘要供展示
	if m.currentPage == PageConfirm {
		m.sanitizedSummary = m.state.SanitizeSummary()
	}

	// 进入新页面时聚焦第一个字段
	m.focusedInput = 0
	return m, m.focusCurrentField()
}

// pushPage 向导航栈压入页面。
func (m *model) pushPage(p Page) {
	m.pageStack = append(m.pageStack, p)
}

// popPage 从导航栈弹出上一页。栈空时返回 PageWelcome。
func (m *model) popPage() Page {
	if len(m.pageStack) == 0 {
		return PageWelcome
	}
	p := m.pageStack[len(m.pageStack)-1]
	m.pageStack = m.pageStack[:len(m.pageStack)-1]
	return p
}

// handleBack 处理 Esc 键——返回上一页或退出。
func (m model) handleBack() (tea.Model, tea.Cmd) {
	if m.currentPage == PageWelcome || len(m.pageStack) == 0 {
		return m, tea.Quit
	}
	m.pageError = ""
	m.currentPage = m.popPage()
	m.focusedInput = 0
	return m, m.focusCurrentField()
}

// handlePageKey 将按键事件路由到当前页面的组件。
func (m model) handlePageKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentPage {
	case PageDBType:
		var cmd tea.Cmd
		m.dbTypeList, cmd = m.dbTypeList.Update(msg)
		return m, cmd
	case PageConnectionMethod:
		var cmd tea.Cmd
		m.connMethodList, cmd = m.connMethodList.Update(msg)
		return m, cmd
	case PageDSNForm:
		return m.updateTextInputPage(msg, []*textinput.Model{&m.dsnInput, &m.schemaInput})
	case PageParamsForm:
		return m.updateTextInputPage(msg, []*textinput.Model{
			&m.hostInput, &m.portInput, &m.dbNameInput,
			&m.usernameInput, &m.passwordInput, &m.schemaInput,
		})
	case PageTableFilter:
		return m.updateTextInputPage(msg, []*textinput.Model{
			&m.tablesInput, &m.excludeInput, &m.patternsInput,
		})
	case PageOutputSettings:
		return m.updateOutputSettingsKey(msg)
	case PageExportContent:
		var cmd tea.Cmd
		m.contentCheckboxes, cmd = m.contentCheckboxes.update(msg)
		return m, cmd
	}
	return m, nil
}

// updateTextInputPage 更新文本输入页面的按键。
func (m *model) updateTextInputPage(msg tea.KeyMsg, inputs []*textinput.Model) (tea.Model, tea.Cmd) {
	if m.focusedInput >= 0 && m.focusedInput < len(inputs) {
		var cmd tea.Cmd
		*inputs[m.focusedInput], cmd = inputs[m.focusedInput].Update(msg)
		return m, cmd
	}
	return m, nil
}

// updateOutputSettingsKey 处理输出设置页的按键。
func (m model) updateOutputSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.focusedInput {
	case 0:
		var cmd tea.Cmd
		m.outputDirInput, cmd = m.outputDirInput.Update(msg)
		return m, cmd
	case 1:
		// 格式多选
		m.formatCheckboxes, _ = m.formatCheckboxes.update(msg)
		return m, nil
	case 2:
		// 分文件开关
		if msg.String() == " " {
			m.state.SplitFiles = !m.state.SplitFiles
		}
		return m, nil
	}
	return m, nil
}

// syncStateFromInputs 从组件读取值到 TuiState。
func (m *model) syncStateFromInputs() {
	// 列表选择
	if sel, ok := m.dbTypeList.SelectedItem().(item); ok {
		m.state.DBType = sel.text
	}
	if sel, ok := m.connMethodList.SelectedItem().(item); ok {
		if strings.Contains(sel.text, "DSN") {
			m.state.ConnectionMethod = ConnectionMethodDSN
		} else {
			m.state.ConnectionMethod = ConnectionMethodParams
		}
	}

	// 文本输入
	m.state.DSN = m.dsnInput.Value()
	m.state.Host = m.hostInput.Value()
	m.state.Port = m.portInput.Value()
	m.state.Database = m.dbNameInput.Value()
	m.state.Username = m.usernameInput.Value()
	m.state.Password = m.passwordInput.Value()
	m.state.Schema = m.schemaInput.Value()
	m.state.Tables = m.tablesInput.Value()
	m.state.Exclude = m.excludeInput.Value()
	m.state.Patterns = m.patternsInput.Value()
	m.state.OutputDir = m.outputDirInput.Value()

	// 复选框
	m.state.applyContentCheckboxes(m.contentCheckboxes)
	m.state.applyFormatCheckboxes(m.formatCheckboxes)
}

// startExport 启动导出（在 goroutine 中执行）。
func (m model) startExport() tea.Cmd {
	return func() tea.Msg {
		cfg := m.state.ToConfig()
		if err := cfg.Validate(); err != nil {
			return exportFinishedMsg{err: fmt.Errorf("配置校验失败: %w", err)}
		}
		log, err := m.exportFunc(cfg)
		return exportFinishedMsg{err: err, log: log}
	}
}

func (m *model) resetExportState() {
	m.exporting = true
	m.exportDone = false
	m.exportError = nil
	m.exportResult = ""
}

// --- 焦点管理 ---

// nextFocus 切换到下一个输入字段。
func (m *model) nextFocus() {
	count := pageInputCount(m.currentPage)
	if count <= 1 {
		return
	}
	m.focusedInput++
	if m.focusedInput >= count {
		m.focusedInput = 0
	}
}

// prevFocus 切换到上一个输入字段。
func (m *model) prevFocus() {
	count := pageInputCount(m.currentPage)
	if count <= 1 {
		return
	}
	m.focusedInput--
	if m.focusedInput < 0 {
		m.focusedInput = count - 1
	}
}

// focusCurrentField 聚焦当前页面的第 focusedInput 个字段。
func (m *model) focusCurrentField() tea.Cmd {
	m.blurAllInputs()

	var cmd tea.Cmd
	switch m.currentPage {
	case PageDSNForm:
		if m.focusedInput == 0 {
			cmd = m.dsnInput.Focus()
		} else {
			cmd = m.schemaInput.Focus()
		}
	case PageParamsForm:
		inputs := []*textinput.Model{&m.hostInput, &m.portInput, &m.dbNameInput, &m.usernameInput, &m.passwordInput, &m.schemaInput}
		if m.focusedInput >= 0 && m.focusedInput < len(inputs) {
			cmd = inputs[m.focusedInput].Focus()
		}
	case PageTableFilter:
		inputs := []*textinput.Model{&m.tablesInput, &m.excludeInput, &m.patternsInput}
		if m.focusedInput >= 0 && m.focusedInput < len(inputs) {
			cmd = inputs[m.focusedInput].Focus()
		}
	case PageOutputSettings:
		if m.focusedInput == 0 {
			cmd = m.outputDirInput.Focus()
		}
	}

	return cmd
}

// blurAllInputs 取消所有文本输入的焦点。
func (m *model) blurAllInputs() {
	m.dsnInput.Blur()
	m.schemaInput.Blur()
	m.hostInput.Blur()
	m.portInput.Blur()
	m.dbNameInput.Blur()
	m.usernameInput.Blur()
	m.passwordInput.Blur()
	m.tablesInput.Blur()
	m.excludeInput.Blur()
	m.patternsInput.Blur()
	m.outputDirInput.Blur()
}

// parsePort 解析端口字符串。
func parsePort(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}
