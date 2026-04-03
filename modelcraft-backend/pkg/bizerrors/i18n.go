package bizerrors

import (
	"fmt"
	"strings"
)

// Language 语言类型
const (
	LangEN = "en" // 英文
	LangZH = "zh" // 中文
)

var ConstLangArray = [2]string{"en", "zh"}

// GetMessage 获取指定语言的错误消息
func GetMessage(code ErrorDefinition, lang string) string {
	// 直接使用ErrorDefinition的GetMessage方法
	return code.GetMessageTemplate(lang)
}

// GetMessageWithParams 获取带参数的错误消息
func GetMessageWithParams(code ErrorDefinition, lang string, params []any) string {
	template := code.GetMessageTemplate(lang)

	// 参数替换，支持 {key} 格式
	message := template
	for idx, value := range params {
		placeholder := fmt.Sprintf("{%d}", idx)
		message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
	}

	return message
}

// ParseLanguage 解析语言字符串
func ParseLanguage(langStr string) string {
	switch strings.ToLower(langStr) {
	case "en", "english":
		return LangEN
	case "zh", "chinese", "zh-cn", "zh_cn", "chinese-simplified":
		return LangZH
	default:
		return LangEN // 默认返回英文
	}
}
