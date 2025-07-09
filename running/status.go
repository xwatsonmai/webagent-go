package running

import (
	"encoding/json"
	"github.com/xwatsonmai/webagent-go/event"
)

type EventType string

const (
	Thinking           EventType = "thinking"            // 正在思考中
	Click              EventType = "click"               // 点击事件
	Filling            EventType = "filling"             // 填充事件
	Waiting            EventType = "waiting"             // 等待事件
	Goto               EventType = "goto"                // 跳转事件
	NewPage            EventType = "new_page"            // 新页面事件
	Collect            EventType = "collect"             // 收集数据事件
	Switch             EventType = "switch"              // 切换标签事件
	Fail               EventType = "fail"                // 失败事件
	QrCode             EventType = "qrcode"              // 二维码事件
	End                EventType = "end"                 // 结束
	GoBack             EventType = "goback"              // 返回上一步
	CloseLabel         EventType = "close_label"         // 关闭标签页事件
	Keyboard           EventType = "keyboard"            // 键盘输入事件
	ScreenshotIdentify EventType = "screenshot_identify" // 截图识别事件
	Experience         EventType = "experience"          // 管理使用经验事件
	MouseWheel         EventType = "mouse_wheel"         // 鼠标滚轮事件
)

type EventStaus string

const (
	StatusRunning EventStaus = "running" // 运行中
	StatusSuccess EventStaus = "success" // 成功
	StatusFailed  EventStaus = "failed"  // 失败
)

type Event struct {
	step   int       // 当前轮次
	event  EventType // 事件名称
	status EventStaus
	info   string
}

func NewRunningEvent(step int, event EventType, status EventStaus, info string) Event {
	return Event{
		step:   step,
		event:  event,
		status: status,
		info:   info,
	}
}

func (r Event) Type() event.Type {
	return event.TypeRunning
}

func (r Event) Content() string {
	data := map[string]any{
		"step":   r.step,
		"event":  r.event,
		"status": r.status,
		"info":   r.info,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
