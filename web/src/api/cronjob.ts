import api from './axios';

export interface CronJobInfo {
  name: string;
  namespace: string;
  schedule: string;
  suspend: boolean;
  lastScheduleTime?: string;
  lastSuccessfulTime?: string;
  activeJobs: number;
  creationTime: string;
  labels?: Record<string, string>;
  concurrencyPolicy?: string;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listCronJobs = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/cronjobs`
    : `/clusters/${clusterName}/cronjobs`;
  return api.get<ApiResponse<{ cronjobs: CronJobInfo[] }>>(path);
};

export const getCronJob = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ cronjob: CronJobInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/cronjobs/${name}`,
  );

export const createCronJob = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/cronjobs`, data);

export const updateCronJob = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/cronjobs/${name}`, data);

export const suspendCronJob = (clusterName: string, namespace: string, name: string, suspend: boolean) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/cronjobs/${name}/suspend`, { suspend });

export const deleteCronJob = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/cronjobs/${name}`);
