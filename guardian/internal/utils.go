package guardian

// TruncateString 截断字符串显示，只显示开头和结尾部分
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	// 计算头尾各取多少字符
	headLen := maxLength / 2
	tailLen := maxLength - headLen - 3 // 减去省略号的长度

	if tailLen < 0 {
		return s[:maxLength]
	}

	head := s[:headLen]
	tail := s[len(s)-tailLen:]

	return head + "..." + tail
}
