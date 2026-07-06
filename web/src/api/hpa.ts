import api from './axios';

export interface HPATargetRef {
  kind: string;
  name: string;
}

export interface HPAMetricInfo {
  type: string;
  name?: string;
  average?: string;
  utilization?: number;
}

export interface HPAInfo {
  name: string;
  namespace: string;
  minReplicas?: number;
  maxReplicas: number;
  currentReplicas: number;
  desiredReplicas: number;
  targetRef: HPATargetRef;
  metrics?: HPAMetricInfo[];
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listHPAs = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/hpas`
    : `/clusters/${clusterName}/hpas`;
  return api.get<ApiResponse<{ hpas: HPAInfo[] }>>(path);
};

export const getHPA = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ hpa: HPAInfo }>>(`/clusters/${clusterName}/namespaces/${namespace}/hpas/${name}`);

export const createHPA = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/hpas`, data);

export const updateHPA = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/hpas/${name}`, data);

export const deleteHPA = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/hpas/${name}`);
