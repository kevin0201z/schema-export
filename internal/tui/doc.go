// Package tui 提供基于 Bubble Tea 的终端交互式导出界面。
//
// 通过分步向导引导用户完成数据库导出配置，适用于：
//   - 不熟悉命令行参数的用户
//   - 需要频繁调整导出范围的场景
//   - 无 Web UI 但可使用终端的环境
//
// 本包状态模型 TuiState 可通过 ToConfig 方法转换为
// config.Config 后复用现有的导出服务。
package tui
