import api from './axios';

export interface ServiceResponse {
  code: number;
  message: string;
  data: {
    services: Array<{
      name: string;
      namespace: string;
      type: string;
      clusterIP: string;
      externalIPs: string[];
      ports: Array<{
        port: number;
        targetPort: number;
        protocol: string;
      }>;
      createdAt: string;
    }>;
  };
}

export interface ServiceDetailResponse {
  code: number;
  message: string;
  data: {
    service: {
      name: string;
      namespace: string;
      type: string;
      clusterIP: string;
      externalIPs: string[];
      ports: Array<{
        name?: string;
        port: number;
        targetPort: number;
        protocol: string;
        nodePort?: number;
      }>;
      selector?: { [key: string]: string };
      labels?: { [key: string]: string };
      annotations?: { [key: string]: string };
      createdAt: string;
    };
  };
}

export interface ServicePort {
  name?: string;
  port: number;
  targetPort: number;
  protocol: string;
  nodePort?: number;
}

export interface CreateServiceRequest {
  name: string;
  namespace: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer';
  ports: Array<{
    name?: string;
    port: number;
    targetPort: number;
    protocol: 'TCP' | 'UDP' | 'SCTP';
    nodePort?: number;
  }>;
  selector?: { [key: string]: string };
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
}

export interface UpdateServiceRequest {
  type?: string;
  ports?: ServicePort[];
  selector?: { [key: string]: string };
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
}

export const getServiceList = (clusterName: string) => {
  return api.get<ServiceResponse>(`/clusters/${clusterName}/services`);
};

export const getServicesByNamespace = (clusterName: string, namespace: string) => {
  return api.get<ServiceResponse>(`/clusters/${clusterName}/namespaces/${namespace}/services`);
};

export const getServiceDetails = (clusterName: string, namespace: string, serviceName: string) => {
  return api.get<ServiceDetailResponse>(`/clusters/${clusterName}/namespaces/${namespace}/services/${serviceName}`);
};

export const createService = (clusterName: string, namespace: string,service: CreateServiceRequest) => {
  return api.post<{code: number; message: string}>(`/clusters/${clusterName}/namespaces/${namespace}/services`, service);
};

export const updateService = (clusterName: string, namespace: string, serviceName: string, data: UpdateServiceRequest) => {
  return api.put<{code: number; message: string}>(`/clusters/${clusterName}/namespaces/${namespace}/services/${serviceName}`, data);
};

export const deleteService = (clusterName: string, namespace: string, serviceName: string) => {
  return api.delete<{code: number; message: string}>(`/clusters/${clusterName}/namespaces/${namespace}/services/${serviceName}`);
};