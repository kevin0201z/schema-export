package tui

import (
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

// NewCommand 创建 tui 子命令。
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "通过交互式界面导出数据库结构",
		Long: `以交互式向导方式配置数据库连接和导出选项，无需记忆命令行参数。

支持 DSN 和分离参数两种连接方式，
支持多种导出格式和对象类型选择。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunTUI()
		},
	}
}

// RunTUI 启动 TUI 程序。
func RunTUI() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
