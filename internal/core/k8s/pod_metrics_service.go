package k8s

import (
	"context"
)

// GetPodMetrics 获取Pod的CPU和内存监控指标
func (s *PodService) GetPodMetrics(ctx context.Context, clusterName, namespace, podName string) (*PodMetrics, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	metrics, err := GetPodMetrics(client, namespace, podName)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
