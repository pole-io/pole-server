package v1

// Principal 策略相关人
type Principal struct {
	Id   string
	Name string
}

// Principal 策略相关人
type Principals struct {
	Users  []Principal `json:"users"`
	Groups []Principal `json:"groups"`
	Roles  []Principal `json:"roles"`
}

// AuthorizeResources 授权资源请求
type AuthorizeResources struct {
	// 资源类型
	ResourceType string `json:"resource_type"`
	// 资源名称
	ResourceName string `json:"resource_name"`
	// 资源ID
	ResourceID string `json:"resource_id"`
	// 资源操作
	Principals Principals `json:"principals"`
}
