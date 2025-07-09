package instruction

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type Fill struct {
	ins Instruction
	err error
}

func (f *Fill) instruction(instruction Instruction) {
	f.ins = instruction
}

func (f *Fill) Result() []string {
	if f.err != nil {
		return []string{f.err.Error()}
	}
	return []string{fmt.Sprintf("浏览器已填充表单[%s]:%s", f.ins.Target, f.ins.Content)}
}

func (f *Fill) ReadyInfo() (running.EventType, string) {
	return running.Filling, fmt.Sprintf("浏览器填充表单: %s", f.ins.Target)
}

func (f *Fill) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	defer func() {
		if err != nil {
			f.err = err
		}
	}()
	if err := nowOpenPage.Fill(f.ins.Target, gconv.String(f.ins.Content), playwright.PageFillOptions{
		Timeout: playwright.Float(5000), // 设置超时时间为 5 秒
	}); err != nil {
		//return nil, errors.Wrapf(err, "浏览器填充表单失败: %s", ins.Target)
		return nowOpenPage, nil, errors.New(fmt.Sprintf("浏览器填充[%s]失败：%s", f.ins.Target, err.Error())), exit
	}
	return nowOpenPage, nil, err, exit
}
