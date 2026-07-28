package model

type TypeInfo struct {
	Type string `json:"type"`
	Desc string `json:"desc"`
}

// InstanceEventType 探测事件类型
type InstanceEventType string

const (
	// EventDiscoverNone empty discover event
	EventDiscoverNone InstanceEventType = "EventDiscoverNone"
	// EventInstanceOnline instance becoming online
	EventInstanceOnline InstanceEventType = "InstanceOnline"
	// EventInstanceTurnUnHealth Instance becomes unhealthy
	EventInstanceTurnUnHealth InstanceEventType = "InstanceTurnUnHealth"
	// EventInstanceTurnHealth Instance becomes healthy
	EventInstanceTurnHealth InstanceEventType = "InstanceTurnHealth"
	// EventInstanceOpenIsolate Instance is in isolation
	EventInstanceOpenIsolate InstanceEventType = "InstanceOpenIsolate"
	// EventInstanceCloseIsolate Instance shutdown isolation state
	EventInstanceCloseIsolate InstanceEventType = "InstanceCloseIsolate"
	// EventInstanceOffline Instance offline
	EventInstanceOffline InstanceEventType = "InstanceOffline"
	// EventInstanceSendHeartbeat Instance send heartbeat package to server
	EventInstanceSendHeartbeat InstanceEventType = "InstanceSendHeartbeat"
)

type EventRecordLogResponse struct {
	Code    uint32         `json:"code"`
	Info    string         `json:"info"`
	Total   uint64         `json:"total"`
	Size    uint32         `json:"size"`
	HasNext bool           `json:"has_next"`
	Data    []*EventRecord `json:"data"`
	Cursor  string         `json:"cursor"`
}

type EventRecord struct {
	Cursor    string `json:"cursor"`
	EventType string `json:"event_type"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
	Resource  string `json:"resource"`
	EventTime string `json:"event_time"`
	Server    string `json:"server"` // 事件发生的服务器
}
