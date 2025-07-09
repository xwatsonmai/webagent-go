package instruction

import (
	"context"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type Fail struct {
	ins Instruction
}

func (f *Fail) ReadyInfo() (running.EventType, string) {
	return running.Fail, ""
}

func (f *Fail) instruction(instruction Instruction) {
	f.ins = instruction
}

func (f *Fail) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	return nowOpenPage, nil, errors.New("执行失败：" + gconv.String(f.ins.Content)), true
}

func (f *Fail) Result() []string {
	return nil
}
