import api from './axios';

export interface PDBInfo {
  name: string;
  namespace: string;
  minAvailable?: string;
  maxUnavailable?: string;
  selector?: Record<string, string>;
  currentHealthy: number;
  desiredHealthy: number;
  disruptionsAllowed: number;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listPDBs = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/pdbs`
    : `/clusters/${clusterName}/pdbs`;
  return api.get<ApiResponse<{ pdbs: PDBInfo[] }>>(path);
};

export const getPDB = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ pdb: PDBInfo }>>(`/clusters/${clusterName}/namespaces/${namespace}/pdbs/${name}`);

export const createPDB = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/pdbs`, data);

export const updatePDB = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/pdbs/${name}`, data);

export const deletePDB = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/pdbs/${name}`);
