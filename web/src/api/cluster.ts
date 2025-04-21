import api from './axios';

interface Cluster {
  name: string;
  kubeconfigPath?: string;
  kubeconfigContent?: string;
  addType?: 'path' | 'content'; // addType: 'path' (file path) or 'content' (content)
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
  addType?: 'path' | 'content' | 'unknown'; // addType: 'path' (file path) or 'content' (content) or 'unknown'
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
  cpuUsage: number;       // CPU usage (percentage)
  memoryUsage: number;    // Memory usage (percentage)
  cpuRequestsPercentage: number;   // CPU requests percentage
  cpuLimitsPercentage: number;     // CPU limits percentage
  memoryRequestsPercentage: number; // Memory requests percentage
  memoryLimitsPercentage: number;   // Memory limits percentage
  podCount: number;       // Total pods
  nodeCounts: {
    ready: number;
    notReady: number;
  };
  deploymentReadiness: {
    available: number;
    total: number;
  };
  // monitoring data for the last 24 hours
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

export interface ClusterAddTypeResponse {
  code: number;
  message: string;
  data: {
    name: string;
    addType: 'path' | 'content' | 'unknown';
  };
}

export const getClusterAddType = (clusterName: string) => {
  return api.get<ClusterAddTypeResponse>(`/clusters/${clusterName}/add-type`);
};