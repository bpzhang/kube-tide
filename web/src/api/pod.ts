import api from './axios';

// Pod response interfaces
export interface PodResponse {
  code: number;
  message: string;
  data: {
    pod: any;
  };
}

export interface PodListResponse {
  code: number;
  message: string;
  data: {
    pods: any[];
  };
}

export interface PodLogsResponse {
  code: number;
  message: string;
  data: {
    logs: string;
  };
}

export interface PodExistsResponse {
  code: number;
  message: string;
  data: {
    exists: boolean;
    pod?: any;
  };
}

export interface PodEventsResponse {
  code: number;
  message: string;
  data: {
    events: any[];
  };
}

export interface PodRestartPolicyResponse {
  code: number;
  message: string;
  data: {
    restartPolicy: string;
  };
}

export interface UpdateRestartPolicyRequest {
  restartPolicy: string;
  deleteOriginal?: boolean;
}

export interface UpdateRestartPolicyResponse {
  code: number;
  message: string;
  data: {
    message: string;
  };
}

// Pod list
export const getPodsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<PodListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods`);
};

// Pod detail
export const getPodDetails = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}`);
};

// Pod logs
export const getPodLogs = (
  clusterName: string,
  namespace: string,
  podName: string,
  containerName: string,
  tailLines?: number
) => {
  return api.get<PodLogsResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/logs`,
    {
      params: {
        container: containerName,
        tailLines,
      },
    }
  );
};

/**
 * get pod logs stream
 * @param clusterName cluster name
 * @param namespace namespace
 * @param podName pod name
 * @param containerName container name
 * @param tailLines log lines
 * @param follow whether to follow
 * @param onMessage callback function for received log messages
 * @returns returns an object containing a close method to close the EventSource connection
 */
export const streamPodLogs = (
  clusterName: string,
  namespace: string,
  podName: string,
  containerName: string,
  tailLines: number = 100,
  follow: boolean = true,
  onMessage: (logLine: string) => void
) => {
  // build the base URL for the API
  const baseUrl = window.location.origin + (api.defaults.baseURL || '/api');
  const url = new URL(`${baseUrl}/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/logs/stream`);
  
  // Add query parameters
  url.searchParams.append('container', containerName || '');
  url.searchParams.append('tailLines', tailLines.toString());
  url.searchParams.append('follow', follow.toString());
  
  // Create EventSource object to handle server-sent events
  const eventSource = new EventSource(url.toString());
  
  // Handle received messages
  eventSource.onmessage = (event) => {
    onMessage(event.data);
  };
  
  // Handle errors - add more detailed error handling
  eventSource.onerror = (error) => {
    console.error('Pod log stream error:', error);
    // Display more detailed error information
    onMessage(`[Error] Log stream connection failed or interrupted. Please check the connection or the server status.`);
    // Optionally, you can also close the EventSource connection here
    eventSource.close();
  };
  
  // Handle connection open event
  eventSource.onopen = () => {
    console.log('Pod log stream connection established');
    onMessage('[System] Real-time log connection established...');
  };
  
  // return an object containing a close method to close the EventSource connection
  return {
    close: () => {
      console.log('closeing pod log stream connection');
      eventSource.close();
    }
  };
};

// Pod deletion
export const deletePod = (clusterName: string, namespace: string, podName: string) => {
  return api.delete<any>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}`);
};

// Pod terminal WebSocket URL generation
export const getPodTerminalWebSocketUrl = (
  clusterName: string,
  namespace: string,
  podName: string,
  containerName: string
) => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/api/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/exec?container=${containerName}`;
};

/**
 * Get Pod list by label selector
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param selector Label selector
 * @returns Pod list response
 */
export const getPodsBySelector = (clusterName: string, namespace: string, selector: { [key: string]: string }) => {
  return api.post<PodListResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/selector`, 
    selector
  );
};

/**
 * Check if the Pod exists and its status
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @returns Pod existence status response
 */
export const checkPodExists = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodExistsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/exists`);
};

/**
 * Check if the Pod related events
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @returns Pod events list response
 */
export const getPodEvents = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodEventsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/events`);
};

/**
 * Get Pod restart policy
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @returns Pod restart policy response
 */
export const getPodRestartPolicy = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodRestartPolicyResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/restart-policy`);
};

/**
 * Update Pod restart policy
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @param restartPolicy Restart policy (Always, OnFailure, Never)
 * @param deleteOriginal Whether to delete the original Pod
 * @returns Update result response
 */
export const updatePodRestartPolicy = (
  clusterName: string, 
  namespace: string, 
  podName: string, 
  restartPolicy: string,
  deleteOriginal?: boolean
) => {
  return api.put<UpdateRestartPolicyResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/restart-policy`,
    { restartPolicy, deleteOriginal }
  );
};

// Pod Lifecycle Management interfaces
export interface PodLifecycleRequest {
  action: 'start' | 'stop' | 'restart' | 'pause' | 'resume';
  gracePeriod?: number;
  force?: boolean;
  waitTimeout?: number;
}

export interface ContainerLifecycleStatus {
  name: string;
  ready: boolean;
  restartCount: number;
  state: {
    running?: { startedAt: string };
    waiting?: { reason: string; message: string };
    terminated?: {
      exitCode: number;
      signal: number;
      reason: string;
      message: string;
      startedAt: string;
      finishedAt: string;
    };
  };
  lastState: {
    terminated?: {
      exitCode: number;
      signal: number;
      reason: string;
      message: string;
      startedAt: string;
      finishedAt: string;
    };
  };
}

export interface PodLifecycleStatus {
  phase: string;
  ready: boolean;
  containerStatuses: ContainerLifecycleStatus[];
  startTime?: string;
  restartCount: number;
}

export interface PodLifecycleResponse {
  success: boolean;
  message: string;
  podStatus: PodLifecycleStatus;
  duration: number;
}

export interface PodLifecycleEvent {
  timestamp: string;
  type: string;
  reason: string;
  message: string;
  source: string;
}

export interface PodLifecycleHistoryResponse {
  code: number;
  message: string;
  data: {
    history: PodLifecycleEvent[];
  };
}

export interface PodLifecycleStatusResponse {
  code: number;
  message: string;
  data: {
    status: PodLifecycleStatus;
  };
}

/**
 * Manage Pod lifecycle
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @param request Lifecycle request
 * @returns Lifecycle operation response
 */
export const managePodLifecycle = (
  clusterName: string,
  namespace: string,
  podName: string,
  request: PodLifecycleRequest
) => {
  return api.post<{ data: PodLifecycleResponse }>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/lifecycle`,
    request
  );
};

/**
 * Get Pod lifecycle history
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @returns Lifecycle history response
 */
export const getPodLifecycleHistory = (
  clusterName: string,
  namespace: string,
  podName: string
) => {
  return api.get<PodLifecycleHistoryResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/lifecycle/history`
  );
};

/**
 * Get Pod lifecycle status
 * @param clusterName Cluster name
 * @param namespace Namespace
 * @param podName Pod name
 * @returns Lifecycle status response
 */
export const getPodLifecycleStatus = (
  clusterName: string,
  namespace: string,
  podName: string
) => {
  return api.get<PodLifecycleStatusResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/lifecycle/status`
  );
};