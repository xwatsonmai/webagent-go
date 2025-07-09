package instruction

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type Goto struct {
	err error
	ins Instruction
}

func (g *Goto) instruction(instruction Instruction) {
	g.ins = instruction
}

func (g *Goto) Result() []string {
	if g.err != nil {
		return []string{fmt.Sprintf("跳转失败: 【%s】", g.err.Error())}
	}
	return []string{fmt.Sprintf("浏览器已跳转到: %s", g.ins.Target)}
}

func (g *Goto) ReadyInfo() (running.EventType, string) {
	return running.Goto, fmt.Sprintf("浏览器跳转到: %s", g.ins.Target)
}

func (g *Goto) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	defer func() {
		if err != nil {
			g.err = err
		}
	}()
	pages := browser.Pages()
	var page playwright.Page
	if len(pages) == 0 {
		page, _ = browser.NewPage()
	} else {
		page = pages[0]
	}
	if _, err = page.Goto(g.ins.Target); err != nil {
		return nowOpenPage, nil, errors.New("打开页面失败: " + err.Error()), exit
	}
	//if err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
	//	State:   playwright.LoadStateNetworkidle,
	//	Timeout: playwright.Float(10000), // 10 seconds
	//}); err != nil {
	//	return nowOpenPage, nil, errors.New("等待页面加载失败: " + err.Error()), exit
	//}
	openPage = page
	return openPage, nil, nil, exit
}
