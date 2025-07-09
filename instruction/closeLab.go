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

type CloseLabel struct {
	ins    Instruction // 关闭实验室指令
	err    bool        // 是否发生错误
	result []string    // 结果信息
}

func (c *CloseLabel) ReadyInfo() (running.EventType, string) {
	return running.CloseLabel, "关闭标签页"
}

func (c *CloseLabel) instruction(instruction Instruction) {
	c.ins = instruction
}

func (c *CloseLabel) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	defer func() {
		if err != nil {
			c.err = true
		}
	}()
	if c.ins.Target == "" {
		// 如果没有指定标签页，则关闭当前标签页
		if err := nowOpenPage.Close(); err != nil {
			return nowOpenPage, nil, errors.Wrap(err, "关闭当前标签页失败"), exit
		}
		runInfoChan <- "当前标签页已关闭"
		c.result = append(c.result, "当前标签页已关闭")
		// 查看浏览器中是否还有其他标签页，如果有的话，返回第一个标签页
		pages := browser.Pages()
		if len(pages) > 0 {
			openPage = pages[0]
			title, _ := openPage.Title()
			c.result = append(c.result, fmt.Sprintf("浏览器已切换至标签页【%d】: %s", 1, title))
		}
		return openPage, nil, nil, exit
	}
	// 如果指定了标签页，则关闭指定的标签页
	index := gconv.Int(c.ins.Target) - 1 // 索引从1开始
	pages := browser.Pages()
	if index < 0 || index >= len(pages) {
		return nowOpenPage, nil, errors.New("关闭标签页失败，所选标签索引超出当前标签页数量"), exit
	}
	if err := pages[index].Close(); err != nil {
		return nowOpenPage, nil, errors.Wrap(err, "关闭指定标签页失败"), exit
	}
	runInfoChan <- fmt.Sprintf("标签页【%d】已关闭", index+1)
	c.result = append(c.result, fmt.Sprintf("标签页【%d】已关闭", index+1))
	return nowOpenPage, nil, nil, exit
}

func (c *CloseLabel) Result() []string {
	if c.err {
		return []string{"关闭标签页失败"}
	}
	return c.result
}
