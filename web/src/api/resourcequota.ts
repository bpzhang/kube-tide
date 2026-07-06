import api from './axios';

export interface ResourceQuotaInfo {
  name: string;
  namespace: string;
  hard?: Record<string, string>;
  used?: Record<string, string>;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listResourceQuotas = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/resourcequotas`
    : `/clusters/${clusterName}/resourcequotas`;
  return api.get<ApiResponse<{ resourcequotas: ResourceQuotaInfo[] }>>(path);
};

export const getResourceQuota = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ resourcequota: ResourceQuotaInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/resourcequotas/${name}`,
  );

export const createResourceQuota = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/resourcequotas`, data);

export const updateResourceQuota = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/resourcequotas/${name}`, data);

export const deleteResourceQuota = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/resourcequotas/${name}`);
