import api from './axios';

export interface RBACRoleRef {
  kind: string;
  name: string;
  apiGroup?: string;
}

export interface RBACSubject {
  kind: string;
  name: string;
  namespace?: string;
}

export interface RoleInfo {
  name: string;
  namespace?: string;
  ruleCount: number;
  creationTime: string;
  labels?: Record<string, string>;
}

export interface RoleBindingInfo {
  name: string;
  namespace?: string;
  roleRef: RBACRoleRef;
  subjectCount: number;
  subjects?: RBACSubject[];
  creationTime: string;
  labels?: Record<string, string>;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const listRoles = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/roles`
    : `/clusters/${clusterName}/roles`;
  return api.get<ApiResponse<{ roles: RoleInfo[] }>>(path);
};

export const getRole = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ role: RoleInfo }>>(`/clusters/${clusterName}/namespaces/${namespace}/roles/${name}`);

export const listClusterRoles = (clusterName: string) =>
  api.get<ApiResponse<{ clusterroles: RoleInfo[] }>>(`/clusters/${clusterName}/clusterroles`);

export const getClusterRole = (clusterName: string, name: string) =>
  api.get<ApiResponse<{ clusterrole: RoleInfo }>>(`/clusters/${clusterName}/clusterroles/${name}`);

export const listRoleBindings = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/rolebindings`
    : `/clusters/${clusterName}/rolebindings`;
  return api.get<ApiResponse<{ rolebindings: RoleBindingInfo[] }>>(path);
};

export const getRoleBinding = (clusterName: string, namespace: string, name: string) =>
  api.get<ApiResponse<{ rolebinding: RoleBindingInfo }>>(
    `/clusters/${clusterName}/namespaces/${namespace}/rolebindings/${name}`,
  );

export const createRoleBinding = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/rolebindings`, data);

export const deleteRoleBinding = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/rolebindings/${name}`);

export const listClusterRoleBindings = (clusterName: string) =>
  api.get<ApiResponse<{ clusterrolebindings: RoleBindingInfo[] }>>(`/clusters/${clusterName}/clusterrolebindings`);

export const getClusterRoleBinding = (clusterName: string, name: string) =>
  api.get<ApiResponse<{ clusterrolebinding: RoleBindingInfo }>>(
    `/clusters/${clusterName}/clusterrolebindings/${name}`,
  );

export const createClusterRoleBinding = (clusterName: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/clusterrolebindings`, data);

export const deleteClusterRoleBinding = (clusterName: string, name: string) =>
  api.delete(`/clusters/${clusterName}/clusterrolebindings/${name}`);
