package event

import "encoding/json"

type IEvent interface {
	Type() Type
	Content() string
}

type Type string

const (
	TypeMessage Type = "message" // 消息事件
	TypeDebug   Type = "debug"   // 调试事件
	TypeEnd     Type = "end"     // 结束事件
	TypeRunning Type = "running" // 运行事件
	TypeQrCode  Type = "qrcode"  // 二维码事件
	TypeError   Type = "error"   // 错误事件
	TypeInit    Type = "init"    // 初始化事件
)

func ToJson(event IEvent) string {
	// 这里可以使用json.Marshal(event)将event转换为JSON字符串
	// 但为了简化示例，这里直接返回一个占位符字符串
	data := map[string]any{
		"type":    event.Type(),
		"content": event.Content(),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}" // 如果转换失败，返回空的JSON对象
	}
	return string(jsonData)
}
