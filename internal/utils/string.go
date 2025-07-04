package utils

import (
	"fmt"
	"strings"
	"time"
)

// BuildPrompt 构建提示词，替换模板中的变量
func BuildPrompt(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// FormatMessage 格式化消息
func FormatMessage(name, text string, timestamp time.Time) string {
	timeStr := timestamp.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s %s: %s", timeStr, name, text)
}

// PickRandom 从字符串切片中随机选择一个
func PickRandom(items []string) string {
	if len(items) == 0 {
		return ""
	}
	// 简单的时间基础随机
	index := int(time.Now().UnixNano()) % len(items)
	return items[index]
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// DefaultString 如果字符串为空则返回默认值
func DefaultString(s, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
} 