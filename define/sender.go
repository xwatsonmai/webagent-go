package define

import (
	"context"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/running"
)

type ISender interface {
	Send(ctx context.Context, event event.IEvent) error    // 发送事件
	SendMessage(ctx context.Context, message string) error // 发送消息
	//SendTakeOver(ctx context.Context, message string) error // 发送接管事件
	SendEnd(ctx context.Context, result []collect.Data) error
	SendRunning(ctx context.Context, step int, eventName running.EventType, status running.EventStaus, info string) error
	SendError(ctx context.Context, err error) error
}
