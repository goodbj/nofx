package utils

import (
	"strings"
)

// ReasoningProcessor 定义reasoning字段处理器接口
type ReasoningProcessor interface {
	Process(content string) string
}

// DefaultReasoningHook 默认的reasoning字段处理钩子
type DefaultReasoningHook struct {
	// 可以在这里添加配置选项
	Enabled bool
	Filters []ReasoningFilter
}

// ReasoningFilter 定义过滤器函数类型
type ReasoningFilter func(string) string

// Process 处理reasoning字段内容
func (h *DefaultReasoningHook) Process(content string) string {
	if !h.Enabled {
		// 如果钩子未启用，直接返回原始内容
		return content
	}

	result := content
	for _, filter := range h.Filters {
		result = filter(result)
	}
	return result
}

// AddFilter 添加过滤器
func (h *DefaultReasoningHook) AddFilter(filter ReasoningFilter) {
	h.Filters = append(h.Filters, filter)
}

// NewDefaultReasoningHook 创建默认的reasoning钩子
func NewDefaultReasoningHook(enabled bool) *DefaultReasoningHook {
	hook := &DefaultReasoningHook{
		Enabled: enabled,
		Filters: make([]ReasoningFilter, 0),
	}

	// 可以根据需要添加默认过滤器
	if enabled {
		// 示例：移除多余的空白字符
		hook.AddFilter(func(s string) string {
			// 只在启用时添加默认处理逻辑
			return strings.TrimSpace(s)
		})
	}

	return hook
}

// GlobalReasoningHook 全局reasoning钩子实例
var GlobalReasoningHook = NewDefaultReasoningHook(false) // 默认禁用

// ProcessReasoning 使用全局钩子处理reasoning字段
func ProcessReasoning(content string) string {
	return GlobalReasoningHook.Process(content)
}

// EnableReasoningHook 启用reasoning钩子
func EnableReasoningHook() {
	GlobalReasoningHook.Enabled = true
}

// DisableReasoningHook 禁用reasoning钩子
func DisableReasoningHook() {
	GlobalReasoningHook.Enabled = false
}

// RegisterReasoningFilter 注册自定义过滤器
func RegisterReasoningFilter(filter ReasoningFilter) {
	GlobalReasoningHook.AddFilter(filter)
}

// SanitizeForJSON 专门用于JSON安全的清理函数
func SanitizeForJSON(content string) string {
	// 仅进行必要的JSON安全转义
	content = strings.ReplaceAll(content, `\`, `\\`)   // 转义反斜杠
	content = strings.ReplaceAll(content, `"`, `\"`)   // 转义双引号
	content = strings.ReplaceAll(content, "\n", "\\n") // 转义换行符
	content = strings.ReplaceAll(content, "\r", "\\r") // 转义回车符
	content = strings.ReplaceAll(content, "\t", "\\t") // 转义制表符
	return content
}
