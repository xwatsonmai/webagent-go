package instruction

import (
	"context"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type DataCollect struct {
	ins Instruction
}

func (d *DataCollect) ReadyInfo() (running.EventType, string) {
	return running.Collect, "浏览器数据采集"
}

func (d *DataCollect) instruction(instruction Instruction) {
	d.ins = instruction
}

func (d *DataCollect) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	result := make([]collect.Data, 0)
	if d.ins.Content != "" {
		data := collect.Data{
			Title:   "",
			Content: gconv.String(d.ins.Content),
		}
		result = append(result, data)
	}
	return nowOpenPage, result, nil, exit
}

func (d *DataCollect) Result() []string {
	return []string{"数据采集指令已执行"}
}
