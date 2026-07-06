import api from './axios';

export interface LimitRangeInfo {
  name: string;
  namespace: string;
  limits?: Array<{
    type: string;
    default?: Record<string, string>;
    defaultRequest?: Record<string, string>;
    max?: Record<string, string>;
    min?: Record<string, string>;
  }>;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listLimitRanges = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/limitranges`
    : `/clusters/${clusterName}/limitranges`;
  return api.get<ApiResponse<{ limitranges: LimitRangeInfo[] }>>(path);
};

export const getLimitRange = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ limitrange: LimitRangeInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/limitranges/${name}`,
  );

export const createLimitRange = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/limitranges`, data);

export const updateLimitRange = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/limitranges/${name}`, data);

export const deleteLimitRange = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/limitranges/${name}`);
