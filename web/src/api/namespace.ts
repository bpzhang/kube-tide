import axios from './axios';

export interface NamespaceListResponse {
  code: number;
  message: string;
  data: {
    namespaces: string[];
  };
}

/**
 * 获取指定集群的命名空间列表
 * @param clusterName 集群名称
 * @returns 命名空间列表响应
 */
export const getNamespaceList = (clusterName: string) => {
  return axios.get<NamespaceListResponse>(`/clusters/${clusterName}/namespaces`);
};