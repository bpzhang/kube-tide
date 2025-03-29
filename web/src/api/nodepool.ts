import api from './axios';

export interface NodePool {
  name: string;
  labels?: { [key: string]: string };
  taints?: Array<{
    key: string;
    value?: string;
    effect: string;
  }>;
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

// 获取节点池列表
export const getNodePools = (clusterName: string) => {
  return api.get<NodePoolResponse>(`/clusters/${clusterName}/nodepools`);
};

// 获取节点池详情
export const getNodePool = (clusterName: string, poolName: string) => {
  return api.get<NodePoolDetailResponse>(`/clusters/${clusterName}/nodepools/${poolName}`);
};

// 创建节点池
export const createNodePool = (clusterName: string, pool: NodePool) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/nodepools`, pool);
};

// 更新节点池
export const updateNodePool = (clusterName: string, poolName: string, pool: NodePool) => {
  return api.put<OperationResponse>(`/clusters/${clusterName}/nodepools/${poolName}`, pool);
};

// 删除节点池
export const deleteNodePool = (clusterName: string, poolName: string) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/nodepools/${poolName}`);
};