package server

// CompatEndpointStatusResponse 是兼容接口实现状态响应。
type CompatEndpointStatusResponse struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
}
