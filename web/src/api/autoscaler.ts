import api from './axios';

export interface AutoScalerConfig {
  enabled: boolean;
  image?: string;
  scaleDownDelay?: string;
  scaleDownThreshold?: string;
  scaleUpThreshold?: string;
  scaleDownUnneededTime?: string;
  scaleDownDelayAfterAdd?: string;
  nodeGroups?: NodeGroupConfig[];
}

export interface NodeGroupConfig {
  name: string;
  minNodes: number;
  maxNodes: number;
}

export interface AutoScalerStatus {
  enabled: boolean;
  status: string; // Disabled, NotDeployed, Starting, Running
  replicas: number;
  readyReplicas: number;
  availableReplicas: number;
}

export interface AutoScalerConfigResponse {
  code: number;
  message: string;
  data: {
    config: AutoScalerConfig;
  };
}

export interface AutoScalerStatusResponse {
  code: number;
  message: string;
  data: {
    status: AutoScalerStatus;
  };
}

export interface OperationResponse {
  code: number;
  message: string;
  data?: {
    message: string;
  };
}

// get autoscaler config
export const getAutoScalerConfig = (clusterName: string) => {
  return api.get<AutoScalerConfigResponse>(`/clusters/${clusterName}/autoscaler/config`);
};

// update autoscaler config
export const updateAutoScalerConfig = (clusterName: string, config: AutoScalerConfig) => {
  return api.put<OperationResponse>(`/clusters/${clusterName}/autoscaler/config`, config);
};

// get autoscaler status
export const getAutoScalerStatus = (clusterName: string) => {
  return api.get<AutoScalerStatusResponse>(`/clusters/${clusterName}/autoscaler/status`);
};
