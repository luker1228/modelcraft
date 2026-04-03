package bizutils

import "encoding/json"

// JSONParser 定义了 JSON 序列化和反序列化的接口。
// 实现此接口的类型可以用于处理 JSON 数据的编解码操作。
type JSONParser interface {
	// Marshal 将对象序列化为JSON字节数组
	Marshal(v any) ([]byte, error)

	// Unmarshal 将JSON字节数组反序列化为对象
	Unmarshal(data []byte, v any) error

	// MarshalToString 将对象序列化为JSON字符串
	MarshalToString(v any) (string, error)

	// UnmarshalFromString 将JSON字符串反序列化为对象
	UnmarshalFromString(data string, v any) error
}

// StandardJSONParser 是 JSONParser 接口的标准实现，使用 Go 标准库的 encoding/json 包。
type StandardJSONParser struct{}

// NewStandardJSONParser 创建标准库JSON解析器
func NewStandardJSONParser() JSONParser {
	return &StandardJSONParser{}
}

// Marshal 将对象序列化为JSON字节数组
func (p *StandardJSONParser) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将JSON字节数组反序列化为对象
func (p *StandardJSONParser) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MarshalToString 将对象序列化为JSON字符串
func (p *StandardJSONParser) MarshalToString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalFromString 将JSON字符串反序列化为对象
func (p *StandardJSONParser) UnmarshalFromString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

// defaultJSONParser 全局默认 JSON 解析器实例，使用标准库实现。
var defaultJSONParser JSONParser = NewStandardJSONParser()

// Marshal 使用默认解析器将对象序列化为 JSON 字节数组。
func Marshal(v any) ([]byte, error) {
	return defaultJSONParser.Marshal(v)
}

// Unmarshal 使用默认解析器将 JSON 字节数组反序列化为对象。
func Unmarshal(data []byte, v any) error {
	return defaultJSONParser.Unmarshal(data, v)
}

// MarshalToString 使用默认解析器将对象序列化为 JSON 字符串。
func MarshalToString(v any) (string, error) {
	return defaultJSONParser.MarshalToString(v)
}

// UnmarshalFromString 使用默认解析器将 JSON 字符串反序列化为对象。
func UnmarshalFromString(data string, v any) error {
	return defaultJSONParser.UnmarshalFromString(data, v)
}

// MarshalToStringIgnoreErr 使用默认解析器将对象序列化为 JSON 字符串。
// 注意：此函数故意忽略错误，适用于调试日志等非关键场景。
func MarshalToStringIgnoreErr(v any) string {
	str, _ := defaultJSONParser.MarshalToString(v)
	return str
}
