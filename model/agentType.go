package model

type AgentType string

const (
	PageCollecter AgentType = "page_collector" // 页面采集器
	DataCollector AgentType = "data_collector" // 数据采集器
	Browser       AgentType = "browser"        // 浏览者
)
