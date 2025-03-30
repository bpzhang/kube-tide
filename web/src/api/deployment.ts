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
 * 获取集群所有Deployment列表
 * @param clusterName 集群名称
 * @returns Deployments列表响应
 */
export const listAllDeployments = (clusterName: string) => {
  return api.get<DeploymentResponse>(`/clusters/${clusterName}/deployments`);
};

/**
 * 获取指定命名空间的Deployment列表
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @returns Deployments列表响应
 */
export const listDeploymentsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<DeploymentResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments`);
};

/**
 * 获取Deployment详情
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentName Deployment名称
 * @returns Deployment详情响应
 */
export const getDeploymentDetails = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<DeploymentDetailResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}`);
};

/**
 * 调整Deployment副本数
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentName Deployment名称
 * @param replicas 副本数
 * @returns 操作结果
 */
export const scaleDeployment = (clusterName: string, namespace: string, deploymentName: string, replicas: number) => {
  return api.put(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/scale`, { replicas });
};

/**
 * 重启Deployment
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentName Deployment名称
 * @returns 操作结果
 */
export const restartDeployment = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/restart`);
};

/**
 * 更新Deployment配置
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentName Deployment名称
 * @param updateData 更新数据
 * @returns 操作结果
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
 * 创建Deployment
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentData 部署数据
 * @returns 操作结果
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
  serviceAccountName?: string;
  hostNetwork?: boolean;
  dnsPolicy?: string;
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
 * 获取Deployment相关的事件
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param deploymentName Deployment名称
 * @returns Deployment事件列表响应
 */
export const getDeploymentEvents = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<DeploymentEventsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/events`);
};