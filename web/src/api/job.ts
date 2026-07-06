import api from './axios';

export interface JobInfo {
  name: string;
  namespace: string;
  completions?: number;
  parallelism?: number;
  succeeded: number;
  failed: number;
  active: number;
  startTime?: string;
  completionTime?: string;
  creationTime: string;
  labels?: Record<string, string>;
  images?: string[];
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listJobs = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/jobs`
    : `/clusters/${clusterName}/jobs`;
  return api.get<ApiResponse<{ jobs: JobInfo[] }>>(path);
};

export const getJob = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ job: JobInfo }>>(`/clusters/${clusterName}/namespaces/${namespace}/jobs/${name}`);

export const createJob = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/jobs`, data);

export const deleteJob = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/jobs/${name}`);
