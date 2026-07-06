import api from './axios';

export interface NetworkPolicyInfo {
  name: string;
  namespace: string;
  policyTypes?: string[];
  podSelector?: Record<string, string>;
  ingressRuleCount: number;
  egressRuleCount: number;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listNetworkPolicies = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/networkpolicies`
    : `/clusters/${clusterName}/networkpolicies`;
  return api.get<ApiResponse<{ networkpolicies: NetworkPolicyInfo[] }>>(path);
};

export const getNetworkPolicy = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ networkpolicy: NetworkPolicyInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/networkpolicies/${name}`,
  );

export const createNetworkPolicy = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/networkpolicies`, data);

export const updateNetworkPolicy = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/networkpolicies/${name}`, data);

export const deleteNetworkPolicy = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/networkpolicies/${name}`);
