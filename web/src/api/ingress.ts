import api from './axios';

export interface IngressListResponse {
  code: number;
  message: string;
  data: {
    ingresses: Array<{
      name: string;
      namespace: string;
      ingressClassName?: string;
      rules: Array<{
        host?: string;
        paths: Array<{
          path?: string;
          pathType?: string;
          backend: {
            serviceName?: string;
            servicePort?: string;
          };
        }>;
      }>;
      tls: Array<{
        hosts?: string[];
        secretName?: string;
      }>;
    }>;
  };
}

export const getIngressesByNamespace = (clusterName: string, namespace: string) => {
  return api.get<IngressListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/ingresses`);
};