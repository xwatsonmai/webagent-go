package instruction

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collec
)

type Click struct {
	ins    Instruction
	err    error
	result []string
}

func (c *Click) ReadyInfo() (running.EventType, string) {
	return running.Click, fmt.Sprintf("浏览器点击元素: %s", c.ins.Target)
}

func (c *Click) instruction(instruction Instruction) {
	c.ins = instruction
}

func (c *Click) Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool) {
	initialPages := len(browser.Pages())
	page := nowOpenPage
	var button playwright.Locator
	if c.ins.Content != "" {
		button = page.Locator(c.ins.Target, playwright.PageLocatorOptions{
			HasText: c.ins.Content,
		}).First()
	} else {
		button = page.Locator(c.ins.Target).First()
	}
	buttonText := ""
	// 获取button的文本
	if buttonText, err = button.TextContent(); err != nil || buttonText == "" {
		buttonText = c.ins.Target
	}
	runInfoChan <- fmt.Sprintf("浏览器点击元素: %s", buttonText)
	if err := button.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(1000), // 设置超时时间为 10 秒
	}); err != nil {
		if errors.Is(err, playwright.ErrTimeout) {
			c.result = append(c.result, fmt.Sprintf("浏览器点击元素[%s][%s]失败：无法选中该目标", c.ins.Target, buttonText))
			return nowOpenPage, nil, errors.New(fmt.Sprintf("浏览器点击[%s][%s]失败：无法选中该目标", c.ins.Target, buttonText)), exit
		}
		errInfo := err.Error()
		// 如果errInfo字数超过200个字符，截断
		if rune(len(errInfo)) > 200 {
			errInfo = string([]rune(errInfo)[:200]) + "..."
		}
		c.result = append(c.result, fmt.Sprintf("浏览器点击元素[%s]失败：%s", c.ins.Target, errInfo))
		return nowOpenPage, nil, errors.New(fmt.Sprintf("浏览器点击[%s]失败：%s", c.ins.Target, errInfo)), exit
	}
	page.WaitForTimeout(1000) // 等待 2 秒
	pages := browser.Pages()
	if len(pages) > initialPages {
		nowOpenPage = pages[len(pages)-1] // 获取最新打开的页面
		c.result = append(c.result, fmt.Sprintf("浏览器已打开新标签页: %s", page.URL()))
	}
	c.result = append(c.result, fmt.Sprintf("已点击元素[%s]: %s", c.ins.Target, buttonText))
	return nowOpenPage, nil, nil, exit
}

func (c *Click) Result() []string {
	return c.result
}
