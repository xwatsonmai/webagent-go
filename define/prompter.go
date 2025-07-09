package define

import (
	"context"
	"github.com/xwatsonmai/webagent-go/instruction"
)

type IPrompter interface {
	SystemPrompt(ctx context.Context, userIntention string, targetUrl string) (string, error) // 获取系统提示
	StartUserPrompt(ctx context.Context, userIntention string, targetUrl string) string       // 获取开始用户提示
	// AgentAnswer 根据Agent的返回，解析出指令
	// 由于Agent的回答格式是与Prompter相关的，所以需要由Prompter解析处理
	AgentAnswer(ctx context.Context, agentAnswer string) ([]instruction.Instruction, error)
}

type IAnswer interface {
	Instruction() []instruction.Instruction // 获取需执行指令
}
