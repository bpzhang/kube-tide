import api from './axios';

export interface NodeResponse {
  code: number;
  message: string;
  data: {
    nodes: any[];
    pagination?: {
      total: number;
      page: number;
      limit: number;
      pages: number;
    }
  };
}

export interface NodeMetricsResponse {
  code: number;
  message: string;
  data: {
    metrics: {
      cpu_capacity: string;
      memory_capacity: string;
      cpu_allocatable: string;
      memory_allocatable: string;
      cpu_requests: string;
      memory_requests: string;
      cpu_limits: string;
      memory_limits: string;
      cpu_usage: string;
      memory_usage: string;
    };
  };
}

export interface OperationResponse {
  code: number;
  message: string;
  data: {
    message: string;
  };
}

export interface DrainNodeParams {
  gracePeriodSeconds?: number;
  deleteLocalData?: boolean;
  ignoreDaemonSets?: boolean;
}

export interface NodeTaint {
  key: string;
  value?: string;
  effect: string; // 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute'
}

export interface NodeLabel {
  key: string;
  value: string;
}

export interface TaintsResponse {
  code: number;
  message: string;
  data: {
    taints: NodeTaint[];
  };
}

export interface LabelsResponse {
  code: number;
  message: string;
  data: {
    labels: { [key: string]: string };
  };
}

export interface AddNodeRequest {
  name: string;
  ip: string;
  nodePool?: string;
  role?: string;
  labels?: { [key: string]: string };
  taints?: Array<{
    key: string;
    value?: string;
    effect: string;
  }>;
  sshPort?: number;
  sshUser?: string;
  authType?: string; // "key" 或 "password"
  sshKeyFile?: string;
  sshPassword?: string;
}

export interface RemoveNodeRequest {
  force: boolean;
}

export const getNodeList = (clusterName: string, page: number = 1, limit: number = 10) => {
  return api.get<NodeResponse>(`/clusters/${clusterName}/nodes?page=${page}&limit=${limit}`);
};

export const getNodeDetails = (clusterName: string, nodeName: string) => {
  return api.get<NodeResponse>(`/clusters/${clusterName}/nodes/${nodeName}`);
};

export const getNodeMetrics = (clusterName: string, nodeName: string) => {
  return api.get<NodeMetricsResponse>(`/clusters/${clusterName}/nodes/${nodeName}/metrics`);
};

// 节点排水操作
export const drainNode = (clusterName: string, nodeName: string, params: DrainNodeParams) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}/drain`, params);
};

// 禁止节点调度
export const cordonNode = (clusterName: string, nodeName: string) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}/cordon`);
};

// 启用节点调度
export const uncordonNode = (clusterName: string, nodeName: string) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}/uncordon`);
};

// 污点管理
export const getNodeTaints = (clusterName: string, nodeName: string) => {
  return api.get<TaintsResponse>(`/clusters/${clusterName}/nodes/${nodeName}/taints`);
};

export const addNodeTaint = (clusterName: string, nodeName: string, taint: NodeTaint) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}/taints`, taint);
};

export const removeNodeTaint = (clusterName: string, nodeName: string, key: string, effect: string) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}/taints`, {
    data: { key, effect }
  });
};

// 标签管理
export const getNodeLabels = (clusterName: string, nodeName: string) => {
  return api.get<LabelsResponse>(`/clusters/${clusterName}/nodes/${nodeName}/labels`);
};

export const addNodeLabel = (clusterName: string, nodeName: string, key: string, value: string) => {
  return api.post<OperationResponse>(
    `/clusters/${clusterName}/nodes/${nodeName}/labels`,
    { key, value }
  );
};

export const removeNodeLabel = (clusterName: string, nodeName: string, key: string) => {
  return api.delete<OperationResponse>(
    `/clusters/${clusterName}/nodes/${nodeName}/labels`,
    { data: { key } }
  );
};

// 添加新节点
export const addNode = (clusterName: string, params: AddNodeRequest) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodes`, params);
};

// 移除节点
export const removeNode = (clusterName: string, nodeName: string, params: RemoveNodeRequest) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/nodes/${nodeName}`, {
    data: params
  });
};

// 获取节点上运行的Pod列表
export const getNodePods = (clusterName: string, nodeName: string) => {
  return api.get<any>(`/clusters/${clusterName}/nodes/${nodeName}/pods`);
};