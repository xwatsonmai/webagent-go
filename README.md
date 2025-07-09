# webagent-go
使用大模型驱动的Web自动化程序/Using Large Model Driven Web Automation Programs

# 项目初衷 / Project Motivation
- 构建一个基于大模型的，低使用门槛、高泛化性的Web自动化程序/Build a Large Model Based, Low Usage Threshold, Highly Generalized Web Automation Program
- 它可以替代传统的Web自动化程序，如爬虫、自动化测试等/It Can Replace Traditional Web Automation Programs, Such as Crawlers, Automated Testing, etc.
- 它可以在不需要编写代码的情况下，完成大部分Web自动化任务/It Can Complete Most Web Automation Tasks Without Writing Code

# 体验地址/ demo
http://114.132.90.16/#/

# 能力/Capabilities
- 基于大模型的Web自动化/ Web Automation Based on Large Models
- 支持Agent多轮自动化操作/Supports Agent Multi-Round Automated Operations
- 支持点击、输入、滚动等操作/Supports Click, Input, Scroll and Other Operations
- 可扩展，支持自定义操作/Extensible, Supports Custom Operations

# 当前的问题
- 凌乱的日志打印/ Messy Log Printing
- 部分写死的逻辑/ Partially Hardcoded Logic
- 缺乏错误处理/ Lack of Error Handling

# 预期的未来
- 可高泛化性的Web自动化/ Highly Generalized Web Automation
- 可实现大部分页面的无干预自动化/ Can Achieve Unattended Automation of Most Pages
- 高可用性和可扩展性/ High Usability and Extensibility
- 支持mcp协议/ Supports MCP Protocol

# 使用说明/Usage Instructions
```go
package main

import (
	"context"
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
```
