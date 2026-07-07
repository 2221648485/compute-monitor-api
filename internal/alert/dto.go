package alert

// RuleOperationResponse 是告警规则更新、删除等操作响应。
type RuleOperationResponse struct {
	ID      string `json:"id"`
	Updated bool   `json:"updated,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
}
