package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/schema-export/schema-export/internal/config"
)

// testModel 创建用于测试的模型。
func testModel() model {
	m := initialModel()
	m.exportFunc = func(cfg *config.Config) (string, error) {
		return "", nil
	}
	return m
}

// advanceTo 模拟用户连续按 Enter 直到到达目标页面。
func advanceTo(m model, target Page) model {
	for m.currentPage != target {
		// 页面有输入时预填必要值
		fillRequiredFields(&m)
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = result.(model)
	}
	return m
}

// fillRequiredFields 为当前页预填必填字段值。
func fillRequiredFields(m *model) {
	switch m.currentPage {
	case PageDSNForm:
		if m.dsnInput.Value() == "" {
			m.dsnInput.SetValue("dm://user:pass@host:5236")
		}
	case PageParamsForm:
		if m.hostInput.Value() == "" {
			m.hostInput.SetValue("localhost")
		}
		if m.usernameInput.Value() == "" {
			m.usernameInput.SetValue("root")
		}
	}
}

func TestConfirmPageShowsSummary(t *testing.T) {
	m := testModel()

	// 模拟用户从 Welcome 一路按 Enter 到 Confirm 页
	m = advanceTo(m, PageConfirm)

	// 到达确认页时应已预构建脱敏摘要
	if m.sanitizedSummary == "" {
		t.Fatal("sanitizedSummary should be populated when entering Confirm page")
	}
}

func TestNavigationBackFromConfirm(t *testing.T) {
	m := testModel()

	// 到达确认页
	m = advanceTo(m, PageConfirm)

	// 按 Esc 应该回到 OutputSettings
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := result.(model)

	if m2.currentPage != PageOutputSettings {
		t.Fatalf("Esc from Confirm should go to OutputSettings, got %v", m2.currentPage)
	}
}

func TestNavigationBackMultipleSteps(t *testing.T) {
	m := testModel()

	// 到达确认页
	m = advanceTo(m, PageConfirm)

	// 连续按 Esc 3 次应该能回到前面的页面
	for i := 0; i < 3; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = result.(model)
	}

	// 不应该崩溃，currentPage 应该是一个有效的早期页面
	if m.currentPage == PageExecution || m.currentPage == PageConfirm {
		t.Fatalf("after 3 Esc presses should be before Confirm, got %v", m.currentPage)
	}
}

func TestFailedExportReturnsToConfirmWithoutStacking(t *testing.T) {
	m := testModel()
	// 到达确认页
	m = advanceTo(m, PageConfirm)
	stackBefore := len(m.pageStack)

	// 模拟进入执行页
	m = advanceTo(m, PageExecution)

	// 模拟导出失败
	m.exportDone = true
	m.exportError = errSentinel

	// 按 Enter —— 应返回确认页
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := result.(model)

	if m2.currentPage != PageConfirm {
		t.Fatalf("failed export Enter should go to Confirm, got %v", m2.currentPage)
	}

	// 栈应减少一层（失败返回时弹出了栈顶的 Confirm 页）
	if len(m2.pageStack) != stackBefore {
		t.Fatalf("stack should be at pre-confirm level after failed-export return, got %d, want %d", len(m2.pageStack), stackBefore)
	}
}

func TestFailedExportEscGoesToEditablePage(t *testing.T) {
	m := testModel()
	m = advanceTo(m, PageConfirm)

	// 记录当前栈顶（应该是 OutputSettings）
	lastEditablePage := m.pageStack[len(m.pageStack)-1]

	m = advanceTo(m, PageExecution)

	// 模拟导出失败
	m.exportDone = true
	m.exportError = errSentinel

	// Enter 返回确认页
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(model)

	// Esc 应弹出到可编辑页面（OutputSettings），而非 Execution
	result2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := result2.(model)

	if m2.currentPage != lastEditablePage {
		t.Fatalf("Esc after failed export should return to last editable page %v, got %v", lastEditablePage, m2.currentPage)
	}
}

func TestSuccessfulExportClearsStack(t *testing.T) {
	m := testModel()
	m = advanceTo(m, PageConfirm)
	m = advanceTo(m, PageExecution)

	// 模拟导出成功
	m.exportDone = true
	m.exportError = nil

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := result.(model)

	if m2.currentPage != PageWelcome {
		t.Fatalf("successful export should return to Welcome, got %v", m2.currentPage)
	}
	if len(m2.pageStack) != 0 {
		t.Fatalf("successful export should clear stack, got %d items", len(m2.pageStack))
	}
}

func TestRetryExportResetsExecutionStateImmediately(t *testing.T) {
	m := testModel()
	m = advanceTo(m, PageConfirm)
	m = advanceTo(m, PageExecution)

	m.exportDone = true
	m.exportError = errSentinel
	m.exportResult = "stale result"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(model)

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(model)

	if m.currentPage != PageExecution {
		t.Fatalf("retry should navigate to execution page, got %v", m.currentPage)
	}
	if !m.exporting {
		t.Fatal("retry should mark export as in progress immediately")
	}
	if m.exportDone {
		t.Fatal("retry should clear exportDone before async completion")
	}
	if m.exportError != nil {
		t.Fatalf("retry should clear previous exportError, got %v", m.exportError)
	}
	if m.exportResult != "" {
		t.Fatalf("retry should clear previous exportResult, got %q", m.exportResult)
	}
	if cmd == nil {
		t.Fatal("retry should schedule a new export command")
	}
}

func TestNextPageMapping(t *testing.T) {
	tests := []struct {
		from, to Page
	}{
		{PageWelcome, PageDBType},
		{PageDBType, PageConnectionMethod},
		{PageExportContent, PageTableFilter},
		{PageTableFilter, PageOutputSettings},
		{PageOutputSettings, PageConfirm},
		{PageConfirm, PageExecution},
	}
	for _, tt := range tests {
		if got := nextPage(tt.from); got != tt.to {
			t.Fatalf("nextPage(%v): got %v, want %v", tt.from, got, tt.to)
		}
	}
}

// errSentinel 用于测试的错误哨兵。
var errSentinel = &testError{msg: "export failed"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
