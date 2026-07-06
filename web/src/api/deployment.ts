import api from './axios';

export interface LabelSelectorExpression {
  key: string;
  operator: string;
  values?: string[];
}

export interface LabelSelectorConfig {
  matchLabels?: { [key: string]: string };
  matchExpressions?: LabelSelectorExpression[];
}

export interface PodAffinityTermConfig {
  topologyKey: string;
  namespaces?: string[];
  labelSelector?: LabelSelectorConfig;
}

export interface PodAffinityRuleConfig {
  requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTermConfig[];
  preferredDuringSchedulingIgnoredDuringExecution?: Array<{
    weight: number;
    podAffinityTerm: PodAffinityTermConfig;
  }>;
}

export interface NodeAffinityRuleConfig {
  requiredDuringSchedulingIgnoredDuringExecution?: {
    nodeSelectorTerms: Array<{
      matchExpressions?: LabelSelectorExpression[];
      matchFields?: LabelSelectorExpression[];
    }>;
  };
  preferredDuringSchedulingIgnoredDuringExecution?: Array<{
    weight: number;
    preference: {
      matchExpressions?: LabelSelectorExpression[];
      matchFields?: LabelSelectorExpression[];
    };
  }>;
}

export interface AffinityConfig {
  nodeAffinity?: NodeAffinityRuleConfig;
  podAffinity?: PodAffinityRuleConfig;
  podAntiAffinity?: PodAffinityRuleConfig;
}

export interface TolerationConfig {
  key?: string;
  operator?: string;
  value?: string;
  effect?: string;
  tolerationSeconds?: number;
}

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
      nodeSelector?: { [key: string]: string };
      tolerations?: TolerationConfig[];
      affinity?: AffinityConfig;
      serviceAccountName?: string;
      hostNetwork?: boolean;
      dnsPolicy?: string;
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
  tolerations?: TolerationConfig[];
  affinity?: AffinityConfig;
  serviceAccountName?: string;
  hostNetwork?: boolean;
  dnsPolicy?: string;
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
  tolerations?: TolerationConfig[];
  affinity?: AffinityConfig;
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

// Deployment版本管理相关接口

/**
 * RevisionInfo 版本信息
 */
export interface RevisionInfo {
  revision: number;
  changeReason?: string;
  creationTime: string;
  labels: { [key: string]: string };
  annotations: { [key: string]: string };
  podTemplateSpec: any;
  replicaSetName: string;
  replicas?: number;
  readyReplicas: number;
  availableReplicas: number;
  conditions: any[];
}

/**
 * 获取Deployment版本历史
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @returns version history
 */
export const getDeploymentRolloutHistory = (clusterName: string, namespace: string, deploymentName: string) => {
  return api.get<{ code: number; message: string; data: { revisions: RevisionInfo[] } }>(
    `/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/history`
  );
};

/**
 * 获取指定版本详情
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @param revision revision number
 * @returns revision details
 */
export const getDeploymentRevisionDetails = (clusterName: string, namespace: string, deploymentName: string, revision: number) => {
  return api.get<{ revision: RevisionInfo }>(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/revisions/${revision}`);
};

/**
 * 回滚Deployment到指定版本
 * @param clusterName cluster name
 * @param namespace namespace
 * @param deploymentName Deployment name
 * @param revision revision number (可选，不提供则回滚到上一个版本)
 * @returns operation result
 */
export const rollbackDeployment = (clusterName: string, namespace: string, deploymentName: string, revision?: number) => {
  const data = revision !== undefined ? { revision } : {};
  return api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/rollback`, data);
};

export interface MetricDataPoint {
  timestamp: string;
  value: number;
}

export interface WorkloadAlert {
  level: string;
  source: string;
  metric: string;
  value: number;
  message: string;
}

export interface ContainerGroupMetrics {
  name: string;
  podCount: number;
  avgCpuUsage: number;
  maxCpuUsage: number;
  minCpuUsage: number;
  avgMemoryUsage: number;
  maxMemoryUsage: number;
  minMemoryUsage: number;
  avgDiskUsed: string;
  maxDiskUsed: string;
  minDiskUsed: string;
  avgDiskUsedBytes: number;
  maxDiskUsedBytes: number;
  minDiskUsedBytes: number;
  cpuRequests: string;
  cpuLimits: string;
  memoryRequests: string;
  memoryLimits: string;
  diskRequests: string;
  diskLimits: string;
  healthStatus: string;
}

export interface WorkloadPodMetrics {
  name: string;
  phase: string;
  ready: boolean;
  cpuUsage: number;
  memoryUsage: number;
  diskUsed: string;
  diskUsedBytes: number;
  healthStatus: string;
  restarts: number;
}

export interface WorkloadMetrics {
  workloadType: string;
  name: string;
  namespace: string;
  summary: {
    replicas: number;
    readyReplicas: number;
    availableReplicas: number;
    podCount: number;
    runningPods: number;
    metricsPodCount: number;
    avgCpuUsage: number;
    maxCpuUsage: number;
    avgMemoryUsage: number;
    maxMemoryUsage: number;
    avgDiskUsed: string;
    maxDiskUsed: string;
    totalDiskUsed: string;
    avgDiskUsedBytes: number;
    maxDiskUsedBytes: number;
    totalDiskUsedBytes: number;
    cpuRequests: string;
    cpuLimits: string;
    memoryRequests: string;
    memoryLimits: string;
    diskRequests: string;
    diskLimits: string;
    healthStatus: string;
    alerts: WorkloadAlert[];
  };
  monitoringStrategy: {
    policy: string;
    description: string;
    thresholds: {
      cpuWarning: number;
      cpuCritical: number;
      memoryWarning: number;
      memoryCritical: number;
    };
    podCoverage: string;
    recommendation: string;
  };
  pods: WorkloadPodMetrics[];
  containerGroups: ContainerGroupMetrics[];
  historicalData: {
    cpuUsage: MetricDataPoint[];
    memoryUsage: MetricDataPoint[];
    diskUsage: MetricDataPoint[];
  };
}

export interface DeploymentMetricsResponse {
  code: number;
  message: string;
  data: {
    metrics: WorkloadMetrics;
  };
}

export const getDeploymentMetrics = (
  clusterName: string,
  namespace: string,
  deploymentName: string,
) => {
  return api.get<DeploymentMetricsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/metrics`,
    { timeout: 60000 },
  );
};

export interface RolloutStatus {
  updatedReplicas: number;
  readyReplicas: number;
  availableReplicas: number;
  unavailableReplicas: number;
  replicas: number;
  observedGeneration: number;
  paused: boolean;
  conditions?: Array<{
    type: string;
    status: string;
    lastUpdateTime: string;
    lastTransitionTime: string;
    reason: string;
    message: string;
  }>;
}

export interface CreateCanaryDeploymentRequest {
  name: string;
  replicas?: number;
  labels?: Record<string, string>;
  canaryLabelKey?: string;
  canaryLabelValue?: string;
}

export const pauseRollout = (clusterName: string, namespace: string, deploymentName: string) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/pause`);

export const resumeRollout = (clusterName: string, namespace: string, deploymentName: string) =>
  api.post(`/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/resume`);

export const getRolloutStatus = (clusterName: string, namespace: string, deploymentName: string) =>
  api.get<{ code: number; message: string; data: { rollout: RolloutStatus } }>(
    `/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/rollout`,
  );

export const createCanaryDeployment = (
  clusterName: string,
  namespace: string,
  deploymentName: string,
  data: CreateCanaryDeploymentRequest,
) =>
  api.post(
    `/clusters/${clusterName}/namespaces/${namespace}/deployments/${deploymentName}/canary`,
    data,
  );