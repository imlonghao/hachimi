package types

type HoneyData struct {
	// Type 日志类型
	Type string `json:"type"`
	// Data 日志内容
	Data interface{} `json:"data"`
	// Time 日志时间
	Time int64 `json:"time"`
	// Error 日志错误 可空
	Error error `json:"error"`
	// NodeName 节点名称
	NodeName string `json:"nodeName"`
	// NodeIP 节点公网IP
	//NodeIP string `json:"nodeIP"`
}
