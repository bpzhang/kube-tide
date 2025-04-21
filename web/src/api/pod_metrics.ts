import api from './axios';

export interface PodMetricsResponse {
  code: number;
  message: string;
  data: {
    metrics: PodMetrics;
  };
}

// PodMetrics Pod structure
export interface PodMetrics {
  // currentcpu usage (unit: m)
  cpuUsage: number;
  // current memory usage (percentage)
  memoryUsage: number;
  // current disk usage (percentage)
  diskUsage: number;
  // CPU request value (unit: m)
  cpuRequests: string;
  // CPU limit value (unit: m)
  cpuLimits: string;
  // Memory request value (e.g., 100Mi)
  memoryRequests: string;
  // Memory limit value (e.g., 200Mi)
  memoryLimits: string;
  // Disk request value (e.g., 1Gi) 
  diskRequests: string;
  // Disk limit value (e.g., 10Gi)
  diskLimits: string;
  // history data
  historicalData: {
    cpuUsage: MetricDataPoint[];
    memoryUsage: MetricDataPoint[];
    diskUsage: MetricDataPoint[];
  };
  // containers metrics
  containers: ContainerMetrics[];
}

// MetricDataPoint 
export interface MetricDataPoint {
  timestamp: string;
  value: number;
}

// ContainerMetrics 
export interface ContainerMetrics {
  name: string;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  cpuRequests: string;
  cpuLimits: string;
  memoryRequests: string;
  memoryLimits: string;
  diskRequests: string;
  diskLimits: string;
  // container history data
  historicalData?: {
    cpuUsage: MetricDataPoint[];
    memoryUsage: MetricDataPoint[];
    diskUsage: MetricDataPoint[];
  };
}

// Get Pod CPU and memory metrics
export const getPodMetrics = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodMetricsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/metrics`);
};
