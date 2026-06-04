package compat

// fixtures 中的数据只服务阶段 1 mock 联调；真实数据接入后应逐步减少这里的依赖。
var demoClusters = []Cluster{
	{
		ID:          "demo-cluster",
		Name:        "demo-cluster",
		DisplayName: "Demo GPU Cluster",
		Status:      "Running",
		Type:        "kubernetes",
	},
}

var demoNodes = []Node{
	{
		Name:                "demo-node-01",
		InternalIP:          "192.168.1.10",
		Status:              "Ready",
		Role:                "worker",
		CPUCapacity:         32,
		MemoryCapacityBytes: 137438953472,
		GPUCount:            4,
		OSImage:             "Ubuntu 22.04",
		ContainerRuntime:    "containerd",
	},
	{
		Name:                "demo-node-02",
		InternalIP:          "192.168.1.11",
		Status:              "Ready",
		Role:                "worker",
		CPUCapacity:         32,
		MemoryCapacityBytes: 137438953472,
		GPUCount:            4,
		OSImage:             "Ubuntu 22.04",
		ContainerRuntime:    "containerd",
	},
}

var demoApps = []App{
	{
		ID:            "default/nginx-demo",
		Name:          "nginx-demo",
		Namespace:     "default",
		Kind:          "Deployment",
		Status:        "Running",
		Replicas:      1,
		ReadyReplicas: 1,
		CreatedAt:     "2026-06-06T10:00:00+08:00",
	},
}

var demoInstances = []Instance{
	{
		ID:           "default/nginx-demo-6d4cf56db6-x7k2p",
		Name:         "nginx-demo-6d4cf56db6-x7k2p",
		Namespace:    "default",
		AppName:      "nginx-demo",
		NodeName:     "demo-node-01",
		Phase:        "Running",
		PodIP:        "10.244.1.5",
		HostIP:       "192.168.1.10",
		RestartCount: 0,
	},
}

var demoDeployments = []Deployment{
	{
		Name:              "nginx-demo",
		Namespace:         "default",
		Replicas:          1,
		ReadyReplicas:     1,
		AvailableReplicas: 1,
		Labels: map[string]string{
			"app": "nginx",
		},
	},
}
