package trader

import (
	"fmt"
	"strings"
	"time"
)

// ErrorCategory 错误分类枚举
type ErrorCategory int

const (
	CategoryUnknown        ErrorCategory = iota
	CategoryRiskControl                  // 风控机制
	CategorySystemBug                    // 系统bug
	CategoryRuntimeError                 // 运行时错误
	CategoryUserInput                    // 用户输入错误
	CategoryNetwork                      // 网络问题
	CategoryDataValidation               // 数据验证错误
)

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// CategorizedError 结构化错误信息
type CategorizedError struct {
	Category     ErrorCategory
	Severity     ErrorSeverity
	ErrorCode    string
	UserMessage  string
	TechnicalMsg string
	Suggestions  []string
	Timestamp    time.Time
	Context      map[string]interface{}
}

func (e *CategorizedError) Error() string {
	return e.TechnicalMsg
}

// ErrorClassifier 错误分类器
type ErrorClassifier struct {
	strictMode bool
}

// NewErrorClassifier 创建错误分类器
func NewErrorClassifier(strictMode bool) *ErrorClassifier {
	return &ErrorClassifier{
		strictMode: strictMode,
	}
}

// ClassifyError 对错误进行分类
func (ec *ErrorClassifier) ClassifyError(err error, context map[string]interface{}) *CategorizedError {
	if err == nil {
		return nil
	}

	// 根据错误内容和上下文进行分类
	category := ec.determineCategory(err, context)
	severity := ec.determineSeverity(category, err)
	errorCode := ec.generateErrorCode(category, err)

	userMsg, suggestions := ec.createUserMessage(category, err, context)

	return &CategorizedError{
		Category:     category,
		Severity:     severity,
		ErrorCode:    errorCode,
		UserMessage:  userMsg,
		TechnicalMsg: err.Error(),
		Suggestions:  suggestions,
		Timestamp:    time.Now(),
		Context:      context,
	}
}

// determineCategory 确定错误分类
func (ec *ErrorClassifier) determineCategory(err error, context map[string]interface{}) ErrorCategory {
	errMsg := strings.ToLower(err.Error())

	// 风控机制相关错误
	riskKeywords := []string{
		"risk", "风控", "limit", "exceed", "超出", "maximum", "最大",
		"position size", "仓位", "leverage", "杠杆", "margin", "保证金",
		"balance", "余额", "insufficient", "不足",
	}

	for _, keyword := range riskKeywords {
		if strings.Contains(errMsg, keyword) {
			return CategoryRiskControl
		}
	}

	// 系统bug相关错误
	bugKeywords := []string{
		"nil pointer", "空指针", "index out of range", "数组越界",
		"panic", "runtime error", "运行时错误", "unexpected",
		"unhandled", "未处理", "assertion", "断言失败",
	}

	for _, keyword := range bugKeywords {
		if strings.Contains(errMsg, keyword) {
			return CategorySystemBug
		}
	}

	// 用户输入错误
	inputKeywords := []string{
		"invalid", "无效", "format", "格式", "parse", "解析",
		"missing", "缺少", "required", "必需",
	}

	for _, keyword := range inputKeywords {
		if strings.Contains(errMsg, keyword) {
			return CategoryUserInput
		}
	}

	// 网络错误
	networkKeywords := []string{
		"network", "网络", "timeout", "超时", "connection", "连接",
		"dial", "dns", "resolve", "解析", "unreachable", "不可达",
	}

	for _, keyword := range networkKeywords {
		if strings.Contains(errMsg, keyword) {
			return CategoryNetwork
		}
	}

	// 数据验证错误
	validationKeywords := []string{
		"validation", "验证", "mismatch", "不匹配", "constraint", "约束",
		"range", "范围", "bounds", "边界",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errMsg, keyword) {
			return CategoryDataValidation
		}
	}

	return CategoryRuntimeError
}

// determineSeverity 确定错误严重程度
func (ec *ErrorClassifier) determineSeverity(category ErrorCategory, err error) ErrorSeverity {
	switch category {
	case CategorySystemBug:
		return SeverityCritical
	case CategoryRiskControl:
		return SeverityHigh
	case CategoryNetwork, CategoryDataValidation:
		return SeverityMedium
	case CategoryUserInput:
		return SeverityLow
	default:
		// 根据错误信息长度和关键词判断
		if strings.Contains(strings.ToLower(err.Error()), "fatal") ||
			strings.Contains(strings.ToLower(err.Error()), "critical") {
			return SeverityCritical
		}
		return SeverityMedium
	}
}

// generateErrorCode 生成错误代码
func (ec *ErrorClassifier) generateErrorCode(category ErrorCategory, err error) string {
	prefix := ""
	switch category {
	case CategoryRiskControl:
		prefix = "RISK"
	case CategorySystemBug:
		prefix = "BUG"
	case CategoryRuntimeError:
		prefix = "RUNTIME"
	case CategoryUserInput:
		prefix = "INPUT"
	case CategoryNetwork:
		prefix = "NETWORK"
	case CategoryDataValidation:
		prefix = "VALID"
	default:
		prefix = "UNKNOWN"
	}

	// 简单的错误代码生成（可以根据需要复杂化）
	timestamp := time.Now().Unix() % 1000
	return fmt.Sprintf("%s-%03d", prefix, timestamp)
}

// createUserMessage 创建用户友好的消息
func (ec *ErrorClassifier) createUserMessage(category ErrorCategory, err error, context map[string]interface{}) (string, []string) {
	var message string
	var suggestions []string

	switch category {
	case CategoryRiskControl:
		message = "🛡️  风控机制触发"
		suggestions = []string{
			"检查仓位规模是否超出限制",
			"确认账户余额充足",
			"验证杠杆倍数设置",
			"查看风险控制参数配置",
		}

	case CategorySystemBug:
		message = "🐛 系统内部错误"
		suggestions = []string{
			"请联系技术支持团队",
			"提供错误代码和操作步骤",
			"稍后重试操作",
			"检查系统更新状态",
		}

	case CategoryRuntimeError:
		message = "⚡ 运行时异常"
		suggestions = []string{
			"检查输入参数是否正确",
			"确认系统资源充足",
			"重启相关服务",
			"查看详细日志信息",
		}

	case CategoryUserInput:
		message = "📝 输入参数错误"
		suggestions = []string{
			"检查参数格式是否正确",
			"确认必填项已填写",
			"参考使用文档说明",
			"验证数据范围限制",
		}

	case CategoryNetwork:
		message = "🌐 网络连接问题"
		suggestions = []string{
			"检查网络连接状态",
			"确认防火墙设置",
			"稍后重试操作",
			"联系网络管理员",
		}

	case CategoryDataValidation:
		message = "🔍 数据验证失败"
		suggestions = []string{
			"检查数据格式和范围",
			"确认数据完整性",
			"验证约束条件",
			"参考数据规范文档",
		}

	default:
		message = "⚠️  未知错误"
		suggestions = []string{
			"请稍后重试",
			"联系技术支持",
			"查看系统日志",
		}
	}

	// 添加错误详情（如果有的话）
	if detail, ok := context["detail"]; ok {
		message += fmt.Sprintf(" - %v", detail)
	}

	return message, suggestions
}

// FormatErrorForDisplay 格式化错误显示
func (ec *ErrorClassifier) FormatErrorForDisplay(categorizedErr *CategorizedError) string {
	if categorizedErr == nil {
		return ""
	}

	var builder strings.Builder

	// 根据分类使用不同的视觉标识
	switch categorizedErr.Category {
	case CategoryRiskControl:
		builder.WriteString("🛡️  [风控警告] ")
	case CategorySystemBug:
		builder.WriteString("🐛 [系统错误] ")
	case CategoryRuntimeError:
		builder.WriteString("⚡ [运行异常] ")
	case CategoryUserInput:
		builder.WriteString("📝 [输入错误] ")
	case CategoryNetwork:
		builder.WriteString("🌐 [网络问题] ")
	case CategoryDataValidation:
		builder.WriteString("🔍 [验证失败] ")
	default:
		builder.WriteString("⚠️  [未知错误] ")
	}

	builder.WriteString(categorizedErr.UserMessage)
	builder.WriteString("\n")

	// 添加错误代码
	builder.WriteString(fmt.Sprintf("   错误代码: %s\n", categorizedErr.ErrorCode))

	// 添加建议
	if len(categorizedErr.Suggestions) > 0 {
		builder.WriteString("   💡 建议操作:\n")
		for i, suggestion := range categorizedErr.Suggestions {
			builder.WriteString(fmt.Sprintf("      %d. %s\n", i+1, suggestion))
		}
	}

	// 严重程度指示
	severityIndicator := ""
	switch categorizedErr.Severity {
	case SeverityLow:
		severityIndicator = "🟢 低"
	case SeverityMedium:
		severityIndicator = "🟡 中"
	case SeverityHigh:
		severityIndicator = "🟠 高"
	case SeverityCritical:
		severityIndicator = "🔴 严重"
	}
	builder.WriteString(fmt.Sprintf("   严重程度: %s\n", severityIndicator))

	// 时间戳
	builder.WriteString(fmt.Sprintf("   发生时间: %s\n", categorizedErr.Timestamp.Format("2006-01-02 15:04:05")))

	return builder.String()
}

// GetTechnicalDetails 获取技术详情（用于调试）
func (ec *ErrorClassifier) GetTechnicalDetails(categorizedErr *CategorizedError) string {
	if categorizedErr == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("=== 技术详情 ===\n")
	builder.WriteString(fmt.Sprintf("错误代码: %s\n", categorizedErr.ErrorCode))
	builder.WriteString(fmt.Sprintf("分类: %v\n", categorizedErr.Category))
	builder.WriteString(fmt.Sprintf("严重程度: %v\n", categorizedErr.Severity))
	builder.WriteString(fmt.Sprintf("技术信息: %s\n", categorizedErr.TechnicalMsg))
	builder.WriteString(fmt.Sprintf("时间戳: %v\n", categorizedErr.Timestamp))

	if len(categorizedErr.Context) > 0 {
		builder.WriteString("上下文信息:\n")
		for key, value := range categorizedErr.Context {
			builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return builder.String()
}
