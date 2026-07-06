import api from './axios';

export interface ConfigMapInfo {
  name: string;
  namespace: string;
  labels?: Record<string, string>;
  dataKeys: string[];
  creationTime: string;
}

export interface ConfigMapDetail extends ConfigMapInfo {
  data: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listConfigMaps = (clusterName: string) =>
  api.get<ApiResponse<{ configmaps: ConfigMapInfo[] }>>(`/clusters/${clusterName}/configmaps`);

export const listConfigMapsByNamespace = (clusterName: string, namespace: string) =>
  api.get<ApiResponse<{ configmaps: ConfigMapInfo[] }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/configmaps`,
  );

export const getConfigMap = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ configmap: ConfigMapDetail }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/configmaps/${name}`,
  );

export const createConfigMap = (
  clusterName: string,
  namespace: string,
  data: { name: string; labels?: Record<string, string>; data: Record<string, string> },
) => api.post(`/clusters/${clusterName}/namespaces/${namespace}/configmaps`, data);

export const updateConfigMap = (
  clusterName: string,
  namespace: string,
  name: string,
  data: { data: Record<string, string>; labels?: Record<string, string> },
) => api.put(`/clusters/${clusterName}/namespaces/${namespace}/configmaps/${name}`, data);

export const deleteConfigMap = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/configmaps/${name}`);
