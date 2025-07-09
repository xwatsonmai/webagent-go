package v1

import (
	"context"
	"fmt"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/running"
)

type TestSender struct {
}

func (t TestSender) Send(ctx context.Context, event event.IEvent) error {
	fmt.Println("TestSender Send called with event:", event)
	return nil
}

func (t TestSender) SendMessage(ctx context.Context, message string) error {
	fmt.Println("TestSender SendMessage called with message:", message)
	return nil
}

func (t TestSender) SendEnd(ctx context.Context, result []collect.Data) error {
	fmt.Println("TestSender SendEnd called with result:", result)
	return nil
}

func (t TestSender) SendRunning(ctx context.Context, step int, eventName running.EventType, status running.EventStaus, info string) error {
	//fmt.Println("TestSender SendRunning called with step:", step, "eventName:", eventName, "status:", status, "info:", info)
	return nil
}

func (t TestSender) SendError(ctx context.Context, err error) error {
	fmt.Println("TestSender SendError called with error:", err)
	return nil
}
