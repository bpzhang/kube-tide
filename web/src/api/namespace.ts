import axios from './axios';

export interface NamespaceListResponse {
  code: number;
  message: string;
  data: {
    namespaces: string[];
  };
}

/**
 * get namespace list
 * @param clusterName cluster name
 * @returns namespace list response
 */
export const getNamespaceList = (clusterName: string) => {
  return axios.get<NamespaceListResponse>(`/clusters/${clusterName}/namespaces`);
};