package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	"github.com/xwatsonmai/webagent-go/instruction"
	"github.com/xwatsonmai/webagent-go/model"
	"strings"
	"text/template"
)

const (
	systemPrompt = "你是一个浏览器操作助手，能够帮助用户进行浏览器相关的操作。\n\n# 你运行的环境\n- 你允许在一个go所编写的进程中，并使用`github.com/playwright-community/playwright-go`包来操作浏览器。\n- 你所输出的操作指令会被转换为对应的playwright-go的API调用。\n- 与你沟通的是一个系统，它负责把你的输出转换为实际的浏览器操作，并把浏览器操作的结果反馈给你。\n- 你可以与系统进行多轮对话，直到任务完成。\n\n# 你的任务\n- {{ .mainTask }}\n- 你需要先深入理解用户的需求，然后根据页面的实际情况，输出一系列操作指令来完成用户的任务。\n- 你需要根据用户的指令，在对应的站点中检索对应的信息。\n- 如果用户有一些说明性的描述，那么你需要严格按照用户的描述来执行操作。\n- 由于与你进行交互的是一个系统，它只能代理执行你所输出的操作指令，所以你在输出内容时只需要输出操作指令即可，不需要输出任何解释或说明。\n- 你所面对的是一个会响应你操作的浏览器，你不需要一步到位地完成任务，可以分开多轮操作，根据当前的浏览器实时状态来决定下一步的操作。不要一次性输出所有的操作指令，而是根据当前的浏览器状态，逐步输出操作指令。\n- 注意：你只能给根据当前的浏览器实际情况来输出这一轮的操作指令，而不能输出下一轮的操作指令，更不要把预测的下一轮指令在这一轮输出。你需要在每轮中根据当前的浏览器状态来决定当下的操作指令。比如：打开页面与输入信息，由于还未打开页面，所以你不能在这一轮中输出输入信息的指令，而是先打开页面，等页面加载完成后，根据页面的实际情况来决定下一步的操作。\n- 你并不是只能采集一个页面的信息，你可以在多个页面中进行操作，直到你认为已经采集到足够的信息为止。\n\n# 你的能力\n- 你可以在浏览器中打开指定的URL。\n- 你可以在页面中查找指定的元素。\n- 你可以点击页面中的元素。\n- 你可以输入文本到页面中的输入框。\n- 你可以切换/关闭标签页。\n- 你可以上下滚动鼠标，实现页面的滚动。遇到一些瀑布流、或者是需要滚动触发加载的页面时，你可以通过滚动鼠标来加载更多内容。\n- 你可以使用点击+鼠标滚轮的方式来实现在页面上的某个元素上进行滚动，用来解决如表格、列表等元素无法直接滚动的问题。\n- 你可以通过与系统进行多轮交互，来完成复杂的任务。\n- 当你需要进行键盘输入，比如按下回车键或者回退键时，你可以使用`keyboard`操作来模拟键盘输入。\n\n# 注意\n- 请注意，如果页面出现输入手机号、邮箱等需要填入接收验证码的输入框时，你禁止向输入框输入任何内容，并且严禁点击发送验证码的按钮。\n- 当你进行选择器操作时，应该尽可能使用唯一的属性作为选择器，比如id。如果无法唯一定位到一个元素，你可以通过`value`字段来补充文字内容来唯一定位元素。总而言之，你需要确保你的选择器能够唯一定位到一个元素。\n- 由于你在进行下一轮操作前是需要进行思考的，这个思考会占用一定的时间，所以正常情况下你是不需要使用`wait`操作的，除非你发现页面还没有加载完成，或者需要等待某个元素出现时，你可以使用`wait`操作来等待一段时间。\n- 如果你在使用某个选择器失败时，你可以尝试使用其他选择器来定位元素。你可以通过查看页面的HTML结构，找到其他可能的选择器。而不是一直反复使用同一个选择器进行重试。\n- 当你的多次操作都失败时，你应该尝试其他策略，如后退、点击ESC等方式来尝试解决问题，而不是一味地重复同样的操作。\n- 如果页面有弹窗、模态框等元素，它可能会有遮罩层导致你的操作无效。你可以尝试先关闭弹窗或模态框，比如点击关闭按钮、或者使用`keyboard`操作模拟按下`Esc`键来关闭弹窗或模态框，又或者是尝试使用`goBack`操作让页面回到上一个状态。\n- 当你在页面上进行搜索查询时，可以多尝试不同的搜索关键词。\n\n# 名词解释\n- 交互轮数：指的是你与系统之间的对话轮数。每次你输出操作指令，系统执行后反馈结果，你再根据结果输出下一轮的操作指令，这样的一轮对话称为一次交互轮数。\n- 模块框、弹窗：指的是页面上弹出的对话框或提示框，通常用于显示信息、警告或确认操作等。它们可能会阻止你与页面其他元素进行交互，直到你关闭它们。你可以通过它们是否带有\"modal\"、\"dialog\"、\"popup\"等关键class来判断它们是否是模态框或弹窗。通常情况下，你可以通过点击它们上的关闭按钮来关闭它们，或者使用`keyboard`操作模拟按下`Esc`键来关闭模态框或弹窗。\n- 遮罩层：指的是页面上覆盖在其他元素之上的透明或半透明层，通常用于模态框、弹窗等场景。遮罩层会阻止你与被遮罩的元素进行交互。如果页面上有close按钮，你可以尝试点击它来关闭遮罩层，或者使用`keyboard`操作模拟按下`Esc`键来关闭模态框或弹窗。\n\n# 输出格式\n## 输出格式：\n- 在输入框中输入内容，并点击包含“搜索”字样的a标签\n```json\n[\n   {\n      \"action\": \"fill\",\n      \"target\": \"input[name=\\\"q\\\"]\",\n      \"value\": \"playwright-go\"\n   },\n   {\n      \"action\": \"click\",\n      \"target\": \"a\",\n      \"value\": \"搜索\"\n   }\n]\n```\n\n{{ .collectInstruction }}\n\n## 输出说明\n1. `action`：表示你要执行的操作类型，可选值包括：\n   - `goto`：浏览器操作：打开指定的URL\n   - `fill`：浏览器操作：在输入框中输入文本\n   - `click`：浏览器操作：点击页面中的元素\n   - `switch`: 浏览器操作：切换到指定的标签页\n   - `goBack`: 浏览器操作：后退到上一个页面\n   - `close_label`: 浏览器操作：关闭标签页\n   - `keyboard`: 浏览器操作：模拟键盘输入\n   - `collect`: 系统指令：{{ .collectActionDesc }}\n   - `wait`: 系统指令：等待一段时间，通常用于等待页面加载完成或元素出现。如果你发现页面还没有加载完成，你可以选择等待几秒钟再进行下一步操作。\n   - `end`: 系统指令：结束本轮操作，不再执行任何操作\n   - `mouse_wheel`: 浏览器操作：模拟鼠标滚轮\n   - `fail`: 系统指令：表示检索不到用户需要的信息，通知系统结束本轮操作并返回错误信息\n   - `qrcode`: 系统指令：当页面出现二维码登录等需要用户扫描二维码时，通知系统进行截图并发送给用户扫描。当你发起`qrcode`操作时，系统会将当前页面的二维码截图发送给用户，然后持续等待最多60秒，直到用户扫描二维码登录成功或超时。\n   - `html_slice_select`: 系统指令，由于html内容可能过多，所以html会被切割为2000个字符为一个单位的切片，你所看到的html是其中的某个片段。如果你需要查看其他片段的html内容，你可以使用`html_slice_select`操作来选择其他片段。系统会返回对应片段的html内容。注意：该指令并不是对浏览器页面进行滚动，而是对html内容进行切片选择。你可以通过`taget`字段来指定要查看的html片段的索引值。\n2. `target`：\n   - 当`action`为`goto`时，表示要打开的URL。\n   - 当`action`为`fill`、`click`时，表示你要操作的元素的选择器。它会作为输入直接传递给playwright-go的API调用，如：`page.Click(`a[href=\"/info\"]`)`，所以target的值需要是符合给playwright-go的API调用的选择器格式。\n   - 当`action`为`switch`时，表示要切换到的标签页的索引值。\n   - 当`action`为`close_label`时，如果该值为空，则表示关闭当前标签页；如果该值为数字，则表示关闭指定索引的标签页。如果关闭的是当前标签页，那么该标签页会被关闭，浏览器会自动切换到索引值为1的标签页。\n   - 当`action`为`keyboard`时，表示要模拟的键盘输入内容。\n   - 当`action`为`html_slice_select`时，表示要查看的html片段的索引值，索引值从0开始计数。\n   - 当`action`为`mouse_wheel`时，这个字段为滚动的距离，其中正数表示向下滚动，负数表示向上滚动。并且只能上下滚动，不能左右滚动。\n   - 当`action`为`collect`时，{{ .collectTargetDesc }}。\n   - 当`action`为其他操作时，这个字段可以不填。\n3. `value`：表示执行`action`时所需要的值。\n   - value必须是一个字符串。\n   - 对于`fill`操作，表示要填充到输入框中的文本。\n   - 对于`click`操作，如果通过选择器无法唯一定位到一个元素（如存在多个a标签，它们没有id，href都等于`javascript:void(0)`，且具有统一样式），只能通过文字内容来定位元素时，则可以通过value中补充文字内容来唯一定位元素。\n   - 对于`collect`操作，{{ .collectValueDesc }}。\n   - 对于`end`操作，这个字段可以不填。\n   - 对于`fail`操作，这个字段是失败的错误信息，可以总结一下之前执行过的操作，并表示无法找到用户需要的信息。\n   - 对于`wait`操作，这个字段为等待的时间，时间为秒。\n   - 对于其他操作，这个字段可以不填。\n4. 你可以输出多个操作指令，它们会被依次执行，但是`end`指令意味着：\n   - 你已经完成了当前的任务，不再需要执行任何操作。\n   - 如果你输出了`end`指令，后续的操作将不会被执行。\n\n## 要求：\n1. 你只能输出操作指令，不能输出任何解释或说明。\n2. 必须严格按照上述格式输出操作指令。\n3. 你的输出必须只能是一个JSON数组，数组中的每个元素都是一个操作指令对象。"
)

const (
	pageCollecterMainTask    = "你的核心任务是检索用户所需要的信息。"
	pageCollecterInstruction = "- 告知系统采集该页面数据\n```json\n[\n  {\n    \"action\": \"collect\",\n    \"target\": \"第一章：Playwright-Go简介\"\n  }\n]\n```"
	pageCollectActionDesc    = "该页面中包含了用户需要的信息，通知系统把该页面中的信息收集起来进行进一步处理"
	pageCollectTargetDesc    = "表示要收集的页面信息的标题或描述"
	pageCollectValueDesc     = "这个字段可以不填"
)

const (
	dataCollecterMainTask    = "你的核心任务是浏览相关页面，并采集到足够的用户所需的信息。"
	dataCollecterInstruction = "- 向系统上报你生成的信息\n```json\n[\n  {\n    \"action\": \"collect\",\n    \"target\": \"\",\n    \"value\": \"用户id：12345\\n用户名：testuser\\n用户需求：获取Playwright-Go的使用方法\"\n  }\n]\n```"
	dataCollectActionDesc    = "该页面中包含了用户所需的信息，你需要把这些信息按用户需要的格式总结好，然后上报给系统"
	dataCollectTargetDesc    = "这个字段可以不填"
	dataCollectValueDesc     = "你从页面中提取的信息内容，按用户需要的格式进行总结"
)

const (
	BrowserMainTask    = "你的核心任务是在帮助用户在目标站点上浏览相关的信息，并收集一些页面。在最后这些页面会被汇总起来给另外一个智能体，完成一份总结报告。"
	BrowserInstruction = "- 向系统上报你收集到的信息\n```json\n[\n  {\n    \"action\": \"collect\",\n    \"target\": \"简短的总结\",\n    \"value\": \"该页面描述了xxx\"\n  }\n]\n```"
	BrowserActionDesc  = "该页面中包含了可以用于参考的资料，你需要把这些信息尽可能详细地总结起来，方便后续的“总结智能体”进行了解和使用"
	BrowserTargetDesc  = "该信息的一个简单总结"
	BrowserValueDesc   = "该信息的详细汇总，需要保留重要的信息点。"
)

type McpServerPrompter struct {
	agentType model.AgentType
}

func NewMcpServerPrompter(agentType model.AgentType) *McpServerPrompter {
	if agentType == "" {
		agentType = model.PageCollecter // 默认是页面采集者
	}
	return &McpServerPrompter{
		agentType: agentType,
	}
}

func (m McpServerPrompter) SystemPrompt(ctx context.Context, userIntention string, targetUrl string) (string, error) {

	//data := map[string]string{
	//	"mainTask":           "",
	//	"collectInstruction": "",
	//	"collectActionDesc":  "",
	//	"collectTargetDesc":  "",
	//	"collectValueDesc":   "",
	//	"targetSite":         targetUrl,
	//	"userRequest":        userIntention,
	//}
	//
	//switch m.agentType {
	//case model.PageCollecter:
	//	// 页面采集者，负责采集某些目标页面的，是把整个页面内容采集下来，并转换为md
	//	data["mainTask"] = pageCollecterMainTask
	//	data["collectInstruction"] = pageCollecterInstruction
	//	data["collectActionDesc"] = pageCollectActionDesc
	//	data["collectTargetDesc"] = pageCollectTargetDesc
	//	data["collectValueDesc"] = pageCollectValueDesc
	//case model.DataCollector:
	//	// 数据采集者，负责采集用户所需的信息，可能需要多次交互
	//	data["mainTask"] = dataCollecterMainTask
	//	data["collectInstruction"] = dataCollecterInstruction
	//	data["collectActionDesc"] = dataCollectActionDesc
	//	data["collectTargetDesc"] = dataCollectTargetDesc
	//	data["collectValueDesc"] = dataCollectValueDesc
	//}
	//t, err := template.New("example").Parse(systemPrompt)
	//if err != nil {
	//	fmt.Println("解析模板失败:", err)
	//	return "", err
	//}
	//var result bytes.Buffer
	//err = t.Execute(&result, data)
	//if err != nil {
	//	fmt.Println("执行模板失败:", err)
	//	return "", err
	//}
	//
	//// 输出结果
	//finalString := result.String()
	return "", nil
}

func (m McpServerPrompter) StartUserPrompt(ctx context.Context, userIntention string, targetUrl string) string {
	data := map[string]string{
		"mainTask":           "",
		"collectInstruction": "",
		"collectActionDesc":  "",
		"collectTargetDesc":  "",
		"collectValueDesc":   "",
		"targetSite":         targetUrl,
		"userRequest":        userIntention,
	}

	switch m.agentType {
	case model.PageCollecter:
		// 页面采集者，负责采集某些目标页面的，是把整个页面内容采集下来，并转换为md
		data["mainTask"] = pageCollecterMainTask
		data["collectInstruction"] = pageCollecterInstruction
		data["collectActionDesc"] = pageCollectActionDesc
		data["collectTargetDesc"] = pageCollectTargetDesc
		data["collectValueDesc"] = pageCollectValueDesc
	case model.DataCollector:
		// 数据采集者，负责采集用户所需的信息，可能需要多次交互
		data["mainTask"] = dataCollecterMainTask
		data["collectInstruction"] = dataCollecterInstruction
		data["collectActionDesc"] = dataCollectActionDesc
		data["collectTargetDesc"] = dataCollectTargetDesc
		data["collectValueDesc"] = dataCollectValueDesc
	case model.Browser:
		data["mainTask"] = BrowserMainTask
		data["collectInstruction"] = BrowserInstruction
		data["collectActionDesc"] = BrowserActionDesc
		data["collectTargetDesc"] = BrowserTargetDesc
		data["collectValueDesc"] = BrowserValueDesc
	}
	t, err := template.New("example").Parse(systemPrompt)
	if err != nil {
		fmt.Println("解析模板失败:", err)
		return ""
	}
	var result bytes.Buffer
	err = t.Execute(&result, data)
	if err != nil {
		fmt.Println("执行模板失败:", err)
		return ""
	}

	// 输出结果
	finalString := result.String()

	finalString += "\n- 目标站点：\n  %s\n- 站点使用提示：\n```\n%s\n```\n- 站点使用经验：\n```\n%s\n```\n- 用户需求：\n```\n  %s\n```"
	experienceDataStr := "暂无可参考的经验"
	fixExp := "无可用提示"
	return fmt.Sprintf(finalString, targetUrl, fixExp, experienceDataStr, userIntention)
}

func (m McpServerPrompter) AgentAnswer(ctx context.Context, agentAnswer string) ([]instruction.Instruction, error) {
	// 把前后的```json，```去掉
	agentAnswer = strings.ReplaceAll(agentAnswer, "```json", "")
	agentAnswer = strings.ReplaceAll(agentAnswer, "```", "")
	type Data struct {
		Action string `json:"action"`
		Target any    `json:"target"`
		Value  any    `json:"value,omitempty"`
	}

	var aiResponse []Data
	if err := json.Unmarshal([]byte(agentAnswer), &aiResponse); err != nil {
		fmt.Println("解析Agent的回答失败:", agentAnswer)
		return nil, fmt.Errorf("无法解析Agent的回答: %w", err)
	}
	var instructions []instruction.Instruction
	for _, item := range aiResponse {
		instructions = append(instructions, instruction.Instruction{
			ID:      uuid.New().String(),
			Type:    instruction.Type(item.Action),
			Target:  gconv.String(item.Target),
			Content: item.Value,
		})
	}
	return instructions, nil
}
