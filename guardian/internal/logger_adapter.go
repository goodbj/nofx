package guardian

import (
	"nofx/logger"
)

// LoggerAdapter 适配器，使Guardian能够使用主项目的日志系统
type LoggerAdapter struct{}

// NewLoggerAdapter 创建新的日志适配器
func NewLoggerAdapter() *LoggerAdapter {
	return &LoggerAdapter{}
}

// Debugf 输出调试日志
func (l *LoggerAdapter) Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Infof 输出信息日志
func (l *LoggerAdapter) Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warnf 输出警告日志
func (l *LoggerAdapter) Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Errorf 输出错误日志
func (l *LoggerAdapter) Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Printf 输出普通日志（对应Info级别）
func (l *LoggerAdapter) Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}
