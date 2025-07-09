package instruction

import (
	"context"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/htmlHandler"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type PageCollect struct {
	ins Instruction // 收集指令
}

func (p *PageCollect) ReadyInfo() (running.EventType, string) {
	return running.Collect, "收集页面数据"
}

func (p *PageCollect) instruction(instruction Instruction) {
	p.ins = instruction
}

func (p *PageCollect) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	page := nowOpenPage
	content, err := page.Content()
	if err != nil {
		return nowOpenPage, nil, errors.Wrap(err, "收集失败"), exit
	}
	md, err := htmlHandler.ToMarkdown(content)
	if err != nil {
		return nowOpenPage, nil, errors.Wrap(err, "转换为Markdown失败"), exit
	}
	collectData = append(collectData, collect.Data{
		Title:   p.ins.Target,
		URL:     nowOpenPage.URL(),
		Content: md,
	})
	return nowOpenPage, collectData, nil, exit
}

func (p *PageCollect) Result() []string {
	return []string{p.ins.Target + " 页面数据已收集"}
}
