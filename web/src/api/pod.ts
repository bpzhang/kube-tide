import api from './axios';

// Pod 相关接口响应类型
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

// Pod 列表查询
export const getPodsByNamespace = (clusterName: string, namespace: string) => {
  return api.get<PodListResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods`);
};

// Pod 详情查询
export const getPodDetails = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}`);
};

// Pod 日志查询
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
 * 获取Pod的实时日志流
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param podName Pod名称
 * @param containerName 容器名称
 * @param tailLines 日志行数
 * @param follow 是否持续跟踪
 * @param onMessage 收到日志消息的回调函数
 * @returns 返回一个对象，包含close方法用于关闭EventSource连接
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
  // 构建URL，确保使用绝对路径
  const baseUrl = window.location.origin + (api.defaults.baseURL || '/api');
  const url = new URL(`${baseUrl}/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/logs/stream`);
  
  // 添加查询参数
  url.searchParams.append('container', containerName || '');
  url.searchParams.append('tailLines', tailLines.toString());
  url.searchParams.append('follow', follow.toString());
  
  // 创建EventSource对象来处理服务器发送的事件
  const eventSource = new EventSource(url.toString());
  
  // 处理接收到的消息
  eventSource.onmessage = (event) => {
    onMessage(event.data);
  };
  
  // 处理错误 - 添加更详细的错误处理
  eventSource.onerror = (error) => {
    console.error('Pod日志流错误:', error);
    // 显示更详细的错误信息
    onMessage(`[错误] 日志流连接失败或中断`);
    eventSource.close();
  };
  
  // 添加打开连接的处理程序
  eventSource.onopen = () => {
    console.log('Pod日志流连接已建立');
    onMessage('[系统] 实时日志连接已建立...');
  };
  
  // 返回一个对象，包含关闭连接的方法
  return {
    close: () => {
      console.log('关闭Pod日志流连接');
      eventSource.close();
    }
  };
};

// Pod 删除
export const deletePod = (clusterName: string, namespace: string, podName: string) => {
  return api.delete<any>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}`);
};

// Pod 终端 WebSocket URL 生成
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
 * 通过标签选择器获取Pod列表
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param selector 标签选择器
 * @returns Pod列表响应
 */
export const getPodsBySelector = (clusterName: string, namespace: string, selector: { [key: string]: string }) => {
  return api.post<PodListResponse>(
    `/clusters/${clusterName}/namespaces/${namespace}/pods/selector`, 
    selector
  );
};

/**
 * 检查Pod是否存在及其状态
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param podName Pod名称
 * @returns Pod存在状态响应
 */
export const checkPodExists = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodExistsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/exists`);
};

/**
 * 获取Pod相关的事件
 * @param clusterName 集群名称
 * @param namespace 命名空间
 * @param podName Pod名称
 * @returns Pod事件列表响应
 */
export const getPodEvents = (clusterName: string, namespace: string, podName: string) => {
  return api.get<PodEventsResponse>(`/clusters/${clusterName}/namespaces/${namespace}/pods/${podName}/events`);
};