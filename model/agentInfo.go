package model

type AgentInfo struct {
	AgentId   string `json:"agent_id"`
	ThinkTime int    `json:"think_time"` // 思考时间，单位为秒
}
