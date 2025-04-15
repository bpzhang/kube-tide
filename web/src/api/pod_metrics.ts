// PodMetricsResponse 定义Pod指标响应
import api from './axios';

export interface PodMetricsResponse {
  code: number;
  message: string;
  data: {
    metrics: PodMetrics;
  };
}

// PodMetrics Pod指标结构
export interface PodMetrics {
  // 当前CPU使用率（百分比）
  cpuUsage: number;
  // 当前内存使用率（百分比）
  memoryUsage: number;
  // 当前硬盘使用率（百分比）
  diskUsage: number;
  // CPU请求值（单位：m）
  cpuRequests: string;
  // CPU限制值（单位：m）
  cpuLimits: string;
  // 内存请求值（例如：100Mi）
  memoryRequests: string;
  // 内存限制值（例如：200Mi）
  memoryLimits: string;
  // 硬盘请求值（例如：1Gi）
  diskRequests: string;
  // 硬盘限制值（例如：10Gi）
  diskLimits: string;
  // 历史数据（24小时内每小时一个数据点）
  historicalData: {
    cpuUsage: MetricDataPoint[];
    memoryUsage: MetricDataPoint[];
    diskUsage: MetricDataPoint[];
  };
  // 容器指标
  containers: ContainerMetrics[];
}

// MetricDataPoint 指标数据点
export interface MetricDataPoint {
  timestamp: string;
  value: number;
}

// ContainerMetrics 容器指标结构
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
  // 容器级别的历史数据
  historicalData?: {
    cpuUsage: MetricDataPoint[];
    memoryUsage: MetricDataPoint[];
    diskUsage: MetricDataPoint[];
  };
}

// 获取Pod的CPU和内存指标
export const getPodMetrics = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodMetricsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/metrics`);
};
