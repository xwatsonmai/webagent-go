package instruction

import (
	"context"
	"github.com/playwright-community/playwright-go"
	"github.com/xwatsonmai/webagent-go/collect"
	"github.com/xwatsonmai/webagent-go/event"
	"github.com/xwatsonmai/webagent-go/model"
	"github.com/xwatsonmai/webagent-go/running"
)

type Type string

func (i Type) ToString() string {
	switch i {
	case TypeGoto:
		return "跳转指令"
	case TypeFill:
		return "填充指令"
	case TypeClick:
		return "点击指令"
	case TypeCollect:
		return "收集指令"
	case TypeEnd:
		return "结束指令"
	case TypeSwitch:
		return "切换指令"
	case TypeFail:
		return "失败指令"
	case TypeWait:
		return "等待指令"
	case TypeCloseLabel:
		return "关闭标签页指令"
	case TypeGoBack:
		return "返回上一步指令"
	case TypeKeyboard:
		return "键盘输入指令"
	case HtmlSliceSelect:
		return "HTML切片选择器"
	}
	return string(i)
}

const (
	TypeGoto        Type = "goto"              // 跳转指令
	TypeFill        Type = "fill"              // 填充指令
	TypeClick       Type = "click"             // 点击指令
	TypeCollect     Type = "collect"           // 收集指令
	TypeEnd         Type = "end"               // 结束指令
	TypeSwitch      Type = "switch"            // 切换指令
	TypeFail        Type = "fail"              // 失败指令
	TypeWait        Type = "wait"              // 等待指令
	TypeCloseLabel  Type = "close_label"       // 关闭标签页指令
	TypeGoBack      Type = "goBack"            // 返回上一步指令
	TypeKeyboard    Type = "keyboard"          // 键盘输入指令
	HtmlSliceSelect Type = "html_slice_select" // HTML切片选择器，用于在HTML中选择特定的元素
	TypeMouseWheel  Type = "mouse_wheel"       // 鼠标滚轮指令
)

type Instruction struct {
	ID      string `json:"id"`
	Type    Type   `json:"type"`    // 指令类型: message, takeover, mcp_tool
	Target  string `json:"target"`  // 指令目标: 如mcp工具调用目标
	Content any    `json:"content"` // 指令内容: 消息内容、接管信息或MCP工具的入参
}

func (i *Instruction) IsEnd() bool {
	return i.Type == TypeEnd
}

func (i *Instruction) ToString() string {
	return i.Type.ToString() + ": " + i.Target
}

type List []Instruction

func (l List) ToString() string {
	str := ""
	for _, instruction := range l {
		str += instruction.ToString() + "\n"
	}
	return str
}

type IInstruction interface {
	ReadyInfo() (running.EventType, string)
	instruction(instruction Instruction)
	Handler(ctx context.Context, agentInfo model.AgentInfo, browser playwright.BrowserContext, nowOpenPage playwright.Page, sendChan chan event.IEvent, runInfoChan chan string) (openPage playwright.Page, collectData []collect.Data, err error, exit bool)
	Result() []string
}

type Map map[Type]IInstruction

var DefaultInstructionMap = map[Type]IInstruction{
	TypeGoto:       &Goto{},
	TypeFill:       &Fill{},
	TypeClick:      &Click{},
	TypeSwitch:     &Switch{},
	TypeFail:       &Fail{},
	TypeWait:       &Wait{},
	TypeCloseLabel: &CloseLabel{},
	TypeGoBack:     &GoBack{},
	TypeKeyboard:   &Keyboard{},
	TypeMouseWheel: &MouseWheel{},
}

func DefaultAgentInstructionMap(agentType model.AgentType) Map {
	dataMap := DefaultInstructionMap
	switch agentType {
	case model.PageCollecter:
		dataMap[TypeCollect] = &PageCollect{}
	case model.DataCollector:
		dataMap[TypeCollect] = &DataCollect{}
	case model.Browser:
		dataMap[TypeCollect] = &DataCollect{}
	}

	return dataMap
}

type GetInstructionMapFunc func(agentType model.AgentType) Map

func Builder(agentType model.AgentType, getIMap GetInstructionMapFunc, ins Instruction) IInstruction {
	if instruction, ok := getIMap(agentType)[ins.Type]; ok {
		instruction.instruction(ins)
		return instruction
	}
	return nil
}
