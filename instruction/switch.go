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
)

type Switch struct {
	ins    Instruction
	result []string
}

func (s *Switch) ReadyInfo() (running.EventType, string) {
	return running.Switch, fmt.Sprintf("浏览器切换标签页: %s", s.ins.Target)
}

func (s *Switch) instruction(instruction Instruction) {
	s.ins = instruction
}

func (s *Switch) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	pages := browser.Pages()
	if len(pages) < gconv.Int(s.ins.Target) {
		s.result = append(s.result, "切换失败，所选标签索引超出当前标签页数量")
	}
	index := gconv.Int(s.ins.Target) - 1 // 索引从1开始
	nowOpenPage = pages[index]
	title, _ := nowOpenPage.Title()
	s.result = append(s.result, fmt.Sprintf("浏览器已切换到标签页[%d]: %s", index+1, title))
	return nowOpenPage, nil, err, exit
}

func (s *Switch) Result() []string {
	return s.result
}
