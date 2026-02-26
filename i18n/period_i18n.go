package i18n

import (
	"fmt"
	"strings"
)

// Language 语言类型
type Language string

const (
	Chinese  Language = "zh-CN"
	English  Language = "en-US"
	Japanese Language = "ja-JP"
)

// APIErrorTranslations API错误消息国际化文本
var APIErrorTranslations = map[Language]map[string]string{
	Chinese: {
		"-2027":            "超出当前杠杆下的最大允许持仓",
		"-2026":            "账户余额不足",
		"-2025":            "订单金额过小",
		"-2024":            "订单金额过大",
		"-2023":            "价格超出限制范围",
		"-2022":            "交易对暂停交易",
		"-2021":            "系统维护中",
		"-2020":            "参数错误",
		"-2019":            "API权限不足",
		"-2018":            "请求频率超限",
		"-2017":            "服务器内部错误",
		"-2016":            "网络连接异常",
		"-2015":            "订单不存在",
		"-2014":            "订单状态不允许此操作",
		"-2013":            "保证金不足",
		"-2012":            "持仓方向冲突",
		"-2011":            "价格精度超限",
		"-2010":            "数量精度超限",
		"unknown_error":    "未知错误",
		"api_error_prefix": "API错误",
		"code_label":       "错误码",
		"message_label":    "错误信息",
	},
	English: {
		"-2027":            "Exceeded the maximum allowable position at current leverage",
		"-2026":            "Insufficient account balance",
		"-2025":            "Order amount too small",
		"-2024":            "Order amount too large",
		"-2023":            "Price exceeds limit range",
		"-2022":            "Trading pair suspended",
		"-2021":            "System maintenance",
		"-2020":            "Invalid parameters",
		"-2019":            "Insufficient API permissions",
		"-2018":            "Request rate limit exceeded",
		"-2017":            "Internal server error",
		"-2016":            "Network connection error",
		"-2015":            "Order not found",
		"-2014":            "Order status does not allow this operation",
		"-2013":            "Insufficient margin",
		"-2012":            "Position direction conflict",
		"-2011":            "Price precision exceeded",
		"-2010":            "Quantity precision exceeded",
		"unknown_error":    "Unknown error",
		"api_error_prefix": "API Error",
		"code_label":       "Error Code",
		"message_label":    "Error Message",
	},
	Japanese: {
		"-2027":            "現在のレバレッジでの最大許容ポジションを超過しました",
		"-2026":            "口座残高不足",
		"-2025":            "注文金額が小さすぎます",
		"-2024":            "注文金額が大きすぎます",
		"-2023":            "価格が制限範囲を超えています",
		"-2022":            "取引ペアが一時停止中",
		"-2021":            "システムメンテナンス中",
		"-2020":            "パラメータエラー",
		"-2019":            "API権限不足",
		"-2018":            "リクエスト頻度制限超過",
		"-2017":            "サーバー内部エラー",
		"-2016":            "ネットワーク接続エラー",
		"-2015":            "注文が存在しません",
		"-2014":            "注文状態によりこの操作は許可されません",
		"-2013":            "証拠金不足",
		"-2012":            "ポジション方向の衝突",
		"-2011":            "価格精度超過",
		"-2010":            "数量精度超過",
		"unknown_error":    "不明なエラー",
		"api_error_prefix": "APIエラー",
		"code_label":       "エラーコード",
		"message_label":    "エラーメッセージ",
	},
}

// PeriodTranslations 周期相关国际化文本
var PeriodTranslations = map[Language]map[string]string{
	Chinese: {
		"cycle_prefix":       "第",
		"cycle_suffix":       "周期",
		"cycle_running":      "运行中",
		"cycle_completed":    "已完成",
		"cycle_waiting":      "等待中",
		"cycle_error":        "错误",
		"runtime_minutes":    "运行分钟",
		"position_count":     "持仓数",
		"total_profit":       "总收益",
		"max_drawdown":       "最大回撤",
		"win_rate":           "胜率",
		"start_time":         "开始时间",
		"end_time":           "结束时间",
		"symbol":             "交易对",
		"action":             "操作",
		"leverage":           "杠杆",
		"position_size":      "仓位规模",
		"stop_loss":          "止损",
		"take_profit":        "止盈",
		"confidence":         "置信度",
		"reasoning":          "理由",
		"open_long":          "开多",
		"open_short":         "开空",
		"close_long":         "平多",
		"close_short":        "平空",
		"hold":               "持有",
		"wait":               "等待",
		"update_stop_loss":   "更新止损",
		"update_take_profit": "更新止盈",
	},
	English: {
		"cycle_prefix":       "Cycle #",
		"cycle_suffix":       "",
		"cycle_running":      "Running",
		"cycle_completed":    "Completed",
		"cycle_waiting":      "Waiting",
		"cycle_error":        "Error",
		"runtime_minutes":    "Runtime (min)",
		"position_count":     "Positions",
		"total_profit":       "Total Profit",
		"max_drawdown":       "Max Drawdown",
		"win_rate":           "Win Rate",
		"start_time":         "Start Time",
		"end_time":           "End Time",
		"symbol":             "Symbol",
		"action":             "Action",
		"leverage":           "Leverage",
		"position_size":      "Position Size",
		"stop_loss":          "Stop Loss",
		"take_profit":        "Take Profit",
		"confidence":         "Confidence",
		"reasoning":          "Reasoning",
		"open_long":          "Open Long",
		"open_short":         "Open Short",
		"close_long":         "Close Long",
		"close_short":        "Close Short",
		"hold":               "Hold",
		"wait":               "Wait",
		"update_stop_loss":   "Update Stop Loss",
		"update_take_profit": "Update Take Profit",
	},
	Japanese: {
		"cycle_prefix":       "第",
		"cycle_suffix":       "サイクル",
		"cycle_running":      "実行中",
		"cycle_completed":    "完了",
		"cycle_waiting":      "待機中",
		"cycle_error":        "エラー",
		"runtime_minutes":    "実行時間(分)",
		"position_count":     "ポジション数",
		"total_profit":       "総利益",
		"max_drawdown":       "最大ドローダウン",
		"win_rate":           "勝率",
		"start_time":         "開始時間",
		"end_time":           "終了時間",
		"symbol":             "取引ペア",
		"action":             "アクション",
		"leverage":           "レバレッジ",
		"position_size":      "ポジションサイズ",
		"stop_loss":          "ストップロス",
		"take_profit":        "テイクプロフィット",
		"confidence":         "信頼度",
		"reasoning":          "理由",
		"open_long":          "ロング",
		"open_short":         "ショート",
		"close_long":         "ロング決済",
		"close_short":        "ショート決済",
		"hold":               "保持",
		"wait":               "待機",
		"update_stop_loss":   "ストップロス更新",
		"update_take_profit": "テイクプロフィット更新",
	},
}

// APIErrorI18n API错误国际化处理器
type APIErrorI18n struct {
	lang Language
}

// PeriodI18n 周期国际化处理器
type PeriodI18n struct {
	lang Language
}

// NewAPIErrorI18n 创建API错误国际化处理器
func NewAPIErrorI18n(lang Language) *APIErrorI18n {
	// 默认使用中文
	if _, exists := APIErrorTranslations[lang]; !exists {
		lang = Chinese
	}
	return &APIErrorI18n{lang: lang}
}

// GetAPIError 获取API错误翻译
func (a *APIErrorI18n) GetAPIError(code string, message string) string {
	// 获取错误码对应的翻译
	errorCode := code
	if strings.HasPrefix(code, "code=") {
		errorCode = strings.TrimPrefix(code, "code=")
	}

	// 移除msg=前缀
	originalMsg := message
	if strings.HasPrefix(message, "msg=") {
		originalMsg = strings.TrimPrefix(message, "msg=")
	}

	// 查找翻译
	var translatedMsg string
	if translations, exists := APIErrorTranslations[a.lang]; exists {
		if text, found := translations[errorCode]; found {
			translatedMsg = text
		} else {
			translatedMsg = translations["unknown_error"]
		}
	} else {
		// 回退到中文
		if translations, exists := APIErrorTranslations[Chinese]; exists {
			if text, found := translations[errorCode]; found {
				translatedMsg = text
			} else {
				translatedMsg = translations["unknown_error"]
			}
		} else {
			translatedMsg = "未知错误"
		}
	}

	// 格式化输出
	switch a.lang {
	case Chinese:
		return fmt.Sprintf("❌ %s: <API错误> 错误码=%s, 错误信息=%s", translatedMsg, errorCode, originalMsg)
	case English:
		return fmt.Sprintf("❌ %s: <API Error> Code=%s, Message=%s", translatedMsg, errorCode, originalMsg)
	case Japanese:
		return fmt.Sprintf("❌ %s: <APIエラー> コード=%s, メッセージ=%s", translatedMsg, errorCode, originalMsg)
	default:
		return fmt.Sprintf("❌ %s: <API Error> Code=%s, Message=%s", translatedMsg, errorCode, originalMsg)
	}
}

// FormatAPIError 简化版API错误格式化
func (a *APIErrorI18n) FormatAPIError(errorString string) string {
	// 解析错误字符串格式: "failed to open long position: <APIError> code=-2027, msg=Exceeded the maximum allowable position at current leverage."

	// 提取code和msg
	var code, msg string

	// 查找code=
	codeStart := strings.Index(errorString, "code=")
	if codeStart != -1 {
		codeEnd := strings.Index(errorString[codeStart:], ",")
		if codeEnd != -1 {
			code = errorString[codeStart : codeStart+codeEnd]
		} else {
			code = errorString[codeStart:]
		}
	}

	// 查找msg=
	msgStart := strings.Index(errorString, "msg=")
	if msgStart != -1 {
		msg = errorString[msgStart:]
		// 移除末尾的点号
		msg = strings.TrimSuffix(msg, ".")
	}

	// 如果解析失败，返回原错误
	if code == "" || msg == "" {
		return errorString
	}

	return a.GetAPIError(code, msg)
}

// NewPeriodI18n 创建周期国际化处理器
func NewPeriodI18n(lang Language) *PeriodI18n {
	// 默认使用中文
	if _, exists := PeriodTranslations[lang]; !exists {
		lang = Chinese
	}
	return &PeriodI18n{lang: lang}
}

// Get 获取翻译文本
func (p *PeriodI18n) Get(key string) string {
	if translations, exists := PeriodTranslations[p.lang]; exists {
		if text, found := translations[key]; found {
			return text
		}
	}
	// 回退到中文
	if translations, exists := PeriodTranslations[Chinese]; exists {
		if text, found := translations[key]; found {
			return text
		}
	}
	return key // 如果都找不到，返回key本身
}

// FormatCycle 格式化周期显示
func (p *PeriodI18n) FormatCycle(cycleNumber int) string {
	prefix := p.Get("cycle_prefix")
	suffix := p.Get("cycle_suffix")

	if p.lang == English {
		return fmt.Sprintf("%s%d", prefix, cycleNumber)
	}
	return fmt.Sprintf("%s%d%s", prefix, cycleNumber, suffix)
}

// FormatRuntime 格式化运行时间显示
func (p *PeriodI18n) FormatRuntime(minutes int) string {
	label := p.Get("runtime_minutes")
	return fmt.Sprintf("%s: %d", label, minutes)
}

// FormatPositionCount 格式化持仓数显示
func (p *PeriodI18n) FormatPositionCount(count int) string {
	label := p.Get("position_count")
	return fmt.Sprintf("%s: %d", label, count)
}

// TranslateAction 翻译交易动作
func (p *PeriodI18n) TranslateAction(action string) string {
	action = strings.ToLower(action)

	// 特殊处理复合动作
	switch action {
	case "open_long":
		return p.Get("open_long")
	case "open_short":
		return p.Get("open_short")
	case "close_long":
		return p.Get("close_long")
	case "close_short":
		return p.Get("close_short")
	case "update_stop_loss":
		return p.Get("update_stop_loss")
	case "update_take_profit":
		return p.Get("update_take_profit")
	default:
		// 尝试直接翻译
		translated := p.Get(action)
		if translated != action {
			return translated
		}
		// 如果没有翻译，返回原值并首字母大写
		return strings.ToUpper(action[:1]) + action[1:]
	}
}

// GetStatusText 获取状态文本
func (p *PeriodI18n) GetStatusText(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return p.Get("cycle_running")
	case "completed":
		return p.Get("cycle_completed")
	case "waiting":
		return p.Get("cycle_waiting")
	case "error":
		return p.Get("cycle_error")
	default:
		return status
	}
}
