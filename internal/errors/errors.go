package errors

import (
	"errors"
	"fmt"
)

// 通用错误类型
var (
	// 连接相关错误
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrConnectionTimeout = errors.New("database connection timeout")
	ErrInvalidDSN        = errors.New("invalid database DSN")
	ErrDriverNotFound    = errors.New("database driver not found")

	// 查询相关错误
	ErrQueryFailed    = errors.New("query execution failed")
	ErrTableNotFound  = errors.New("table not found")
	ErrColumnNotFound = errors.New("column not found")
	ErrSchemaNotFound = errors.New("schema not found")

	// 导出相关错误
	ErrExportFailed          = errors.New("export failed")
	ErrInvalidFormat         = errors.New("invalid export format")
	ErrFileCreateFailed      = errors.New("failed to create output file")
	ErrDirectoryCreateFailed = errors.New("failed to create output directory")

	// 配置相关错误
	ErrInvalidConfig        = errors.New("invalid configuration")
	ErrMissingRequiredParam = errors.New("missing required parameter")
)

// Wrap 包装错误，添加上下文信息
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf 包装错误，使用格式化字符串
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New 创建新错误
func New(msg string) error {
	return errors.New(msg)
}

// Newf 使用格式化字符串创建新错误
func Newf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Is 检查错误是否匹配
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为特定类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
