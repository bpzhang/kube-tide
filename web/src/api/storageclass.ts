import api from './axios';

export interface StorageClassInfo {
  name: string;
  provisioner: string;
  reclaimPolicy?: string;
  volumeBindingMode?: string;
  allowVolumeExpansion?: boolean;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listStorageClasses = (clusterName: string) =>
  api.get<ApiResponse<{ storageclasses: StorageClassInfo[] }>>(`/clusters/${clusterName}/storageclasses`);

export const getStorageClass = (clusterName: string, name: string) =>
  api.get<ApiResponse<{ storageclass: StorageClassInfo }>>(`/clusters/${clusterName}/storageclasses/${name}`);
