package cluster

import "errors"

var (
	ErrClusterNotFound      = errors.New("cluster not found")
	ErrClusterExists        = errors.New("cluster already exists")
	ErrClusterIDRequired    = errors.New("cluster id required")
	ErrClusterNameRequired  = errors.New("cluster name required")
	ErrInvalidClusterStatus = errors.New("invalid cluster status")
	ErrInvalidPrometheusURL = errors.New("invalid prometheus url")
	ErrInvalidAccessMode    = errors.New("invalid cluster access mode")
	ErrKubeconfigRequired   = errors.New("cluster kubeconfig required")
	ErrAPIServerRequired    = errors.New("cluster api server required")
	ErrCredentialRequired   = errors.New("cluster credential required")
)

func IsClusterError(err error) bool {
	return errors.Is(err, ErrClusterNotFound) ||
		errors.Is(err, ErrClusterExists) ||
		errors.Is(err, ErrClusterIDRequired) ||
		errors.Is(err, ErrClusterNameRequired) ||
		errors.Is(err, ErrInvalidClusterStatus) ||
		errors.Is(err, ErrInvalidPrometheusURL) ||
		errors.Is(err, ErrInvalidAccessMode) ||
		errors.Is(err, ErrKubeconfigRequired) ||
		errors.Is(err, ErrAPIServerRequired) ||
		errors.Is(err, ErrCredentialRequired)
}

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
	case errors.Is(err, ErrInvalidAccessMode):
		return "invalid cluster access mode"
	case errors.Is(err, ErrKubeconfigRequired):
		return "cluster kubeconfig required"
	case errors.Is(err, ErrAPIServerRequired):
		return "cluster api server required"
	case errors.Is(err, ErrCredentialRequired):
		return "cluster credential required"
	default:
		return "cluster operation failed"
	}
}
