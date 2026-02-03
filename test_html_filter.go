package main

import (
	"fmt"
	"regexp"
	"strings"
)

// cleanAndDecodeContent 清理和解码AI输出内容，处理HTML实体编码和其他特殊字符
func cleanAndDecodeContent(content string) string {
	// 移除多余的空白字符
	cleaned := strings.TrimSpace(content)

	// 处理常见的HTML实体编码
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&#x27;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&#x2F;", "/")

	// 使用更高级的HTML清理方法，针对不同类型的标签进行不同的处理
	// 首先处理块级标签，用换行符替换它们
	cleaned = regexp.MustCompile(`</?(p|div|section|article|h[1-6]|ul|ol|li|tr|td|th)[^>]*>`).ReplaceAllString(cleaned, "\n")
	// 处理行内标签（用空格替换，确保文本间有适当间距）
	cleaned = regexp.MustCompile(`</?(span|strong|b|em|i|u)[^>]*>`).ReplaceAllString(cleaned, " ")
	// 处理换行相关标签
	cleaned = regexp.MustCompile(`<(/?)(br|hr)[^>]*>`).ReplaceAllString(cleaned, "\n")
	// 处理表格和列表相关标签
	cleaned = regexp.MustCompile(`</?(dl|dt|dd)[^>]*>`).ReplaceAllString(cleaned, "\n")
	// 移除所有剩余的HTML标签
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned = tagRegex.ReplaceAllString(cleaned, "")

	// 替换连续的空白字符为单个空格
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	// 替换多个连续换行符为两个换行符（保持段落结构）
	for strings.Contains(cleaned, "\n\n\n") {
		cleaned = strings.ReplaceAll(cleaned, "\n\n\n", "\n\n")
	}

	// 移除多余的空格和换行
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

func main() {
	// 测试您提供的HTML内容
	htmlContent := `<div class="ds-markdown" style="--ds-md-zoom: 1.143;"><p class="ds-markdown-paragraph"><span>这个问题看起来简单，但可以从多个层面来理解：</span></p><p class="ds-markdown-paragraph"><strong><span>数学层面</span></strong><br><span>在十进制算术中：</span><br><strong><span>1 + 1 = 2</span></strong></p><p class="ds-markdown-paragraph"><strong><span>其他可能含义</span></strong></p><ul><li><p class="ds-markdown-paragraph"><span>二进制中：1 + 1 = 10（读作"一零"）</span></p></li><li><p class="ds-markdown-paragraph"><span>在布尔逻辑中：1 + 1 = 1（这里的"+"表示逻辑或）</span></p></li><li><p class="ds-markdown-paragraph"><span>在某些脑筋急转弯中：1 + 1 = "王"或"田"（汉字角度）</span></p></li><li><p class="ds-markdown-paragraph"><span>有时也被用来讨论哥德巴赫猜想（1+1问题）——这里的"1+1"是指"一个质数加一个质数"的简写。</span></p></li></ul><p class="ds-markdown-paragraph"><strong><span>所以最常见的答案就是：2</span></strong><span> ✅</span></p></div>`

	fmt.Println("原始HTML内容:")
	fmt.Println(htmlContent)
	fmt.Println("\n清理后的内容:")
	cleanedContent := cleanAndDecodeContent(htmlContent)
	fmt.Println(cleanedContent)
}
