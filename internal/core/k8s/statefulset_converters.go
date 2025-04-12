package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// convertResourceRequirementsToK8s 将自定义资源需求转换为K8s资源需求
func convertResourceRequirementsToK8s(resources map[string]map[string]string) corev1.ResourceRequirements {
	k8sResources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if resources == nil {
		return k8sResources
	}

	if limits, ok := resources["limits"]; ok {
		for key, value := range limits {
			parsedQuantity, err := resource.ParseQuantity(value)
			if err == nil {
				k8sResources.Limits[corev1.ResourceName(key)] = parsedQuantity
			}
		}
	}

	if requests, ok := resources["requests"]; ok {
		for key, value := range requests {
			parsedQuantity, err := resource.ParseQuantity(value)
			if err == nil {
				k8sResources.Requests[corev1.ResourceName(key)] = parsedQuantity
			}
		}
	}

	return k8sResources
}

// convertEnvVarsToK8s 将自定义环境变量转换为K8s环境变量
func convertEnvVarsToK8s(envVars []map[string]string) []corev1.EnvVar {
	if envVars == nil {
		return nil
	}

	k8sEnvVars := make([]corev1.EnvVar, 0, len(envVars))
	for _, env := range envVars {
		name, nameOk := env["name"]
		value, valueOk := env["value"]

		if nameOk {
			envVar := corev1.EnvVar{
				Name: name,
			}

			if valueOk {
				envVar.Value = value
			}

			// 处理valueFrom配置，如果需要的话
			// 这里可以扩展支持configMapKeyRef、secretKeyRef等

			k8sEnvVars = append(k8sEnvVars, envVar)
		}
	}

	return k8sEnvVars
}

// convertContainerPortsToK8s 将自定义容器端口转换为K8s容器端口
func convertContainerPortsToK8s(ports []map[string]interface{}) []corev1.ContainerPort {
	if ports == nil {
		return nil
	}

	k8sPorts := make([]corev1.ContainerPort, 0, len(ports))
	for _, port := range ports {
		containerPort, ok := port["containerPort"].(int32)
		if !ok {
			// 尝试转换其他数值类型
			if containerPortFloat, ok := port["containerPort"].(float64); ok {
				containerPort = int32(containerPortFloat)
			} else {
				continue // 跳过这个端口定义
			}
		}

		k8sPort := corev1.ContainerPort{
			ContainerPort: containerPort,
		}

		if name, ok := port["name"].(string); ok {
			k8sPort.Name = name
		}

		if protocol, ok := port["protocol"].(string); ok {
			k8sPort.Protocol = corev1.Protocol(protocol)
		}

		if hostPort, ok := port["hostPort"].(int32); ok {
			k8sPort.HostPort = hostPort
		} else if hostPortFloat, ok := port["hostPort"].(float64); ok {
			k8sPort.HostPort = int32(hostPortFloat)
		}

		if hostIP, ok := port["hostIP"].(string); ok {
			k8sPort.HostIP = hostIP
		}

		k8sPorts = append(k8sPorts, k8sPort)
	}

	return k8sPorts
}

// convertK8sProbeToCustomProbe 将K8s探针转换为自定义探针
func convertK8sProbeToCustomProbe(probe *corev1.Probe) map[string]interface{} {
	if probe == nil {
		return nil
	}

	customProbe := map[string]interface{}{
		"initialDelaySeconds": probe.InitialDelaySeconds,
		"timeoutSeconds":      probe.TimeoutSeconds,
		"periodSeconds":       probe.PeriodSeconds,
		"successThreshold":    probe.SuccessThreshold,
		"failureThreshold":    probe.FailureThreshold,
	}

	// 处理HTTP探针
	if probe.HTTPGet != nil {
		httpGet := map[string]interface{}{
			"path":   probe.HTTPGet.Path,
			"port":   probe.HTTPGet.Port.String(),
			"scheme": string(probe.HTTPGet.Scheme),
		}

		if len(probe.HTTPGet.HTTPHeaders) > 0 {
			headers := make([]map[string]string, 0, len(probe.HTTPGet.HTTPHeaders))
			for _, header := range probe.HTTPGet.HTTPHeaders {
				headers = append(headers, map[string]string{
					"name":  header.Name,
					"value": header.Value,
				})
			}
			httpGet["httpHeaders"] = headers
		}

		customProbe["httpGet"] = httpGet
	}

	// 处理TCP探针
	if probe.TCPSocket != nil {
		customProbe["tcpSocket"] = map[string]interface{}{
			"port": probe.TCPSocket.Port.String(),
		}
	}

	// 处理Exec探针
	if probe.Exec != nil {
		customProbe["exec"] = map[string]interface{}{
			"command": probe.Exec.Command,
		}
	}

	return customProbe
}
