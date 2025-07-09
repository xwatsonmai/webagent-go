package event

import "encoding/json"

type InitEvent struct {
	agentId string // 代理ID
}

func NewInitEvent(agentId string) IEvent {
	return &InitEvent{
		agentId: agentId,
	}
}

func (i InitEvent) Type() Type {
	return TypeInit
}

func (i InitEvent) Content() string {
	data := map[string]any{
		"agent_id": i.agentId,
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
