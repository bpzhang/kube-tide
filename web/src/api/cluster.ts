import api from './axios';

interface Cluster {
  name: string;
  kubeconfigPath: string;
}

export interface ClusterResponse {
  code: number;
  message: string;
  data: {
    clusters: string[];
  };
}

export interface ClusterDetail {
  name: string;
  version: string;
  totalNodes: number;
  totalNamespaces: number;
  namespaces: any[];
  totalCPU: number;
  totalMemory: string;
  platform: string;
}

export interface ClusterDetailResponse {
  code: number;
  message: string;
  data: {
    cluster: ClusterDetail;
  };
}

export interface ClusterMetrics {
  timestamp: string;
  cpuUsage: number;       // CPU使用率（百分比）
  memoryUsage: number;    // 内存使用率（百分比）
  cpuRequestsPercentage: number;   // CPU请求百分比
  cpuLimitsPercentage: number;     // CPU限制百分比
  memoryRequestsPercentage: number; // 内存请求百分比
  memoryLimitsPercentage: number;   // 内存限制百分比
  podCount: number;       // Pod总数
  nodeCounts: {
    ready: number;
    notReady: number;
  };
  deploymentReadiness: {
    available: number;
    total: number;
  };
  // 过去24小时的监控数据点（每小时一个数据点）
  historicalData: {
    cpuUsage: Array<{ timestamp: string; value: number }>;
    memoryUsage: Array<{ timestamp: string; value: number }>;
    podCount: Array<{ timestamp: string; value: number }>;
  };
}

export interface ClusterMetricsResponse {
  code: number;
  message: string;
  data: {
    metrics: ClusterMetrics;
  };
}

export interface ClusterEventsResponse {
  code: number;
  message: string;
  data: {
    events: any[];
  };
}

export const getClusterList = () => {
  return api.get<ClusterResponse>('/clusters');
};

export const getClusterDetails = (clusterName: string) => {
  return api.get<ClusterDetailResponse>(`/clusters/${clusterName}`);
};

export const addCluster = (cluster: Cluster) => {
  return api.post<{code: number; message: string}>('/clusters', cluster);
};

export const removeCluster = (clusterName: string) => {
  return api.delete<{code: number; message: string}>(`/clusters/${clusterName}`);
};

export const testClusterConnection = (clusterName: string) => {
  return api.get<{code: number; message: string; data: {status: string}}>(`/clusters/${clusterName}/test`);
};

export const getClusterMetrics = (clusterName: string) => {
  return api.get<ClusterMetricsResponse>(`/clusters/${clusterName}/metrics`);
};

export const getClusterNamespaces = (clusterName: string) => {
  return api.get<{code: number; message: string; data: {namespaces: string[]}}>(`/clusters/${clusterName}/namespaces`);
};

export const getClusterEvents = (clusterName: string) => {
  return api.get<ClusterEventsResponse>(`/clusters/${clusterName}/events`);
};