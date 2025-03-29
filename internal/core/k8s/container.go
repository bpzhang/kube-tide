package k8s

import (
	v1 "k8s.io/api/core/v1"
)

// ContainerInfo 包含容器的基本信息
type ContainerInfo struct {
	Name           string                  `json:"name"`
	Image          string                  `json:"image"`
	Resources      v1.ResourceRequirements `json:"resources"`
	Ports          []v1.ContainerPort      `json:"ports"`
	Env            []v1.EnvVar             `json:"env"`
	LivenessProbe  *v1.Probe               `json:"livenessProbe,omitempty"`
	ReadinessProbe *v1.Probe               `json:"readinessProbe,omitempty"`
	StartupProbe   *v1.Probe               `json:"startupProbe,omitempty"`
}
