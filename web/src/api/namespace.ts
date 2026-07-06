import api from './axios';

export interface NamespaceInfo {
  name: string;
  status: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  creationTime: string;
}

export interface NamespaceListResponse {
  code: number;
  message: string;
  data: {
    namespaces: string[];
    items?: NamespaceInfo[];
  };
}

export const getNamespaceList = (clusterName: string) =>
  api.get<NamespaceListResponse>(`/clusters/${clusterName}/namespaces`);

export const getNamespace = (clusterName: string, namespace: string) =>
  api.get<{ code: number; message: string; data: { namespace: NamespaceInfo } }>(
    `/clusters/${clusterName}/namespaces/${namespace}`,
  );

export const createNamespace = (
  clusterName: string,
  data: { name: string; labels?: Record<string, string>; annotations?: Record<string, string> },
) => api.post(`/clusters/${clusterName}/namespaces`, data);

export const deleteNamespace = (clusterName: string, namespace: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}`);

export const patchNamespaceLabels = (clusterName: string, namespace: string, labels: Record<string, string>) =>
  api.patch(`/clusters/${clusterName}/namespaces/${namespace}/labels`, { labels });
