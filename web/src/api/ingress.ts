import api from './axios';

export interface IngressRule {
  host?: string;
  paths: Array<{
    path?: string;
    pathType?: string;
    backend: {
      serviceName?: string;
      servicePort?: string;
    };
  }>;
}

export interface IngressInfo {
  name: string;
  namespace: string;
  ingressClassName?: string;
  rules: IngressRule[];
  tls: Array<{
    hosts?: string[];
    secretName?: string;
  }>;
}

export interface IngressListResponse {
  code: number;
  message: string;
  data: {
    ingresses: IngressInfo[];
  };
}

export const getIngressesByNamespace = (clusterName: string, namespace: string) =>
  api.get<IngressListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/ingresses`);

export const getIngress = (clusterName: string, namespace: string, name: string) =>
  api.get<{ code: number; message: string; data: { ingress: IngressInfo } }>(
    `/clusters/${clusterName}/namespaces/${namespace}/ingresses/${name}`,
  );

export const createIngress = (clusterName: string, namespace: string, data: Record<string, unknown>) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/ingresses`, data);

export const updateIngress = (clusterName: string, namespace: string, name: string, data: Record<string, unknown>) =>
  api.put(`/clusters/${clusterName}/namespaces/${namespace}/ingresses/${name}`, data);

export const deleteIngress = (clusterName: string, namespace: string, name: string) =>
  api.delete(`/clusters/${clusterName}/namespaces/${namespace}/ingresses/${name}`);
