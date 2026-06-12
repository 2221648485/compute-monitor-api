package cluster

import "errors"

var (
	// ErrClusterNotFound 表示要操作的集群配置不存在。
	ErrClusterNotFound = errors.New("cluster not found")
	// ErrClusterExists 表示要创建的集群 ID 已经存在。
	ErrClusterExists = errors.New("cluster already exists")
	// ErrClusterIDRequired 表示请求中缺少集群 ID。
	ErrClusterIDRequired = errors.New("cluster id required")
	// ErrClusterNameRequired 表示请求中缺少集群名称。
	ErrClusterNameRequired = errors.New("cluster name required")
	// ErrInvalidClusterStatus 表示集群状态不在允许范围内。
	ErrInvalidClusterStatus = errors.New("invalid cluster status")
	// ErrInvalidPrometheusURL 表示 Prometheus 地址格式不合法。
	ErrInvalidPrometheusURL = errors.New("invalid prometheus url")
)

// IsClusterError 判断 err 是否是集群模块定义的业务错误。
func IsClusterError(err error) bool {
	return errors.Is(err, ErrClusterNotFound) ||
		errors.Is(err, ErrClusterExists) ||
		errors.Is(err, ErrClusterIDRequired) ||
		errors.Is(err, ErrClusterNameRequired) ||
		errors.Is(err, ErrInvalidClusterStatus) ||
		errors.Is(err, ErrInvalidPrometheusURL)
}

// ErrorMessage 把集群模块内部错误转换成可以返回给前端的提示。
func ErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrClusterNotFound):
		return "cluster not found"
	case errors.Is(err, ErrClusterExists):
		return "cluster already exists"
	case errors.Is(err, ErrClusterIDRequired):
		return "cluster id required"
	case errors.Is(err, ErrClusterNameRequired):
		return "cluster name required"
	case errors.Is(err, ErrInvalidClusterStatus):
		return "invalid cluster status"
	case errors.Is(err, ErrInvalidPrometheusURL):
		return "invalid prometheus url"
	default:
		return "cluster operation failed"
	}
}
