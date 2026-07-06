import api from './axios';

export interface PVInfo {
  name: string;
  status: string;
  capacity?: string;
  accessModes?: string[];
  reclaimPolicy?: string;
  storageClassName?: string;
  claimRef?: { namespace: string; name: string };
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listPVs = (clusterName: string) =>
  api.get<ApiResponse<{ pvs: PVInfo[] }>>(`/clusters/${clusterName}/pvs`);

export const getPV = (clusterName: string, name: string) =>
  api.get<ApiResponse<{ pv: PVInfo }>>(`/clusters/${clusterName}/pvs/${name}`);
