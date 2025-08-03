import api from './axios';

export interface NodePool {
  name: string;
  labels?: { [key: string]: string };
  taints?: Array<{
    key: string;
    value?: string;
    effect: string;
  }>;
  autoScaling?: AutoScalingConfig;
}

export interface AutoScalingConfig {
  enabled: boolean;
  minNodes: number;
  maxNodes: number;
  scaleDownDelay?: string;
  scaleDownThreshold?: string;
  scaleUpThreshold?: string;
  scaleDownUnneededTime?: string;
  scaleDownDelayAfterAdd?: string;
}

export interface NodePoolResponse {
  code: number;
  message: string;
  data: {
    pools: NodePool[];
  };
}

export interface NodePoolDetailResponse {
  code: number;
  message: string;
  data: {
    pool: NodePool;
  };
}

export interface OperationResponse {
  code: number;
  message: string;
  data?: {
    message: string;
  };
}

// get nodes pool list
export const getNodePools = (clusterName: string) => {
  return api.get<NodePoolResponse>(`/clusters/${clusterName}/nodepools`);
};

// get nodes pool details
export const getNodePool = (clusterName: string, poolName: string) => {
  return api.get<NodePoolDetailResponse>(`/clusters/${clusterName}/nodepools/${poolName}`);
};

// create node pool
export const createNodePool = (clusterName: string, pool: NodePool) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodepools`, pool);
};

// update node pool
export const updateNodePool = (clusterName: string, poolName: string, pool: NodePool) => {
  return api.put<OperationResponse>(`/clusters/${clusterName}/nodepools/${poolName}`, pool);
};

// delete node pool
export const deleteNodePool = (clusterName: string, poolName: string) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/nodepools/${poolName}`);
};