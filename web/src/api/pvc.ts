import api from './axios';

export interface PVCInfo {
  name: string;
  namespace: string;
  status: string;
  volumeName?: string;
  storageClassName?: string;
  accessModes?: string[];
  capacity?: string;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listPVCs = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/pvcs`
    : `/clusters/${clusterName}/pvcs`;
  return api.get<ApiResponse<{ pvcs: PVCInfo[] }>>(path);
};

export const getPVC = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ pvc: PVCInfo }>>(`/clusters/${clusterName}/namespaces/${namespace}/pvcs/${name}`);

export const createPVC = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/pvcs`, data);

export const deletePVC = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/pvcs/${name}`);
