package tui

import (
	"fmt"
	"net/url"
	"strings"
)

// SanitizeDSN 脱敏 DSN 中的密码部分。
//
// 支持标准 URL 格式（如 dm://user:pass@host）和原始连接字符串
// （如 MySQL user:pass@tcp(host:port)/db、SQL Server user:pass@host?params）。
func SanitizeDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return sanitizeDSNFallback(dsn)
	}

	// 已解析到用户信息且含密码：从标准化 URL 中掩码
	if u.User != nil {
		if pw, ok := u.User.Password(); ok && pw != "" {
			u.User = url.User(u.User.Username())
			s := u.String()
			protoEnd := strings.Index(s, "://")
			if protoEnd == -1 {
				return s
			}
			userInfoStart := protoEnd + 3
			atIdx := strings.Index(s[userInfoStart:], "@")
			if atIdx == -1 {
				return s
			}
			atIdx += userInfoStart
			return s[:atIdx] + ":******" + s[atIdx:]
		}
		// 有用户信息但没有密码——安全
		return dsn
	}

	// url.Parse 成功但未解析到用户信息（如原始 DSN 格式）
	// 回退到字符串级别检查并掩码
	return sanitizeDSNFallback(dsn)
}

// sanitizeDSNFallback 无法解析 URL 时的密码脱敏降级方案。
func sanitizeDSNFallback(dsn string) string {
	idx := strings.Index(dsn, "@")
	if idx == -1 {
		return dsn
	}
	userInfo := dsn[:idx]
	colonIdx := strings.LastIndex(userInfo, ":")
	if colonIdx == -1 {
		return dsn
	}
	// 检查冒号前面是否已经有协议分隔符（如 dm://）
	protoEnd := strings.Index(userInfo, "://")
	if protoEnd != -1 && colonIdx <= protoEnd+1 {
		return dsn
	}
	return dsn[:colonIdx+1] + "******" + dsn[idx:]
}

// SanitizeSummary 构建确认页的脱敏摘要。
func (s *TuiState) SanitizeSummary() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("数据库类型: %s\n", s.DBType))

	if s.ConnectionMethod == ConnectionMethodDSN {
		b.WriteString("连接方式: DSN\n")
		b.WriteString(fmt.Sprintf("DSN: %s\n", SanitizeDSN(s.DSN)))
		if s.Schema != "" {
			b.WriteString(fmt.Sprintf("Schema: %s\n", s.Schema))
		}
	} else {
		b.WriteString("连接方式: 分离参数\n")
		b.WriteString(fmt.Sprintf("主机: %s\n", s.Host))
		if s.Port != "" {
			b.WriteString(fmt.Sprintf("端口: %s\n", s.Port))
		}
		if s.Database != "" {
			b.WriteString(fmt.Sprintf("数据库: %s\n", s.Database))
		}
		b.WriteString(fmt.Sprintf("用户名: %s\n", s.Username))
		b.WriteString("密码: ******\n")
		if s.Schema != "" {
			b.WriteString(fmt.Sprintf("Schema: %s\n", s.Schema))
		}
	}

	b.WriteString("\n导出内容:\n")
	b.WriteString("  表: 是\n")
	b.WriteString(fmt.Sprintf("  视图: %s\n", boolToYesNo(s.IncludeViews)))
	b.WriteString(fmt.Sprintf("  存储过程: %s\n", boolToYesNo(s.IncludeProcedures)))
	b.WriteString(fmt.Sprintf("  函数: %s\n", boolToYesNo(s.IncludeFunctions)))
	b.WriteString(fmt.Sprintf("  触发器: %s\n", boolToYesNo(s.IncludeTriggers)))
	b.WriteString(fmt.Sprintf("  序列: %s\n", boolToYesNo(s.IncludeSequences)))

	if s.Tables != "" || s.Exclude != "" || s.Patterns != "" {
		b.WriteString("\n表过滤:\n")
		if s.Tables != "" {
			b.WriteString(fmt.Sprintf("  指定表: %s\n", s.Tables))
		}
		if s.Exclude != "" {
			b.WriteString(fmt.Sprintf("  排除表: %s\n", s.Exclude))
		}
		if s.Patterns != "" {
			b.WriteString(fmt.Sprintf("  正则匹配: %s\n", s.Patterns))
		}
	}

	b.WriteString("\n输出设置:\n")
	b.WriteString(fmt.Sprintf("  输出目录: %s\n", s.OutputDir))
	b.WriteString(fmt.Sprintf("  格式: %s\n", strings.Join(s.Formats, ", ")))
	b.WriteString(fmt.Sprintf("  分文件: %s\n", boolToYesNo(s.SplitFiles)))

	return b.String()
}

func boolToYesNo(v bool) string {
	if v {
		return "是"
	}
	return "否"
}
