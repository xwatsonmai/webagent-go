package event

import (
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xwatsonmai/webagent-go/collect"
)

type EndEvent struct {
	result []collect.Data // 结束事件的结果数据
}

func NewEndEvent(result []collect.Data) EndEvent {
	return EndEvent{
		result: result,
	}
}

func (e EndEvent) Type() Type {
	return TypeEnd
}

func (e EndEvent) Content() string {
	return gconv.String(e.result)
}
