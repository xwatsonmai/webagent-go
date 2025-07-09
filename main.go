package main

import (
	"context"
	"fmt"
	"github.com/xwatsonmai/webagent-go/agent"
	v1 "github.com/xwatsonmai/webagent-go/agent/v1"
	"github.com/xwatsonmai/webagent-go/aimodel"
	"github.com/xwatsonmai/webagent-go/define"
	"github.com/xwatsonmai/webagent-go/instruction"
	"github.com/xwatsonmai/webagent-go/model"
)

func main() {
	ctx := context.Background()
	ai, _ := aimodel.Builder("deepseek", false, "your-key")
	var agentType = model.PageCollecter // 选择使用PageCollecter类型的Agent
	// 实现你自己的Sender和Prompter接口
	var (
		sender   define.ISender   = v1.TestSender{}
		prompter define.IPrompter = v1.NewMcpServerPrompter(agentType)
	)
	webAgent, _ := agent.NewAgent(ctx, agentType, instruction.DefaultAgentInstructionMap, sender, prompter, ai, true, false)
	webAgent.Do(ctx, "redis怎么用", "https://goframe.org/")
	fmt.Println("Agent execution completed", "result:", webAgent.GetResult(ctx))
}
