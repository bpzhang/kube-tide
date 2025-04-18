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

// StatefulSet详情响应
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

// StatefulSet列表响应
export interface StatefulSetListResponse {
  code: number;
  message: string;
  data: {
    statefulsets: StatefulSetInfo[];
  };
}

// StatefulSet事件响应
export interface StatefulSetEventsResponse {
  code: number;
  message: string;
  data: {
    events: any[];
  };
}

// StatefulSet所有关联事件响应
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

// 操作响应
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

// 创建StatefulSet请求
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

// 更新StatefulSet请求
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
 * 获取StatefulSet列表
 * @param clusterName 集群名称
 * @returns 所有StatefulSet
 */
export const getStatefulSetList = (clusterName: string) => {
  return api.get<StatefulSetListResponse>(`/clusters/${clusterName}/statefulsets`);
};

/**
 * 获取指定命名空间的StatefulSet列表
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @returns 命名空间内的StatefulSet
 */
export const listStatefulSetsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<StatefulSetListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets`);
};

/**
 * 获取StatefulSet详情
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns StatefulSet详情
 */
export const getStatefulSetDetails = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<StatefulSetDetailResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`);
};

/**
 * 创建StatefulSet
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetData StatefulSet数据
 * @returns 操作结果
 */
export const createStatefulSet = (clusterName: string, namespace: string, statefulsetData: CreateStatefulSetRequest) => {
  return api.post<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets`, statefulsetData);
};

/**
 * 更新StatefulSet
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @param updateData 更新数据
 * @returns 操作结果
 */
export const updateStatefulSet = (clusterName: string, namespace: string, statefulsetName: string, updateData: UpdateStatefulSetRequest) => {
  return api.put<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`, updateData);
};

/**
 * 删除StatefulSet
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns 操作结果
 */
export const deleteStatefulSet = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.delete<OperationResponse>(`/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}`);
};

/**
 * 扩缩容StatefulSet
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @param replicas 副本数
 * @returns 操作结果
 */
export const scaleStatefulSet = (clusterName: string, namespace: string, statefulsetName: string, replicas: number) => {
  return api.put<OperationResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/scale`,
    { replicas }
  );
};

/**
 * 重启StatefulSet
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns 操作结果
 */
export const restartStatefulSet = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.post<OperationResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/restart`
  );
};

/**
 * 获取StatefulSet事件
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns 事件列表
 */
export const getStatefulSetEvents = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<StatefulSetEventsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/events`
  );
};

/**
 * 获取StatefulSet及其关联Pod的所有事件
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns 所有相关事件
 */
export const getAllStatefulSetEvents = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<AllStatefulSetEventsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/all-events`
  );
};

/**
 * 获取StatefulSet关联的Pod
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param statefulsetName StatefulSet名称
 * @returns Pod列表
 */
export const getStatefulSetPods = (clusterName: string, namespace: string, statefulsetName: string) => {
  return api.get<PodListResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/statefulsets/${statefulsetName}/pods`
  );
};
