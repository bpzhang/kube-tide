import api from './axios';

export interface DeploymentResponse {
  code: number;
  message: string;
  data: {
    deployments: Array<{
      name: string;
      namespace: string;
      replicas: number;
      readyReplicas: number;
      strategy: string;
      creationTime: string;
      labels: { [key: string]: string };
      selector: { [key: string]: string };
      containerCount: number;
      images: string[];
    }>;
  };
}

export interface DeploymentDetailResponse {
  code: number;
  message: string;
  data: {
    deployment: {
      name: string;
      namespace: string;
      replicas: number;
      readyReplicas: number;
      strategy: string;
      creationTime: string;
      labels: { [key: string]: string };
      selector: { [key: string]: string };
      containerCount: number;
      images: string[];
      annotations: { [key: string]: string };
      containers: Array<{
        name: string;
        image: string;
        resources: any;
        ports: any[];
        env: any[];
        livenessProbe?: {
          httpGet?: {
            path: string;
            port: number | string;
            scheme: string;
          };
          tcpSocket?: {
            port: number | string;
          };
          exec?: {
            command: string[];
          };
          initialDelaySeconds?: number;
          timeoutSeconds?: number;
          periodSeconds?: number;
          successThreshold?: number;
          failureThreshold?: number; 
        };
        readinessProbe?: {
          httpGet?: {
            path: string;
            port: number | string;
            scheme: string;
          };
          tcpSocket?: {
            port: number | string;
          };
          exec?: {
            command: string[];
          };
          initialDelaySeconds?: number;
          timeoutSeconds?: number;
          periodSeconds?: number;
          successThreshold?: number;
          failureThreshold?: number;
        };
        startupProbe?: {
          httpGet?: {
            path: string;
            port: number | string;
            scheme: string;
          };
          tcpSocket?: {
            port: number | string;
          };
          exec?: {
            command: string[];
          };
          initialDelaySeconds?: number;
          timeoutSeconds?: number;
          periodSeconds?: number;
          successThreshold?: number;
          failureThreshold?: number;
        };
      }>;
      conditions: Array<{
        type: string;
        status: string;
        lastUpdateTime: string;
        lastTransitionTime: string;
        reason: string;
        message: string;
      }>;
      minReadySeconds?: number;
      revisionHistoryLimit?: number;
      paused: boolean;
    };
  };
}

/**
 * get all Deployments list
 * @param clusterName cluster name
 * @returns Deployments list response
 */
export const listAllDeployments = (clusterName: string) => {
  return api.get<DeploymentResponse>(`/clusters/${clusterName}/deployments`);
};

/**
 * get Deployments list by namespace
 * @param clusterName cluster name
 * @param namespace namespace
 * @returns Deployments list response
 */
export const listDeploymentsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<DeploymentResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments`);
};

/**
 * get Deployment details
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @returns Deployment details response
 */
export const getDeploymentDetails = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<DeploymentDetailResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}`);
};

/**
 * scale Deployment replicas
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @param replicas replicas count
 * @returns operation result
 */
export const scaleDeployment = (clusterName: string, namespace: string, deploymentName: string, replicas: number) => {
  return api.put(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/scale`, { replicas });
};

/**
 * restart Deployment
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @returns operation result
 */
export const restartDeployment = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/restart`);
};

/**
 * update Deployment configuration
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @param updateData update data
 * @returns operation result
 */
export interface UpdateDeploymentRequest {
  replicas?: number;
  image?: { [containerName: string]: string };
  env?: { [containerName: string]: Array<{ name: string; value: string; valueFrom?: any }> };
  resources?: { 
    [containerName: string]: { 
      limits?: { [key: string]: string }; 
      requests?: { [key: string]: string } 
    } 
  };
  livenessProbe?: { [containerName: string]: any };
  readinessProbe?: { [containerName: string]: any };
  startupProbe?: { [containerName: string]: any };
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
  strategy?: {
    type: 'RollingUpdate' | 'Recreate';
    rollingUpdate?: {
      maxSurge?: string;
      maxUnavailable?: string;
    };
  };
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
  paused?: boolean;
  volumes?: Array<{
    name: string;
    configMap?: {
      name: string;
      items?: Array<{
        key: string;
        path: string;
        mode?: string;
      }>;
      defaultMode?: number;
    };
    secret?: {
      secretName: string;
      items?: Array<{
        key: string;
        path: string;
        mode?: string;
      }>;
      defaultMode?: number;
    };
    persistentVolumeClaim?: {
      claimName: string;
      readOnly?: boolean;
    };
    emptyDir?: {
      medium?: string;
      sizeLimit?: string;
    };
    hostPath?: {
      path: string;
      type?: string;
    };
  }>;
  volumeMounts?: {
    [containerName: string]: Array<{
      name: string;
      mountPath: string;
      subPath?: string;
      readOnly?: boolean;
    }>;
  };
  nodeSelector?: { [key: string]: string };
  affinity?: {
    nodeAffinity?: {
      requiredDuringSchedulingIgnoredDuringExecution?: {
        nodeSelectorTerms: Array<{
          matchExpressions?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
          matchFields?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
        }>;
      };
      preferredDuringSchedulingIgnoredDuringExecution?: Array<{
        weight: number;
        preference: {
          matchExpressions?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
          matchFields?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
        };
      }>;
    };
  };
}

export const updateDeployment = (
  clusterName: string, 
  namespace: string, 
  deploymentName: string, 
  updateData: UpdateDeploymentRequest
) => {
  return api.put(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}`, updateData);
};

/**
 * create a new Deployment
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentData deployment data
 * @returns operation result
 */
export interface CreateDeploymentRequest {
  name: string;
  clusterName?: string;
  namespace?: string;
  replicas?: number;
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
  paused?: boolean;
  strategy?: {
    type: 'RollingUpdate' | 'Recreate';
    rollingUpdate?: {
      maxSurge?: string;
      maxUnavailable?: string;
    };
  };
  containers: Array<{
    name: string;
    image: string;
    command?: string[];
    args?: string[];
    workingDir?: string;
    ports?: Array<{
      name?: string;
      containerPort: number;
      hostPort?: number;
      protocol?: string;
    }>;
    env?: Array<{
      name: string;
      value?: string;
      valueFrom?: {
        configMapKeyRef?: {
          name: string;
          key: string;
        };
        secretKeyRef?: {
          name: string;
          key: string;
        };
      };
    }>;
    resources?: {
      limits?: { [key: string]: string };
      requests?: { [key: string]: string };
    };
    volumeMounts?: Array<{
      name: string;
      mountPath: string;
      subPath?: string;
      readOnly?: boolean;
    }>;
    livenessProbe?: any;
    readinessProbe?: any;
    startupProbe?: any;
    imagePullPolicy?: string;
    securityContext?: any;
  }>;
  volumes?: Array<{
    name: string;
    configMap?: {
      name: string;
      items?: Array<{
        key: string;
        path: string;
        mode?: string;
      }>;
      defaultMode?: number;
    };
    secret?: {
      secretName: string;
      items?: Array<{
        key: string;
        path: string;
        mode?: string;
      }>;
      defaultMode?: number;
    };
    persistentVolumeClaim?: {
      claimName: string;
      readOnly?: boolean;
    };
    emptyDir?: {
      medium?: string;
      sizeLimit?: string;
    };
    hostPath?: {
      path: string;
      type?: string;
    };
  }>;
  nodeSelector?: { [key: string]: string };
  affinity?: {
    nodeAffinity?: {
      requiredDuringSchedulingIgnoredDuringExecution?: {
        nodeSelectorTerms: Array<{
          matchExpressions?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
          matchFields?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
        }>;
      };
      preferredDuringSchedulingIgnoredDuringExecution?: Array<{
        weight: number;
        preference: {
          matchExpressions?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
          matchFields?: Array<{
            key: string;
            operator: string;
            values?: string[];
          }>;
        };
      }>;
    };
  };
  serviceAccountName?: string;
}

export const createDeployment = (
  clusterName: string,
  namespace: string,
  deploymentData: CreateDeploymentRequest
) => {
  return api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments`, deploymentData);
};

export interface DeploymentEventsResponse {
  code: number;
  message: string;
  data: {
    events: any[];
  };
}

/**
 * get Deployment events
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @returns Deployment events list response
 */
export const getDeploymentEvents = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<DeploymentEventsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/events`);
};

export interface AllDeploymentEventsResponse {
  code: number;
  message: string;
  data: {
    events: {
      deployment: any[];
      replicaSet: any[];
      pod: any[];
    };
  };
}

/**
 * get all events related to Deployment and its associated resources (ReplicaSet and Pod)
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @returns all related events response
 */
export const getAllDeploymentEvents = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<AllDeploymentEventsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/all-events`);
};