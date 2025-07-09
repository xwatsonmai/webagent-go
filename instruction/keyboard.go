package instruction

import (
	"context"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type Keyboard struct {
	ins Instruction
}

func (k *Keyboard) ReadyInfo() (running.EventType, string) {
	return running.Keyboard, "浏览器键盘输入"
}

func (k *Keyboard) instruction(instruction Instruction) {
	k.ins = instruction
}

func (k *Keyboard) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	if err := nowOpenPage.Keyboard().Press(k.ins.Target, playwright.KeyboardPressOptions{
		Delay: playwright.Float(100), // 设置按键延迟为 100 毫秒
	}); err != nil {
		return nowOpenPage, nil, err, exit
	}
	runInfoChan <- fmt.Sprintf("浏览器键盘输入: %s", k.ins.Target)
	return nowOpenPage, nil, nil, exit
}

func (k *Keyboard) Result() []string {
	return []string{fmt.Sprintf("浏览器已键盘输入: %s", k.ins.Target)}
}
