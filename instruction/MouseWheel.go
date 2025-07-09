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

type MouseWheel struct {
	//// 方向
	//Direction string `json:"direction"` // up or down
	// 滚动的距离
	Distance float64 `json:"distance"` // 滚动的距离，单位为像素
}

func (m *MouseWheel) ReadyInfo() (running.EventType, string) {
	info := ""
	if m.Distance > 0 {
		info = "向下滚动" + gconv.String(m.Distance) + "像素"
	} else {
		info = "向上滚动" + gconv.String(-m.Distance) + "像素"
	}
	return running.MouseWheel, info
}

func (m *MouseWheel) instruction(instruction Instruction) {
	m.Distance = gconv.Float64(instruction.Target)
}

func (m *MouseWheel) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	if nowOpenPage == nil {
		return nil, nil, nil, false
	}

	// 执行鼠标滚轮操作
	err = nowOpenPage.Mouse().Wheel(0, m.Distance)
	if err != nil {
		return nil, nil, err, false
	}

	// 等待页面加载
	_ = nowOpenPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateDomcontentloaded,
		Timeout: playwright.Float(10000), // 设置超时时间为10秒
	})

	return nowOpenPage, nil, nil, false
}

func (m *MouseWheel) Result() []string {
	return []string{
		"鼠标滚轮操作完成",
	}
}
