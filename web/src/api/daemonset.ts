import api from './axios';

export interface DaemonSetInfo {
  name: string;
  namespace: string;
  desiredNumberScheduled: number;
  currentNumberScheduled: number;
  numberReady: number;
  numberAvailable: number;
  updateStrategy: string;
  creationTime: string;
  labels?: Record<string, string>;
  selector?: Record<string, string>;
  containerCount: number;
  images: string[];
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listDaemonSets = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/daemonsets`
    : `/clusters/${clusterName}/daemonsets`;
  return api.get<ApiResponse<{ daemonsets: DaemonSetInfo[] }>>(path);
};

export const getDaemonSet = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ daemonset: DaemonSetInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/daemonsets/${name}`,
  );

export const getDaemonSetPods = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ pods: unknown[] }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/daemonsets/${name}/pods`,
  );

export const createDaemonSet = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/daemonsets`, data);

export const updateDaemonSet = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/daemonsets/${name}`, data);

export const deleteDaemonSet = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/daemonsets/${name}`);
