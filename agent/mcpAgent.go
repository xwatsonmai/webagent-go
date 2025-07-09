package agent

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/safeGoroutine/goroutine"
	"github.com/xwatsonmai/webagent-go/aimodel"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/define"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/htmlHandler"
	"github.com/xwatsonmai/webagent-go/instruction"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
	"io"
	"log"
	"slices"
	"strings"
	"time"
)

type instructionHandler struct {
	Type instruction.Type // 指令类型
	Err  error            // 错误信息
}

type Agent struct {
	AgentId        string `json:"agent_id"`
	agentType      model.AgentType
	sender         define.ISender
	prompter       define.IPrompter // 提示词处理器
	aiModel        aimodel.IAiModel // AI模型接口
	agentFlowChat  bool             // Agent是否流式调用
	browser        playwright.BrowserContext
	nowOpenPage    playwright.Page // 当前打开的页面
	pageSliceIndex int             // 当前打开的页面索引，从0开始
	collect        []collect.Data
	lastThinkTime  int64                // 上次思考时间戳，用于控制思考间隔
	chrome         playwright.Browser   // 浏览器实例
	handlerFlow    []instructionHandler // 指令处理器列表，按顺序处理指令
	getIMapFunc    instruction.GetInstructionMapFunc
	userIntention  string // 用户意图，表示用户想要做什么
}

func NewAgent(ctx context.Context, agentType model.AgentType, getIMapFunc instruction.GetInstructionMapFunc, sender define.ISender, prompter define.IPrompter, aiModel aimodel.IAiModel, openHeadLess, agentFlowChat bool) (*Agent, error) {
	pw, err := playwright.Run()
	if err != nil {
		log.Printf("Agent[%s]启动Playwright失败: %v", uuid.New().String(), err)
		return nil, errors.Wrap(err, "启动Playwright失败")
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Args: []string{
			"--disable-blink-features=AutomationControlled", // 禁用自动化标识
		},
		Headless: playwright.Bool(!openHeadLess), // 设置为非无头模式
	})
	if err != nil {
		log.Printf("Agent[%s]启动浏览器失败: %v", uuid.New().String(), err)
		return nil, errors.Wrap(err, "启动浏览器失败")
	}

	c, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"), // 设置用户代理
	})
	if err != nil {
		log.Printf("Agent[%s]创建浏览器上下文失败: %v", uuid.New().String(), err)
		return nil, errors.Wrap(err, "创建浏览器上下文失败")
	}
	c.SetDefaultTimeout(5 * 1000) // 设置默认超时时间为5秒

	AgentId := uuid.New().String()
	sender.Send(ctx, event.NewInitEvent(AgentId))
	a := &Agent{
		AgentId:       AgentId,
		sender:        sender,
		prompter:      prompter,
		aiModel:       aiModel,
		agentFlowChat: agentFlowChat,
		collect:       []collect.Data{},
		browser:       c,
		agentType:     agentType,
		chrome:        browser,
		getIMapFunc:   getIMapFunc,
	}
	return a, nil
}

func (a *Agent) Do(ctx context.Context, userIntention string, targetUrl string) playwright.BrowserContext {
	a.userIntention = userIntention
	page, err := a.browser.NewPage()
	if err != nil {
		log.Printf("Agent[%s]创建浏览器页面失败: %v", uuid.New().String(), err)
		//return
	}
	page.Goto(targetUrl) // 打开小红书首页
	// 等待页面加载完成
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateDomcontentloaded,
		Timeout: playwright.Float(10 * 1000), // 设置超时时间为30秒
	}); err != nil {
		log.Printf("Agent[%s]等待页面加载失败: %v", a.AgentId, err)
		a.sender.SendError(ctx, errors.Wrapf(err, "Agent[%s]等待页面加载失败", a.AgentId))
		return a.browser
	}
	a.nowOpenPage = page // 设置当前打开的页面
	topDomain, _ := htmlHandler.GetTopLevelDomain(page)
	if topDomain == "" {
		topDomain = targetUrl
	}
	a.agentHandler(ctx, userIntention, targetUrl)
	return a.browser
}

func (a *Agent) agentHandler(ctx context.Context, userIntention string, targetUrl string) {
	//log.Printf("Agent[%s]开始处理对话列表: %s", a.AgentId, message)
	var instructionList []instruction.Instruction
	aiChat := aimodel.ChatList{}
	systemPrompt, _ := a.prompter.SystemPrompt(ctx, userIntention, targetUrl)
	startUserPrompt := a.prompter.StartUserPrompt(ctx, userIntention, targetUrl)
	aiChat = aimodel.ChatList{
		aimodel.Chat{
			Role:    aimodel.EAIChatRoleSystem,
			Content: systemPrompt,
		},
		aimodel.Chat{
			Role: aimodel.EAIChatRoleUser,
			Content: []aimodel.UserContent{
				{
					Type: "text",
					Text: startUserPrompt,
				},
			},
		},
	}
	//a.sender.SendDebug(ctx, event.NewAgentInitDebugEvent(g.Map{
	//	"system":          systemPrompt,
	//	"startUserPrompt": startUserPrompt,
	//}))
	// 开始调用模型，模型会返回指令，获取指令执行：
	// 如果是查询指令，则调用mcpService查询，并把结果设置进背景知识后，把结果以user角色发送给模型。
	// 如果是话术发送/接管指令，则把话术通过sender发送出去。
	// 重复以上的步骤，直到模型返回结束指令。
	// 注意：接管指令总是代表着结束本轮执行，它与结束指令相同的是表示结束本轮执行，但不同的是，结束指令表示的是本轮执行正常结束，接管指令表示的是本轮执行异常结束，执行中止
	step := 0
	for {
		step++
		var thisRoundUserChatContent []aimodel.UserContent
		//log.Printf("Agent[%s]当前对话内容: %s", a.AgentId, aiChat.ToString())
		a.lastThinkTime = time.Now().Unix()
		a.sender.SendRunning(ctx, step, running.Thinking, running.StatusRunning, "Agent正在思考...")
		//if len(aiChat) > 10 {
		//	// 保留第0,1条，与最后一条对话内容，其余的对话内容都删除
		//	aiChat = append(aimodel.ChatList{aiChat[0], aiChat[1]}, aiChat[len(aiChat)-1])
		//}

		toAiChat := append(aimodel.ChatList{}, aiChat...) // 深拷贝一份当前对话内容
		latestUserChat := toAiChat[len(toAiChat)-1].Content.([]aimodel.UserContent)

		latestUserChat = append(latestUserChat, aimodel.UserContent{
			Type: "text",
			Text: fmt.Sprintf("当前浏览器打开的标签页列表:\n%s", a.getPagesInfo()),
		})
		latestUserChat = append(latestUserChat, aimodel.UserContent{
			Type: "text",
			Text: fmt.Sprintf("当前正在浏览的页面:\n%s", a.getNowOpenPageInfo()),
		})
		//if dialogHtml, has := a.CheckModelDialog(); has {
		//	latestUserChat = append(latestUserChat, aimodel.UserContent{
		//		Type: "text",
		//		Text: fmt.Sprintf("当前页面有高zindex浮层元素:\n%s", dialogHtml),
		//	})
		//}

		// 检查执行流水，如果连续3次操作失败
		//if len(a.handlerFlow) >= 3 {
		//	lastThree := a.handlerFlow[len(a.handlerFlow)-3:]
		//	if lastThree[0].Err != nil && lastThree[1].Err != nil && lastThree[2].Err != nil {
		//		identify := a.ErrIdentify(ctx, []string{lastThree[0].Type.ToString(), lastThree[1].Type.ToString(), lastThree[2].Type.ToString()})
		//		if identify != "" {
		//			latestUserChat = append(latestUserChat, aimodel.UserContent{
		//				Type: "text",
		//				Text: fmt.Sprintf("Agent检测到连续3次操作失败，可能是因为：%s。", identify),
		//			})
		//			//log.Printf("Agent[%s]检测到连续3次操作失败，可能是因为：%s，请检查后重试。", a.AgentId, identify)
		//		} else {
		//			latestUserChat = append(latestUserChat, aimodel.UserContent{
		//				Type: "text",
		//				Text: "连续3次操作失败，请尝试重新分析页面、求求助AI识图专家，或者检查页面是否有遮罩层、模态框等元素遮挡。",
		//			})
		//			//log.Printf("Agent[%s]检测到连续3次操作失败，但无法识别原因，请检查后重试。", a.AgentId)
		//		}
		//	}
		//}

		toAiChat[len(toAiChat)-1].Content = latestUserChat // 更新最新的用户对话内容

		// 调用AI模型
		res, err := a.aiChat(ctx, toAiChat)
		if err != nil {
			a.sender.SendRunning(ctx, step, running.Thinking, running.StatusFailed, "Agent调用AI模型失败："+err.Error())
			log.Printf("Agent[%s]调用AI模型失败: %v", a.AgentId, err)
			//a.sender.SendRunning(ctx, step, event.RunningThinking, event.RunningStatusFailed, "")
			a.sender.SendError(ctx, errors.Wrapf(err, "Agent[%s]调用AI模型失败", a.AgentId))
			// 处理错误
			return
		}
		a.sender.SendRunning(ctx, step, running.Thinking, running.StatusSuccess, "Agent调用AI模型成功")
		aiChat = append(aiChat, aimodel.Chat{
			Role:    aimodel.EAIChatRoleAssistant,
			Content: res,
		})
		log.Printf("Agent[%s]调用AI模型成功: %s", a.AgentId, res)
		// 把Agent的回答转为Answer类型
		instructionList, err = a.prompter.AgentAnswer(ctx, res)
		if err != nil {
			//a.sender.SendRunning(ctx, step, event.RunningThinking, event.RunningStatusFailed, "")
			a.sender.SendError(ctx, errors.Wrapf(err, "Agent[%s]解析Agent回答失败：%s", a.AgentId, res))
			log.Printf("Agent[%s]解析Agent回答失败: %v", a.AgentId, err)
			// 处理错误
			return
		}
		//var err error
		//a.sender.SendRunning(ctx, step, event.RunningThinking, event.RunningStatusSuccess, "")
		for _, i := range instructionList {
			log.Printf("Agent[%s]处理指令: %s", a.AgentId, i)
			thisRoundUserChatContent, err = a.agentInstructionHandler(ctx, step, i, thisRoundUserChatContent)
			if errors.Is(err, io.EOF) {
				//a.sender.SendRunning(ctx, step, running.End, running.StatusRunning, "正在获取数据结果...")
				a.sender.SendEnd(ctx, a.GetResult(ctx))
				//a.sender.SendRunning(ctx, step, running.GetResult, running.StatusSuccess, "数据结果获取成功")
				// 结束本轮执行
				return
			}
			if err != nil {
				log.Printf("Agent[%s]处理指令失败: %v", a.AgentId, err)
				//a.sender.SendEnd(ctx)
				a.sender.SendError(ctx, errors.Wrapf(err, "Agent[%s]处理指令失败：%s", a.AgentId, i.Type.ToString()))
				//return
			}
		}
		instructionList = []instruction.Instruction{}

		thisRoundUserChat := aimodel.Chat{
			Role:    aimodel.EAIChatRoleUser,
			Content: thisRoundUserChatContent,
		}
		aiChat = append(aiChat, thisRoundUserChat)

	}
}

func (a *Agent) agentInstructionHandler(ctx context.Context, step int, ins instruction.Instruction, thisRoundUserChat []aimodel.UserContent) (newUserContent []aimodel.UserContent, err error) {
	if ins.IsEnd() {
		log.Printf("Agent[%s]收到结束指令: %s", a.AgentId, gconv.String(ins))
		return thisRoundUserChat, io.EOF
	}

	if ins.Type == instruction.HtmlSliceSelect {
		h, _ := a.nowOpenPage.Content()
		h, _ = htmlHandler.CleanHTML(strings.NewReader(h))
		htmlSlice, count := htmlHandler.Slice(h)
		if gconv.Int(ins.Target) >= len(htmlSlice) {
			log.Printf("Agent[%s]html切片索引超出范围(索引从0开始计数): %d, 总切片数: %d", a.AgentId, gconv.Int(ins.Target), count)
			thisRoundUserChat = append(thisRoundUserChat, aimodel.UserContent{
				Type: "text",
				Text: fmt.Sprintf("html切片索引超出范围(索引从0开始计数): %d, 总切片数: %d", gconv.Int(ins.Target), len(htmlSlice)),
			})
			return thisRoundUserChat, nil
		}
		a.pageSliceIndex = gconv.Int(ins.Target)
		log.Printf("Agent[%s]设置当前html切片索引为: %d, 总切片数: %d", a.AgentId, a.pageSliceIndex, count)
		thisRoundUserChat = append(thisRoundUserChat, aimodel.UserContent{
			Type: "text",
			Text: fmt.Sprintf("已设置当前html切片索引为: %d, 总切片数: %d", a.pageSliceIndex, count),
		})
		return thisRoundUserChat, nil
	}

	insHandler := instruction.Builder(a.agentType, a.getIMapFunc, ins)
	if insHandler == nil {
		log.Printf("Agent[%s]无法处理指令: %s", a.AgentId, gconv.String(ins))
		return thisRoundUserChat, errors.New("无法处理的指令类型: " + ins.Type.ToString())
	}
	sendChan := make(chan event.IEvent, 10)
	runInfoChan := make(chan string, 10)
	log.Printf("Agent[%s]处理指令: %s", a.AgentId, gconv.String(ins))
	runType, info := insHandler.ReadyInfo()
	a.sender.SendRunning(ctx, step, runType, running.StatusRunning, info)
	var insHandlerErr error
	goroutine.Go(func() {
		defer func() {
			close(sendChan)
			close(runInfoChan)
		}()
		var collectData []collect.Data
		var needExit = false
		a.nowOpenPage, collectData, insHandlerErr, needExit = insHandler.Handler(ctx, model.AgentInfo{
			AgentId:   a.AgentId,
			ThinkTime: 0,
		}, a.browser, a.nowOpenPage, sendChan, runInfoChan)
		if len(collectData) > 0 {
			a.collect = append(a.collect, collectData...)
			log.Printf("Agent[%s]收集到数据: %s", a.AgentId, gconv.String(collectData))
		}
		if needExit {
			log.Printf("Agent[%s]指令处理需要退出: %s", a.AgentId, gconv.String(ins))
			insHandlerErr = io.EOF // 结束本轮执行
		}
	}, goroutine.WithPanicHandler(func(err error) {
		g.Log().Errorf(ctx, "Agent[%s]指令处理发生异常: %v", a.AgentId, err)
		insHandlerErr = errors.Wrap(err, "指令处理发生异常")
	}))
L:
	for {
		select {
		case e, ok := <-sendChan:
			if !ok {
				log.Printf("Agent[%s]指令处理事件通道已关闭", a.AgentId)
				break L
			}
			a.sender.Send(ctx, e)
		case info, ok := <-runInfoChan:
			if !ok {
				log.Printf("Agent[%s]指令处理运行信息通道已关闭", a.AgentId)
				break L
			}
			if info != "" {
				log.Printf("Agent[%s]指令处理运行信息: %s", a.AgentId, info)
				a.sender.SendRunning(ctx, step, runType, running.StatusRunning, info)
			}
		case <-ctx.Done():
			log.Printf("Agent[%s]指令处理被取消: %v", a.AgentId, ctx.Err())
			break L
		}
	}

	if insHandlerErr == io.EOF {
		return thisRoundUserChat, io.EOF
	}

	if insHandlerErr != nil {
		a.sender.SendRunning(ctx, step, runType, running.StatusFailed, insHandlerErr.Error())
	} else {
		a.sender.SendRunning(ctx, step, runType, running.StatusSuccess, "")
	}
	if ins.Type != instruction.TypeWait {
		a.handlerFlow = append(a.handlerFlow, instructionHandler{
			Type: ins.Type,
			Err:  insHandlerErr,
		})
	}

	for _, resultInfo := range insHandler.Result() {
		thisRoundUserChat = append(thisRoundUserChat, aimodel.UserContent{
			Type: "text",
			Text: resultInfo,
		})
	}

	if ins.Type == instruction.TypeCollect {
		thisRoundUserChat = append(thisRoundUserChat, aimodel.UserContent{
			Type: "text",
			Text: fmt.Sprintf("已收集到共计【%d】条数据", len(a.collect)),
		})
	}
	a.pageSliceIndex = 0
	return thisRoundUserChat, nil
}

func (a *Agent) getPagesInfo() string {
	pages := a.browser.Pages()
	if len(pages) == 0 {
		return "没有打开的标签页"
	}
	info := ""
	for i, page := range pages {
		title, _ := page.Title()
		url := page.URL()
		info += fmt.Sprintf("标签页 【%d】: %s - %s\n", i+1, title, url)
	}
	return info
}

func (a *Agent) getNowOpenPageInfo() string {
	if a.nowOpenPage == nil {
		if len(a.browser.Pages()) > 0 {
			a.nowOpenPage = a.browser.Pages()[0] // 默认使用第一个页面
		} else {
			return "没有当前打开的页面"
		}
	}
	title, _ := a.nowOpenPage.Title()
	url := a.nowOpenPage.URL()
	h, _ := a.nowOpenPage.Content()
	h, _ = htmlHandler.CleanHTML(strings.NewReader(h))
	htmlSlice, count := htmlHandler.Slice(h)
	h = htmlSlice[a.pageSliceIndex]
	return fmt.Sprintf("当前打开的页面:\n标题: %s\nURL: %s\n内容: %s\n当前在html切片第【%d】片，共有【%d】个切片", title, url, h, a.pageSliceIndex, count)
}

func (a *Agent) aiChat(ctx context.Context, chat aimodel.ChatList) (res string, err error) {
	if a.agentFlowChat {
		return a.flow(ctx, chat)
	}
	return a.chat(ctx, chat)
}

func (a *Agent) chat(ctx context.Context, chat aimodel.ChatList) (res string, err error) {
	log.Printf("Agent[%s]开始调用AI模型: %s", a.AgentId, gconv.String(chat))
	result, err := a.aiModel.Chat(ctx, slices.Clone(chat))
	if err != nil {
		log.Printf("Agent[%s]调用AI模型失败: %v", a.AgentId, err)
		return "", errors.Wrap(err, "调用AI模型失败")
	}
	log.Printf("Agent[%s]调用AI模型成功: %s", a.AgentId, gconv.String(result))
	return result.Result, nil
}
func (a *Agent) flow(ctx context.Context, chat aimodel.ChatList) (res string, err error) {
	log.Printf("Agent[%s]开始流式调用AI模型: %s", a.AgentId, gconv.String(chat))
	resultChan, thinkChan, errChan, err := a.aiModel.ChatFlow(ctx, slices.Clone(chat))
	if err != nil {
		log.Printf("Agent[%s]流式调用AI模型失败: %v", a.AgentId, err)
		return "", errors.Wrap(err, "流式调用AI模型失败")
	}
	var result aimodel.StringResult
	defer func() {
		log.Printf("思考结果：%s", result.Reason)
	}()
	for {
		thisRoundResult := aimodel.StringResult{}
		select {
		case res, ok := <-resultChan:
			if !ok {
				log.Printf("Agent[%s]流式调用AI模型结果通道已关闭", a.AgentId)
				return result.Result, nil
			}
			if res == aimodel.DoneKey {
				continue
			}
			thisRoundResult.Result = res
			result.Result += res
			//a.sender.SendDebug(ctx, event.NewAgentAIDebugEvent(thisRoundResult))
			//log.Printf("Agent[%s]流式调用AI模型结果: %s", a.AgentId, gconv.String(res))
		case think, ok := <-thinkChan:
			if !ok {
				log.Printf("Agent[%s]流式调用AI模型思考通道已关闭", a.AgentId)
				return result.Result, nil
			}
			thisRoundResult.Reason = think
			result.Reason += think
			//a.sender.SendDebug(ctx, event.NewAgentAIDebugEvent(thisRoundResult))
			//log.Printf("Agent[%s]流式调用AI模型思考: %s", a.AgentId, think)
		case err := <-errChan:
			if err != nil {
				log.Printf("Agent[%s]流式调用AI模型错误: %v", a.AgentId, err)
				return "", errors.Wrap(err, "流式调用AI模型错误")
			}
		}
	}
}

func (a *Agent) GetResult(ctx context.Context) []collect.Data {
	return a.collect
}
