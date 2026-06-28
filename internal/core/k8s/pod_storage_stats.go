package k8s

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	diskStatsSourceKubelet = "kubelet-stats"
	diskStatsSourceMetrics = "metrics-api"
	diskStatsSourceExec    = "exec"
)

// PodDiskStats Pod 磁盘占用统计（字节）
type PodDiskStats struct {
	PodUsedBytes  int64
	ContainerUsed map[string]int64
	VolumeUsed    map[string]int64
	Source        string
}

// GetPodDiskStats 获取 Pod 磁盘占用，优先 kubelet stats，其次 metrics-server ephemeral-storage，最后 exec
func GetPodDiskStats(
	client *kubernetes.Clientset,
	config *rest.Config,
	pod *corev1.Pod,
	podMetrics *metricsv1beta1.PodMetrics,
) *PodDiskStats {
	if pod == nil {
		return nil
	}

	if stats := getDiskStatsFromKubelet(client, config, pod); stats != nil && stats.PodUsedBytes > 0 {
		return stats
	}

	if podMetrics != nil {
		if stats := getDiskStatsFromMetricsAPI(pod, podMetrics); stats != nil && stats.PodUsedBytes > 0 {
			return stats
		}
	}

	if usageMap, err := getPodDiskUsageByExec(client, config, pod.Namespace, pod.Name); err == nil && len(usageMap) > 0 {
		stats := buildDiskStatsFromExec(pod, usageMap)
		if stats.PodUsedBytes > 0 {
			return stats
		}
	}

	return &PodDiskStats{
		ContainerUsed: make(map[string]int64),
		VolumeUsed:    make(map[string]int64),
	}
}

func (s *PodDiskStats) containerBytes(name string) int64 {
	if s == nil || s.ContainerUsed == nil {
		return 0
	}
	return s.ContainerUsed[name]
}

type kubeletStatsSummary struct {
	Pods []kubeletPodStats `json:"pods"`
}

type kubeletPodStats struct {
	PodRef           corev1.ObjectReference `json:"podRef"`
	EphemeralStorage *kubeletFsStats        `json:"ephemeral-storage,omitempty"`
	Containers       []kubeletContainerStats `json:"containers"`
	VolumeStats      []kubeletVolumeStats   `json:"volumeStats,omitempty"`
}

type kubeletFsStats struct {
	UsedBytes *uint64 `json:"usedBytes,omitempty"`
}

type kubeletContainerStats struct {
	Name   string          `json:"name"`
	Rootfs *kubeletFsStats `json:"rootfs,omitempty"`
	Logs   *kubeletFsStats `json:"logs,omitempty"`
}

type kubeletVolumeStats struct {
	Name      string  `json:"name"`
	UsedBytes *uint64 `json:"usedBytes,omitempty"`
}

func getDiskStatsFromKubelet(client *kubernetes.Clientset, config *rest.Config, pod *corev1.Pod) *PodDiskStats {
	if pod.Spec.NodeName == "" {
		return nil
	}

	data, err := client.CoreV1().RESTClient().Get().
		Resource("nodes").
		Name(pod.Spec.NodeName).
		SubResource("proxy").
		Suffix("stats/summary").
		DoRaw(context.Background())
	if err != nil {
		return nil
	}

	var summary kubeletStatsSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil
	}

	for _, podStats := range summary.Pods {
		ref := podStats.PodRef
		if ref.Name != pod.Name || ref.Namespace != pod.Namespace {
			continue
		}
		if pod.UID != "" && ref.UID != "" && ref.UID != pod.UID {
			continue
		}
		return buildDiskStatsFromKubeletPod(pod, podStats)
	}

	return nil
}

func buildDiskStatsFromKubeletPod(pod *corev1.Pod, podStats kubeletPodStats) *PodDiskStats {
	stats := &PodDiskStats{
		ContainerUsed: make(map[string]int64),
		VolumeUsed:    make(map[string]int64),
		Source:        diskStatsSourceKubelet,
	}

	var containerSum int64
	for _, containerStats := range podStats.Containers {
		used := fsStatsBytes(containerStats.Rootfs) + fsStatsBytes(containerStats.Logs)
		if used > 0 {
			stats.ContainerUsed[containerStats.Name] = used
			containerSum += used
		}
	}

	var volumeSum int64
	for _, volumeStats := range podStats.VolumeStats {
		used := fsStatsBytes(&kubeletFsStats{UsedBytes: volumeStats.UsedBytes})
		if used > 0 {
			stats.VolumeUsed[volumeStats.Name] = used
			volumeSum += used
		}
	}

	attributeVolumeUsageToContainers(pod, stats.VolumeUsed, stats.ContainerUsed)

	if podStats.EphemeralStorage != nil {
		stats.PodUsedBytes = fsStatsBytes(podStats.EphemeralStorage)
	}
	if stats.PodUsedBytes == 0 {
		stats.PodUsedBytes = containerSum + volumeSum
	}
	if stats.PodUsedBytes == 0 {
		for _, used := range stats.ContainerUsed {
			stats.PodUsedBytes += used
		}
	}

	return stats
}

func getDiskStatsFromMetricsAPI(pod *corev1.Pod, podMetrics *metricsv1beta1.PodMetrics) *PodDiskStats {
	stats := &PodDiskStats{
		ContainerUsed: make(map[string]int64),
		VolumeUsed:    make(map[string]int64),
		Source:        diskStatsSourceMetrics,
	}

	for _, containerMetrics := range podMetrics.Containers {
		if qty, ok := containerMetrics.Usage[corev1.ResourceEphemeralStorage]; ok {
			used := qty.Value()
			if used > 0 {
				stats.ContainerUsed[containerMetrics.Name] = used
				stats.PodUsedBytes += used
			}
		}
	}

	if stats.PodUsedBytes == 0 {
		return nil
	}

	// metrics API 不提供卷级明细，按挂载关系把 PVC 配额记为参考（不占 used）
	_ = pod
	return stats
}

func buildDiskStatsFromExec(pod *corev1.Pod, usageMap map[string]int64) *PodDiskStats {
	stats := &PodDiskStats{
		ContainerUsed: make(map[string]int64),
		VolumeUsed:    make(map[string]int64),
		Source:        diskStatsSourceExec,
	}

	for path, used := range usageMap {
		if used <= 0 {
			continue
		}
		stats.PodUsedBytes += used
		assigned := false
		for _, container := range pod.Spec.Containers {
			for _, mount := range container.VolumeMounts {
				if path == mount.MountPath || hasPathPrefix(path, mount.MountPath) {
					stats.ContainerUsed[container.Name] += used
					assigned = true
					break
				}
			}
		}
		if !assigned {
			if len(pod.Spec.Containers) > 0 {
				name := pod.Spec.Containers[0].Name
				stats.ContainerUsed[name] += used
			}
		}
	}

	return stats
}

func attributeVolumeUsageToContainers(pod *corev1.Pod, volumeUsed, containerUsed map[string]int64) {
	for volumeName, used := range volumeUsed {
		if used <= 0 {
			continue
		}

		var mountContainers []string
		for _, container := range pod.Spec.Containers {
			for _, mount := range container.VolumeMounts {
				if mount.Name == volumeName {
					mountContainers = append(mountContainers, container.Name)
					break
				}
			}
		}
		if len(mountContainers) == 0 {
			continue
		}

		share := used / int64(len(mountContainers))
		if share == 0 {
			share = used
		}
		for _, containerName := range mountContainers {
			containerUsed[containerName] += share
		}
	}
}

func fsStatsBytes(stats *kubeletFsStats) int64 {
	if stats == nil || stats.UsedBytes == nil {
		return 0
	}
	return int64(*stats.UsedBytes)
}

func hasPathPrefix(path, prefix string) bool {
	if prefix == "" || path == prefix {
		return path == prefix
	}
	if len(path) <= len(prefix) {
		return false
	}
	return path[:len(prefix)] == prefix && (path[len(prefix)] == '/')
}

// sumEphemeralStorageResources 汇总 Pod 内容器 ephemeral-storage 请求/限制及 PVC 容量
func sumEphemeralStorageResources(pod *corev1.Pod, client *kubernetes.Clientset, namespace string) (requests, limits int64) {
	for _, container := range pod.Spec.Containers {
		if req, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
			requests += req.Value()
		}
		if lim, ok := container.Resources.Limits[corev1.ResourceEphemeralStorage]; ok {
			limits += lim.Value()
		}
	}

	ctx := context.Background()
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, volume.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
		if err != nil {
			continue
		}
		if storage, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			requests += storage.Value()
		}
		if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			limits += storage.Value()
		}
	}

	if limits == 0 && requests > 0 {
		limits = requests
	}
	return requests, limits
}

func containerStorageQuota(
	pod *corev1.Pod,
	containerName string,
	client *kubernetes.Clientset,
	namespace string,
) (requests, limits int64) {
	if pod == nil {
		return 0, 0
	}

	for _, container := range pod.Spec.Containers {
		if container.Name != containerName {
			continue
		}
		if req, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
			requests += req.Value()
		}
		if lim, ok := container.Resources.Limits[corev1.ResourceEphemeralStorage]; ok {
			limits += lim.Value()
		}

		ctx := context.Background()
		for _, mount := range container.VolumeMounts {
			for _, volume := range pod.Spec.Volumes {
				if volume.Name != mount.Name || volume.PersistentVolumeClaim == nil {
					continue
				}
				pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(
					ctx, volume.PersistentVolumeClaim.ClaimName, metav1.GetOptions{},
				)
				if err != nil {
					continue
				}
				if storage, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
					requests += storage.Value()
				}
				if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
					limits += storage.Value()
				}
			}
		}
		break
	}

	if limits == 0 && requests > 0 {
		limits = requests
	}
	return requests, limits
}
