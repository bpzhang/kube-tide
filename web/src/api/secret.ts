import api from './axios';

export interface SecretInfo {
  name: string;
  namespace: string;
  type: string;
  labels?: Record<string, string>;
  dataKeys: string[];
  creationTime: string;
}

export interface SecretDetail extends SecretInfo {
  data: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listSecrets = (clusterName: string) =>
  api.get<ApiResponse<{ secrets: SecretInfo[] }>>(`/clusters/${clusterName}/secrets`);

export const listSecretsByNamespace = (clusterName: string, namespace: string) =>
  api.get<ApiResponse<{ secrets: SecretInfo[] }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/secrets`,
  );

export const getSecret = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ secret: SecretDetail }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/secrets/${name}`,
  );

export const createSecret = (
  clusterName: string,
  namespace: string,
  data: { name: string; type?: string; labels?: Record<string, string>; stringData: Record<string, string> },
) => api.post(`/clusters/${clusterName}/namespaces/${namespace}/secrets`, data);

export const updateSecret = (
  clusterName: string,
  namespace: string,
  name: string,
  data: { stringData: Record<string, string>; labels?: Record<string, string> },
) => api.put(`/clusters/${clusterName}/namespaces/${namespace}/secrets/${name}`, data);

export const deleteSecret = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/secrets/${name}`);
