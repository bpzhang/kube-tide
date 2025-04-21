import api from './axios';
import { PodListResponse } from './pod';

// StatefulSet 相关接口响应类型
export interface StatefulSetInfo {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  serviceName: string;
  creationTime: string;
  labels: { [key: string]: string };
  selector: { [key: string]: string };
  containerCount: number;
  images: string[];
  updateStrategy: string;
  volumeClaimTemplates: string[];
}

// StatefulSet detail response
export interface StatefulSetDetailResponse {
  code: number;
  message: string;
  data: {
    statefulset: {
      name: string;
      namespace: string;
      replicas: number;
      readyReplicas: number;
      serviceName: string;
      creationTime: string;
      labels: { [key: string]: string };
      selector: { [key: string]: string };
      containerCount: number;
      images: string[];
      annotations: { [key: string]: string };
      updateStrategy: string;
      paused: boolean;
      podManagementPolicy: string;
      minReadySeconds: number;
      volumeClaimTemplates: Array<{
        name: string;
        storageClassName: string;
        accessModes: string[];
        storage: string;
        labels?: { [key: string]: string };
      }>;
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
        lastTransitionTime: string;
        reason: string;
        message: string;
      }>;
    };
  };
}

// StatefulSet list response
export interface StatefulSetListResponse {
  code: number;
  message: string;
  data: {
    statefulsets: StatefulSetInfo[];
  };
}

// StatefulSet events response
export interface StatefulSetEventsResponse {
  code: number;
  message: string;
  data: {
    events: any[];
  };
}

// StatefulSet all associated events response
export interface AllStatefulSetEventsResponse {
  code: number;
  message: string;
  data: {
    events: {
      statefulset: any[];
      pod: any[];
    };
  };
}

// Operation response 
export interface OperationResponse {
  code: number;
  message: string;
  data: {
    message: string;
    statefulset?: {
      name: string;
      namespace: string;
    };
  };
}

// Create StatefulSet request
export interface CreateStatefulSetRequest {
  name: string;
  namespace: string;
  replicas?: number;
  serviceName: string;
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
  podManagementPolicy?: string;
  updateStrategy?: string;
  containers: Array<{
    name: string;
    image: string;
    command?: string[];
    args?: string[];
    workingDir?: string;
    ports?: Array<{
      name?: string;
      containerPort: number;
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
  }>;
  volumeClaimTemplates?: Array<{
    name: string;
    storageClassName?: string;
    accessModes?: string[];
    storage?: string;
    labels?: { [key: string]: string };
  }>;
}

// Update StatefulSet request
export interface UpdateStatefulSetRequest {
  replicas?: number;
  image?: { [containerName: string]: string };
  env?: { [containerName: string]: Array<{ name: string; value: string; valueFrom?: any }> };
  resources?: { 
    [containerName: string]: { 
      limits?: { [key: string]: string }; 
      requests?: { [key: string]: string } 
    } 
  };
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
  strategy?: {
    type: 'RollingUpdate' | 'OnDelete';
    rollingUpdate?: {
      partition?: number;
    };
  };
  paused?: boolean;
}

/**
 * Get StatefulSet list
 * @param clusterName Cluster name
 * @returns All StatefulSets
 */
export const getStatefulSetList = (clusterName: string) => {
  return api.get<StatefulSetListResponse>(`/clusters/${clusterName}/statefulsets`);
};

/**
 * Get StatefulSet list by namespace
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @returns StatefulSets in the namespace
 */
export const listStatefulSetsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<StatefulSetListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets`);
};

/**
 * Get StatefulSet details
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param statefulsetName StatefulSet name
 * @returns StatefulSet details
 */
export const getStatefulSetDetails = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<StatefulSetDetailResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`);
};

/**
 * Create StatefulSet
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param statefulsetData StatefulSet data
 * @returns Operation result
 */
export const createStatefulSet = (clusterName: string, namespace: string, statefulsetData: CreateStatefulSetRequest) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets`, statefulsetData);
};

/**
 * Update StatefulSet
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param statefulsetName StatefulSet name
 * @param updateData Update data
 * @returns Operation result
 */
export const updateStatefulSet = (clusterName: string, namespace: string, statefulsetName: string, updateData: UpdateStatefulSetRequest) => {
  return api.put<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`, updateData);
};

/**
 * Delete StatefulSet
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param statefulsetName StatefulSet name
 * @returns Operation result
 */
export const deleteStatefulSet = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`);
};

/**
 * Scale StatefulSet
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param statefulsetName StatefulSet name
 * @param replicas Replicas
 * @returns Operation result
 */
export const scaleStatefulSet = (clusterName: string, namespace: string, statefulsetName: string, replicas: number) => {
  return api.put<OperationResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/scale`,
    { replicas }
  );
};

/**
 * restart StatefulSet
 * @param clusterName cluster name
 * @param namespace namespace
 * @param statefulsetName StatefulSet name
 * @returns operation result
 */
export const restartStatefulSet = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.post<OperationResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/restart`
  );
};

/**
 * Get StatefulSet events
 * @param clusterName cluster name
 * @param namespace namespace
 * @param statefulsetName StatefulSet name
 * @returns event list
 */
export const getStatefulSetEvents = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<StatefulSetEventsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/events`
  );
};

/**
 * Get StatefulSet and its associated Pod events
 * @param clusterName cluster name
 * @param namespace namespace
 * @param statefulsetName StatefulSet name
 * @returns all related events
 */
export const getAllStatefulSetEvents = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<AllStatefulSetEventsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/all-events`
  );
};

/**
 * Get StatefulSet and its associated Pod
 * @param clusterName cluster name
 * @param namespace namespace
 * @param statefulsetName StatefulSet name
 * @returns Pod list
 */
export const getStatefulSetPods = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<PodListResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/pods`
  );
};
