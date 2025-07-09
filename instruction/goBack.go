package instruction

import (
	"context"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type GoBack struct {
	err bool
}

func (g *GoBack) ReadyInfo() (running.EventType, string) {
	return running.GoBack, "浏览器后退"
}

func (g *GoBack) instruction(instruction Instruction) {
	return
}

func (g *GoBack) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	defer func() {
		if err != nil {
			g.err = true
		}
	}()
	if _, err := nowOpenPage.GoBack(); err != nil {
		return nowOpenPage, nil, errors.Wrap(err, "浏览器后退失败"), exit
	}
	title, _ := nowOpenPage.Title()
	runInfoChan <- "浏览器已后退到页面: " + title
	return nowOpenPage, nil, nil, exit
}

func (g *GoBack) Result() []string {
	if g.err {
		return []string{"浏览器后退失败"}
	}
	return []string{"浏览器已后退到上一个页面"}
}
