package instruction

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
	"time"
)

type Wait struct {
	ins Instruction
}

func (w *Wait) ReadyInfo() (running.EventType, string) {
	return running.Waiting, fmt.Sprintf("浏览器等待: %s", w.ins.Content)
}

func (w *Wait) instruction(instruction Instruction) {
	w.ins = instruction
}

func (w *Wait) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	time.Sleep(gconv.Duration(w.ins.Content) * time.Second) // 等待指定的时间
	return nowOpenPage, nil, nil, exit
}

func (w *Wait) Result() []string {
	return []string{fmt.Sprintf("等待了 %s 秒", w.ins.Content)}
}
